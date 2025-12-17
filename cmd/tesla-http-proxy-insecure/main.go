package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/teslamotors/vehicle-command/internal/log"
	"github.com/teslamotors/vehicle-command/pkg/cli"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
	"github.com/teslamotors/vehicle-command/pkg/proxy"
)

const (
	cacheSize   = 10000 // Number of cached vehicle sessions
	defaultPort = 8080
)

const (
	EnvHost    = "TESLA_HTTP_PROXY_HOST"
	EnvPort    = "TESLA_HTTP_PROXY_PORT"
	EnvTimeout = "TESLA_HTTP_PROXY_TIMEOUT"
	EnvVerbose = "TESLA_VERBOSE"
)

// HTTPProxyConfig holds configuration for the HTTP-only proxy server.
type HTTPProxyConfig struct {
	verbose bool
	host    string
	port    int
	timeout time.Duration
}

var (
	httpConfig = &HTTPProxyConfig{}
)

func init() {
	flag.BoolVar(&httpConfig.verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&httpConfig.host, "host", "localhost", "Proxy server `hostname`")
	flag.IntVar(&httpConfig.port, "port", defaultPort, "`Port` to listen on")
	flag.DurationVar(&httpConfig.timeout, "timeout", proxy.DefaultTimeout, "Timeout interval when sending commands")
}

// Usage prints help text for the command.
func Usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: %s [OPTION...]\n", os.Args[0])
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "A server that exposes a REST API for sending commands to Tesla vehicles over HTTP.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "WARNING: This proxy does NOT encrypt client traffic. Use only behind TLS-terminating")
	fmt.Fprintln(out, "infrastructure (Cloud Run, nginx, Traefik, K8s ingress) or in local development.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Options:")
	flag.PrintDefaults()
}

func main() {
	config, err := cli.NewConfig(cli.FlagPrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load credential configuration: %s\n", err)
		os.Exit(1)
	}

	defer func() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	}()

	flag.Usage = Usage
	config.RegisterCommandLineFlags()
	flag.Parse()

	err = readFromEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading environment: %s\n", err)
		os.Exit(1)
	}
	config.ReadFromEnvironment()

	if httpConfig.verbose {
		log.SetLevel(log.LevelDebug)
	}

	var skey protocol.ECDHPrivateKey
	skey, err = config.PrivateKey()
	if err != nil {
		return
	}

	log.Debug("Creating proxy")
	p, err := proxy.New(context.Background(), skey, cacheSize)
	if err != nil {
		log.Error("Error initializing proxy service: %v", err)
		return
	}
	p.Timeout = httpConfig.timeout
	addr := fmt.Sprintf("%s:%d", httpConfig.host, httpConfig.port)
	log.Info("Listening on %s (HTTP, no TLS)", addr)

	log.Error("Server stopped: %s", http.ListenAndServe(addr, p))
}

// readFromEnvironment applies configuration from environment variables.
// Values set by command-line flags are not overwritten.
func readFromEnvironment() error {
	if httpConfig.host == "localhost" {
		if host, ok := os.LookupEnv(EnvHost); ok {
			httpConfig.host = host
		}
	}

	if !httpConfig.verbose {
		if verbose, ok := os.LookupEnv(EnvVerbose); ok {
			httpConfig.verbose = verbose != "false" && verbose != "0"
		}
	}

	var err error
	if httpConfig.port == defaultPort {
		if port, ok := os.LookupEnv(EnvPort); ok {
			httpConfig.port, err = strconv.Atoi(port)
			if err != nil {
				return fmt.Errorf("invalid port: %s", port)
			}
		}
	}

	if httpConfig.timeout == proxy.DefaultTimeout {
		if timeoutEnv, ok := os.LookupEnv(EnvTimeout); ok {
			httpConfig.timeout, err = time.ParseDuration(timeoutEnv)
			if err != nil {
				return fmt.Errorf("invalid timeout: %s", timeoutEnv)
			}
		}
	}

	return nil
}
