# Gouija - Two-Stage Encrypted Proxy

Gouija is a high-performance two-stage encrypted proxy that provides secure communication through encrypted channels. It splits the proxy functionality into two components that communicate over an encrypted channel.

> **Note:** This project is a complete rethinking and continuation of the deprecated [Ouija](https://github.com/neurophant/ouija) project, rewritten from scratch in Go with improved architecture, security, and performance.

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Automated Installation (Ubuntu)](#automated-installation-ubuntu)
- [Docker Deployment](#docker-deployment)
- [Server Deployment (Ubuntu)](#server-deployment-ubuntu)
  - [Deploying Proxy Node](#deploying-proxy-node)
  - [Deploying Relay Node](#deploying-relay-node)
  - [WSS Deployment with Let's Encrypt](#wss-deployment-with-lets-encrypt)
- [Operating Modes](#operating-modes)
- [Configuration Reference](#configuration-reference)
  - [Relay Mode Options](#relay-mode-options)
  - [Proxy Mode Options](#proxy-mode-options)
  - [Environment Variables](#environment-variables)
- [Client Configuration Examples](#client-configuration-examples)
- [Performance Tips](#performance-tips)
- [Security Considerations](#security-considerations)
  - [Generating Secure Credentials](#generating-secure-credentials)
  - [Maximum Security Configuration](#maximum-security-configuration)
- [Troubleshooting](#troubleshooting)
- [Changelog](#changelog)
- [License](#license)

## Architecture

```
[Client] ←HTTP CONNECT→ [Relay Node] ←Encrypted Channel→ [Proxy Node] ←TCP→ [Target]
```

- **Relay Node**: Acts as a standard HTTP proxy that encrypts and forwards traffic to the proxy node.
- **Proxy Node**: Decrypts traffic and connects to target resources.

## Features

- **ChaCha20-Poly1305 encryption** with Argon2id key derivation
- **Two transport modes**: Raw TCP and WebSocket
- **Connection pooling** with multiplexing (multiple tunnels per connection)
- **HTTP Proxy Authentication** (Basic Auth) for relay node access control
- **Challenge-Response Authentication** (HMAC-SHA256) for proxy node protection
- **Connection Management** - Connection pacing, exponential backoff, adaptive pool sizing
- **Bandwidth Throttling** - Optional rate limiting for traffic smoothing
- **Graceful shutdown** with proper connection cleanup

## Prerequisites

### macOS

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Verify installation
go version
```

### Ubuntu/Debian

**Option 1: Install via Snap (recommended)**
```bash
# Install Go via Snap
sudo snap install go --classic

# Verify installation
go version
```

**Option 2: Install from official binary**
```bash
# Update package list
sudo apt update

# Download Go
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz

# Remove any previous Go installation and extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# Add Go to PATH (add to ~/.bashrc or ~/.zshrc for persistence)
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

### Requirements

- **Go 1.24.4** or later
- **Git** for cloning the repository

All Go dependencies are managed via Go modules and will be automatically downloaded during build:
- `github.com/gorilla/websocket` - WebSocket support
- `golang.org/x/crypto` - ChaCha20-Poly1305 encryption and Argon2id key derivation

## Installation

```bash
# Clone and build
git clone https://github.com/neurophant/gouija.git
cd gouija
go build -o gouija ./cmd/gouija

# Or install directly
go install github.com/neurophant/gouija/cmd/gouija@latest
```

## Quick Start

### Basic Setup (TCP Mode)

**On your external server:**
```bash
./gouija proxy --listen :9000 --key "your-secret-key-here"
```

**On your local machine:**
```bash
./gouija relay --listen :8080 --proxy-addr your-server-ip:9000 --key "your-secret-key-here"
```

**Use the proxy:**
```bash
curl -x http://localhost:8080 https://example.com
```

---

# Automated Installation (Ubuntu)

For quick deployment on Ubuntu Server, use the provided installation scripts. These scripts automatically:
- Install Go if not present
- Build Gouija from source
- Create system user and directories
- Generate secure credentials
- Configure systemd service
- Set up firewall rules

## Install Proxy Node

```bash
# Download and run the installation script
curl -fsSL https://raw.githubusercontent.com/neurophant/gouija/main/scripts/install-proxy.sh | sudo bash

# Or if you have cloned the repository:
sudo bash scripts/install-proxy.sh
```

The script will:
1. Prompt for listen port (default: 9000)
2. Prompt for transport mode (tcp/ws)
3. Generate secure key and salt
4. Install and start the service

**Important:** Save the generated credentials - you'll need them for the relay node!

## Install Relay Node

```bash
# Download and run the installation script
curl -fsSL https://raw.githubusercontent.com/neurophant/gouija/main/scripts/install-relay.sh | sudo bash

# Or if you have cloned the repository:
sudo bash scripts/install-relay.sh
```

The script will:
1. Prompt for listen port (default: 8080)
2. Prompt for proxy server address and port
3. Request key and salt from proxy installation
4. Generate HTTP auth credentials for clients
5. Install and start the service

---

# Docker Deployment

Deploy Gouija using Docker for containerized environments. The unified docker-compose.yml allows running proxy and relay nodes separately or together.

## Prerequisites

- Docker and Docker Compose installed
- OpenSSL for credential generation

## Quick Start

```bash
# Clone the repository
git clone https://github.com/neurophant/gouija.git
cd gouija/docker

# Generate environment configuration (interactive)
chmod +x generate-env.sh
./generate-env.sh

# Build and deploy both services (for local testing)
docker compose up -d

# Or deploy only proxy node (on external server)
docker compose up -d gouija-proxy

# Or deploy only relay node (on local machine)
docker compose up -d gouija-relay
```

## Deploy Proxy Node Only

On your external server:

```bash
git clone https://github.com/neurophant/gouija.git
cd gouija/docker
./generate-env.sh
docker compose up -d gouija-proxy
```

**Save the generated credentials** for the relay node setup!

## Deploy Relay Node Only

On your local machine (update PROXY_ADDR in .env first):

```bash
git clone https://github.com/neurophant/gouija.git
cd gouija/docker
./generate-env.sh
# Edit .env and set PROXY_ADDR to your proxy server IP:port
docker compose up -d gouija-relay
```

## Docker Management Commands

```bash
# View logs
docker compose logs -f
docker compose logs -f gouija-proxy
docker compose logs -f gouija-relay

# Stop services
docker compose down

# Restart
docker compose restart gouija-proxy
docker compose restart gouija-relay

# View running containers
docker ps
```

---

# Server Deployment (Ubuntu)

## Deploying Proxy Node

### 1. Install and Build

```bash
# Install Go
sudo snap install go --classic

# Clone and build
git clone https://github.com/neurophant/gouija.git
cd gouija
go build -o gouija ./cmd/gouija

# Move binary to system location
sudo mv gouija /usr/local/bin/
sudo chmod +x /usr/local/bin/gouija
```

### 2. Create System User

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin gouija
```

### 3. Create Configuration Directory

```bash
sudo mkdir -p /etc/gouija
sudo chown gouija:gouija /etc/gouija
```

### 4. Create Environment File

```bash
sudo tee /etc/gouija/proxy.env << 'EOF'
GOUIJA_KEY=your-very-long-secret-key-at-least-32-chars
GOUIJA_SALT=your-custom-salt-2024
GOUIJA_QUIET=1
EOF

sudo chmod 600 /etc/gouija/proxy.env
sudo chown gouija:gouija /etc/gouija/proxy.env
```

### 5. Create Systemd Service

```bash
sudo tee /etc/systemd/system/gouija-proxy.service << 'EOF'
[Unit]
Description=Gouija Proxy Node
After=network.target

[Service]
Type=simple
User=gouija
Group=gouija
EnvironmentFile=/etc/gouija/proxy.env
ExecStart=/usr/local/bin/gouija proxy --listen :9000 --require-auth
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
```

### 6. Configure Firewall (UFW)

```bash
# Allow proxy port
sudo ufw allow 9000/tcp

# Enable firewall
sudo ufw enable
```

### 7. Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable gouija-proxy
sudo systemctl start gouija-proxy

# Check status
sudo systemctl status gouija-proxy

# View logs
sudo journalctl -u gouija-proxy -f
```

---

## Deploying Relay Node

### 1. Install and Build

```bash
# Install Go
sudo snap install go --classic

# Clone and build
git clone https://github.com/neurophant/gouija.git
cd gouija
go build -o gouija ./cmd/gouija

# Move binary to system location
sudo mv gouija /usr/local/bin/
sudo chmod +x /usr/local/bin/gouija
```

### 2. Create System User

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin gouija
```

### 3. Create Configuration Directory

```bash
sudo mkdir -p /etc/gouija
sudo chown gouija:gouija /etc/gouija
```

### 4. Create Environment File

```bash
sudo tee /etc/gouija/relay.env << 'EOF'
GOUIJA_KEY=your-very-long-secret-key-at-least-32-chars
GOUIJA_SALT=your-custom-salt-2024
GOUIJA_AUTH_USER=proxyuser
GOUIJA_AUTH_PASS=proxypassword
GOUIJA_QUIET=1
EOF

sudo chmod 600 /etc/gouija/relay.env
sudo chown gouija:gouija /etc/gouija/relay.env
```

### 5. Create Systemd Service

```bash
sudo tee /etc/systemd/system/gouija-relay.service << 'EOF'
[Unit]
Description=Gouija Relay Node
After=network.target

[Service]
Type=simple
User=gouija
Group=gouija
EnvironmentFile=/etc/gouija/relay.env
ExecStart=/usr/local/bin/gouija relay --listen :8080 --proxy-addr YOUR_PROXY_SERVER_IP:9000 --require-auth --pool-size 5
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
```

**Note:** Replace `YOUR_PROXY_SERVER_IP` with your actual proxy server IP address.

### 6. Configure Firewall (UFW)

```bash
# Allow relay port (only if you need external access)
sudo ufw allow 8080/tcp

# Enable firewall
sudo ufw enable
```

### 7. Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable gouija-relay
sudo systemctl start gouija-relay

# Check status
sudo systemctl status gouija-relay

# View logs
sudo journalctl -u gouija-relay -f
```

---

## WSS Deployment with Let's Encrypt

For production WSS (WebSocket Secure) deployment:

### 1. Install Certbot

```bash
sudo apt update
sudo apt install certbot
```

### 2. Obtain Certificate

```bash
sudo certbot certonly --standalone -d your-domain.com
```

### 3. Update Proxy Service for WSS

```bash
sudo tee /etc/systemd/system/gouija-proxy.service << 'EOF'
[Unit]
Description=Gouija Proxy Node (WSS)
After=network.target

[Service]
Type=simple
User=root
EnvironmentFile=/etc/gouija/proxy.env
ExecStart=/usr/local/bin/gouija proxy --listen :443 --transport ws --require-auth \
  --tls-cert /etc/letsencrypt/live/your-domain.com/fullchain.pem \
  --tls-key /etc/letsencrypt/live/your-domain.com/privkey.pem
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl restart gouija-proxy
```

### 4. Update Relay to Use WSS

Update the relay service to connect via WSS:

```bash
ExecStart=/usr/local/bin/gouija relay --listen :8080 --proxy-addr your-domain.com:443 \
  --require-auth --transport ws --tls --pool-size 5
```

---

# Operating Modes

## Mode 1: Basic TCP Mode

The simplest mode using raw TCP connections. Best performance.

**Proxy Node:**
```bash
./gouija proxy --listen :9000 --key "secret"
```

**Relay Node:**
```bash
./gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret"
```

## Mode 2: TCP with HTTP Proxy Authentication

Protects your relay node from unauthorized users. Clients must provide credentials.

**Proxy Node:**
```bash
./gouija proxy --listen :9000 --key "secret"
```

**Relay Node:**
```bash
./gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" \
  --auth-user myuser --auth-pass mypassword
```

**Client Usage:**
```bash
curl -x http://myuser:mypassword@localhost:8080 https://example.com

# Or configure in browser/system proxy settings:
# Proxy: localhost:8080
# Username: myuser
# Password: mypassword
```

## Mode 3: TCP with Challenge-Response Authentication

Protects your proxy node from unauthorized relay nodes. Both sides must enable `--require-auth`.

**Proxy Node:**
```bash
./gouija proxy --listen :9000 --key "secret" --require-auth
```

**Relay Node:**
```bash
./gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" --require-auth
```

## Mode 4: WebSocket Transport

Uses WebSocket protocol instead of raw TCP.

**Proxy Node:**
```bash
./gouija proxy --listen :9000 --key "secret" --transport ws
```

**Relay Node:**
```bash
./gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" --transport ws
```

## Mode 5: WebSocket with TLS (WSS)

Most secure mode using HTTPS protocol.

**Proxy Node (requires SSL certificate):**
```bash
./gouija proxy --listen :443 --key "secret" --transport ws \
  --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem
```

**Relay Node:**
```bash
./gouija relay --listen :8080 --proxy-addr proxy.example.com:443 --key "secret" \
  --transport ws --tls
```

## Mode 6: Full Security (All Features)

Combines all authentication and encryption features.

**Proxy Node:**
```bash
./gouija proxy --listen :443 --key "secret" --transport ws --require-auth \
  --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem
```

**Relay Node:**
```bash
./gouija relay --listen :8080 --proxy-addr proxy.example.com:443 --key "secret" \
  --transport ws --tls --require-auth \
  --auth-user myuser --auth-pass mypassword
```

---

# Configuration Reference

## Relay Mode Options

### Basic Options

| Flag | Description | Default |
|------|-------------|---------|
| `--listen` | Address to listen on | `:8080` |
| `--proxy-addr` | Proxy node address (required) | - |
| `--key` | Pre-shared encryption key | `$GOUIJA_KEY` |
| `--salt` | Custom salt for key derivation | `$GOUIJA_SALT` |
| `--pool-size` | Target number of connections to proxy | `5` |
| `--timeout` | Connection timeout | `30s` |
| `--transport` | Transport mode: `tcp` or `ws` | `tcp` |
| `--ws-path` | WebSocket endpoint path | `/ws` |
| `--tls` | Use TLS for WebSocket | `false` |
| `--auth-user` | HTTP Proxy auth username | - |
| `--auth-pass` | HTTP Proxy auth password | - |
| `--require-auth` | Enable challenge-response auth | `false` |
| `--quiet` | Suppress informational logging | `false` |

### Connection Management - Connection Pacing

| Flag | Description | Default |
|------|-------------|---------|
| `--conn-pacing` | Delay between new connections | `500ms` |
| `--max-concurrent-dials` | Max simultaneous dial attempts | `2` |

### Connection Management - Exponential Backoff

| Flag | Description | Default |
|------|-------------|---------|
| `--backoff-initial` | Initial backoff interval | `5s` |
| `--backoff-max` | Maximum backoff interval | `120s` |
| `--backoff-multiplier` | Backoff multiplier | `2.0` |
| `--backoff-jitter` | Jitter factor (0-1) | `0.2` |

### Adaptive Pool Sizing

| Flag | Description | Default |
|------|-------------|---------|
| `--min-pool-size` | Minimum connections | `1` |
| `--max-pool-size` | Maximum connections | `10` |
| `--growth-threshold` | Grow when utilization exceeds | `0.8` |
| `--shrink-threshold` | Shrink when utilization below | `0.2` |

### TCP Settings

| Flag | Description | Default |
|------|-------------|---------|
| `--keepalive` | Enable TCP keepalive | `true` |
| `--keepalive-interval` | Keepalive probe interval | `30s` |

### Bandwidth Throttling

| Flag | Description | Default |
|------|-------------|---------|
| `--max-bandwidth` | Max bandwidth (bytes/sec) | `0` (unlimited) |
| `--burst-size` | Max burst size (bytes) | `65536` |

## Proxy Mode Options

| Flag | Description | Default |
|------|-------------|---------|
| `--listen` | Address to listen on | `:9000` |
| `--key` | Pre-shared encryption key | `$GOUIJA_KEY` |
| `--salt` | Custom salt for key derivation | `$GOUIJA_SALT` |
| `--timeout` | Connection timeout | `30s` |
| `--transport` | Transport mode: `tcp` or `ws` | `tcp` |
| `--ws-path` | WebSocket endpoint path | `/ws` |
| `--tls-cert` | TLS certificate file | - |
| `--tls-key` | TLS private key file | - |
| `--require-auth` | Enable challenge-response auth | `false` |
| `--quiet` | Suppress informational logging | `false` |
| `--keepalive` | Enable TCP keepalive | `true` |
| `--keepalive-interval` | Keepalive probe interval | `30s` |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GOUIJA_KEY` | Pre-shared encryption key |
| `GOUIJA_SALT` | Custom salt for key derivation |
| `GOUIJA_AUTH_USER` | HTTP Proxy auth username |
| `GOUIJA_AUTH_PASS` | HTTP Proxy auth password |
| `GOUIJA_QUIET` | Set to "1" or "true" to enable quiet mode |

---

# Client Configuration Examples

## Firefox
1. Settings → Network Settings → Manual proxy configuration
2. HTTP Proxy: `localhost`, Port: `8080`
3. Check "Use this proxy server for all protocols"
4. If using auth, Firefox will prompt for credentials

## Chrome/Chromium
```bash
chromium --proxy-server="http://localhost:8080"
```

## System-wide (Linux)
```bash
export http_proxy="http://user:pass@localhost:8080"
export https_proxy="http://user:pass@localhost:8080"
```

## curl
```bash
curl -x http://user:pass@localhost:8080 https://example.com
```

## wget
```bash
wget -e use_proxy=yes -e http_proxy=http://user:pass@localhost:8080 https://example.com
```

## Python requests
```python
import requests
proxies = {
    'http': 'http://user:pass@localhost:8080',
    'https': 'http://user:pass@localhost:8080'
}
response = requests.get('https://example.com', proxies=proxies)
```

---

# Performance Tips

1. **Use `--quiet` flag** in production to disable informational logging and reduce I/O overhead
2. **Adjust pool size** based on your use case:
   - Light usage: 3-5 connections
   - Heavy usage: 5-10 connections
3. **Use TCP mode** when possible - it has lower overhead than WebSocket
4. **Consider geographic proximity** - choose server locations close to your target services

```bash
# High-performance production configuration
./gouija proxy --listen :9000 --key "secret" --quiet
./gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" --quiet --pool-size 10
```

```bash
# Conservative configuration
./gouija relay --listen :8080 --proxy-addr proxy.example.com:9000 --key "secret" --quiet \
  --pool-size 3 --conn-pacing 1s --max-concurrent-dials 1
```

---

# Security Considerations

1. **Use a strong key**: At least 32 characters, random
2. **Use custom salt**: The `--salt` option allows customizing the salt for key derivation, adding another layer of security
3. **Enable both auth types** for maximum security
4. **Use HTTPS (WSS)** in WebSocket mode
5. **Firewall your proxy node**: Only allow connections from known relay IPs
6. **Monitor logs** for suspicious activity (don't use `--quiet` when debugging)
7. **Rotate keys** periodically
8. **Keep software updated**

## Generating Secure Credentials

Use these commands to generate cryptographically secure random strings for key, salt, and password:

**Using OpenSSL (recommended):**
```bash
# Generate 32-character key (base64, alphanumeric-safe)
openssl rand -base64 32 | tr -d '/+=' | head -c 32

# Generate 16-character salt
openssl rand -base64 16 | tr -d '/+=' | head -c 16

# Generate 24-character password
openssl rand -base64 24 | tr -d '/+=' | head -c 24
```

**Using /dev/urandom:**
```bash
# Generate 32-character key
head -c 32 /dev/urandom | base64 | tr -d '/+=' | head -c 32

# Generate 16-character salt
head -c 16 /dev/urandom | base64 | tr -d '/+=' | head -c 16

# Generate 24-character password
head -c 24 /dev/urandom | base64 | tr -d '/+=' | head -c 24
```

**Generate all credentials at once:**
```bash
echo "GOUIJA_KEY=$(openssl rand -base64 32 | tr -d '/+=' | head -c 32)"
echo "GOUIJA_SALT=$(openssl rand -base64 16 | tr -d '/+=' | head -c 16)"
echo "GOUIJA_AUTH_USER=proxyuser"
echo "GOUIJA_AUTH_PASS=$(openssl rand -base64 24 | tr -d '/+=' | head -c 24)"
```

## Maximum Security Configuration

```bash
# Maximum security configuration with custom salt
./gouija proxy --listen :443 --key "your-very-long-secret-key" --salt "custom-salt-2024" \
  --transport ws --require-auth --tls-cert cert.pem --tls-key key.pem

./gouija relay --listen :8080 --proxy-addr proxy.example.com:443 \
  --key "your-very-long-secret-key" --salt "custom-salt-2024" \
  --transport ws --tls --require-auth --auth-user myuser --auth-pass mypassword
```

---

# Troubleshooting

## "connection refused"
- Check if proxy node is running
- Check firewall rules on server
- Verify port is correct

## "authentication failed"
- Verify keys match on both sides
- Check for typos in credentials

## "timeout"
- Network latency too high
- Try different server location
- Check if server is reachable

## "certificate error" (WSS mode)
- Verify certificate paths are correct
- Check certificate is not expired
- Ensure domain matches certificate

## Connection issues
- Reduce `--pool-size` to 3-5
- Increase `--conn-pacing` to 1-2 seconds
- Set `--max-concurrent-dials 1`
- Increase `--backoff-initial` and `--backoff-max`
- Try using WebSocket transport mode

## Streaming issues
- Use bandwidth throttling: `--max-bandwidth 10485760` (10 MB/s)
- Reduce pool size to avoid overwhelming network
- Consider using WSS mode

---

# Changelog

## v0.1.0 (2026-01-16)
- Initial release
- **Encryption**: ChaCha20-Poly1305 with Argon2id key derivation
- **Transport Modes**: Raw TCP and WebSocket (WS/WSS)
- **Connection Pooling**: Multiplexing with adaptive pool sizing
- **Authentication**: HTTP Proxy auth (Basic) and Challenge-Response (HMAC-SHA256)
- **Connection Management**: Connection pacing, exponential backoff with jitter
- **Bandwidth Throttling**: Token bucket rate limiting
- **TCP Keepalive**: Configurable keepalive settings
- **Graceful Shutdown**: Proper connection cleanup
- **Installation Scripts**: Automated installation for Ubuntu Server (proxy and relay nodes)
- **Docker Support**: Dockerfiles and Docker Compose for containerized deployment

---

# License

MIT License - See LICENSE file for details.

---

# Disclaimer

This software is provided for educational and legitimate purposes only. Users are responsible for complying with applicable laws and regulations.
