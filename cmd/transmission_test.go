package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/hekmon/transmissionrpc/v3"
)

func TestAddTorrent(t *testing.T) {
	// Test configuration (can be overridden via env vars)
	trURL := getEnvOrDefault("TEST_TR_URL", "http://localhost:9091/transmission/rpc")
	trUsername := getEnvOrDefault("TEST_TR_USERNAME", "")
	trPassword := getEnvOrDefault("TEST_TR_PASSWORD", "")
	testTorrentPath := getEnvOrDefault("TEST_TORRENT_PATH", "")

	// Ensure test torrent file exists
	if _, err := os.Stat(testTorrentPath); os.IsNotExist(err) {
		t.Fatalf("Test torrent file not found: %s", testTorrentPath)
	}

	// Connect to Transmission
	parsedURL, err := url.Parse(trURL)
	if err != nil {
		t.Fatalf("Failed to parse Transmission URL: %v", err)
	}

	// Set custom HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Add basic auth to the request if credentials are provided
	if trUsername != "" || trPassword != "" {
		parsedURL.User = url.UserPassword(trUsername, trPassword)
	}

	trConfig := transmissionrpc.Config{
		CustomClient: httpClient,
	}

	t.Logf("Connecting to Transmission at %s", parsedURL.String())
	trClient, err := transmissionrpc.New(parsedURL, &trConfig)
	if err != nil {
		t.Fatalf("Failed to connect to Transmission: %v", err)
	}
	t.Log("Successfully connected to Transmission")

	// Add torrent to Transmission
	torrentB64, err := transmissionrpc.File2Base64(testTorrentPath)
	if err != nil {
		t.Fatalf("Failed to encode torrent file %s: %v", testTorrentPath, err)
	}

	paused := true // Don't start downloading in test mode
	addOptions := transmissionrpc.TorrentAddPayload{
		MetaInfo: &torrentB64,
		Paused:   &paused,
	}

	t.Log("Attempting to add torrent to Transmission")
	tr, err := trClient.TorrentAdd(context.Background(), addOptions)
	if err != nil {
		t.Fatalf("Failed to add torrent to Transmission: %v", err)
	}

	t.Logf("Added torrent to Transmission: %s (ID: %d)", *tr.Name, *tr.ID)

	// Clean up after test
	err = trClient.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
		IDs:             []int64{*tr.ID},
		DeleteLocalData: true,
	})
	if err != nil {
		t.Logf("Warning: Failed to remove test torrent from Transmission: %v", err)
	} else {
		t.Log("Successfully removed test torrent from Transmission")
	}
}

// Helper function to get environment variable with default fallback
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
