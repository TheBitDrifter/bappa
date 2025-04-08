package drip

import (
	"encoding/binary"
	"io"
	"net"
)

// Connection handles communication with a connected client
type Connection interface {
	// Send transmits data to the client
	Send(data []byte) error

	// Receive reads data from the client
	Receive() ([]byte, error)

	// Close terminates the connection
	Close() error

	Address() string
}

type connectionImpl struct {
	conn net.Conn
}

// NewConnection creates a new connection from a network connection
func NewConnection(conn net.Conn) Connection {
	return &connectionImpl{
		conn: conn,
	}
}

// Send transmits data to the client with a length prefix
func (c *connectionImpl) Send(data []byte) error {
	// First send the length as a uint32
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	if _, err := c.conn.Write(lenBuf); err != nil {
		return err
	}
	_, err := c.conn.Write(data)
	return err
}

// Receive reads length-prefixed data from the client
func (c *connectionImpl) Receive() ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, lenBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)
	data := make([]byte, length)

	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Close terminates the connection
func (c *connectionImpl) Close() error {
	return c.conn.Close()
}

// Close terminates the connection
func (c *connectionImpl) Address() string {
	return c.conn.RemoteAddr().String()
}
