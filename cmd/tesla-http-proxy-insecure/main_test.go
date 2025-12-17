package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/teslamotors/vehicle-command/pkg/proxy"
)

// assertEquals is a test helper for comparing values
func assertEquals(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// resetConfig resets httpConfig to default values for test isolation
func resetConfig() {
	httpConfig.host = "localhost"
	httpConfig.port = defaultPort
	httpConfig.timeout = proxy.DefaultTimeout
	httpConfig.verbose = false
}

func TestDefaultValues(t *testing.T) {
	// Save and restore original environment
	origHost := os.Getenv(EnvHost)
	origPort := os.Getenv(EnvPort)
	origVerbose := os.Getenv(EnvVerbose)
	origTimeout := os.Getenv(EnvTimeout)

	defer func() {
		os.Setenv(EnvHost, origHost)
		os.Setenv(EnvPort, origPort)
		os.Setenv(EnvVerbose, origVerbose)
		os.Setenv(EnvTimeout, origTimeout)
	}()

	// Clear environment variables
	os.Unsetenv(EnvHost)
	os.Unsetenv(EnvPort)
	os.Unsetenv(EnvVerbose)
	os.Unsetenv(EnvTimeout)

	resetConfig()
	err := readFromEnvironment()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertEquals(t, "localhost", httpConfig.host, "host")
	assertEquals(t, defaultPort, httpConfig.port, "port")
	assertEquals(t, proxy.DefaultTimeout, httpConfig.timeout, "timeout")
	assertEquals(t, false, httpConfig.verbose, "verbose")
}

func TestEnvironmentVariables(t *testing.T) {
	// Save and restore original environment
	origHost := os.Getenv(EnvHost)
	origPort := os.Getenv(EnvPort)
	origVerbose := os.Getenv(EnvVerbose)
	origTimeout := os.Getenv(EnvTimeout)

	defer func() {
		os.Setenv(EnvHost, origHost)
		os.Setenv(EnvPort, origPort)
		os.Setenv(EnvVerbose, origVerbose)
		os.Setenv(EnvTimeout, origTimeout)
	}()

	// Set test environment variables
	os.Setenv(EnvHost, "0.0.0.0")
	os.Setenv(EnvPort, "9090")
	os.Setenv(EnvVerbose, "true")
	os.Setenv(EnvTimeout, "30s")

	resetConfig()
	err := readFromEnvironment()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertEquals(t, "0.0.0.0", httpConfig.host, "host")
	assertEquals(t, 9090, httpConfig.port, "port")
	assertEquals(t, 30*time.Second, httpConfig.timeout, "timeout")
	assertEquals(t, true, httpConfig.verbose, "verbose")
}

func TestFlagsPrecedenceOverEnvironment(t *testing.T) {
	// Save and restore original environment and args
	origHost := os.Getenv(EnvHost)
	origPort := os.Getenv(EnvPort)
	origVerbose := os.Getenv(EnvVerbose)
	origTimeout := os.Getenv(EnvTimeout)
	origArgs := os.Args

	defer func() {
		os.Setenv(EnvHost, origHost)
		os.Setenv(EnvPort, origPort)
		os.Setenv(EnvVerbose, origVerbose)
		os.Setenv(EnvTimeout, origTimeout)
		os.Args = origArgs
	}()

	// Set environment variables
	os.Setenv(EnvHost, "envhost")
	os.Setenv(EnvPort, "9090")
	os.Setenv(EnvVerbose, "true")
	os.Setenv(EnvTimeout, "30s")

	// Set command-line args that should take precedence
	os.Args = []string{"cmd", "-host", "flaghost", "-port", "8888", "-timeout", "60s"}

	// Reset flag parsing state and re-register flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	httpConfig = &HTTPProxyConfig{}
	flag.BoolVar(&httpConfig.verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&httpConfig.host, "host", "localhost", "Proxy server `hostname`")
	flag.IntVar(&httpConfig.port, "port", defaultPort, "`Port` to listen on")
	flag.DurationVar(&httpConfig.timeout, "timeout", proxy.DefaultTimeout, "Timeout interval when sending commands")

	flag.Parse()
	err := readFromEnvironment()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Flags should override environment variables
	assertEquals(t, "flaghost", httpConfig.host, "host")
	assertEquals(t, 8888, httpConfig.port, "port")
	assertEquals(t, 60*time.Second, httpConfig.timeout, "timeout")
}

func TestInvalidPortEnvironmentVariable(t *testing.T) {
	// Save and restore original environment
	origPort := os.Getenv(EnvPort)
	defer os.Setenv(EnvPort, origPort)

	os.Setenv(EnvPort, "notanumber")

	resetConfig()
	err := readFromEnvironment()
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}
}

func TestInvalidTimeoutEnvironmentVariable(t *testing.T) {
	// Save and restore original environment
	origTimeout := os.Getenv(EnvTimeout)
	defer os.Setenv(EnvTimeout, origTimeout)

	os.Setenv(EnvTimeout, "notaduration")

	resetConfig()
	err := readFromEnvironment()
	if err == nil {
		t.Error("Expected error for invalid timeout, got nil")
	}
}

func TestVerboseEnvironmentVariableParsing(t *testing.T) {
	// Save and restore original environment
	origVerbose := os.Getenv(EnvVerbose)
	origHost := os.Getenv(EnvHost)
	origPort := os.Getenv(EnvPort)
	origTimeout := os.Getenv(EnvTimeout)
	defer func() {
		os.Setenv(EnvVerbose, origVerbose)
		os.Setenv(EnvHost, origHost)
		os.Setenv(EnvPort, origPort)
		os.Setenv(EnvTimeout, origTimeout)
	}()

	// Clear other env vars to avoid interference
	os.Unsetenv(EnvHost)
	os.Unsetenv(EnvPort)
	os.Unsetenv(EnvTimeout)

	tests := []struct {
		envValue string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"anything", true},
		{"false", false},
		{"0", false},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv(EnvVerbose, tt.envValue)
			resetConfig()
			err := readFromEnvironment()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertEquals(t, tt.expected, httpConfig.verbose, "verbose")
		})
	}
}

func TestDefaultPortConstant(t *testing.T) {
	assertEquals(t, 8080, defaultPort, "defaultPort constant")
}

func TestCacheSizeConstant(t *testing.T) {
	assertEquals(t, 10000, cacheSize, "cacheSize constant")
}
