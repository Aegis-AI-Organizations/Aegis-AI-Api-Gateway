package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const installScriptTemplate = `#!/bin/bash
set -euo pipefail

# Configuration
AGENT_USER="aegis-agent"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/aegis-agent"
VAR_DIR="/var/lib/aegis-agent"
BINARY_NAME="aegis-ai-agent"
SERVICE_NAME="aegis-agent.service"

echo "Installing Aegis AI Agent..."

# Create restricted user
if ! id "$AGENT_USER" &>/dev/null; then
    echo "Creating restricted user: $AGENT_USER"
    useradd --system --shell /usr/sbin/nologin --no-create-home "$AGENT_USER"
fi

# Download binary from release bucket
echo "Downloading Aegis AI Agent binary..."
curl -fsSL -o "$INSTALL_DIR/$BINARY_NAME" "https://storage.aegis-ai.fr/releases/aegis-ai-agent"
chmod 755 "$INSTALL_DIR/$BINARY_NAME"

# Prepare config directory
mkdir -p "$CONFIG_DIR"
echo "Creating environment template at $CONFIG_DIR/agent.env"
cat <<EOF > "$CONFIG_DIR/agent.env"
# Aegis AI Agent Configuration
GATEWAY_URL=https://api.aegis-ai.fr
DEPLOYMENT_TOKEN=TOKEN_VALUE
INGEST_HOST=localhost
INGEST_PORT=7233
# Bind health checks to localhost by default.
HEALTH_BIND_ADDR=127.0.0.1
HEALTH_PORT=8081
EOF

chown -R root:root "$CONFIG_DIR"
chmod 755 "$CONFIG_DIR"
chmod 600 "$CONFIG_DIR/agent.env"

# Create working directory
mkdir -p "$VAR_DIR"
chown "$AGENT_USER:$AGENT_USER" "$VAR_DIR"
chmod 700 "$VAR_DIR"

# Create systemd service
echo "Creating systemd service..."
cat <<EOF > /etc/systemd/system/$SERVICE_NAME
[Unit]
Description=Aegis AI Agent
After=network.target

[Service]
Type=simple
User=$AGENT_USER
WorkingDirectory=$VAR_DIR
EnvironmentFile=$CONFIG_DIR/agent.env
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
DeviceAllow=/dev/null rw
ProtectSystem=full
ProtectHome=true
CapabilityBoundingSet=
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl restart $SERVICE_NAME

echo "-------------------------------------------------------"
echo "Aegis AI Agent installed and started successfully."
echo "-------------------------------------------------------"
`

// InstallScriptHandler serves the installer script dynamically with the token pre-injected.
func (a *API) InstallScriptHandler(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" || !strings.HasPrefix(token, "ag_") {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusBadRequest, "missing or invalid agent deployment token\n")
		return
	}

	scriptContent := strings.Replace(installScriptTemplate, "TOKEN_VALUE", token, 1)

	c.Header("Content-Type", "text/x-shellscript")
	c.String(http.StatusOK, scriptContent)
}
