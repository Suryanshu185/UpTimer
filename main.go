package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"gopkg.in/yaml.v3"
)

func main() {
	log.Println("Starting UpTimer...")
	
	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize components
	metrics := NewMetrics()
	alerter := NewAlerter(config)
	monitor := NewMonitor(config, alerter, metrics)
	webServer := NewWebServer(config, monitor)
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start monitor in a goroutine
	go func() {
		log.Println("Starting monitor...")
		monitor.Start(ctx)
	}()
	
	// Start web server in a goroutine
	go func() {
		log.Println("Starting web server...")
		if err := webServer.Start(); err != nil {
			log.Fatalf("Web server failed: %v", err)
		}
	}()
	
	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal, stopping...")
	
	// Cancel context to stop monitor
	cancel()
	
	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	
	log.Println("UpTimer stopped")
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	
	if config.Monitoring.Interval <= 0 {
		return fmt.Errorf("monitoring interval must be positive")
	}
	
	if config.Monitoring.Timeout <= 0 {
		return fmt.Errorf("monitoring timeout must be positive")
	}
	
	if len(config.Websites) == 0 {
		return fmt.Errorf("at least one website must be configured")
	}
	
	for i, website := range config.Websites {
		if website.Name == "" {
			return fmt.Errorf("website %d: name is required", i)
		}
		if website.URL == "" {
			return fmt.Errorf("website %d: URL is required", i)
		}
		if website.Method == "" {
			config.Websites[i].Method = "GET"
		}
		if website.ExpectedStatus == 0 {
			config.Websites[i].ExpectedStatus = 200
		}
	}
	
	return nil
}