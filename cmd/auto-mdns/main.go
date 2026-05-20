// Package main is the entry point for the Home Assistant addon auto-mDNS service.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ha-addon-auto-mdns/ha"
	"ha-addon-auto-mdns/mdns"
	"ha-addon-auto-mdns/proxy"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Read configuration from environment
	supervisorURL := getEnv("SUPERVISOR_URL", "http://supervisor")
	supervisorToken := getEnv("SUPERVISOR_TOKEN", "")
	listenAddr := getEnv("LISTEN_ADDR", ":80")
	// Also listen on 443 for HTTPS
	httpsListenAddr := getEnv("HTTPS_LISTEN_ADDR", ":443")
	pollInterval := getDurationEnv("POLL_INTERVAL", 30*time.Second)

	log.Printf("Starting ha-addon-auto-mdns")
	log.Printf("Supervisor URL: %s", supervisorURL)
	log.Printf("Listen address: %s, HTTPS: %s", listenAddr, httpsListenAddr)

	// Create the Home Assistant client
	client, err := ha.NewClient(supervisorURL, supervisorToken)
	if err != nil {
		log.Fatalf("Failed to create HA client: %v", err)
	}

	// Create the reverse proxy manager
	proxyMgr := proxy.NewManager()

	// Create the mDNS advertiser
	advertiser := mdns.NewAdvertiser()

	// Get local hostname for mDNS
	localHostname := getEnv("LOCAL_HOSTNAME", "auto-mdns")
	if err := advertiser.Start(fmt.Sprintf("%s.local", localHostname)); err != nil {
		log.Printf("Warning: Failed to start mDNS: %v", err)
	}

	// Start the reverse proxy server
	proxyServer, err := proxyMgr.StartProxyServer(listenAddr)
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}

	log.Printf("Reverse proxy started on %s", listenAddr)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Main polling loop
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Run initial discovery
	services, err := discoverServices(ctx, client)
	if err != nil {
		log.Printf("Error discovering services: %v", err)
	} else {
		updateAll(services, proxyMgr, advertiser)
	}

	// Periodically poll for changes
	for {
		select {
		case <-ticker.C:
			services, err := discoverServices(ctx, client)
			if err != nil {
				log.Printf("Error discovering services: %v", err)
				continue
			}
			updateAll(services, proxyMgr, advertiser)

		case <-ctx.Done():
			log.Println("Shutting down...")
			// Stop mDNS advertiser
			advertiser.Stop()
			// Shut down proxy server
			proxyServer.Shutdown(ctx)
			return
		}
	}
}

// discoverServices queries the HA API and returns a list of addon services.
func discoverServices(ctx context.Context, client *ha.Client) ([]ha.AddonService, error) {
	addons, err := client.ListAddons(ctx)
	if err != nil {
		return nil, fmt.Errorf("list addons: %w", err)
	}

	if len(addons) == 0 {
		log.Println("No addons found")
		return nil, nil
	}

	log.Printf("Found %d addons", len(addons))

	var services []ha.AddonService
	for _, addon := range addons {
		// Generate an mDNS name from the addon slug
		mDNSName := fmt.Sprintf("%s.local", addon.Slug)

		// Assign service ports (use the primary port or the first available)
		servicePorts := make(map[string]int)
		if addon.Ports != nil {
			for portName, port := range addon.Ports {
				servicePorts[portName] = port
			}
		} else if addon.Port > 0 {
			servicePorts["http"] = addon.Port
		}

		// Use the primary port as the proxy port
		proxyPort := 80
		if len(servicePorts) > 0 {
			for _, port := range servicePorts {
				proxyPort = port
				break
			}
		}

		// Determine protocol
		protocol := "http"
		if addon.SSL {
			protocol = "https"
		}

		service := ha.NewAddonService(
			mDNSName,
			addon.Slug,
			servicePorts,
			proxyPort,
			protocol,
			addon,
		)

		services = append(services, service)
	}

	return services, nil
}

// updateAll updates the reverse proxy and mDNS advertiser with the latest services.
func updateAll(services []ha.AddonService, proxyMgr *proxy.Manager, advertiser *mdns.Advertiser) {
	if len(services) == 0 {
		log.Println("No services to update")
		return
	}

	log.Printf("Updating %d services", len(services))

	// Update the reverse proxy manager
	proxyMgr.UpdateServices(services)

	// Update mDNS advertiser
	mdnsServices := make([]mdns.Service, 0, len(services))
	for _, svc := range services {
		mdnsServices = append(mdnsServices, mdns.NewService(
			svc.MdnsName,
			net.ParseIP(svc.Original.IP),
		))
	}
	advertiser.UpdateServices(mdnsServices)
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// getDurationEnv returns the value of a duration environment variable or a default.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return defaultValue
}
