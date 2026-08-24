package network

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/network"
	"MrRSS/internal/utils/httputil"
)

// HandleDetectNetwork detects network speed and updates settings
// @Summary      Detect network speed
// @Description  Detect network speed and latency, then optimize max concurrent refreshes setting
// @Tags         network
// @Accept       json
// @Produce      json
// @Success      200  {object}  network.DetectionResult  "Network detection result (speed_level, bandwidth_mbps, latency_ms, max_concurrency)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /network/detect [post]
func HandleDetectNetwork(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get proxy settings
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetSetting("proxy_username")
	proxyPassword, _ := h.DB.GetSetting("proxy_password")

	// Create HTTP client with proxy if enabled
	var httpClient *http.Client
	if proxyEnabled == "true" {
		proxyURL := httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
		if proxyURL != "" {
			client, err := httputil.CreateHTTPClient(proxyURL, 10*time.Second)
			if err != nil {
				log.Printf("Failed to create HTTP client with proxy: %v", err)
				// Fall back to default client
				httpClient = &http.Client{Timeout: 10 * time.Second}
			} else {
				httpClient = client
			}
		} else {
			httpClient = &http.Client{Timeout: 10 * time.Second}
		}
	} else {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	detector := network.NewDetector(httpClient)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result := detector.DetectSpeed(ctx)

	// Store results in settings
	if result.DetectionSuccess {
		h.DB.SetSetting("network_speed", string(result.SpeedLevel))
		h.DB.SetSetting("network_bandwidth_mbps", fmt.Sprintf("%.2f", result.BandwidthMbps))
		h.DB.SetSetting("network_latency_ms", strconv.FormatInt(result.LatencyMs, 10))
		h.DB.SetSetting("max_concurrent_refreshes", strconv.Itoa(result.MaxConcurrency))
		h.DB.SetSetting("last_network_test", result.DetectionTime.Format(time.RFC3339))
	}

	// Return results to frontend
	response.JSON(w, result)
}

// HandleGetNetworkInfo returns current network detection info from settings
// @Summary      Get network info
// @Description  Get the last network detection information (speed, bandwidth, latency, concurrent refreshes)
// @Tags         network
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Network info (speed_level, bandwidth_mbps, latency_ms, max_concurrent_refreshes, last_network_test)"
// @Router       /network/info [get]
func HandleGetNetworkInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	speedLevel, _ := h.DB.GetSetting("network_speed")
	bandwidthStr, _ := h.DB.GetSetting("network_bandwidth_mbps")
	latencyStr, _ := h.DB.GetSetting("network_latency_ms")
	concurrencyStr, _ := h.DB.GetSetting("max_concurrent_refreshes")
	lastTestStr, _ := h.DB.GetSetting("last_network_test")

	bandwidth, err := strconv.ParseFloat(bandwidthStr, 64)
	if err != nil {
		bandwidth = 0
	}

	latency, err := strconv.ParseInt(latencyStr, 10, 64)
	if err != nil {
		latency = 0
	}

	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		concurrency = 5 // Default
	}

	var lastTest time.Time
	if lastTestStr != "" {
		lastTest, _ = time.Parse(time.RFC3339, lastTestStr)
	}

	result := network.DetectionResult{
		SpeedLevel:       network.SpeedLevel(speedLevel),
		BandwidthMbps:    bandwidth,
		LatencyMs:        latency,
		MaxConcurrency:   concurrency,
		DetectionTime:    lastTest,
		DetectionSuccess: speedLevel != "",
	}

	response.JSON(w, result)
}
