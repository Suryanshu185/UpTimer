package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type WebServer struct {
	config  *Config
	monitor *Monitor
}

func NewWebServer(config *Config, monitor *Monitor) *WebServer {
	return &WebServer{
		config:  config,
		monitor: monitor,
	}
}

func (ws *WebServer) Start() error {
	r := mux.NewRouter()
	
	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	
	// Main dashboard
	r.HandleFunc("/", ws.handleDashboard).Methods("GET")
	
	// API endpoints
	r.HandleFunc("/api/status", ws.handleStatus).Methods("GET")
	r.HandleFunc("/api/health", ws.handleHealth).Methods("GET")
	r.HandleFunc("/api/websites", ws.handleWebsites).Methods("GET")
	r.HandleFunc("/api/website/{name}", ws.handleWebsite).Methods("GET")
	
	// Metrics endpoint
	if ws.config.Metrics.Enabled {
		r.Handle(ws.config.Metrics.Path, promhttp.Handler())
	}
	
	// CORS middleware
	r.Use(corsMiddleware)
	
	// Logging middleware
	r.Use(loggingMiddleware)
	
	addr := fmt.Sprintf("%s:%d", ws.config.Server.Host, ws.config.Server.Port)
	log.Printf("Starting web server on %s", addr)
	
	return http.ListenAndServe(addr, r)
}

func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "index.html"))
}

func (ws *WebServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	results := ws.monitor.GetLatestResults()
	
	response := make(map[string]interface{})
	for name, result := range results {
		response[name] = map[string]interface{}{
			"Website": map[string]interface{}{
				"Name":           result.Website.Name,
				"URL":            result.Website.URL,
				"Method":         result.Website.Method,
				"ExpectedStatus": result.Website.ExpectedStatus,
			},
			"Success":      result.Success,
			"ResponseTime": result.ResponseTime.Nanoseconds(),
			"StatusCode":   result.StatusCode,
			"Error":        result.Error,
			"Timestamp":    result.Timestamp,
		}
	}
	
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	results := ws.monitor.GetLatestResults()
	
	allUp := true
	upCount := 0
	totalCount := len(results)
	
	for _, result := range results {
		if result.Success {
			upCount++
		} else {
			allUp = false
		}
	}
	
	status := "healthy"
	if !allUp {
		status = "degraded"
	}
	if upCount == 0 && totalCount > 0 {
		status = "unhealthy"
	}
	
	response := map[string]interface{}{
		"status":      status,
		"upCount":     upCount,
		"totalCount":  totalCount,
		"uptime":      float64(upCount) / float64(totalCount) * 100,
		"timestamp":   time.Now(),
	}
	
	if allUp {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleWebsites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	websites := make([]map[string]interface{}, 0)
	for _, website := range ws.config.Websites {
		websites = append(websites, map[string]interface{}{
			"name":           website.Name,
			"url":            website.URL,
			"method":         website.Method,
			"expectedStatus": website.ExpectedStatus,
		})
	}
	
	json.NewEncoder(w).Encode(websites)
}

func (ws *WebServer) handleWebsite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	w.Header().Set("Content-Type", "application/json")
	
	allResults := ws.monitor.GetResults()
	results, exists := allResults[name]
	
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Website not found"})
		return
	}
	
	response := make([]map[string]interface{}, len(results))
	for i, result := range results {
		response[i] = map[string]interface{}{
			"success":      result.Success,
			"responseTime": result.ResponseTime.Nanoseconds(),
			"statusCode":   result.StatusCode,
			"error":        result.Error,
			"timestamp":    result.Timestamp,
		}
	}
	
	json.NewEncoder(w).Encode(response)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		next.ServeHTTP(w, r)
		
		log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
	})
}