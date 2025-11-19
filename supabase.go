package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	supa_storage "github.com/supabase-community/storage-go"
)

func storageChartImage(chartBase64 string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	bucket := "uptime-charts"

	if supabaseURL == "" || supabaseKey == "" {
		return "", fmt.Errorf("missing SUPABASE_URL or SUPABASE_KEY environment variables")
	}

	projectURL := fmt.Sprintf("%s/storage/v1", supabaseURL)
	storageClient := supa_storage.NewClient(projectURL, supabaseKey, nil)

	_, err := storageClient.GetBucket(bucket)
	if err != nil {
		_, err := storageClient.CreateBucket(bucket, supa_storage.BucketOptions{
			Public: true,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	data, err := base64.StdEncoding.DecodeString(chartBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode chart image: %w", err)
	}

	dateFolder := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("charts/%s/report_%d.png", dateFolder, time.Now().Unix())

	_, err = storageClient.UploadFile(bucket, filename, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to upload chart: %w", err)
	}

	publicURL := storageClient.GetPublicUrl(bucket, filename)
	return publicURL.SignedURL, nil
}

func queryDataFromSupabase(query string) ([]map[string]interface{}, error) {
	return nil, nil
}
