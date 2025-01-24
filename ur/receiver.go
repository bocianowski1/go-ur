package ur

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
)

type URReceiver struct {
	*URCommon
}

func NewReceiver(ctx context.Context, cfg URConfig) *URReceiver {
	return &URReceiver{
		URCommon: &URCommon{
			Ctx: ctx,
			cfg: cfg,
		},
	}
}

// extend URCommon connect method
func (r *URReceiver) Connect() error {
	err := r.URCommon.Connect()
	if err != nil {
		return err
	}

	// Send the RTDE handshake
	ok, err := r.negotiateProtocolVersion2()
	if err != nil || !ok {
		return fmt.Errorf("failed to negotiate protocol version: %v", err)
	}

	return nil
}

func (r *URReceiver) StartDataExchange() error {
	// Create the payload
	payload := r.createPayload(RTDE_CONTROL_PACKAGE_START, nil)

	// Send the payload to the robot
	_, err := r.conn.Write(payload)
	if err != nil {
		return err
	}

	// Read the robot's response
	response := make([]byte, 1024)
	n, err := r.conn.Read(response)
	if err != nil {
		return err
	}

	// Parse the response
	if n < 3 {
		return fmt.Errorf("invalid response: too short")
	}

	header := Header{
		PkgSize: binary.BigEndian.Uint16(response[0:2]),
		Cmd:     response[2],
	}

	accepted := response[3]

	log.Println("Data exchange started")
	log.Println("Header:", header)
	log.Println("Accepted:", accepted)

	return nil
}

func (r *URReceiver) Recv(command uint8) ([]byte, error) {
	resp, err := r.recvToBuffer(r.conn)
	if err != nil {
		return nil, err
	}

	// Process the buffer
	for len(resp) >= 3 {
		packetHeader := UnpackControlHeader(resp)

		if len(resp) >= int(packetHeader.Size) {
			packet := resp[3:packetHeader.Size]
			resp = resp[packetHeader.Size:]

			if len(resp) >= int(packetHeader.Size) {
				nextPacketHeader := UnpackControlHeader(resp)
				if nextPacketHeader.Command == command {
					log.Println("Skipping package (1)")
					continue
				}
			}

			if packetHeader.Command == command {
				return packet[1:], nil
			} else {
				log.Println("Skipping package (2)")
			}
		} else {
			break // Not enough data for a complete packet
		}
	}

	return nil, fmt.Errorf("recv() Connection lost")
}

type OutputRequest struct {
	*Header
	Freq float64
	Vars string
}

func (r *OutputRequest) ToBytes() []byte {
	// Encode the frequency as a big-endian double
	freqBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(freqBytes, math.Float64bits(r.Freq))

	// Encode the variable names as a UTF-8 string
	varsBytes := []byte(r.Vars)

	// Combine the frequency and variable names into the payload
	payload := append(freqBytes, varsBytes...)

	return payload
}

type OutputResponse struct {
	*Header
	RecipeID uint8
	Types    string
}

func (r *URReceiver) SendOutputSetup(vars string) (*OutputResponse, error) {
	log.Println("Sending output setup...")

	// Create the payload
	req := &OutputRequest{
		Header: &Header{
			PkgSize: 0,
			Cmd:     RTDE_CONTROL_PACKAGE_SETUP_OUTPUTS,
		},
		Freq: MAX_FREQ,
		Vars: vars,
	}
	payload := r.createPayload(RTDE_CONTROL_PACKAGE_SETUP_OUTPUTS, req.ToBytes())

	// Send the payload to the robot
	_, err := r.conn.Write(payload)
	if err != nil {
		return nil, err
	}

	// Read the robot's response
	response := make([]byte, 4096)
	n, err := r.conn.Read(response)
	if err != nil {
		return nil, err
	}

	log.Printf("Received %d bytes\n", n)

	// Parse the response
	if n < 4 {
		return nil, fmt.Errorf("invalid response: too short")
	}

	header := Header{
		PkgSize: binary.BigEndian.Uint16(response[0:2]),
		Cmd:     response[2],
	}

	recipeID := response[3]
	types := string(response[4:n])

	return &OutputResponse{
		Header:   &header,
		RecipeID: recipeID,
		Types:    types,
	}, nil
}

// negotiateProtocolVersion2 sends the RTDE_REQUEST_PROTOCOL_VERSION message and handles the response
func (r *URReceiver) negotiateProtocolVersion2() (bool, error) {
	// Create the payload
	payload := r.createRTDEProtocolRequest(RTDE_PROTOCOL_VERSION_2)

	// Send the payload to the robot
	_, err := r.conn.Write(payload)
	if err != nil {
		return false, err
	}

	// Read the robot's response
	response := make([]byte, 1024)
	n, err := r.conn.Read(response)
	if err != nil {
		return false, err
	}

	// Parse the response
	if n < 3 {
		return false, fmt.Errorf("invalid response: too short")
	}

	// do not need the header from the response,
	// only the accepted flag
	accepted := response[3] == 1
	return accepted, nil
}

func (r *URReceiver) recvToBuffer(conn net.Conn) ([]byte, error) {
	// Read data from the socket into the buffer
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("timeout reading from socket")
		}
		return nil, err
	}

	return buf[:n], nil
}

func (r *URReceiver) createPayload(pkgType uint8, payload []byte) []byte {
	// Calculate the package size (header: 3 bytes + payload size)
	packageSize := uint16(3 + len(payload))

	// Create a buffer to hold the payload
	buf := make([]byte, 0, packageSize)

	// Add the header
	buf = binary.BigEndian.AppendUint16(buf, packageSize) // package size
	buf = append(buf, pkgType)                            // package type

	// Add the payload
	buf = append(buf, payload...)

	return buf
}

// createRTDEProtocolRequest creates the payload for RTDE_REQUEST_PROTOCOL_VERSION
func (r *URReceiver) createRTDEProtocolRequest(protocolVersion uint16) []byte {
	// Calculate the package size (header: 3 bytes, payload: 2 bytes)
	packageSize := uint16(3 + 2)

	// Create a buffer to hold the payload
	buf := make([]byte, 0, packageSize)

	// Add the header
	buf = binary.BigEndian.AppendUint16(buf, packageSize) // package size
	buf = append(buf, RTDE_REQUEST_PROTOCOL_VERSION)      // package type

	// Add the payload (protocol version)
	buf = binary.BigEndian.AppendUint16(buf, protocolVersion) // protocol version

	return buf
}
