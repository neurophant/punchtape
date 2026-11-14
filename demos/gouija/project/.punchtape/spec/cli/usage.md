# cli/usage — the usage page

run: 3

## Story

help prints the whole usage page to stdout and exits zero

## Demands

- R2 (lines D2): the usage page prints

## Claims

- cmd/gouija/main.go:62-217 — 1a6dffe4bb6c
token: 467d33a0da70

## Rows

## row 0 — PROOF

command: ./gouija-demo help
exit: 0
stdout: Gouija - Two-stage encrypted proxy

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

proof: green (8482a630666c)

## Verify

- ./gouija-demo help
