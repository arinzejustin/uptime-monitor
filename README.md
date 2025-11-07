# 🚀 Axiolot Hub Uptime Monitor

A production-ready uptime monitoring service built with Go, designed for automated monitoring with comprehensive health checking, SSL certificate tracking, and multi-channel notifications.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Actions](https://img.shields.io/badge/CI-GitHub%20Actions-2088FF?style=flat&logo=github-actions)](https://github.com/features/actions)
[![GitHub Stars](https://img.shields.io/github/stars/arieduportal/uptime-monitor?style=social)](https://github.com/arieduportal/uptime-monitor)
[![Axiolot Hub logo](https://static.axiolot.com.ng/favicon.ico)](https://axiolot.com.ng)


## ✨ Features

- ✅ **Concurrent Health Checks** - Monitor multiple domains simultaneously with configurable concurrency
- 🔒 **SSL Certificate Monitoring** - Automatic SSL expiry tracking with 30-day advance warnings
- 📊 **Detailed JSON Reports** - Comprehensive metrics including response times, status codes, and uptime percentages
- 🚨 **Smart Notifications** - Slack and Discord webhooks (only alerts on issues)
- 🔄 **API Integration** - Submit monitoring reports to your own API endpoint
- 📈 **Performance Metrics** - Response time categorization (fast, acceptable, degraded)
- 📝 **Structured Logging** - Production-grade JSON logging with configurable levels
- 🎯 **Exit Code Support** - Returns exit code 1 if services are down (perfect for CI/CD)
- 🤖 **GitHub Actions Ready** - Pre-configured workflow for automated scheduling

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [GitHub Actions Setup](#-github-actions-setup)
- [Output Format](#-output-format)
- [Notifications](#-notifications)
- [API Integration](#-api-integration)
- [Development](#-development)
- [Troubleshooting](#-troubleshooting)

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/arieduportal/uptime-monitor.git
cd uptime-monitor

# Install dependencies
go mod download

# Set required environment variable
export MONITOR_DOMAINS="example.com,google.com,github.com"

# Run the monitor
go run .
```

## 📦 Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Local Setup

1. **Clone and navigate to the project:**
   ```bash
   git clone https://github.com/arieduportal/uptime-monitor.git
   cd uptime-monitor
   ```

2. **Install Go dependencies:**
   ```bash
   go mod download
   ```

3. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

4. **Test the monitor:**
   ```bash
   # Load environment variables
   export $(cat .env | xargs)
   
   # Run monitor
   go run .
   ```

### Building Binary

```bash
# Build for your platform
go build -o uptime-monitor .

# Run the binary
./uptime-monitor
```

## ⚙️ Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `MONITOR_DOMAINS` | Comma-separated list of domains to monitor | `example.com,api.example.com` |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_URL` | - | Endpoint to submit monitoring reports |
| `API_KEY` | - | Bearer token for API authentication |
| `SLACK_WEBHOOK_URL` | - | Slack webhook for notifications |
| `DISCORD_WEBHOOK_URL` | - | Discord webhook for notifications |
| `ENVIRONMENT` | `production` | Environment identifier (production, staging, etc.) |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `MONITOR_TIMEOUT` | `30s` | HTTP request timeout |
| `MONITOR_CONCURRENT` | `5` | Number of concurrent health checks |
| `OUTPUT_DIR` | `./reports` | Directory for saving JSON reports |
| `USER_AGENT` | `UptimeMonitor/2.0` | Custom User-Agent header |

### Status Definitions

The monitor categorizes service health into three states:

| Status | Criteria | Description |
|--------|----------|-------------|
| **UP** 🟢 | HTTP 2xx, response time < 1000ms | Service is healthy and responsive |
| **DEGRADED** 🟡 | HTTP 2xx but slow (>1000ms), or 3xx/4xx codes | Service is working but has issues |
| **DOWN** 🔴 | HTTP 5xx, connection errors, timeouts | Service is not accessible |

### SSL Certificate Warnings

- Certificates expiring within **30 days** trigger warnings
- SSL information is automatically collected for HTTPS domains
- Includes expiry date and days remaining in reports

## 📖 Usage

### Local Development

```bash
# Basic usage with minimal config
export MONITOR_DOMAINS="example.com"
go run .

# With debug logging
export LOG_LEVEL=debug
export MONITOR_DOMAINS="example.com,github.com"
go run .

# With all features
export MONITOR_DOMAINS="example.com,api.example.com"
export API_URL="https://api.yourservice.com/monitoring"
export API_KEY="your-secret-key"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export ENVIRONMENT="staging"
go run .
```

### Using the Binary

```bash
# Build once
go build -o uptime-monitor .

# Run anywhere
MONITOR_DOMAINS="example.com" ./uptime-monitor
```

### Exit Codes

| Exit Code | Meaning | Use Case |
|-----------|---------|----------|
| `0` | All services up or degraded | Success in CI/CD |
| `1` | One or more services down | Fail CI/CD pipeline |

## 🤖 GitHub Actions Setup

### 1. Add Repository Secrets

Navigate to **Settings → Secrets and variables → Actions** and add:

**Required:**
- `MONITOR_DOMAINS` - Domains to monitor (e.g., `example.com,api.example.com`)

**Optional:**
- `API_URL` - Your monitoring API endpoint
- `API_KEY` - API authentication token
- `SLACK_WEBHOOK_URL` - Slack incoming webhook
- `DISCORD_WEBHOOK_URL` - Discord webhook
- `ENVIRONMENT` - Environment name (default: production)
- `LOG_LEVEL` - Logging level (default: info)
- `MONITOR_TIMEOUT` - Request timeout (default: 30s)
- `MONITOR_CONCURRENT` - Concurrent checks (default: 5)

### 2. Workflow Configuration

The included workflow (`.github/workflows/monitor.yml`) runs automatically:

- **Schedule**: Every 5 minutes
- **Manual**: Via Actions tab or GitHub CLI
- **Push**: On changes to Go files or workflow

### 3. Customize Schedule

Edit `.github/workflows/monitor.yml`:

```yaml
on:
  schedule:
    # Every 5 minutes (default)
    #- cron: '*/5 * * * *'
    
    # Every 15 minutes
    # - cron: '*/15 * * * *'
    
    # Every hour
    - cron: '0 * * * *'
    
    # Daily at 9 AM
    # - cron: '0 9 * * *'
```

### 4. View Results

- **Actions Tab**: See all monitoring runs and their status
- **Artifacts**: Download JSON reports from each run (retained for 30 days)
- **Repository**: Reports are committed automatically (if enabled)

### 5. Manual Trigger

```bash
# Via GitHub CLI
gh workflow run monitor.yml

# Or use the Actions tab in your repository
```

## 📊 Output Format

### Console Logs (JSON)

```json
{
  "level": "info",
  "timestamp": "2025-11-06T10:30:00.000Z",
  "msg": "Health check completed",
  "domain": "example.com",
  "status": "up",
  "status_code": 200,
  "response_time_ms": 150
}
```

### JSON Report File

Reports are saved to `{OUTPUT_DIR}/uptime_report_{timestamp}.json`:

```json
{
  "service": "Uptime Monitor",
  "environment": "production",
  "total_checks": 3,
  "uptime_count": 2,
  "downtime_count": 1,
  "degraded_count": 0,
  "uptime_percent": 66.67,
  "average_latency_ms": 250.5,
  "timestamp": "2025-11-06T10:30:00Z",
  "results": [
    {
      "domain": "example.com",
      "url": "https://example.com",
      "status": "up",
      "status_code": 200,
      "response_time_ms": 150,
      "is_ssl": true,
      "ssl_expiry": "2025-12-31T23:59:59Z",
      "ssl_days_left": 55,
      "content_length": 1024,
      "timestamp": "2025-11-06T10:30:00Z",
      "checked_at": "2025-11-06T10:30:00Z"
    },
    {
      "domain": "api.example.com",
      "url": "https://api.example.com",
      "status": "down",
      "status_code": 0,
      "response_time_ms": 30000,
      "is_ssl": true,
      "error_message": "Request failed: context deadline exceeded",
      "timestamp": "2025-11-06T10:30:30Z",
      "checked_at": "2025-11-06T10:30:30Z"
    }
  ]
}
```

## 🔔 Notifications

Notifications are sent **only when services are down or degraded** (no spam!).

### Slack Integration

Configure `SLACK_WEBHOOK_URL` to receive rich notifications:

**Create Webhook:**
1. Go to https://api.slack.com/messaging/webhooks
2. Create an incoming webhook for your workspace
3. Copy the webhook URL
4. Add to environment variables or GitHub secrets

**Example Alert:**
```
🚨 Uptime Alert - 1 service(s) down, 0 degraded

Environment: production
Uptime: 66.67%
Down: 1
Degraded: 0

Failed Services:
api.example.com (down)
```

### Discord Integration

Configure `DISCORD_WEBHOOK_URL` for Discord notifications:

**Create Webhook:**
1. Open Server Settings → Integrations → Webhooks
2. Click "New Webhook"
3. Copy the webhook URL
4. Add to environment variables or GitHub secrets

**Example Alert:**
```
🚨 **Uptime Alert**

**Environment:** production
**Uptime:** 66.67%
**Down:** 1 | **Degraded:** 0

**Failed Services:**
🔴 **api.example.com** - down
```

## 🔌 API Integration

Submit monitoring reports to your own API endpoint.

### Configuration

```bash
export API_URL="https://api.yourservice.com/monitoring/reports"
export API_KEY="your-api-key"
```

### Request Format

```http
POST /monitoring/reports HTTP/1.1
Host: api.yourservice.com
Content-Type: application/json
Authorization: Bearer your-api-key

{
  "service": "Uptime Monitor",
  "environment": "production",
  "total_checks": 3,
  "uptime_count": 2,
  ...
}
```

### Expected Response

- **2xx**: Success (report accepted)
- **4xx/5xx**: Error (logged but doesn't fail the run)

### Example API Handler (Go)

```go
func handleMonitoringReport(w http.ResponseWriter, r *http.Request) {
    var report MonitorReport
    if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process report
    log.Printf("Received report: uptime=%.2f%%", report.UptimePercent)
    
    // Store in database, trigger alerts, etc.
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

## 🛠️ Development

### Project Structure

```
uptime-monitor/
├── main.go                # Application entry point
├── monitor.go             # Core monitoring logic
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── .env.example          # Environment template
├── .github/
│   └── workflows/
│       └── monitor.yml    # GitHub Actions workflow
├── reports/              # Generated JSON reports
└── README.md             # Documentation
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestCheckDomain
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

### Adding New Features

1. Add feature to `monitor.go`
2. Update tests
3. Update README documentation
4. Test locally before committing

## 🐛 Troubleshooting

### Common Issues

**Problem: "MONITOR_DOMAINS environment variable not set"**
```bash
# Solution: Set the required environment variable
export MONITOR_DOMAINS="example.com"
```

**Problem: "Failed to submit to API"**
```bash
# Check API URL and key
echo $API_URL
echo $API_KEY

# Test API endpoint manually
curl -X POST $API_URL \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

**Problem: Timeout errors**
```bash
# Increase timeout
export MONITOR_TIMEOUT=60s

# Or reduce concurrent checks
export MONITOR_CONCURRENT=3
```

**Problem: SSL certificate warnings**
```bash
# Check certificate expiry
echo | openssl s_client -connect example.com:443 2>/dev/null | \
  openssl x509 -noout -dates
```

### Debug Mode

Enable detailed logging:

```bash
export LOG_LEVEL=debug
go run .
```

This will show:
- Request/response details
- Timing information
- Configuration values
- Webhook payloads

### GitHub Actions Debugging

1. Check the Actions tab for error messages
2. Download artifacts to see JSON reports
3. Review workflow logs for environment issues
4. Verify secrets are correctly set

## 📈 Performance Tips

1. **Adjust Concurrency**: Increase `MONITOR_CONCURRENT` for faster checks (use with caution)
2. **Optimize Timeout**: Set `MONITOR_TIMEOUT` based on your slowest service
3. **Limit Domains**: Don't monitor too many domains in one instance
4. **Use Multiple Instances**: Split domains across multiple GitHub Actions workflows

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org)
- Logging by [Zap](https://github.com/uber-go/zap)
- Automated with [GitHub Actions](https://github.com/features/actions)

## 📞 Support

- 📧 Open an issue on GitHub
- 💬 Start a discussion
- ⭐ Star the repository if you find it useful!

---

**Made with ❤️ for reliable monitoring**