package display

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"time"
)

const MAX_RETRIES = 10

type Connection struct {
	Hostport string

	net.Conn
}

func (c *Connection) Ensure() error {
	if c.Conn != nil {
		return nil
	}
	var err error
	for i := range MAX_RETRIES {
		c.Conn, err = net.Dial("tcp", c.Hostport)
		if err == nil {
			break
		}
		slog.Warn("connect to display", "attempt", i+1, "error", err)
		if i < MAX_RETRIES-1 { // don't sleep after last attempt, that's just wasting time
			time.Sleep(time.Duration(math.Pow(1.3, float64(i))) * time.Second)
		}
	}
	if err != nil {
		c.Conn = nil
		slog.Error("connect to display (all attempts failed)", "error", err)
	} else {
		slog.Info("connect to display (succeeded)", "hostport", c.Hostport)
	}
	return err
}

func (c *Connection) handleError(err error) {
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		c.Close()
	}
}

func (c *Connection) Read(b []byte) (int, error) {
	if err := c.Ensure(); err != nil {
		return 0, err
	}
	n, err := c.Conn.Read(b)
	if err != nil {
		c.handleError(err)
		return n, fmt.Errorf("read from display (at %s): %w", c.Hostport, err)
	}
	return n, nil
}

func (c *Connection) Write(b []byte) (int, error) {
	if err := c.Ensure(); err != nil {
		return 0, err
	}
	n, err := c.Conn.Write(b)
	if err != nil {
		c.handleError(err)
		return n, fmt.Errorf("write to display (at %s): %w", c.Hostport, err)
	}
	return n, nil
}

func (c *Connection) Close() error {
	if c.Conn != nil {
		err := c.Conn.Close()
		c.Conn = nil
		return err
	}
	return nil
}
