#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="pixabros"
SERVICE_NAME="pixabros"
INSTALL_DIR="/opt/pixabros"
USER="${PIXABROS_USER:-pixabros}"
PORT="${PIXABROS_PORT:-8080}"

# ── Colors ──
RED='\033[0;31m'
GRN='\033[0;32m'
YEL='\033[0;33m'
NC='\033[0m'

say()  { echo -e "${GRN}→${NC} $*"; }
warn() { echo -e "${YEL}⚠${NC}  $*"; }
die()  { echo -e "${RED}✗${NC}  $*" >&2; exit 1; }

# ── Pre-flight ──
[ "$(id -u)" -eq 0 ] || die "Run as root: sudo ./install.sh"

# Detect binary
if [ -f "pixabros-linux" ]; then
  BIN_NAME="pixabros-linux"
elif [ -f "pixabros" ]; then
  BIN_NAME="pixabros"
else
  die "Binary not found. Place pixabros or pixabros-linux in this directory."
fi

# ── Create system user ──
if ! id "$USER" &>/dev/null; then
  say "Creating system user: $USER"
  useradd --system --no-create-home --shell /usr/sbin/nologin "$USER"
else
  say "User $USER already exists"
fi

# ── Install binary ──
say "Installing to $INSTALL_DIR"
mkdir -p "$INSTALL_DIR/dist"
cp "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
chmod 755 "$INSTALL_DIR/$BIN_NAME"

# Copy game embeds (not embedded in binary — read from disk)
if [ -d "dist/embeds" ]; then
  say "Copying game embeds..."
  cp -r dist/embeds "$INSTALL_DIR/dist/embeds"
fi
if [ -d "dist/browser-games" ]; then
  cp -r dist/browser-games "$INSTALL_DIR/dist/browser-games"
fi

chown -R "$USER:$USER" "$INSTALL_DIR"

# ── Write systemd service ──
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
say "Writing systemd service: $SERVICE_FILE"

cat > "$SERVICE_FILE" << SYSTEMD
[Unit]
Description=PixaBros Studio Website
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${USER}
Group=${USER}
WorkingDirectory=${INSTALL_DIR}
Environment=PORT=${PORT}
ExecStart=${INSTALL_DIR}/${BIN_NAME}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

# Security hardening
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=${INSTALL_DIR}
ProtectKernelTunables=yes
ProtectControlGroups=yes
RestrictRealtime=yes

[Install]
WantedBy=multi-user.target
SYSTEMD

# ── Start it ──
say "Reloading systemd & enabling service"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

# ── Done ──
sleep 1
if systemctl is-active --quiet "$SERVICE_NAME"; then
  say "Service is running on http://localhost:${PORT}"
else
  warn "Service might not have started. Check: journalctl -u ${SERVICE_NAME} -f"
fi

echo ""
say "Install complete!"
echo "  Binary : ${INSTALL_DIR}/${BIN_NAME}"
echo "  Service: systemctl {start|stop|restart|status} ${SERVICE_NAME}"
echo "  Port   : ${PORT}"
echo "  Logs   : journalctl -u ${SERVICE_NAME} -f"
