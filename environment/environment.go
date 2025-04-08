package environment

import "runtime"

// Environment holds the current environment (production, staging, development)
var Environment string

// IsProd returns whether the application is running in production mode
func IsProd() bool {
	return Environment == "production"
}

// IsWASM returns true if running in WebAssembly environment
func IsWASM() bool {
	return runtime.GOOS == "js" && runtime.GOARCH == "wasm"
}
