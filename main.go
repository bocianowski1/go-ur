package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/bocianowski1/go-ur/ur"
)

func main() {
	ctx := context.Background()
	r := ur.NewReceiver(ctx, ur.URConfig{
		IP:   "localhost",
		Port: 30004,
	})
	defer r.Disconnect()

	err := r.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to robot: %v", err)
	}

	resp, err := r.SendOutputSetup("actual_q")
	if err != nil {
		log.Fatalf("Failed to send output setup: %v", err)
	}

	log.Println("Output setup:", resp)

	recipeID := resp.RecipeID
	recipeType := strings.Split(resp.Types, ",")
	log.Println("Recipe ID:", recipeID)
	log.Println("Recipe Type:", recipeType)

	err = r.StartDataExchange()
	if err != nil {
		log.Fatalf("Failed to start data exchange: %v", err)
	}

	state, err := r.Recv(ur.RTDE_DATA_PACKAGE)
	if err != nil {
		log.Fatalf("Failed to receive data: %v", err)
	}

	// Parse the state
	stateMap, err := parseState(state, recipeType)
	if err != nil {
		log.Fatalf("Failed to parse state: %v", err)
	}

	log.Println("Parsed state:", stateMap)

	targetQ, ok := stateMap["target_q"].([]float64)
	if !ok {
		log.Fatalf("Failed to get target_q")
	}

	for i, q := range targetQ {
		log.Printf("Joint %d: %.2f deg\n", i+1, ur.RadToDeg(q))
	}
}

func parseState(data []byte, recipeType []string) (map[string]interface{}, error) {
	state := make(map[string]interface{})
	offset := 0

	for _, dataType := range recipeType {
		switch dataType {
		case ur.TYPE_VECTOR_6D:
			// Parse 6 float64 values
			values := make([]float64, 6)
			for i := 0; i < 6; i++ {
				if offset+8 > len(data) {
					return nil, fmt.Errorf("invalid data length for VECTOR6D")
				}
				values[i] = math.Float64frombits(binary.BigEndian.Uint64(data[offset : offset+8]))
				offset += 8
			}
			state["target_q"] = values
		case ur.TYPE_DOUBLE:
			// Parse a single float64 value
			if offset+8 > len(data) {
				return nil, fmt.Errorf("invalid data length for DOUBLE")
			}
			state["double_value"] = math.Float64frombits(binary.BigEndian.Uint64(data[offset : offset+8]))
			offset += 8
		case ur.TYPE_UINT32:
			// Parse a single uint32 value
			if offset+4 > len(data) {
				return nil, fmt.Errorf("invalid data length for UINT32")
			}
			state["uint32_value"] = binary.BigEndian.Uint32(data[offset : offset+4])
			offset += 4
		// Add more cases for other data types as needed
		default:
			return nil, fmt.Errorf("unknown data type: %s", dataType)
		}
	}

	return state, nil
}
