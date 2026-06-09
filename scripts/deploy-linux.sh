#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
CONF_DIR="$ROOT_DIR/.chat2api/conf"
LOG_DIR="$ROOT_DIR/.chat2api/logs"
PROD_CONFIG="$CONF_DIR/app.prod.yaml"
EXAMPLE_CONFIG="$ROOT_DIR/conf/app.prod.example.yaml"

mkdir -p "$CONF_DIR" "$LOG_DIR"

if [ ! -f "$PROD_CONFIG" ]; then
  cp "$EXAMPLE_CONFIG" "$PROD_CONFIG"
  if command -v openssl >/dev/null 2>&1; then
    TOKEN="sk-$(openssl rand -hex 24)"
    sed -i "s/sk-change-me/$TOKEN/g" "$PROD_CONFIG"
    echo "Generated local API key: $TOKEN"
  else
    echo "Created $PROD_CONFIG with placeholder key sk-change-me"
    echo "Edit auth.access_tokens before exposing the service."
  fi
else
  echo "Using existing config: $PROD_CONFIG"
fi

cd "$ROOT_DIR"

if docker compose version >/dev/null 2>&1; then
  docker compose up -d --build
elif command -v docker-compose >/dev/null 2>&1; then
  docker-compose up -d --build
else
  echo "docker compose is required." >&2
  exit 1
fi

echo
echo "WebUI: http://SERVER_IP:${CHAT2API_PORT:-3040}/admin"
echo "OpenAI-compatible base URL: http://SERVER_IP:${CHAT2API_PORT:-3040}/v1"
