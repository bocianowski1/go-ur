package ur

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"
)

type URConfig struct {
	IP      string
	Port    int
	Timeout time.Duration
}

type URCommon struct {
	Ctx  context.Context
	conn net.Conn

	cfg URConfig
}

func (c *URCommon) Connect() error {
	slog.Info("Connecting to robot...")

	addr := fmt.Sprintf("%s:%d", c.cfg.IP, c.cfg.Port)
	conn, err := net.DialTimeout("tcp", addr, c.cfg.Timeout)
	if err != nil {
		return err
	}

	c.conn = conn

	slog.Info("Connected to robot.")
	return nil
}

func (c *URCommon) Disconnect() error {
	slog.Info("Disconnecting from robot...")
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	if err != nil {
		return err
	}

	c.conn = nil
	slog.Info("Disconnected from robot.")
	return nil
}

func (c *URCommon) IsConnected() bool {
	return c.conn != nil
}

func (c *URCommon) SendCommand(cmd string) error {
	_, err := c.conn.Write([]byte(cmd + "\r\n"))
	return err
}
