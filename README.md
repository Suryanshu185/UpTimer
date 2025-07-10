# UpTimer
Uptime Monitoring Tool (Self-hosted Ping Service)

## Features

✅ **Core Monitoring**
- Pings websites/APIs every configurable interval (default: 30 seconds)
- Supports GET, POST, PUT, DELETE HTTP methods
- Configurable expected status codes
- Concurrent monitoring using goroutines
- Logs results (up/down) with timestamps

✅ **Web Dashboard**
- Clean, responsive web UI showing uptime status
- Real-time updates every 30 seconds
- Shows response times, status codes, and error messages
- Overview statistics (total sites, up/down count, uptime percentage)

✅ **Alerting System**
- Email alerts via SMTP
- Slack webhook integration
- Configurable alert recipients
- Alerts sent when sites go down

✅ **Metrics & Monitoring**
- Prometheus metrics endpoint (`/metrics`)
- Uptime gauges, response time metrics
- Error counters for monitoring
- Health check endpoint (`/api/health`)

✅ **Containerized Deployment**
- Docker support with multi-stage builds
- Docker Compose with Prometheus & Grafana
- Health checks and proper logging
- Non-root user for security

## Quick Start

### 1. Local Development

```bash
# Install dependencies
go mod tidy

# Build and run
go build -o uptime-monitor .
./uptime-monitor
```

The web dashboard will be available at `http://localhost:8080`

### 2. Docker Deployment

```bash
# Build Docker image
docker build -t uptime-monitor .

# Run with Docker Compose (includes Prometheus & Grafana)
docker-compose up -d
```

**Services:**
- UpTimer Dashboard: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

## Configuration

Edit `config.yaml` to customize:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

monitoring:
  interval: 30s    # Check interval
  timeout: 10s     # Request timeout
  
websites:
  - name: "My Website"
    url: "https://example.com"
    method: "GET"
    expected_status: 200

alerts:
  email:
    enabled: true
    smtp_server: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    from: "your-email@gmail.com"
    to: ["admin@example.com"]
  
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#uptime-alerts"
```

## API Endpoints

- `GET /` - Web dashboard
- `GET /api/status` - Current status of all websites
- `GET /api/health` - Health check endpoint
- `GET /api/websites` - List configured websites
- `GET /api/website/{name}` - Historical data for specific website
- `GET /metrics` - Prometheus metrics

## Tech Stack

- **Backend**: Go 1.21+ with goroutines for concurrent monitoring
- **HTTP Framework**: Gorilla Mux for routing
- **Metrics**: Prometheus client for metrics collection
- **Configuration**: YAML-based configuration
- **Containerization**: Docker with multi-stage builds
- **Monitoring Stack**: Prometheus + Grafana for visualization

## Prometheus Metrics

The application exposes the following metrics:

- `uptime_website_up{website_name, website_url}` - Whether site is up (1) or down (0)
- `uptime_response_time_seconds{website_name, website_url}` - Response time in seconds
- `uptime_checks_total{website_name, website_url}` - Total number of checks performed
- `uptime_checks_errors_total{website_name, website_url}` - Total number of failed checks

## Deployment Options

### VPS Deployment

1. **Systemd Service** (recommended for production)
2. **Docker Compose** (easiest with monitoring stack)
3. **Kubernetes** (for scalable deployments)

### Environment Variables

You can override configuration using environment variables:

- `UPTIME_CONFIG_FILE` - Path to config file (default: config.yaml)
- `UPTIME_PORT` - Server port (default: 8080)
- `UPTIME_HOST` - Server host (default: 0.0.0.0)

## Screenshots

![UpTimer Dashboard](https://github.com/user-attachments/assets/b1c838ed-41ca-4f1e-9041-195db13de5c4)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details
