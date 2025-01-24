package ur

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"
)

// ControlHeader represents the header of a control message
type ControlHeader struct {
	Size    uint16
	Command uint8
}

// UnpackControlHeader unpacks a ControlHeader from a byte slice
func UnpackControlHeader(buf []byte) ControlHeader {
	return ControlHeader{
		Size:    binary.BigEndian.Uint16(buf[0:2]),
		Command: buf[2],
	}
}

// ControlVersion represents the version information of the controller
type ControlVersion struct {
	Major  uint32
	Minor  uint32
	Bugfix uint32
	Build  uint32
}

// UnpackControlVersion unpacks a ControlVersion from a byte slice
func UnpackControlVersion(buf []byte) ControlVersion {
	return ControlVersion{
		Major:  binary.BigEndian.Uint32(buf[0:4]),
		Minor:  binary.BigEndian.Uint32(buf[4:8]),
		Bugfix: binary.BigEndian.Uint32(buf[8:12]),
		Build:  binary.BigEndian.Uint32(buf[12:16]),
	}
}

// ReturnValue represents a boolean return value
type ReturnValue struct {
	Success bool
}

// UnpackReturnValue unpacks a ReturnValue from a byte slice
func UnpackReturnValue(buf []byte) ReturnValue {
	return ReturnValue{
		Success: buf[0] != 0,
	}
}

// Message represents a message with a level, content, and source
type Message struct {
	Level   uint8
	Message string
	Source  string
}

const (
	ExceptionMessage = 0
	ErrorMessage     = 1
	WarningMessage   = 2
	InfoMessage      = 3
)

// UnpackMessage unpacks a Message from a byte slice
func UnpackMessage(buf []byte) Message {
	msg := Message{}
	offset := 0

	// Unpack message length
	msgLength := buf[offset]
	offset++

	// Unpack message content
	msg.Message = string(buf[offset : offset+int(msgLength)])
	offset += int(msgLength)

	// Unpack source length
	srcLength := buf[offset]
	offset++

	// Unpack source content
	msg.Source = string(buf[offset : offset+int(srcLength)])
	offset += int(srcLength)

	// Unpack level
	msg.Level = buf[offset]

	return msg
}

// GetItemSize returns the size of a data type
func GetItemSize(dataType string) int {
	if strings.HasPrefix(dataType, "VECTOR6") {
		return 6
	} else if strings.HasPrefix(dataType, "VECTOR3") {
		return 3
	}
	return 1
}

// UnpackField unpacks a field from a byte slice based on its data type
func UnpackField(data []byte, offset int, dataType string) (interface{}, int) {
	size := GetItemSize(dataType)
	switch dataType {
	case "VECTOR6D", "VECTOR3D":
		values := make([]float64, size)
		for i := 0; i < size; i++ {
			values[i] = math.Float64frombits(binary.BigEndian.Uint64(data[offset+i*8 : offset+(i+1)*8]))
		}
		return values, offset + size*8
	case "VECTOR6UINT32":
		values := make([]uint32, size)
		for i := 0; i < size; i++ {
			values[i] = binary.BigEndian.Uint32(data[offset+i*4 : offset+(i+1)*4])
		}
		return values, offset + size*4
	case "DOUBLE":
		value := math.Float64frombits(binary.BigEndian.Uint64(data[offset : offset+8]))
		return value, offset + 8
	case "UINT32":
		value := binary.BigEndian.Uint32(data[offset : offset+4])
		return value, offset + 4
	case "UINT64":
		value := binary.BigEndian.Uint64(data[offset : offset+8])
		return value, offset + 8
	case "VECTOR6INT32":
		values := make([]int32, size)
		for i := 0; i < size; i++ {
			values[i] = int32(binary.BigEndian.Uint32(data[offset+i*4 : offset+(i+1)*4]))
		}
		return values, offset + size*4
	case "INT32", "UINT8":
		value := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
		return value, offset + 4
	case "BOOL":
		value := data[offset] != 0
		return value, offset + 1
	default:
		log.Fatalf("UnpackField: unknown data type: %s", dataType)
		return nil, offset
	}
}

// DataObject represents a data object with a recipe ID and fields
type DataObject struct {
	RecipeID uint8
	Fields   map[string]interface{}
}

// Pack packs the DataObject into a byte slice based on the provided names and types
func (d *DataObject) Pack(names []string, types []string) ([]byte, error) {
	if len(names) != len(types) {
		return nil, fmt.Errorf("list sizes are not identical")
	}

	buf := make([]byte, 0)
	if d.RecipeID != 0 {
		buf = append(buf, d.RecipeID)
	}

	for _, name := range names {
		value := d.Fields[name]
		if value == nil {
			return nil, fmt.Errorf("uninitialized parameter: %s", name)
		}

		switch v := value.(type) {
		case []float64:
			for _, f := range v {
				buf = binary.BigEndian.AppendUint64(buf, math.Float64bits(f))
			}
		case []uint32:
			for _, u := range v {
				buf = binary.BigEndian.AppendUint32(buf, u)
			}
		case float64:
			buf = binary.BigEndian.AppendUint64(buf, math.Float64bits(v))
		case uint32:
			buf = binary.BigEndian.AppendUint32(buf, v)
		case uint64:
			buf = binary.BigEndian.AppendUint64(buf, v)
		case []int32:
			for _, i := range v {
				buf = binary.BigEndian.AppendUint32(buf, uint32(i))
			}
		case int32:
			buf = binary.BigEndian.AppendUint32(buf, uint32(v))
		case bool:
			if v {
				buf = append(buf, 1)
			} else {
				buf = append(buf, 0)
			}
		default:
			return nil, fmt.Errorf("unsupported type: %T", v)
		}
	}

	return buf, nil
}

// UnpackDataObject unpacks a DataObject from a byte slice based on the provided names and types
func UnpackDataObject(data []byte, names []string, types []string) (*DataObject, error) {
	if len(names) != len(types) {
		return nil, fmt.Errorf("list sizes are not identical")
	}

	obj := &DataObject{
		Fields: make(map[string]interface{}),
	}
	offset := 0

	// Unpack recipe ID
	obj.RecipeID = data[offset]
	offset++

	// Unpack fields
	for i, name := range names {
		value, newOffset := UnpackField(data, offset, types[i])
		obj.Fields[name] = value
		offset = newOffset
	}

	return obj, nil
}

// CreateEmptyDataObject creates an empty DataObject with the provided names and recipe ID
func CreateEmptyDataObject(names []string, recipeID uint8) *DataObject {
	obj := &DataObject{
		RecipeID: recipeID,
		Fields:   make(map[string]interface{}),
	}
	for _, name := range names {
		obj.Fields[name] = nil
	}
	return obj
}

// DataConfig represents a data configuration with an ID, names, types, and format
type DataConfig struct {
	ID    uint8
	Names []string
	Types []string
	Fmt   string
}

// UnpackRecipe unpacks a DataConfig from a byte slice
func UnpackRecipe(buf []byte) (*DataConfig, error) {
	cfg := &DataConfig{}
	cfg.ID = buf[0]
	cfg.Types = strings.Split(string(buf[1:]), ",")

	// Build the format string
	cfg.Fmt = ">B" // Start with the recipe ID
	for _, t := range cfg.Types {
		switch t {
		case "INT32":
			cfg.Fmt += "i"
		case "UINT32":
			cfg.Fmt += "I"
		case "VECTOR6D":
			cfg.Fmt += "dddddd"
		case "VECTOR3D":
			cfg.Fmt += "ddd"
		case "VECTOR6INT32":
			cfg.Fmt += "iiiiii"
		case "VECTOR6UINT32":
			cfg.Fmt += "IIIIII"
		case "DOUBLE":
			cfg.Fmt += "d"
		case "UINT64":
			cfg.Fmt += "Q"
		case "UINT8":
			cfg.Fmt += "B"
		case "BOOL":
			cfg.Fmt += "?"
		default:
			return nil, fmt.Errorf("unknown data type: %s", t)
		}
	}

	return cfg, nil
}

// Pack packs the DataConfig into a byte slice based on the provided state
func (d *DataConfig) Pack(state *DataObject) ([]byte, error) {
	return state.Pack(d.Names, d.Types)
}

// Unpack unpacks a DataObject from a byte slice using the DataConfig
func (d *DataConfig) Unpack(data []byte) (*DataObject, error) {
	return UnpackDataObject(data, d.Names, d.Types)
}
