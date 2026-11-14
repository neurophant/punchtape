#!/bin/bash
# Gouija Relay Node Installation Script for Ubuntu Server
# This script installs and configures Gouija as a systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
LISTEN_PORT="8080"
PROXY_PORT="9000"
TRANSPORT="tcp"
REQUIRE_AUTH="true"
POOL_SIZE="5"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Gouija Relay Node Installation${NC}"
echo -e "${GREEN}========================================${NC}"
echo

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: Please run as root (sudo)${NC}"
    exit 1
fi

# Check if Ubuntu/Debian
if ! command -v apt &> /dev/null; then
    echo -e "${RED}Error: This script requires apt package manager (Ubuntu/Debian)${NC}"
    exit 1
fi

# Generate credential function
generate_credential() {
    openssl rand -base64 "$1" | tr -d '/+=' | head -c "$1"
}

# Prompt for configuration
echo -e "${YELLOW}Configuration${NC}"
echo "============="
echo

read -p "Listen port [$LISTEN_PORT]: " input_port
LISTEN_PORT="${input_port:-$LISTEN_PORT}"

read -p "Proxy server address (IP or hostname): " PROXY_ADDR
if [ -z "$PROXY_ADDR" ]; then
    echo -e "${RED}Error: Proxy server address is required${NC}"
    exit 1
fi

read -p "Proxy server port [$PROXY_PORT]: " input_proxy_port
PROXY_PORT="${input_proxy_port:-$PROXY_PORT}"

read -p "Transport mode (tcp/ws) [$TRANSPORT]: " input_transport
TRANSPORT="${input_transport:-$TRANSPORT}"

read -p "Connection pool size [$POOL_SIZE]: " input_pool
POOL_SIZE="${input_pool:-$POOL_SIZE}"

echo
echo -e "${YELLOW}Enter credentials from proxy node installation:${NC}"
read -p "GOUIJA_KEY: " GOUIJA_KEY
if [ -z "$GOUIJA_KEY" ]; then
    echo -e "${RED}Error: GOUIJA_KEY is required${NC}"
    exit 1
fi

read -p "GOUIJA_SALT: " GOUIJA_SALT
if [ -z "$GOUIJA_SALT" ]; then
    echo -e "${RED}Error: GOUIJA_SALT is required${NC}"
    exit 1
fi

echo
echo -e "${YELLOW}HTTP Proxy Authentication (for client access):${NC}"
read -p "Auth username [proxyuser]: " AUTH_USER
AUTH_USER="${AUTH_USER:-proxyuser}"

read -p "Generate random password? (y/n) [y]: " gen_pass
gen_pass="${gen_pass:-y}"
if [ "$gen_pass" = "y" ] || [ "$gen_pass" = "Y" ]; then
    AUTH_PASS=$(generate_credential 24)
    echo -e "Generated password: ${GREEN}$AUTH_PASS${NC}"
else
    read -s -p "Enter password: " AUTH_PASS
    echo
fi

echo
read -p "Press Enter to continue or Ctrl+C to cancel..."

# Install Go if not present
echo
echo -e "${YELLOW}Checking Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo "Installing Go via snap..."
    snap install go --classic
else
    echo "Go is already installed: $(go version)"
fi

# Create temporary directory for build
echo
echo -e "${YELLOW}Building Gouija...${NC}"
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

git clone --depth 1 https://github.com/neurophant/gouija.git
cd gouija
go build -o gouija ./cmd/gouija

# Install binary
echo
echo -e "${YELLOW}Installing binary...${NC}"
mv gouija /usr/local/bin/
chmod +x /usr/local/bin/gouija

# Cleanup
cd /
rm -rf "$TEMP_DIR"

# Create system user
echo
echo -e "${YELLOW}Creating system user...${NC}"
if ! id "gouija" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin gouija
    echo "User 'gouija' created"
else
    echo "User 'gouija' already exists"
fi

# Create configuration directory
echo
echo -e "${YELLOW}Creating configuration...${NC}"
mkdir -p /etc/gouija

# Create environment file
cat > /etc/gouija/relay.env << EOF
GOUIJA_KEY=$GOUIJA_KEY
GOUIJA_SALT=$GOUIJA_SALT
GOUIJA_AUTH_USER=$AUTH_USER
GOUIJA_AUTH_PASS=$AUTH_PASS
GOUIJA_QUIET=1
EOF

chmod 600 /etc/gouija/relay.env
chown gouija:gouija /etc/gouija/relay.env

# Prepare ExecStart command
EXEC_START="/usr/local/bin/gouija relay --listen :$LISTEN_PORT --proxy-addr $PROXY_ADDR:$PROXY_PORT --transport $TRANSPORT --pool-size $POOL_SIZE"
if [ "$REQUIRE_AUTH" = "true" ]; then
    EXEC_START="$EXEC_START --require-auth"
fi

# Create systemd service
echo
echo -e "${YELLOW}Creating systemd service...${NC}"
cat > /etc/systemd/system/gouija-relay.service << EOF
[Unit]
Description=Gouija Relay Node
After=network.target

[Service]
Type=simple
User=gouija
Group=gouija
EnvironmentFile=/etc/gouija/relay.env
ExecStart=$EXEC_START
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Configure firewall
echo
echo -e "${YELLOW}Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    read -p "Allow external access to relay port $LISTEN_PORT? (y/n) [n]: " allow_external
    if [ "$allow_external" = "y" ] || [ "$allow_external" = "Y" ]; then
        ufw allow "$LISTEN_PORT/tcp"
        echo "Allowed port $LISTEN_PORT/tcp in UFW"
    else
        echo "Skipping firewall rule (relay only accessible locally)"
    fi
else
    echo "UFW not found, skipping firewall configuration"
fi

# Enable and start service
echo
echo -e "${YELLOW}Starting service...${NC}"
systemctl daemon-reload
systemctl enable gouija-relay
systemctl start gouija-relay

# Wait a moment for service to start
sleep 2

# Check status
echo
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo
echo -e "Service status:"
systemctl status gouija-relay --no-pager || true
echo
echo -e "${YELLOW}Configuration Summary:${NC}"
echo "  Listen: :$LISTEN_PORT"
echo "  Proxy: $PROXY_ADDR:$PROXY_PORT"
echo "  Transport: $TRANSPORT"
echo "  Pool Size: $POOL_SIZE"
echo
echo -e "${YELLOW}Client Credentials:${NC}"
echo "  Username: $AUTH_USER"
echo "  Password: $AUTH_PASS"
echo
echo -e "${YELLOW}Client Usage:${NC}"
echo "  curl -x http://$AUTH_USER:$AUTH_PASS@localhost:$LISTEN_PORT https://example.com"
echo
echo -e "${YELLOW}Browser Configuration:${NC}"
echo "  Proxy: localhost:$LISTEN_PORT"
echo "  Username: $AUTH_USER"
echo "  Password: $AUTH_PASS"
echo
echo -e "${YELLOW}Useful commands:${NC}"
echo "  sudo systemctl status gouija-relay"
echo "  sudo systemctl restart gouija-relay"
echo "  sudo journalctl -u gouija-relay -f"
echo "  sudo cat /etc/gouija/relay.env"
