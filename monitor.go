package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Monitor struct {
	config  *Config
	results map[string][]MonitorResult
	mu      sync.RWMutex
	alerter *Alerter
	metrics *Metrics
}

func NewMonitor(config *Config, alerter *Alerter, metrics *Metrics) *Monitor {
	return &Monitor{
		config:  config,
		results: make(map[string][]MonitorResult),
		alerter: alerter,
		metrics: metrics,
	}
}

func (m *Monitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.Interval)
	defer ticker.Stop()
	
	// Initial check
	m.checkAllWebsites()
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Monitor stopped")
			return
		case <-ticker.C:
			m.checkAllWebsites()
		}
	}
}

func (m *Monitor) checkAllWebsites() {
	var wg sync.WaitGroup
	
	for _, website := range m.config.Websites {
		wg.Add(1)
		go func(site Website) {
			defer wg.Done()
			result := m.checkWebsite(site)
			m.storeResult(result)
			m.handleAlerts(result)
			m.updateMetrics(result)
		}(website)
	}
	
	wg.Wait()
}

func (m *Monitor) checkWebsite(website Website) MonitorResult {
	start := time.Now()
	result := MonitorResult{
		Website:   website,
		Timestamp: start,
	}
	
	client := &http.Client{
		Timeout: m.config.Monitoring.Timeout,
	}
	
	req, err := http.NewRequest(website.Method, website.URL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Success = false
		return result
	}
	
	req.Header.Set("User-Agent", "UpTimer-Monitor/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Success = false
		result.ResponseTime = time.Since(start)
		return result
	}
	defer resp.Body.Close()
	
	result.ResponseTime = time.Since(start)
	result.StatusCode = resp.StatusCode
	
	if resp.StatusCode == website.ExpectedStatus {
		result.Success = true
		log.Printf("✓ %s is UP (Status: %d, Time: %v)", website.Name, resp.StatusCode, result.ResponseTime)
	} else {
		result.Success = false
		result.Error = fmt.Sprintf("Unexpected status code: %d (expected: %d)", resp.StatusCode, website.ExpectedStatus)
		log.Printf("✗ %s is DOWN (Status: %d, Expected: %d, Time: %v)", website.Name, resp.StatusCode, website.ExpectedStatus, result.ResponseTime)
	}
	
	return result
}

func (m *Monitor) storeResult(result MonitorResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	websiteName := result.Website.Name
	results := m.results[websiteName]
	
	// Keep only last 100 results per website
	if len(results) >= 100 {
		results = results[1:]
	}
	
	results = append(results, result)
	m.results[websiteName] = results
}

func (m *Monitor) handleAlerts(result MonitorResult) {
	if !result.Success {
		m.alerter.SendAlert(result)
	}
}

func (m *Monitor) updateMetrics(result MonitorResult) {
	if m.metrics != nil {
		m.metrics.UpdateMetrics(result)
	}
}

func (m *Monitor) GetResults() map[string][]MonitorResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Create a copy to avoid data races
	results := make(map[string][]MonitorResult)
	for name, siteResults := range m.results {
		results[name] = make([]MonitorResult, len(siteResults))
		copy(results[name], siteResults)
	}
	
	return results
}

func (m *Monitor) GetLatestResults() map[string]MonitorResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	latest := make(map[string]MonitorResult)
	for name, results := range m.results {
		if len(results) > 0 {
			latest[name] = results[len(results)-1]
		}
	}
	
	return latest
}