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
	payload := r.createPayload(RTDE_CONTROL_PACKAGE_START, nil)

	_, err := r.conn.Write(payload)
	if err != nil {
		return err
	}

	response := make([]byte, 1024)
	n, err := r.conn.Read(response)
	if err != nil {
		return err
	}

	if n < 3 {
		return fmt.Errorf("invalid response: too short")
	}

	accepted := response[3] == 1
	if !accepted {
		return fmt.Errorf("failed to start data exchange")
	}

	return nil
}

func (r *URReceiver) Listen(command uint8) ([]byte, error) {
	resp, err := r.recvToBuffer(r.conn)
	if err != nil {
		return nil, err
	}

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
			break
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
	freqBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(freqBytes, math.Float64bits(r.Freq))

	varsBytes := []byte(r.Vars)
	payload := append(freqBytes, varsBytes...)

	return payload
}

type OutputResponse struct {
	*Header
	RecipeID uint8
	Types    string
}

func (r *URReceiver) SendOutputSetup(vars string) (*OutputResponse, error) {
	req := &OutputRequest{
		Header: &Header{
			PkgSize: 0,
			Cmd:     RTDE_CONTROL_PACKAGE_SETUP_OUTPUTS,
		},
		Freq: MAX_FREQ,
		Vars: vars,
	}
	payload := r.createPayload(RTDE_CONTROL_PACKAGE_SETUP_OUTPUTS, req.ToBytes())

	_, err := r.conn.Write(payload)
	if err != nil {
		return nil, err
	}

	response := make([]byte, 4096)
	n, err := r.conn.Read(response)
	if err != nil {
		return nil, err
	}

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
	payload := r.createRTDEProtocolRequest(RTDE_PROTOCOL_VERSION_2)

	_, err := r.conn.Write(payload)
	if err != nil {
		return false, err
	}

	response := make([]byte, 1024)
	n, err := r.conn.Read(response)
	if err != nil {
		return false, err
	}

	if n < 3 {
		return false, fmt.Errorf("invalid response: too short")
	}

	accepted := response[3] == 1
	return accepted, nil
}

func (r *URReceiver) recvToBuffer(conn net.Conn) ([]byte, error) {
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
	packageSize := uint16(3 + len(payload))

	buf := make([]byte, 0, packageSize)

	buf = binary.BigEndian.AppendUint16(buf, packageSize)
	buf = append(buf, pkgType)

	buf = append(buf, payload...)
	return buf
}

func (r *URReceiver) createRTDEProtocolRequest(protocolVersion uint16) []byte {
	// Calculate the package size (header: 3 bytes, payload: 2 bytes)
	packageSize := uint16(3 + 2)

	buf := make([]byte, 0, packageSize)
	buf = binary.BigEndian.AppendUint16(buf, packageSize)
	buf = append(buf, RTDE_REQUEST_PROTOCOL_VERSION)

	// Append payload (protocol version)
	buf = binary.BigEndian.AppendUint16(buf, protocolVersion)
	return buf
}
