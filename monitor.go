package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	StatusUp       = "up"
	StatusDown     = "down"
	StatusDegraded = "degraded"

	ThresholdFast    = 1000
	ThresholdAccept  = 3000
	SSLExpiryWarning = 30

	DefaultTimeout    = 30 * time.Second
	DefaultUserAgent  = ""
	DefaultConcurrent = 5
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
	UserAgent      string
	Concurrent     int
	Environment    string
	OutputDir      string
	SlackWebhook   string
	DiscordWebhook string
}

type UptimeMonitor struct {
	config *MonitorConfig
	logger *zap.Logger
	client *http.Client
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
	result := HealthCheckResult{
		Domain:    domain,
		URL:       domain,
		Timestamp: time.Now(),
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
		result.URL = domain
	}

	result.IsSSL = strings.HasPrefix(domain, "https://")

	req, err := http.NewRequestWithContext(ctx, "GET", domain, nil)
	if err != nil {
		result.Status = StatusDown
		result.ErrorMessage = fmt.Sprintf("Failed to create request: %v", err)
		m.logger.Error("Request creation failed",
			zap.String("domain", result.Domain),
			zap.Error(err))
		return result
	}

	req.Header.Set("User-Agent", m.config.UserAgent)

	startTime := time.Now()
	resp, err := m.client.Do(req)
	duration := time.Since(startTime)
	result.ResponseTime = duration.Milliseconds()

	if err != nil {
		result.Status = StatusDown
		result.ErrorMessage = fmt.Sprintf("Request failed: %v", err)
		m.logger.Error("Request failed",
			zap.String("domain", result.Domain),
			zap.Error(err))
		return result
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

	m.logger.Info("Health check completed",
		zap.String("domain", result.Domain),
		zap.String("status", result.Status),
		zap.Int("status_code", result.StatusCode),
		zap.Int64("response_time_ms", result.ResponseTime))

	return result
}

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

func (m *UptimeMonitor) RunCheck(ctx context.Context) (*MonitorReport, error) {
	m.logger.Info("Starting uptime monitoring",
		zap.Int("total_domains", len(m.config.Domains)),
		zap.String("environment", m.config.Environment))

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

	m.logger.Info("Monitoring completed",
		zap.Int("total_checks", report.TotalChecks),
		zap.Int("uptime", report.Uptime),
		zap.Int("downtime", report.Downtime),
		zap.Float64("uptime_percent", report.UptimePercent))

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

func (m *UptimeMonitor) SaveReport(report *MonitorReport) (string, error) {
	if err := os.MkdirAll(m.config.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s/uptime_report_%s.json", m.config.OutputDir, timestamp)

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	m.logger.Info("Report saved", zap.String("file", filename))
	return filename, nil
}

func (m *UptimeMonitor) SubmitToAPI(ctx context.Context, report *MonitorReport) error {
	if m.config.APIURL == "" {
		m.logger.Debug("API URL not configured, skipping submission")
		return nil
	}

	jsonData, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.config.APIURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create API request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if m.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit to API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API submission failed with status %d: %s", resp.StatusCode, string(body))
	}

	m.logger.Info("Report submitted to API", zap.String("url", m.config.APIURL), zap.Int("status", resp.StatusCode))
	return nil
}

func (m *UptimeMonitor) SendNotifications(ctx context.Context, report *MonitorReport) {
	if report.Downtime == 0 && report.Degraded == 0 {
		m.logger.Debug("No issues detected, skipping notifications")
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
