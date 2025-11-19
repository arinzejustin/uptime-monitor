package main

import (
	"bytes"
	"encoding/base64"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func generateUptimeChart(report *MonitorReport) (string, error) {
	Colors := []drawing.Color{
		drawing.ColorGreen,
		drawing.ColorRed,
		drawing.ColorYellow,
	}

	graph := chart.BarChart{
		Title: "Uptime Overview Report",
		TitleStyle: chart.Style{
			FontSize:  16,
			FontColor: drawing.ColorFromHex("2f2e41"),
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
			FillColor: drawing.ColorWhite,
		},
		Width:    800,
		Height:   400,
		BarWidth: 80,
		Bars: []chart.Value{
			{
				Value: float64(report.Uptime),
				Label: "Uptime",
				Style: chart.Style{FillColor: Colors[0], FontSize: 8, FontColor: Colors[0]}},
			{
				Value: float64(report.Downtime),
				Label: "Downtime",
				Style: chart.Style{FillColor: Colors[1], FontSize: 8, FontColor: Colors[1]}},
			{
				Value: float64(report.Degraded),
				Label: "Degraded",
				Style: chart.Style{FillColor: Colors[2], FontSize: 8, FontColor: Colors[2]}},
		},
		XAxis: chart.Style{
			FontSize: 10,
		},
		YAxis: chart.YAxis{
			Name: "Powered By Axiolot Hub",
			Style: chart.Style{
				FontSize: 8,
			},
		},
		Canvas: chart.Style{
			FillColor: drawing.ColorFromHex("f8f9fa"),
		},
	}

	var buf bytes.Buffer
	err := graph.Render(chart.PNG, &buf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
