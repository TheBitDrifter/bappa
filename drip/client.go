package drip

import (
	"log"
	"net"
	"sync"
)

// Client represents a connection to a drip server
type Client interface {
	// Connect establishes a connection to a server
	Connect(address string) error

	// Disconnect closes the connection
	Disconnect() error

	// Send transmits data to the server
	Send(data []byte) error

	// Receive reads data from the server
	Receive() ([]byte, error)

	Buffer() chan []byte
}

type clientImpl struct {
	connection  Connection
	running     bool
	stateBuffer chan []byte
	mutex       sync.RWMutex
}

func NewClient(max int) *clientImpl {
	stateBuffer := make(chan []byte, max)
	return &clientImpl{stateBuffer: stateBuffer}
}

// Connect establishes a connection to a server
func (c *clientImpl) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	c.connection = NewConnection(conn)
	c.running = true
	c.mutex.Unlock()

	// Receive state
	go c.receiveLoop()

	return nil
}

// Disconnect closes the connection
func (c *clientImpl) Disconnect() error {
	c.mutex.Lock()
	if c.connection == nil {
		c.mutex.Unlock()
		return nil
	}
	c.running = false
	conn := c.connection
	c.mutex.Unlock()

	return conn.Close()
}

// Send transmits data to the server
func (c *clientImpl) Send(data []byte) error {
	if c.connection == nil {
		return nil
	}
	return c.connection.Send(data)
}

// Receive reads data from the server
func (c *clientImpl) Receive() ([]byte, error) {
	if c.connection == nil {
		return nil, nil
	}
	return c.connection.Receive()
}

// Receive reads data from the server
func (c *clientImpl) Buffer() chan []byte {
	return c.stateBuffer
}

// receiveLoop continuously receives data from the server until stopped.
func (c *clientImpl) receiveLoop() {
	for {
		// Check running state with lock protection
		c.mutex.Lock()
		if !c.running {
			c.mutex.Unlock()
			return
		}
		conn := c.connection
		c.mutex.Unlock()

		if conn == nil {
			return
		}

		data, err := conn.Receive()
		if err != nil {
			log.Printf("receiveLoop: %v", err)

			c.mutex.Lock()
			c.running = false
			c.mutex.Unlock()

			return // Exit the goroutine when connection error occurs
		}

		select {
		case c.stateBuffer <- data:
			// Successfully buffered the data
		default:
			// Buffer full, discard oldest item
			log.Println("Buffer full, discarding oldest state")
			select {
			case <-c.stateBuffer:
				// Successfully removed oldest item
			default:
				log.Println("Buffer empty when trying to discard")
			}

			// Try again with new data
			select {
			case c.stateBuffer <- data:
				// Successfully buffered after making space
			default:
				log.Println("Failed to buffer new state after discard")
			}
		}
	}
}
