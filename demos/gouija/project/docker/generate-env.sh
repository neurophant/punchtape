#!/bin/bash
# Gouija Docker Environment Generator
# Generates .env file with secure credentials for both proxy and relay nodes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Generate secure random string
generate_credential() {
    openssl rand -base64 "$1" | tr -d '/+=' | head -c "$1"
}

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Gouija Docker Environment Setup${NC}"
echo -e "${GREEN}========================================${NC}"
echo

# Generate credentials
GOUIJA_KEY=$(generate_credential 32)
GOUIJA_SALT=$(generate_credential 16)
GOUIJA_AUTH_USER="proxyuser"
GOUIJA_AUTH_PASS=$(generate_credential 24)

# Prompt for configuration
echo -e "${YELLOW}Configuration${NC}"
echo "============="
echo

read -p "Proxy port [9000]: " PROXY_PORT
PROXY_PORT="${PROXY_PORT:-9000}"

read -p "Relay port [8080]: " RELAY_PORT
RELAY_PORT="${RELAY_PORT:-8080}"

read -p "Transport mode (tcp/ws) [tcp]: " TRANSPORT
TRANSPORT="${TRANSPORT:-tcp}"

read -p "Connection pool size [5]: " POOL_SIZE
POOL_SIZE="${POOL_SIZE:-5}"

echo
echo -e "${YELLOW}Proxy Address Configuration${NC}"
echo "For local testing, use: gouija-proxy:${PROXY_PORT}"
echo "For remote proxy, use: <proxy-server-ip>:${PROXY_PORT}"
read -p "Proxy address [gouija-proxy:${PROXY_PORT}]: " PROXY_ADDR
PROXY_ADDR="${PROXY_ADDR:-gouija-proxy:${PROXY_PORT}}"

# Create .env file
cat > .env << EOF
# Gouija Docker Configuration
# Generated: $(date -Iseconds)

# Encryption credentials
GOUIJA_KEY=$GOUIJA_KEY
GOUIJA_SALT=$GOUIJA_SALT

# HTTP Proxy Authentication (for clients connecting to relay)
GOUIJA_AUTH_USER=$GOUIJA_AUTH_USER
GOUIJA_AUTH_PASS=$GOUIJA_AUTH_PASS

# Ports
PROXY_PORT=$PROXY_PORT
RELAY_PORT=$RELAY_PORT

# Transport mode (tcp or ws)
TRANSPORT=$TRANSPORT

# Relay configuration
PROXY_ADDR=$PROXY_ADDR
POOL_SIZE=$POOL_SIZE
EOF

chmod 600 .env

echo
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Configuration Generated!${NC}"
echo -e "${GREEN}========================================${NC}"
echo
echo -e "File created: ${YELLOW}.env${NC}"
echo
echo -e "${YELLOW}Credentials:${NC}"
echo -e "  GOUIJA_KEY=${GREEN}$GOUIJA_KEY${NC}"
echo -e "  GOUIJA_SALT=${GREEN}$GOUIJA_SALT${NC}"
echo -e "  Auth User: ${GREEN}$GOUIJA_AUTH_USER${NC}"
echo -e "  Auth Pass: ${GREEN}$GOUIJA_AUTH_PASS${NC}"
echo
echo -e "${YELLOW}Usage:${NC}"
echo
echo "  # Deploy proxy node only (on external server):"
echo "  docker compose up -d gouija-proxy"
echo
echo "  # Deploy relay node only (on local machine):"
echo "  docker compose up -d gouija-relay"
echo
echo "  # Deploy both (for local testing):"
echo "  docker compose up -d"
echo
echo "  # View logs:"
echo "  docker compose logs -f"
echo
echo "  # Stop services:"
echo "  docker compose down"
echo
echo -e "${YELLOW}Test with:${NC}"
echo "  curl -x http://$GOUIJA_AUTH_USER:$GOUIJA_AUTH_PASS@localhost:$RELAY_PORT https://example.com"
