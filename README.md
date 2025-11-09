# 🚀 Axiolot Hub Uptime Monitor

A production-ready uptime monitoring service built with Go, featuring intelligent retry logic, rate limiting, automated email fallbacks, and comprehensive health checking with SSL certificate tracking and multi-channel notifications.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Actions](https://img.shields.io/badge/CI-GitHub%20Actions-2088FF?style=flat&logo=github-actions)](https://github.com/features/actions)
[![GitHub Stars](https://img.shields.io/github/stars/arieduportal/uptime-monitor?style=social)](https://github.com/arieduportal/uptime-monitor)
[![Axiolot Hub logo](https://static.axiolot.com.ng/favicon.ico)](https://axiolot.com.ng)

## ✨ Features

### Core Monitoring
- ✅ **Concurrent Health Checks** - Monitor multiple domains simultaneously with configurable concurrency
- 🔒 **SSL Certificate Monitoring** - Automatic SSL expiry tracking with 30-day advance warnings
- 📊 **Detailed JSON Reports** - Comprehensive metrics including response times, status codes, and uptime percentages
- 🎯 **Exit Code Support** - Returns exit code 1 if services are down (perfect for CI/CD)

### Reliability & Performance
- 🔄 **Exponential Backoff Retries** - Automatic retry with intelligent backoff (1s → 2s → 4s → max 30s)
- ⚡ **Rate Limiting** - Built-in rate limiter (10 requests/second, configurable burst)
- 🛡️ **Smart Retry Logic** - Retries only on transient errors (timeouts, 5xx, 429)
- 📈 **Performance Metrics** - Response time categorization (fast, acceptable, degraded)

### Notifications & Integration
- 📧 **Email Fallback** - Automatically emails JSON reports when file writes fail (Gmail SMTP support)
- 🚨 **Smart Notifications** - Slack and Discord webhooks (only alerts on issues)
- 🔌 **API Integration** - Submit monitoring reports to your own API endpoint with retry support
- 📝 **Structured Logging** - Production-grade JSON logging with configurable levels

### Automation
- 🤖 **GitHub Actions Ready** - Pre-configured workflow for automated scheduling
- 🔧 **Flexible Configuration** - Environment-based configuration with sensible defaults

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [Email Configuration](#-email-configuration)
- [GitHub Actions Setup](#-github-actions-setup)
- [Output Format](#-output-format)
- [Notifications](#-notifications)
- [API Integration](#-api-integration)
- [Retry & Rate Limiting](#-retry--rate-limiting)
- [Development](#-development)
- [Troubleshooting](#-troubleshooting)

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/arieduportal/uptime-monitor.git
cd uptime-monitor

# Install dependencies
go mod download
go get golang.org/x/time/rate

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
   go get golang.org/x/time/rate
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

#### Basic Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `production` | Environment identifier (production, staging, etc.) |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `MONITOR_TIMEOUT` | `30s` | HTTP request timeout |
| `MONITOR_CONCURRENT` | `5` | Number of concurrent health checks |
| `OUTPUT_DIR` | `./reports` | Directory for saving JSON reports |
| `USER_AGENT` | `UptimeMonitor/2.0` | Custom User-Agent header |

#### API Integration
| Variable | Default | Description |
|----------|---------|-------------|
| `API_URL` | - | Endpoint to submit monitoring reports |
| `API_KEY` | - | Bearer token for API authentication |

#### Email Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `EMAIL_USER` | - | Gmail address for sending emails |
| `EMAIL_AUTH` | - | Gmail App Password (16-character) |
| `EMAIL_TO` | - | Comma-separated recipient email addresses |
| `SMTP_HOST` | `smtp.gmail.com` | SMTP server hostname |
| `SMTP_PORT` | `587` | SMTP server port (TLS) |

#### Notification Webhooks
| Variable | Default | Description |
|----------|---------|-------------|
| `SLACK_WEBHOOK_URL` | - | Slack webhook for notifications |
| `DISCORD_WEBHOOK_URL` | - | Discord webhook for notifications |

### Status Definitions

The monitor categorizes service health into three states:

| Status | Criteria | Description |
|--------|----------|-------------|
| **UP** 🟢 | HTTP 2xx, response time < 1000ms | Service is healthy and responsive |
| **DEGRADED** 🟡 | HTTP 2xx but slow (>1000ms), or 3xx/4xx codes | Service is working but has issues |
| **DOWN** 🔴 | HTTP 5xx, connection errors, timeouts | Service is not accessible |

### Retry Configuration

The monitor automatically retries failed requests with exponential backoff:

| Parameter | Default | Description |
|-----------|---------|-------------|
| **Max Retries** | 3 | Maximum number of retry attempts |
| **Initial Backoff** | 1s | Starting delay before first retry |
| **Max Backoff** | 30s | Maximum delay between retries |
| **Backoff Multiplier** | 2.0 | Exponential growth factor (1s → 2s → 4s) |

### Rate Limiting

Built-in rate limiting prevents overwhelming external services:

| Parameter | Default | Description |
|-----------|---------|-------------|
| **Requests Per Second** | 10 | Maximum request rate |
| **Burst Size** | 20 | Allowed burst of requests |

### Retryable Errors

The monitor only retries on specific transient errors:

- ✅ Network timeouts and connection errors
- ✅ HTTP 429 (Too Many Requests)
- ✅ HTTP 500, 502, 503, 504 (Server errors)
- ❌ HTTP 4xx client errors (except 429)
- ❌ SSL certificate errors
- ❌ DNS resolution failures

## 📧 Email Configuration

### Setting Up Gmail SMTP

The monitor can automatically email JSON reports when file writes fail.

#### Step 1: Enable 2-Factor Authentication
1. Go to your [Google Account](https://myaccount.google.com/)
2. Navigate to **Security**
3. Enable **2-Step Verification**

#### Step 2: Generate App Password
1. Go to [App Passwords](https://myaccount.google.com/apppasswords)
2. Select **Mail** and your device
3. Click **Generate**
4. Copy the 16-character password

#### Step 3: Configure Environment Variables
```bash
export EMAIL_USER="your-email@gmail.com"
export EMAIL_AUTH="abcd efgh ijkl mnop"  # 16-character App Password
export EMAIL_TO="recipient1@example.com,recipient2@example.com"
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
```

### Email Fallback Behavior

When the monitor cannot save reports to disk (e.g., permission issues, disk full), it automatically:

1. Formats the JSON report data
2. Sends it via email with clear delimiters
3. Logs the action for audit purposes
4. Continues normal operation

**Email Format:**
```
Subject: Uptime Monitor Report Failed

Failed to create JSON file for report

The report data is attached below:

=== BEGIN JSON DATA ===
{
  "service": "Uptime Monitor",
  "environment": "production",
  ...
}
=== END JSON DATA ===
```

### Using Other SMTP Providers

The monitor supports any SMTP provider. Example configurations:

**Outlook/Office 365:**
```bash
export SMTP_HOST="smtp.office365.com"
export SMTP_PORT="587"
```

**SendGrid:**
```bash
export SMTP_HOST="smtp.sendgrid.net"
export SMTP_PORT="587"
export EMAIL_USER="apikey"
export EMAIL_AUTH="your-sendgrid-api-key"
```

**Custom SMTP Server:**
```bash
export SMTP_HOST="mail.yourdomain.com"
export SMTP_PORT="587"
```

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

# With email fallback
export MONITOR_DOMAINS="example.com,api.example.com"
export EMAIL_USER="your-email@gmail.com"
export EMAIL_AUTH="your-app-password"
export EMAIL_TO="alerts@example.com"
go run .

# With all features
export MONITOR_DOMAINS="example.com,api.example.com"
export API_URL="https://api.yourservice.com/monitoring"
export API_KEY="your-secret-key"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export EMAIL_USER="your-email@gmail.com"
export EMAIL_AUTH="your-app-password"
export EMAIL_TO="alerts@example.com"
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

**Optional - Basic:**
- `ENVIRONMENT` - Environment name (default: production)
- `LOG_LEVEL` - Logging level (default: info)
- `MONITOR_TIMEOUT` - Request timeout (default: 30s)
- `MONITOR_CONCURRENT` - Concurrent checks (default: 5)

**Optional - API Integration:**
- `API_URL` - Your monitoring API endpoint
- `API_KEY` - API authentication token

**Optional - Email (NEW):**
- `EMAIL_USER` - Gmail address
- `EMAIL_AUTH` - Gmail App Password
- `EMAIL_TO` - Recipient email addresses
- `SMTP_HOST` - SMTP server (default: smtp.gmail.com)
- `SMTP_PORT` - SMTP port (default: 587)

**Optional - Webhooks:**
- `SLACK_WEBHOOK_URL` - Slack incoming webhook
- `DISCORD_WEBHOOK_URL` - Discord webhook

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
    - cron: '*/5 * * * *'
    
    # Every 15 minutes
    # - cron: '*/15 * * * *'
    
    # Every hour
    # - cron: '0 * * * *'
    
    # Daily at 9 AM
    # - cron: '0 9 * * *'
```

### 4. View Results

- **Actions Tab**: See all monitoring runs and their status
- **Artifacts**: Download JSON reports from each run (retained for 30 days)
- **Email**: Receive reports via email if disk writes fail
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
  "timestamp": "2025-11-09T10:30:00.000Z",
  "msg": "Health check completed",
  "domain": "example.com",
  "status": "up",
  "status_code": 200,
  "response_time_ms": 150
}
```

### Debug Logs (Retry Information)

```json
{
  "level": "warn",
  "timestamp": "2025-11-09T10:30:01.500Z",
  "msg": "Request failed, retrying",
  "domain": "api.example.com",
  "attempt": 1,
  "backoff": "1s",
  "error": "Request failed: context deadline exceeded"
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
  "timestamp": "2025-11-09T10:30:00Z",
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
      "timestamp": "2025-11-09T10:30:00Z",
      "checked_at": "2025-11-09T10:30:00Z"
    },
    {
      "domain": "api.example.com",
      "url": "https://api.example.com",
      "status": "down",
      "status_code": 0,
      "response_time_ms": 30000,
      "is_ssl": true,
      "error_message": "Request failed: context deadline exceeded",
      "timestamp": "2025-11-09T10:30:30Z",
      "checked_at": "2025-11-09T10:30:30Z"
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

### Email Notifications (NEW)

Configure email settings to receive JSON reports when file storage fails:

**Automatic Trigger:**
- Directory creation fails
- JSON marshaling fails
- File write permissions denied
- Disk space exhausted

**Manual Configuration:**
See [Email Configuration](#-email-configuration) section for setup details.

## 🔌 API Integration

Submit monitoring reports to your own API endpoint with automatic retry support.

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

### Retry Behavior

API submissions automatically retry on:
- Network errors (timeouts, connection failures)
- HTTP 429 (Too Many Requests)
- HTTP 500, 502, 503, 504 (Server errors)

Non-retryable errors (4xx except 429) are logged but don't trigger retries.

### Expected Response

- **2xx**: Success (report accepted)
- **4xx/5xx**: Error (logged, retried if applicable)

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

## 🔄 Retry & Rate Limiting

### Exponential Backoff

The monitor uses exponential backoff for retries:

```
Attempt 1: Wait 1 second
Attempt 2: Wait 2 seconds
Attempt 3: Wait 4 seconds
Max wait: 30 seconds
```

**Example Log Sequence:**
```json
{"level":"warn","msg":"Request failed, retrying","attempt":1,"backoff":"1s"}
{"level":"warn","msg":"Request failed, retrying","attempt":2,"backoff":"2s"}
{"level":"info","msg":"Request succeeded after retry","attempts":3}
```

### Rate Limiting

**Default Configuration:**
- **Rate**: 10 requests per second
- **Burst**: 20 requests allowed in burst
- **Scope**: Per monitor instance

**How It Works:**
1. Rate limiter allows 10 requests/second on average
2. Can handle bursts up to 20 requests
3. Additional requests wait until tokens available
4. Prevents overwhelming target services

**Debug Logging:**
```json
{"level":"debug","msg":"Rate limiter active","rate_limit":"10/s","burst":"20"}
```

### Smart Retry Logic

**Retryable Scenarios:**
```go
// Network errors
- Connection timeout
- Connection refused
- DNS resolution failure

// HTTP Status Codes
- 429 Too Many Requests
- 500 Internal Server Error
- 502 Bad Gateway
- 503 Service Unavailable
- 504 Gateway Timeout
```

**Non-Retryable Scenarios:**
```go
// Client errors (user's fault)
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found

// Fatal errors
- SSL certificate errors
- Invalid URL format
- JSON marshaling errors
```

### Monitoring Retry Performance

Enable debug logging to see retry behavior:

```bash
export LOG_LEVEL=debug
go run .
```

**Sample Output:**
```
[INFO] Starting uptime monitoring with rate limiting
[WARN] Request failed, retrying (attempt=1, backoff=1s)
[WARN] Request failed, retrying (attempt=2, backoff=2s)
[INFO] Request succeeded after retry (attempts=3)
[INFO] Health check completed
```

## 🛠️ Development

### Project Structure

```
uptime-monitor/
├── main.go                # Application entry point
├── monitor.go             # Core monitoring logic with retries
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

# Test retry logic
go test -run TestRetryLogic -v
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
5. Update environment variable documentation

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

**Problem: Email not sending**
```bash
# Verify Gmail App Password (not regular password)
# Check 2FA is enabled
# Test SMTP connection
echo "Test" | mail -s "Test" -S smtp=smtp.gmail.com:587 \
  -S smtp-use-starttls -S smtp-auth=login \
  -S smtp-auth-user=your-email@gmail.com \
  -S smtp-auth-password=your-app-password \
  recipient@example.com
```

**Problem: "Rate limiter error"**
```bash
# This usually indicates context cancellation
# Check timeout settings and reduce concurrent checks
export MONITOR_TIMEOUT=60s
export MONITOR_CONCURRENT=3
```

**Problem: Too many retries slowing down monitoring**
```bash
# Reduce max retries (requires code change)
# Or increase timeout to avoid initial failures
export MONITOR_TIMEOUT=45s
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
- Retry attempts and backoff durations
- Rate limiter activity
- Timing information
- Configuration values
- Webhook payloads
- Email sending details

### GitHub Actions Debugging

1. Check the Actions tab for error messages
2. Download artifacts to see JSON reports
3. Review workflow logs for environment issues
4. Verify secrets are correctly set (especially EMAIL_AUTH)
5. Check for rate limiting or API quota issues

## 📈 Performance Tips

1. **Adjust Concurrency**: Increase `MONITOR_CONCURRENT` for faster checks (default: 5)
   ```bash
   export MONITOR_CONCURRENT=10
   ```

2. **Optimize Timeout**: Set `MONITOR_TIMEOUT` based on your slowest service
   ```bash
   export MONITOR_TIMEOUT=45s
   ```

3. **Rate Limiting**: Adjust for your needs (requires code change in constants)
   ```go
   const RequestsPerSecond = 20  // Increase from 10
   ```

4. **Limit Domains**: Don't monitor too many domains in one instance
   - Split into multiple configurations if monitoring 50+ domains

5. **Use Multiple Instances**: Split domains across multiple GitHub Actions workflows
   ```yaml
   # workflow-1.yml - monitors API services
   # workflow-2.yml - monitors web services
   ```

6. **Disable Retries**: For very fast checks, reduce max retries (code change)
   ```go
   const MaxRetries = 1  // Reduce from 3
   ```

## 🔒 Security Best Practices

1. **Never commit secrets**: Use environment variables or GitHub Secrets
2. **Use App Passwords**: Never use your actual Gmail password
3. **Rotate API Keys**: Regularly update API_KEY values
4. **Limit Email Recipients**: Only send to trusted addresses
5. **Review Logs**: Regularly check logs for suspicious activity
6. **HTTPS Only**: Monitor uses HTTPS by default, don't override
7. **Update Dependencies**: Keep Go and packages up to date
   ```bash
   go get -u ./...
   go mod tidy
   ```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Areas for Contribution:**
- Additional notification channels (Teams, PagerDuty)
- Circuit breaker pattern implementation
- Metrics dashboard integration
- Database storage for historical data
- Advanced alerting rules
- Custom retry strategies

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org)
- Logging by [Zap](https://github.com/uber-go/zap)
- Rate limiting by [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)
- Automated with [GitHub Actions](https://github.com/features/actions)

## 📞 Support

- 📧 Open an issue on GitHub
- 💬 Start a discussion
- ⭐ Star the repository if you find it useful!
- 🐛 Report bugs with detailed logs

## 🎯 Roadmap

- [ ] Circuit breaker pattern for cascading failures
- [ ] Prometheus metrics export
- [ ] Database storage for historical trends
- [ ] Web dashboard UI
- [ ] Custom alert rules engine
- [ ] Multi-region health checks
- [ ] Performance benchmarking tools

---

**Written with ❤️ for reliable monitoring of Axiolot Hub Systems**
