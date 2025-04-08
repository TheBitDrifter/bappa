package drip

// ServerConfig contains configuration for the server
type ServerConfig struct {
	// TPS is the number of ticks (updates) per second
	TPS int

	// Port is the TCP port number to listen on
	Port int

	// MaxConnections is the maximum number of simultaneous client connections
	MaxConnections int
}

// DefaultServerConfig returns a server configuration with sensible defaults
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		TPS:            60,
		Port:           8080,
		MaxConnections: 100,
	}
}
