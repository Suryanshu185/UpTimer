package main

import (
	"time"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	
	Monitoring struct {
		Interval time.Duration `yaml:"interval"`
		Timeout  time.Duration `yaml:"timeout"`
	} `yaml:"monitoring"`
	
	Websites []Website `yaml:"websites"`
	
	Alerts struct {
		Email struct {
			Enabled    bool     `yaml:"enabled"`
			SMTPServer string   `yaml:"smtp_server"`
			SMTPPort   int      `yaml:"smtp_port"`
			Username   string   `yaml:"username"`
			Password   string   `yaml:"password"`
			From       string   `yaml:"from"`
			To         []string `yaml:"to"`
		} `yaml:"email"`
		Slack struct {
			Enabled    bool   `yaml:"enabled"`
			WebhookURL string `yaml:"webhook_url"`
			Channel    string `yaml:"channel"`
		} `yaml:"slack"`
	} `yaml:"alerts"`
	
	Metrics struct {
		Enabled bool   `yaml:"enabled"`
		Path    string `yaml:"path"`
	} `yaml:"metrics"`
}

type Website struct {
	Name           string `yaml:"name"`
	URL            string `yaml:"url"`
	Method         string `yaml:"method"`
	ExpectedStatus int    `yaml:"expected_status"`
}

type MonitorResult struct {
	Website      Website
	Success      bool
	ResponseTime time.Duration
	StatusCode   int
	Error        string
	Timestamp    time.Time
}