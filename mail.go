package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func BuildHTMLReport(report *MonitorReport, subject string) (string, error) {
	var chartBase64 string

	jsonBytes, err := json.MarshalIndent(report, "", "  ")

	if err != nil {
		return "", fmt.Errorf("failed to build json data: %w", err)
	}

	chartBase64, err = generateUptimeChart(report)
	if err != nil {
		fmt.Println("err", err)
		chartBase64 = ""
	} else {
		uploadedLink, uploadErr := storageChartImage(chartBase64)
		if uploadErr == nil {
			chartBase64 = uploadedLink
		} else {
			fmt.Println("err", uploadErr)
			chartBase64 = ""
		}
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>%s</title>
<style>
body {
  font-family: "Segoe UI", Roboto, Arial, sans-serif;
  background-color: #f8f9fb;
  margin: 0;
  color: #333;
}
.container {
  max-width: 850px;
  margin: 30px auto;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 3px 10px rgba(0,0,0,0.1);
  overflow: hidden;
}
.header {
  background: linear-gradient(135deg, #2f2e41, #4a47a3);
  color: #fff;
  padding: 20px 30px;
}
.header h1 { margin: 0; font-size: 1.6em; }
.header p {color: #ffffff}
.section { padding: 20px 30px; }
h2 { color: #2f2e41; border-bottom: 2px solid #eee; padding-bottom: 5px; }
.stats {
  display: flex; flex-wrap: wrap; gap: 15px; margin-top: 10px;
}
.stat {
  flex: 1 1 150px; background: #f5f6f9;
  padding: 10px; border-radius: 8px; text-align: center;
}
.stat span {
  display: block; font-size: 1.3em; font-weight: bold; color: #2f2e41;
}
.table-container { overflow-x: auto; margin-top: 15px; }
table { width: 100%%; border-collapse: collapse; }
th, td { padding: 10px; text-align: left; border-bottom: 1px solid #eee; font-size: 0.95em; }
th { background: #fafafa; font-weight: 600; }
tr:hover { background: #f9f9ff; }
.status-up { color: #2ecc71; font-weight: bold; }
.status-down { color: #e74c3c; font-weight: bold; }
.status-degraded { color: #f39c12; font-weight: bold; }
.chart {
  width: 100%%;
  text-align: center;
  margin-top: 15px;
}
.footer {
  background: #f4f4f8; color: #777;
  text-align: center; padding: 15px; font-size: 0.85em;
}
pre {
  background: #1e1e1e; color: #eee; padding: 12px;
  border-radius: 8px; overflow-x: auto; font-size: 0.9em;
}
</style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>📡 %s</h1>
      <p>Generated on %s</p>
    </div>

    <div class="section">
      <h2>Summary</h2>
      <div class="stats">
        <div class="stat"><span>%d</span>Total Checks</div>
        <div class="stat"><span>%d</span>Uptime</div>
        <div class="stat"><span>%d</span>Downtime</div>
        <div class="stat"><span>%d</span>Degraded</div>
        <div class="stat"><span>%.2f%%</span>Uptime %%</div>
        <div class="stat"><span>%.2f ms</span>Avg Latency</div>
      </div>
      <div class="chart">
        <img src="%s" alt="Uptime Chart" style="max-width: 100%%; border-radius: 8px; margin-top: 10px;">
      </div>
    </div>

    <div class="section">
      <h2>Detailed Results</h2>
      <div class="table-container">
        <table>
          <tr>
            <th>Domain</th><th>Status</th><th>Code</th><th>Latency</th>
            <th>SSL Expiry</th><th>Checked At</th>
          </tr>
          %s
        </table>
      </div>
    </div>

    <div class="section">
      <h2>Raw JSON Data</h2>
      <pre>%s</pre>
    </div>

    <div class="footer">
      <p>Powered by <strong>Axiolot Hub</strong> — Reliable monitoring, elegant delivery.</p>
    </div>
  </div>
</body>
</html>`,
		subject,
		subject,
		report.Timestamp.Format(time.RFC1123),
		report.TotalChecks, report.Uptime, report.Downtime, report.Degraded,
		report.UptimePercent, report.AverageLatency,
		chartBase64,
		buildResultsTable(report.Results),
		string(jsonBytes),
	)

	return html, nil
}


func buildResultsTable(results []HealthCheckResult) string {
	rows := ""
	for _, r := range results {
		statusClass := "status-up"
		if strings.ToLower(r.Status) == "down" {
			statusClass = "status-down"
		} else if strings.ToLower(r.Status) == "degraded" {
			statusClass = "status-degraded"
		}
		rows += fmt.Sprintf(`
<tr>
	<td>%s</td>
	<td class="%s">%s</td>
	<td>%d</td>
	<td>%d ms</td>
	<td>%s</td>
	<td>%s</td>
</tr>`, r.Domain, statusClass, strings.ToUpper(r.Status), r.StatusCode, r.ResponseTime, r.SSLExpiry, r.CheckedAt)
	}
	return rows
}
