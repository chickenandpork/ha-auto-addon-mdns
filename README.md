# HA Add-on for Automatic mDNS

This project build an add-on for HomeAssistant that:
 - creates a reverse-proxy
 - monitors add-ons, connects to their published endpoints via revproxy config, and advertises an alias to the revproxy itself as that addon's name

# Build

`docker build -t ha-addon-auto-mdns:latest -f addon/Dockerfile .`
