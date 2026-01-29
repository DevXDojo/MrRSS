// Package monitor provides analytics tracking for the application
package monitor

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"MrRSS/internal/utils"
	"MrRSS/internal/version"
)

const (
	// Default API endpoint
	defaultAPIURL = "https://cf-monitor-api.ch3nyang.workers.dev"
	// Default App ID
	defaultAppID = "mrrss"
)

// MonitorClient handles analytics reporting
type MonitorClient struct {
	apiURL   string
	appID    string
	deviceID string
	client   *http.Client
	devMode  bool
	enabled  bool
}

// ReportPayload represents the analytics event payload
type ReportPayload struct {
	EventType  string                 `json:"eventType"`
	EventName  string                 `json:"eventName,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	DeviceInfo DeviceInfo             `json:"deviceInfo"`
}

// DeviceInfo represents device information
type DeviceInfo struct {
	OSType     string `json:"osType"`
	OSVersion  string `json:"osVersion,omitempty"`
	AppVersion string `json:"appVersion,omitempty"`
}

// NewMonitorClient creates a new monitor client
func NewMonitorClient(apiURL, appID string) *MonitorClient {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	if appID == "" {
		appID = defaultAppID
	}

	// Check if monitoring should be enabled (disabled in dev mode)
	devMode := os.Getenv("MRRSS_DEBUG") != ""
	enabled := !devMode

	return &MonitorClient{
		apiURL:   apiURL,
		appID:    appID,
		deviceID: getOrCreateDeviceID(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		devMode: devMode,
		enabled: enabled,
	}
}

// getOrCreateDeviceID generates or retrieves the device ID
func getOrCreateDeviceID() string {
	// Try to read from existing file first
	deviceIDPath, err := getDeviceIDPath()
	if err == nil {
		if data, err := os.ReadFile(deviceIDPath); err == nil {
			id := strings.TrimSpace(string(data))
			if id != "" {
				return id
			}
		}
	}

	// Generate new device ID based on machine characteristics
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = "unknown"
	}

	// Create a stable hash based on machine characteristics
	data := strings.Join([]string{hostname, user, runtime.GOOS}, "-")
	hash := md5.Sum([]byte(data))
	deviceID := "device-" + hex.EncodeToString(hash[:])

	// Save to file for future use
	if deviceIDPath != "" {
		_ = os.WriteFile(deviceIDPath, []byte(deviceID), 0666)
	}

	return deviceID
}

// getDeviceIDPath returns the path to store the device ID
func getDeviceIDPath() (string, error) {
	configDir, err := utils.GetDataDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config path: %w", err)
	}
	return configDir + "/monitor_device_id.txt", nil
}

// ReportAppStart reports application startup event
func (c *MonitorClient) ReportAppStart(ctx context.Context) error {
	payload := ReportPayload{
		EventType: "app_start",
		DeviceInfo: DeviceInfo{
			OSType:     getOSType(),
			OSVersion:  getOSVersion(),
			AppVersion: getAppVersion(),
		},
	}
	return c.report(ctx, payload)
}

// ReportPageView reports a page view event
func (c *MonitorClient) ReportPageView(ctx context.Context, page string) error {
	payload := ReportPayload{
		EventType: "page_view",
		EventName: page,
		DeviceInfo: DeviceInfo{
			OSType:     getOSType(),
			AppVersion: getAppVersion(),
		},
	}
	return c.report(ctx, payload)
}

// ReportEvent reports a custom event
func (c *MonitorClient) ReportEvent(ctx context.Context, eventName string, properties map[string]interface{}) error {
	payload := ReportPayload{
		EventType:  "custom_event",
		EventName:  eventName,
		Properties: properties,
		DeviceInfo: DeviceInfo{
			OSType:     getOSType(),
			AppVersion: getAppVersion(),
		},
	}
	return c.report(ctx, payload)
}

// report sends the analytics event to the server
func (c *MonitorClient) report(ctx context.Context, payload ReportPayload) error {
	// Skip reporting in development mode
	if !c.enabled {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/api/report", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-ID", c.appID)
	req.Header.Set("X-Device-ID", c.deviceID)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Silently consume response - don't log to avoid clutter
	_, _ = io.ReadAll(resp.Body)

	// Don't fail the application even if report fails
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	return nil
}

// getOSType returns the OS type in the format expected by the monitor API
func getOSType() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	default:
		return "unknown"
	}
}

// getOSVersion returns the OS version
func getOSVersion() string {
	// For simplicity, we're not implementing detailed OS version detection
	// This could be extended later if needed
	switch runtime.GOOS {
	case "windows":
		// Could use registry to get exact version
		return "windows"
	case "darwin":
		// Could use sw_vers command to get exact version
		return "macos"
	case "linux":
		// Linux versions vary widely
		return "linux"
	default:
		return "unknown"
	}
}

// getAppVersion returns the application version
func getAppVersion() string {
	return version.Version
}
