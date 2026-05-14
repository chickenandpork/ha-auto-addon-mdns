#!/usr/bin/env bash
set -e

# Read add-on configuration from Home Assistant Supervisor
if [ -f /data/options.json ]; then
    export LISTEN_ADDR=$(jq -r '.listen_address // ":80"' /data/options.json)
    export HTTPS_LISTEN_ADDR=$(jq -r '.https_listen_address // ":443"' /data/options.json)
    export POLL_INTERVAL=$(jq -r '.poll_interval // "30s"' /data/options.json)
    export LOCAL_HOSTNAME=$(jq -r '.local_hostname // "auto-mdns"' /data/options.json)
fi

# Supervisor URL - HA injects this automatically, but provide a sensible default
if [ -z "$SUPERVISOR_URL" ]; then
    export SUPERVISOR_URL="http://supervisor"
fi

# Supervisor token - injected automatically by HA
if [ -z "$SUPERVISOR_TOKEN" ]; then
    echo "Warning: SUPERVISOR_TOKEN not set. The addon may not work correctly."
fi

echo "Starting ha-addon-auto-mdns..."
echo "  Supervisor URL: $SUPERVISOR_URL"
echo "  Listen addr: $LISTEN_ADDR"
echo "  HTTPS addr: $HTTPS_LISTEN_ADDR"
echo "  Poll interval: $POLL_INTERVAL"
echo "  Local hostname: $LOCAL_HOSTNAME"

exec /usr/local/bin/auto-mdns
