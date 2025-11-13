package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

const (
	StatusUp       = "up"
	StatusDown     = "down"
	StatusDegraded = "degraded"

	ThresholdFast    = 1000
	ThresholdAccept  = 3000
	SSLExpiryWarning = 30

	DefaultTimeout    = 30 * time.Second
	DefaultUserAgent  = "Monitoring Client/1.0"
	DefaultConcurrent = 5

	MaxRetries        = 3
	InitialBackoff    = 1 * time.Second
	MaxBackoff        = 30 * time.Second
	BackoffMultiplier = 2.0

	RequestsPerSecond = 10
	BurstSize         = 20
	DefaultSMTPHost   = "smtp.gmail.com"
)

type HealthCheckResult struct {
	Domain        string    `json:"domain"`
	URL           string    `json:"url"`
	Status        string    `json:"status"`
	StatusCode    int       `json:"status_code"`
	ResponseTime  int64     `json:"response_time_ms"`
	IsSSL         bool      `json:"is_ssl"`
	SSLExpiry     string    `json:"ssl_expiry,omitempty"`
	SSLDaysLeft   int       `json:"ssl_days_left,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	ContentLength int64     `json:"content_length"`
	Timestamp     time.Time `json:"timestamp"`
	CheckedAt     string    `json:"checked_at"`
}

type MonitorReport struct {
	Service        string              `json:"service"`
	Environment    string              `json:"environment,omitempty"`
	TotalChecks    int                 `json:"total_checks"`
	Uptime         int                 `json:"uptime_count"`
	Downtime       int                 `json:"downtime_count"`
	Degraded       int                 `json:"degraded_count"`
	UptimePercent  float64             `json:"uptime_percent"`
	AverageLatency float64             `json:"average_latency_ms"`
	Timestamp      time.Time           `json:"timestamp"`
	Results        []HealthCheckResult `json:"results"`
}

type MonitorConfig struct {
	Domains        []string
	APIURL         string
	APIKey         string
	Timeout        time.Duration
	UserAgent      string // Monitor User-Agent
	Concurrent     int
	Environment    string
	OutputDir      string
	SlackWebhook   string
	DiscordWebhook string
	EmailAuth      string
	EmailTo        []string
	EmailUser      string
	SMTPHost       string // smtp.gmail.com
	SMTPPort       string // 587
	MaxRetries     int
	RateLimiter    *rate.Limiter
}

type UptimeMonitor struct {
	config *MonitorConfig
	logger *zap.Logger
	client *http.Client
}

type RetryConfig struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns a default RetryConfig
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        MaxRetries,
		InitialBackoff:    InitialBackoff,
		MaxBackoff:        MaxBackoff,
		BackoffMultiplier: BackoffMultiplier,
	}
}

// CalculateBackoff calculates exponential backoff duration
func (rc RetryConfig) CalculateBackoff(attempt int) time.Duration {
	backoff := float64(rc.InitialBackoff) * math.Pow(rc.BackoffMultiplier, float64(attempt))
	if backoff > float64(rc.MaxBackoff) {
		return rc.MaxBackoff
	}
	return time.Duration(backoff)
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(err error, statusCode int) bool {
	if err != nil {
		if strings.Contains(err.Error(), "marshal") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "context cancelled") {
			return false
		}
		return true
	}

	// Retry on specific status codes
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

func NewMonitorConfig() (*MonitorConfig, error) {
	domainsStr := os.Getenv("MONITOR_DOMAINS")
	if domainsStr == "" {
		return nil, fmt.Errorf("MONITOR_DOMAINS environment variable not set")
	}

	domains := strings.Split(domainsStr, ",")
	for i, domain := range domains {
		domains[i] = strings.TrimSpace(domain)
	}

	emailTo := strings.Split(os.Getenv("EMAIL_TO"), ",")
	for i, email := range emailTo {
		emailTo[i] = strings.TrimSpace(email)
	}

	timeout := DefaultTimeout
	if timeoutStr := os.Getenv("MONITOR_TIMEOUT"); timeoutStr != "" {
		if d, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = d
		}
	}

	concurrent := DefaultConcurrent
	if concurrentStr := os.Getenv("MONITOR_CONCURRENT"); concurrentStr != "" {
		fmt.Sscanf(concurrentStr, "%d", &concurrent)
	}

	rateLimiter := rate.NewLimiter(rate.Limit(RequestsPerSecond), BurstSize)

	return &MonitorConfig{
		Domains:        domains,
		APIURL:         getEnvOrDefault("API_URL", ""),
		APIKey:         os.Getenv("API_KEY"),
		Timeout:        timeout,
		UserAgent:      getEnvOrDefault("USER_AGENT", DefaultUserAgent),
		Concurrent:     concurrent,
		Environment:    getEnvOrDefault("ENVIRONMENT", "production"),
		OutputDir:      getEnvOrDefault("OUTPUT_DIR", "./reports"),
		SlackWebhook:   os.Getenv("SLACK_WEBHOOK_URL"),
		DiscordWebhook: os.Getenv("DISCORD_WEBHOOK_URL"),
		EmailAuth:      os.Getenv("EMAIL_AUTH"),
		EmailTo:        emailTo,
		EmailUser:      os.Getenv("EMAIL_USER"),
		SMTPHost:       getEnvOrDefault("SMTP_HOST", DefaultSMTPHost),
		SMTPPort:       os.Getenv("SMTP_PORT"),
		MaxRetries:     MaxRetries,
		RateLimiter:    rateLimiter,
	}, nil
}

func NewUptimeMonitor(config *MonitorConfig, logger *zap.Logger) *UptimeMonitor {
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &UptimeMonitor{
		config: config,
		logger: logger,
		client: client,
	}
}

func (m *UptimeMonitor) CheckDomain(ctx context.Context, domain string) HealthCheckResult {
	retryConfig := DefaultRetryConfig()

	var lastResult HealthCheckResult

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {

		if err := m.config.RateLimiter.Wait(ctx); err != nil {

			return HealthCheckResult{
				Domain:       domain,
				URL:          domain,
				Status:       StatusDown,
				ErrorMessage: fmt.Sprintf("Rate limiter error: %v", err),
				Timestamp:    time.Now(),
				CheckedAt:    time.Now().UTC().Format(time.RFC3339),
			}
		}

		result := HealthCheckResult{
			Domain:    domain,
			URL:       domain,
			Timestamp: time.Now(),
			CheckedAt: time.Now().UTC().Format(time.RFC3339),
		}

		checkURL := domain
		if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
			checkURL = "https://" + domain
			result.URL = checkURL
		}

		result.IsSSL = strings.HasPrefix(checkURL, "https://")

		req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
		if err != nil {
			result.Status = StatusDown
			result.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
			lastResult = result

			if !IsRetryableError(err, 0) || attempt == retryConfig.MaxRetries {
				m.logger.Error("Request creation failed",
					zap.String("domain", result.Domain),
					zap.Error(err))
				return result
			}

			backoff := retryConfig.CalculateBackoff(attempt)

			select {
			case <-ctx.Done():
				result.ErrorMessage = "Context cancelled during retry"
				return result
			case <-time.After(backoff):
				continue
			}
		}

		req.Header.Set("User-Agent", m.config.UserAgent)

		startTime := time.Now()
		resp, err := m.client.Do(req)
		duration := time.Since(startTime)
		result.ResponseTime = duration.Milliseconds()

		if err != nil {
			result.Status = StatusDown
			result.ErrorMessage = fmt.Sprintf("Request failed: %v", err)
			lastResult = result

			if !IsRetryableError(err, 0) {
				return result
			}

			if attempt == retryConfig.MaxRetries {
				m.logger.Warn("Max retries reached",
					zap.String("domain", domain),
					zap.Int("attempts", attempt+1))
				return result
			}

			backoff := retryConfig.CalculateBackoff(attempt)

			select {
			case <-ctx.Done():
				result.ErrorMessage = "Context cancelled during retry"
				return result
			case <-time.After(backoff):
				continue
			}
		}
		defer resp.Body.Close()

		io.Copy(io.Discard, resp.Body)

		result.StatusCode = resp.StatusCode
		result.ContentLength = resp.ContentLength

		if result.IsSSL && resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
			cert := resp.TLS.PeerCertificates[0]
			result.SSLExpiry = cert.NotAfter.UTC().Format(time.RFC3339)
			daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
			result.SSLDaysLeft = daysLeft

			if daysLeft < SSLExpiryWarning {
				m.logger.Warn("SSL certificate expiring soon",
					zap.String("domain", result.Domain),
					zap.Int("days_left", daysLeft))
			}
		}

		result.Status = m.determineStatus(resp.StatusCode, result.ResponseTime)
		lastResult = result

		if result.Status == StatusUp {
			return result
		}

		if !IsRetryableError(nil, result.StatusCode) {
			return result
		}

		if attempt == retryConfig.MaxRetries {
			break
		}

		backoff := retryConfig.CalculateBackoff(attempt)

		select {
		case <-ctx.Done():
			result.ErrorMessage = "Context cancelled during retry"
			return result
		case <-time.After(backoff):
			// Continue to next attempt dont wait
		}
	}

	return lastResult
}

// determineStatus determines the status of a domain based on the response code and response time
func (m *UptimeMonitor) determineStatus(statusCode int, responseTime int64) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		if responseTime >= ThresholdAccept {
			return StatusDegraded
		}
		return StatusUp
	case statusCode >= 300 && statusCode < 400:
		return StatusUp
	case statusCode >= 400 && statusCode < 500:
		return StatusDegraded
	default:
		return StatusDown
	}
}

// RunCheck runs a health check on all domains in the configuration
func (m *UptimeMonitor) RunCheck(ctx context.Context) (*MonitorReport, error) {

	results := make([]HealthCheckResult, len(m.config.Domains))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, m.config.Concurrent)

	for i, domain := range m.config.Domains {
		wg.Add(1)
		go func(index int, d string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[index] = m.CheckDomain(ctx, d)
		}(i, domain)
	}

	wg.Wait()

	report := m.generateReport(results)

	return report, nil
}

func (m *UptimeMonitor) generateReport(results []HealthCheckResult) *MonitorReport {
	var totalLatency int64
	var upCount, downCount, degradedCount int

	for _, result := range results {
		totalLatency += result.ResponseTime

		switch result.Status {
		case StatusUp:
			upCount++
		case StatusDown:
			downCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	avgLatency := float64(0)
	if len(results) > 0 {
		avgLatency = float64(totalLatency) / float64(len(results))
	}

	uptimePercent := float64(0)
	if len(results) > 0 {
		uptimePercent = float64(upCount) / float64(len(results)) * 100
	}

	return &MonitorReport{
		Service:        "Uptime Monitor",
		Environment:    m.config.Environment,
		TotalChecks:    len(results),
		Uptime:         upCount,
		Downtime:       downCount,
		Degraded:       degradedCount,
		UptimePercent:  uptimePercent,
		AverageLatency: avgLatency,
		Timestamp:      time.Now().UTC(),
		Results:        results,
	}
}

// SaveReport saves the report to a file and sends an email if the directory creation fails.
func (m *UptimeMonitor) SaveReport(report *MonitorReport) (string, error) {
	if err := os.MkdirAll(m.config.OutputDir, 0755); err != nil {
		m.logger.Error("Failed to create output directory, sending via email", zap.Error(err))
		if emailErr := m.SendEmailOnFailure(report, nil); emailErr != nil {
			m.logger.Error("Failed to send email", zap.Error(emailErr))
		}
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s/uptime_report_%s.json", m.config.OutputDir, timestamp)

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		m.logger.Error("Failed to marshal JSON, sending via email", zap.Error(err))
		if emailErr := m.SendEmailOnFailure(report, nil); emailErr != nil {
			m.logger.Error("Failed to send email", zap.Error(emailErr))
		}
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		m.logger.Error("Failed to write file, sending via email", zap.Error(err))
		if emailErr := m.SendEmailOnFailure(report, nil); emailErr != nil {
			m.logger.Error("Failed to send email", zap.Error(emailErr))
		}
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	m.logger.Info("Report saved", zap.String("file", filename))
	return filename, nil
}

// BuildEmailMessage builds a multipart email message with both plain text and HTML parts.
func BuildEmailMessage(from string, to []string, subject string, htmlBody string, plainBody string) []byte {
	boundary := "boundary_" + fmt.Sprint(time.Now().UnixNano())

	var msg []byte
	msg = fmt.Appendf(msg, "From: Uptime Monitor <%s>\r\n", from)
	msg = fmt.Appendf(msg, "To: %s\r\n", strings.Join(to, ","))
	msg = fmt.Appendf(msg, "Subject: %s\r\n", subject)
	msg = fmt.Appendf(msg, "MIME-Version: 1.0\r\n")
	msg = fmt.Appendf(msg, "Content-Type: multipart/alternative; boundary=%s\r\n", boundary)
	msg = fmt.Appendf(msg, "\r\n")

	// Plain text section (for clients that don't support HTML)
	msg = fmt.Appendf(msg, "--%s\r\n", boundary)
	msg = fmt.Appendf(msg, "Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg = fmt.Appendf(msg, "%s\r\n", plainBody)

	// HTML section
	msg = fmt.Appendf(msg, "\r\n--%s\r\n", boundary)
	msg = fmt.Appendf(msg, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	msg = fmt.Appendf(msg, "%s\r\n", htmlBody)

	// Closing boundary
	msg = fmt.Appendf(msg, "\r\n--%s--\r\n", boundary)

	return msg
}

// SendEmailOnFailure sends report via email when JSON file creation fails
func (m *UptimeMonitor) SendEmailOnFailure(report *MonitorReport, head *string) error {
	if m.config.EmailAuth == "" || len(m.config.EmailTo) == 0 || m.config.EmailUser == "" {
		return nil
	}

	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON data: %w", err)
	}

	var subject string

	if head == nil {
		subject = "Uptime Monitor File Report Creation Failed"
	} else {
		subject = *head
	}

	plainBody := fmt.Sprintf(
		"Failed to create JSON file for report\n\n"+
			"The report data is attached below:\n\n"+
			"=== BEGIN JSON DATA ===\n"+
			"%s\n"+
			"=== END JSON DATA ===\n",
		string(jsonBytes),
	)

	htmlBody, err := BuildHTMLReport(report, subject)

	if err != nil {
		htmlBody = "<pre>" + plainBody + "</pre>"
	}

	message := BuildEmailMessage(
		m.config.EmailUser,
		m.config.EmailTo,
		subject,
		htmlBody,
		plainBody,
	)

	auth := smtp.PlainAuth("", m.config.EmailUser, m.config.EmailAuth, m.config.SMTPHost)

	err = smtp.SendMail(
		m.config.SMTPHost+":"+m.config.SMTPPort,
		auth,
		m.config.EmailUser,
		m.config.EmailTo,
		message,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	m.logger.Info("Email sent with JSON data",
		zap.Int("data_size", len(jsonBytes)),
	)
	return nil
}

// SubmitToAPI submits the monitoring report to external API with rate limiting and retries
func (m *UptimeMonitor) SubmitToAPI(ctx context.Context, report *MonitorReport) error {
	if m.config.APIURL == "" {
		return fmt.Errorf("failed to provide backend url")
	}

	retryConfig := DefaultRetryConfig()
	var lastErr error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if err := m.config.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		jsonData, err := json.Marshal(report)
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", m.config.APIURL, strings.NewReader(string(jsonData)))
		if err != nil {
			lastErr = fmt.Errorf("failed to create API request: %w", err)

			if attempt == retryConfig.MaxRetries {
				return fmt.Errorf("API submission failed after %d attempts: %w", retryConfig.MaxRetries+1, lastErr)
			}

			backoff := retryConfig.CalculateBackoff(attempt)

			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
				continue
			}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", m.config.UserAgent)
		if m.config.APIKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))
		}

		resp, err := m.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to submit to API: %w", err)

			if attempt == retryConfig.MaxRetries {
				return fmt.Errorf("API submission failed after %d attempts: %w", retryConfig.MaxRetries+1, lastErr)
			}

			backoff := retryConfig.CalculateBackoff(attempt)

			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
				continue
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API submission failed with status %d: %s", resp.StatusCode, string(body))

			if !IsRetryableError(lastErr, resp.StatusCode) {
				return lastErr
			}

			if attempt < retryConfig.MaxRetries {
				backoff := retryConfig.CalculateBackoff(attempt)
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(backoff):
					continue
				}
			}

			return lastErr
		}

		return nil
	}

	return fmt.Errorf("API submission failed after %d attempts: %w", retryConfig.MaxRetries+1, lastErr)
}

// SendNotifications sends notifications for the given report
func (m *UptimeMonitor) SendNotifications(ctx context.Context, report *MonitorReport) {
	if report.Downtime == 0 && report.Degraded == 0 {
		return
	}

	if m.config.SlackWebhook != "" {
		if err := m.sendSlackNotification(ctx, report); err != nil {
			m.logger.Error("Failed to send Slack notification", zap.Error(err))
		}
	}

	if m.config.DiscordWebhook != "" {
		if err := m.sendDiscordNotification(ctx, report); err != nil {
			m.logger.Error("Failed to send Discord notification", zap.Error(err))
		}
	}
}

// sendSlackNotification sends a notification to Slack
func (m *UptimeMonitor) sendSlackNotification(ctx context.Context, report *MonitorReport) error {
	color := "danger"
	if report.Downtime == 0 {
		color = "warning"
	}

	var failedServices []string
	for _, result := range report.Results {
		if result.Status == StatusDown || result.Status == StatusDegraded {
			failedServices = append(failedServices, fmt.Sprintf("%s (%s)", result.Domain, result.Status))
		}
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("🚨 Uptime Alert - %d service(s) down, %d degraded", report.Downtime, report.Degraded),
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"fields": []map[string]interface{}{
					{"title": "Environment", "value": report.Environment, "short": true},
					{"title": "Uptime", "value": fmt.Sprintf("%.2f%%", report.UptimePercent), "short": true},
					{"title": "Down", "value": fmt.Sprintf("%d", report.Downtime), "short": true},
					{"title": "Degraded", "value": fmt.Sprintf("%d", report.Degraded), "short": true},
					{"title": "Failed Services", "value": strings.Join(failedServices, "\n"), "short": false},
				},
				"footer": "Uptime Monitor",
				"ts":     report.Timestamp.Unix(),
			},
		},
	}

	return m.sendWebhook(ctx, m.config.SlackWebhook, payload)
}

func (m *UptimeMonitor) sendDiscordNotification(ctx context.Context, report *MonitorReport) error {
	var failedServices []string
	for _, result := range report.Results {
		if result.Status == StatusDown || result.Status == StatusDegraded {
			emoji := "🔴"
			if result.Status == StatusDegraded {
				emoji = "🟡"
			}
			failedServices = append(failedServices, fmt.Sprintf("%s **%s** - %s", emoji, result.Domain, result.Status))
		}
	}

	content := fmt.Sprintf("🚨 **Uptime Alert**\n\n"+
		"**Environment:** %s\n"+
		"**Uptime:** %.2f%%\n"+
		"**Down:** %d | **Degraded:** %d\n\n"+
		"**Failed Services:**\n%s",
		report.Environment,
		report.UptimePercent,
		report.Downtime,
		report.Degraded,
		strings.Join(failedServices, "\n"))

	payload := map[string]interface{}{
		"content":  content,
		"username": "Uptime Monitor",
	}

	return m.sendWebhook(ctx, m.config.DiscordWebhook, payload)
}

func (m *UptimeMonitor) sendWebhook(ctx context.Context, url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook failed with status %d: %s", resp.StatusCode, string(body))
	}

	m.logger.Info("Notification sent successfully", zap.String("webhook", url))
	return nil
}

func setupMonitorLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
