package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func main() {
	logger, err := setupMonitorLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	config, err := NewMonitorConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	monitor := NewUptimeMonitor(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := monitor.RunCheck(ctx)
	if err != nil {
		logger.Fatal("Monitoring failed", zap.Error(err))
	}

	if _, err := monitor.SaveReport(report); err != nil {
		logger.Error("Failed to save report", zap.Error(err))
	}

	if err := monitor.SubmitToAPI(ctx, report); err != nil {
		logger.Error("Failed to submit report to API", zap.Error(err))
	}

	monitor.SendNotifications(ctx, report)

	exitCode := 0
	if report.Downtime > 0 {
		exitCode = 1
		logger.Error("Services are down", zap.Int("count", report.Downtime))
	}

	logger.Info("Monitoring completed successfully",
		zap.Int("exit_code", exitCode),
		zap.Float64("uptime_percent", report.UptimePercent),
		zap.Int("total_checks", report.TotalChecks))

	os.Exit(exitCode)
}
