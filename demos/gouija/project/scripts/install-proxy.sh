#!/bin/bash
# Gouija Proxy Node Installation Script for Ubuntu Server
# This script installs and configures Gouija as a systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
LISTEN_PORT="9000"
TRANSPORT="tcp"
REQUIRE_AUTH="true"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Gouija Proxy Node Installation${NC}"
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

# Generate credentials function
generate_credential() {
    openssl rand -base64 "$1" | tr -d '/+=' | head -c "$1"
}

# Prompt for configuration
echo -e "${YELLOW}Configuration${NC}"
echo "============="
echo

read -p "Listen port [$LISTEN_PORT]: " input_port
LISTEN_PORT="${input_port:-$LISTEN_PORT}"

read -p "Transport mode (tcp/ws) [$TRANSPORT]: " input_transport
TRANSPORT="${input_transport:-$TRANSPORT}"

echo
echo -e "${YELLOW}Generating secure credentials...${NC}"
GOUIJA_KEY=$(generate_credential 32)
GOUIJA_SALT=$(generate_credential 16)

echo -e "Generated Key: ${GREEN}$GOUIJA_KEY${NC}"
echo -e "Generated Salt: ${GREEN}$GOUIJA_SALT${NC}"
echo
echo -e "${YELLOW}IMPORTANT: Save these credentials! You will need them for the relay node.${NC}"
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
cat > /etc/gouija/proxy.env << EOF
GOUIJA_KEY=$GOUIJA_KEY
GOUIJA_SALT=$GOUIJA_SALT
GOUIJA_QUIET=1
EOF

chmod 600 /etc/gouija/proxy.env
chown gouija:gouija /etc/gouija/proxy.env

# Prepare ExecStart command
EXEC_START="/usr/local/bin/gouija proxy --listen :$LISTEN_PORT --transport $TRANSPORT"
if [ "$REQUIRE_AUTH" = "true" ]; then
    EXEC_START="$EXEC_START --require-auth"
fi

# Create systemd service
echo
echo -e "${YELLOW}Creating systemd service...${NC}"
cat > /etc/systemd/system/gouija-proxy.service << EOF
[Unit]
Description=Gouija Proxy Node
After=network.target

[Service]
Type=simple
User=gouija
Group=gouija
EnvironmentFile=/etc/gouija/proxy.env
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
    ufw allow "$LISTEN_PORT/tcp"
    echo "Allowed port $LISTEN_PORT/tcp in UFW"
else
    echo "UFW not found, skipping firewall configuration"
fi

# Enable and start service
echo
echo -e "${YELLOW}Starting service...${NC}"
systemctl daemon-reload
systemctl enable gouija-proxy
systemctl start gouija-proxy

# Wait a moment for service to start
sleep 2

# Check status
echo
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo
echo -e "Service status:"
systemctl status gouija-proxy --no-pager || true
echo
echo -e "${YELLOW}Configuration Summary:${NC}"
echo "  Listen: :$LISTEN_PORT"
echo "  Transport: $TRANSPORT"
echo "  Auth Required: $REQUIRE_AUTH"
echo
echo -e "${YELLOW}Credentials (save these for relay node):${NC}"
echo "  GOUIJA_KEY=$GOUIJA_KEY"
echo "  GOUIJA_SALT=$GOUIJA_SALT"
echo
echo -e "${YELLOW}Useful commands:${NC}"
echo "  sudo systemctl status gouija-proxy"
echo "  sudo systemctl restart gouija-proxy"
echo "  sudo journalctl -u gouija-proxy -f"
echo "  sudo cat /etc/gouija/proxy.env"
