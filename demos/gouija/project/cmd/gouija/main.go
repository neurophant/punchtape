// Gouija is a two-stage encrypted proxy.
//
// It can run in two modes:
//   - relay: Acts as an HTTP proxy, encrypts traffic and forwards to proxy node
//   - proxy: Accepts encrypted traffic from relay, forwards to target resources
//
// Usage:
//
//	gouija relay --proxy-addr host:9000 --key "secret"
//	gouija proxy --key "secret"
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neurophant/gouija/internal/logging"
	"github.com/neurophant/gouija/internal/pool"
	"github.com/neurophant/gouija/internal/proxy"
	"github.com/neurophant/gouija/internal/relay"
	"github.com/neurophant/gouija/internal/transport"
)

const (
	defaultRelayAddr = ":8080"
	defaultProxyAddr = ":9000"
	defaultPoolSize  = 5
	defaultTimeout   = 30 * time.Second
	defaultWSPath    = "/ws"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	mode := os.Args[1]

	switch mode {
	case "relay":
		runRelay(os.Args[2:])
	case "proxy":
		runProxy(os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
	case "-v", "--version", "version":
		fmt.Println("gouija version 1.3.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n\n", mode)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Gouija - Two-stage encrypted proxy

Usage:
  gouija <mode> [options]

Modes:
  relay    Run as HTTP proxy, forward encrypted traffic to proxy node
  proxy    Run as proxy node, decrypt and forward to target resources

Common Options:
  --quiet              Suppress informational logging (errors still logged)

Relay Mode:
  gouija relay [options]

  Basic Options:
    --listen <addr>      Address to listen on (default: :8080)
    --proxy-addr <addr>  Address of proxy node (required)
    --key <key>          Pre-shared encryption key (or set GOUIJA_KEY env)
    --salt <salt>        Custom salt for key derivation (or set GOUIJA_SALT env)
    --pool-size <n>      Target number of connections to proxy (default: 5)
    --timeout <duration> Connection timeout (default: 30s)
    --transport <mode>   Transport mode: tcp or ws (default: tcp)
    --ws-path <path>     WebSocket endpoint path (default: /ws)
    --tls                Use TLS for WebSocket transport (wss://)
    --auth-user <user>   HTTP Proxy authentication username (optional)
    --auth-pass <pass>   HTTP Proxy authentication password (optional)
    --require-auth       Require challenge-response authentication with proxy node
    --quiet              Suppress informational logging

  Connection Management - Connection Pacing:
    --conn-pacing <duration>    Delay between new connections (default: 500ms)
    --max-concurrent-dials <n>  Max simultaneous dial attempts (default: 2)

  Connection Management - Exponential Backoff:
    --backoff-initial <duration>  Initial backoff interval (default: 5s)
    --backoff-max <duration>      Maximum backoff interval (default: 120s)
    --backoff-multiplier <n>      Backoff multiplier (default: 2.0)
    --backoff-jitter <n>          Jitter factor 0-1 (default: 0.2 = ±20%)

  Adaptive Pool Sizing:
    --min-pool-size <n>      Minimum number of connections (default: 1)
    --max-pool-size <n>      Maximum number of connections (default: 10)
    --growth-threshold <n>   Grow pool when utilization exceeds (default: 0.8)
    --shrink-threshold <n>   Shrink pool when utilization below (default: 0.2)

  TCP Settings:
    --keepalive              Enable TCP keepalive (default: true)
    --keepalive-interval <d> TCP keepalive interval (default: 30s)

  Bandwidth Throttling:
    --max-bandwidth <n>   Maximum bandwidth in bytes/second (default: 0 = unlimited)
    --burst-size <n>      Maximum burst size in bytes (default: 64KB)

Proxy Mode:
  gouija proxy [options]

  Options:
    --listen <addr>      Address to listen on (default: :9000)
    --key <key>          Pre-shared encryption key (or set GOUIJA_KEY env)
    --salt <salt>        Custom salt for key derivation (or set GOUIJA_SALT env)
    --timeout <duration> Connection timeout (default: 30s)
    --transport <mode>   Transport mode: tcp or ws (default: tcp)
    --ws-path <path>     WebSocket endpoint path (default: /ws)
    --tls-cert <path>    TLS certificate file (for wss:// server)
    --tls-key <path>     TLS private key file (for wss:// server)
    --require-auth       Require challenge-response authentication from relay nodes
    --quiet              Suppress informational logging

  TCP Settings:
    --keepalive              Enable TCP keepalive (default: true)
    --keepalive-interval <d> TCP keepalive interval (default: 30s)

Transport Modes:
  tcp    Raw TCP connection (default, best performance)
  ws     WebSocket over HTTP/HTTPS

Authentication:
  There are two types of authentication:
  
  1. HTTP Proxy Authentication (--auth-user/--auth-pass):
     Protects the relay node from unauthorized HTTP clients.
     Clients must provide credentials to use the proxy.
  
  2. Proxy Node Authentication (--require-auth):
     Protects the proxy node from unauthorized relay nodes.
     Uses challenge-response HMAC authentication based on the pre-shared key.
     Both relay and proxy must use --require-auth for this to work.

Connection Management Features:
  1. Connection Pacing: Spreads connection establishment over time instead
     of creating all connections simultaneously at startup.

  2. Exponential Backoff: When connections fail, uses increasing delays
     between retries instead of fixed intervals. Includes jitter to prevent
     "thundering herd" problems.

  3. Adaptive Pool Sizing: Automatically grows/shrinks the connection pool
     based on utilization, reducing unnecessary connections.

  4. Bandwidth Throttling: Optional rate limiting to smooth traffic bursts.

Examples:
  # TCP mode (default) - Start proxy node
  gouija proxy --listen :9000 --key "my-secret-key"

  # TCP mode - Start relay node with HTTP proxy authentication
  gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "my-secret-key" \
    --auth-user myuser --auth-pass mypassword

  # Conservative configuration with connection management
  gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "my-secret-key" \
    --pool-size 3 --conn-pacing 1s --max-concurrent-dials 1

  # With custom salt for key derivation
  gouija proxy --listen :9000 --key "my-secret-key" --salt "my-custom-salt-2024"
  gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "my-secret-key" --salt "my-custom-salt-2024"

  # With proxy node authentication (both sides must enable it)
  gouija proxy --listen :9000 --key "my-secret-key" --require-auth
  gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "my-secret-key" --require-auth

  # WebSocket mode - Start proxy on port 443 with TLS
  gouija proxy --listen :443 --key "secret" --transport ws --tls-cert cert.pem --tls-key key.pem

  # WebSocket mode - Start relay connecting via WSS
  gouija relay --listen :8080 --proxy-addr proxy.example.com:443 --key "secret" --transport ws --tls

  # With bandwidth throttling (10 MB/s limit)
  gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" \
    --max-bandwidth 10485760

  # Using environment variables
  export GOUIJA_KEY="my-secret-key"
  export GOUIJA_SALT="my-custom-salt"
  export GOUIJA_AUTH_USER="myuser"
  export GOUIJA_AUTH_PASS="mypassword"
  gouija relay --proxy-addr proxy.example.com:9000

  # Quiet mode for better performance
  gouija proxy --listen :9000 --key "secret" --quiet
  gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" --quiet

  # Client usage with HTTP proxy authentication
  curl -x http://myuser:mypassword@localhost:8080 https://example.com

Environment Variables:
  GOUIJA_KEY        Pre-shared encryption key (alternative to --key flag)
  GOUIJA_SALT       Custom salt for key derivation (alternative to --salt flag)
  GOUIJA_AUTH_USER  HTTP Proxy authentication username (alternative to --auth-user)
  GOUIJA_AUTH_PASS  HTTP Proxy authentication password (alternative to --auth-pass)
  GOUIJA_QUIET      Set to "1" or "true" to enable quiet mode
`)
}

func runRelay(args []string) {
	fs := flag.NewFlagSet("relay", flag.ExitOnError)

	// Basic options
	listen := fs.String("listen", defaultRelayAddr, "Address to listen on")
	proxyAddr := fs.String("proxy-addr", "", "Address of proxy node (required)")
	key := fs.String("key", "", "Pre-shared encryption key")
	salt := fs.String("salt", "", "Custom salt for key derivation")
	poolSize := fs.Int("pool-size", defaultPoolSize, "Number of connections to proxy")
	timeout := fs.Duration("timeout", defaultTimeout, "Connection timeout")
	transportMode := fs.String("transport", "tcp", "Transport mode: tcp or ws")
	wsPath := fs.String("ws-path", defaultWSPath, "WebSocket endpoint path")
	useTLS := fs.Bool("tls", false, "Use TLS for WebSocket transport")
	authUser := fs.String("auth-user", "", "HTTP Proxy authentication username")
	authPass := fs.String("auth-pass", "", "HTTP Proxy authentication password")
	requireAuth := fs.Bool("require-auth", false, "Require challenge-response authentication with proxy node")
	quiet := fs.Bool("quiet", false, "Suppress informational logging")

	// DoS prevention - Connection pacing
	connPacing := fs.Duration("conn-pacing", pool.DefaultConnectionPacing, "Delay between new connections")
	maxConcurrentDials := fs.Int("max-concurrent-dials", pool.DefaultMaxConcurrentDials, "Max simultaneous dial attempts")

	// DoS prevention - Exponential backoff
	backoffInitial := fs.Duration("backoff-initial", pool.DefaultBackoffInitial, "Initial backoff interval")
	backoffMax := fs.Duration("backoff-max", pool.DefaultBackoffMax, "Maximum backoff interval")
	backoffMultiplier := fs.Float64("backoff-multiplier", pool.DefaultBackoffMultiplier, "Backoff multiplier")
	backoffJitter := fs.Float64("backoff-jitter", pool.DefaultBackoffJitter, "Jitter factor 0-1")

	// Adaptive pool sizing
	minPoolSize := fs.Int("min-pool-size", pool.DefaultMinPoolSize, "Minimum number of connections")
	maxPoolSize := fs.Int("max-pool-size", pool.DefaultMaxPoolSize, "Maximum number of connections")
	growthThreshold := fs.Float64("growth-threshold", pool.DefaultGrowthThreshold, "Grow pool when utilization exceeds")
	shrinkThreshold := fs.Float64("shrink-threshold", pool.DefaultShrinkThreshold, "Shrink pool when utilization below")

	// TCP settings
	keepalive := fs.Bool("keepalive", true, "Enable TCP keepalive")
	keepaliveInterval := fs.Duration("keepalive-interval", 30*time.Second, "TCP keepalive interval")

	// Bandwidth throttling
	maxBandwidth := fs.Int64("max-bandwidth", 0, "Maximum bandwidth in bytes/second (0 = unlimited)")
	burstSize := fs.Int64("burst-size", 64*1024, "Maximum burst size in bytes")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	// Set quiet mode from flag or environment
	quietMode := *quiet
	if !quietMode {
		envQuiet := os.Getenv("GOUIJA_QUIET")
		quietMode = envQuiet == "1" || envQuiet == "true"
	}
	logging.SetQuiet(quietMode)

	// Get key from flag or environment
	encKey := *key
	if encKey == "" {
		encKey = os.Getenv("GOUIJA_KEY")
	}
	if encKey == "" {
		log.Fatal("Encryption key is required. Use --key flag or set GOUIJA_KEY environment variable")
	}

	// Get salt from flag or environment
	keySalt := *salt
	if keySalt == "" {
		keySalt = os.Getenv("GOUIJA_SALT")
	}

	// Get auth credentials from flags or environment
	proxyAuthUser := *authUser
	if proxyAuthUser == "" {
		proxyAuthUser = os.Getenv("GOUIJA_AUTH_USER")
	}
	proxyAuthPass := *authPass
	if proxyAuthPass == "" {
		proxyAuthPass = os.Getenv("GOUIJA_AUTH_PASS")
	}

	// Validate required flags
	if *proxyAddr == "" {
		log.Fatal("--proxy-addr is required for relay mode")
	}

	// Parse transport mode
	var tMode transport.Mode
	switch *transportMode {
	case "tcp":
		tMode = transport.ModeTCP
	case "ws", "websocket":
		tMode = transport.ModeWebSocket
	default:
		log.Fatalf("Invalid transport mode: %s (use tcp or ws)", *transportMode)
	}

	transportCfg := transport.Config{
		Mode:              tMode,
		Path:              *wsPath,
		TLS:               *useTLS,
		KeepAlive:         *keepalive,
		KeepAliveInterval: *keepaliveInterval,
	}

	// Create and start relay server
	srv, err := relay.New(relay.Config{
		ListenAddr:  *listen,
		ProxyAddr:   *proxyAddr,
		Key:         encKey,
		Salt:        keySalt,
		Transport:   transportCfg,
		PoolSize:    *poolSize,
		Timeout:     *timeout,
		AuthUser:    proxyAuthUser,
		AuthPass:    proxyAuthPass,
		RequireAuth: *requireAuth,

		// DoS prevention - Connection pacing
		ConnectionPacing:   *connPacing,
		MaxConcurrentDials: *maxConcurrentDials,

		// DoS prevention - Exponential backoff
		BackoffInitial:    *backoffInitial,
		BackoffMax:        *backoffMax,
		BackoffMultiplier: *backoffMultiplier,
		BackoffJitter:     *backoffJitter,

		// Adaptive pool sizing
		MinPoolSize:     *minPoolSize,
		MaxPoolSize:     *maxPoolSize,
		GrowthThreshold: *growthThreshold,
		ShrinkThreshold: *shrinkThreshold,

		// Bandwidth throttling
		MaxBandwidth: *maxBandwidth,
		BurstSize:    *burstSize,
	})
	if err != nil {
		log.Fatalf("Failed to create relay server: %v", err)
	}

	// Handle shutdown signals
	go handleShutdown(func(ctx context.Context) {
		srv.Shutdown(ctx)
	})

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Relay server error: %v", err)
	}
}

func runProxy(args []string) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)

	listen := fs.String("listen", defaultProxyAddr, "Address to listen on")
	key := fs.String("key", "", "Pre-shared encryption key")
	salt := fs.String("salt", "", "Custom salt for key derivation")
	timeout := fs.Duration("timeout", defaultTimeout, "Connection timeout")
	transportMode := fs.String("transport", "tcp", "Transport mode: tcp or ws")
	wsPath := fs.String("ws-path", defaultWSPath, "WebSocket endpoint path")
	tlsCert := fs.String("tls-cert", "", "TLS certificate file")
	tlsKey := fs.String("tls-key", "", "TLS private key file")
	requireAuth := fs.Bool("require-auth", false, "Require challenge-response authentication from relay nodes")
	quiet := fs.Bool("quiet", false, "Suppress informational logging")

	// TCP settings
	keepalive := fs.Bool("keepalive", true, "Enable TCP keepalive")
	keepaliveInterval := fs.Duration("keepalive-interval", 30*time.Second, "TCP keepalive interval")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	// Set quiet mode from flag or environment
	quietMode := *quiet
	if !quietMode {
		envQuiet := os.Getenv("GOUIJA_QUIET")
		quietMode = envQuiet == "1" || envQuiet == "true"
	}
	logging.SetQuiet(quietMode)

	// Get key from flag or environment
	encKey := *key
	if encKey == "" {
		encKey = os.Getenv("GOUIJA_KEY")
	}
	if encKey == "" {
		log.Fatal("Encryption key is required. Use --key flag or set GOUIJA_KEY environment variable")
	}

	// Get salt from flag or environment
	keySalt := *salt
	if keySalt == "" {
		keySalt = os.Getenv("GOUIJA_SALT")
	}

	// Parse transport mode
	var tMode transport.Mode
	switch *transportMode {
	case "tcp":
		tMode = transport.ModeTCP
	case "ws", "websocket":
		tMode = transport.ModeWebSocket
	default:
		log.Fatalf("Invalid transport mode: %s (use tcp or ws)", *transportMode)
	}

	// Determine if TLS should be used
	useTLS := *tlsCert != "" && *tlsKey != ""

	transportCfg := transport.Config{
		Mode:              tMode,
		Path:              *wsPath,
		TLS:               useTLS,
		TLSCert:           *tlsCert,
		TLSKey:            *tlsKey,
		KeepAlive:         *keepalive,
		KeepAliveInterval: *keepaliveInterval,
	}

	// Create and start proxy server
	srv, err := proxy.New(proxy.Config{
		ListenAddr:  *listen,
		Key:         encKey,
		Salt:        keySalt,
		Transport:   transportCfg,
		Timeout:     *timeout,
		RequireAuth: *requireAuth,
	})
	if err != nil {
		log.Fatalf("Failed to create proxy server: %v", err)
	}

	// Handle shutdown signals
	go handleShutdown(func(ctx context.Context) {
		srv.Shutdown(ctx)
	})

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Proxy server error: %v", err)
	}
}

func handleShutdown(shutdown func(context.Context)) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logging.Info("Received signal %v, shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shutdown(ctx)
	os.Exit(0)
}
