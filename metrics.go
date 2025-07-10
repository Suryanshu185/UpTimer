package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	uptimeGauge      *prometheus.GaugeVec
	responseTimeGauge *prometheus.GaugeVec
	checksTotal      *prometheus.CounterVec
	checksErrors     *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		uptimeGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "uptime_website_up",
			Help: "Whether the website is up (1) or down (0)",
		}, []string{"website_name", "website_url"}),
		
		responseTimeGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "uptime_response_time_seconds",
			Help: "Response time of the website in seconds",
		}, []string{"website_name", "website_url"}),
		
		checksTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "uptime_checks_total",
			Help: "Total number of uptime checks performed",
		}, []string{"website_name", "website_url"}),
		
		checksErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "uptime_checks_errors_total",
			Help: "Total number of failed uptime checks",
		}, []string{"website_name", "website_url"}),
	}
}

func (m *Metrics) UpdateMetrics(result MonitorResult) {
	labels := prometheus.Labels{
		"website_name": result.Website.Name,
		"website_url":  result.Website.URL,
	}
	
	// Update uptime gauge
	if result.Success {
		m.uptimeGauge.With(labels).Set(1)
	} else {
		m.uptimeGauge.With(labels).Set(0)
		m.checksErrors.With(labels).Inc()
	}
	
	// Update response time
	m.responseTimeGauge.With(labels).Set(result.ResponseTime.Seconds())
	
	// Increment total checks
	m.checksTotal.With(labels).Inc()
}