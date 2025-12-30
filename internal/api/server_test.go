package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"test": {
				Type:       "discord",
				WebhookURL: "http://example.com/webhook",
			},
		},
	}

	server := NewServer(cfg, "test-version")
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp models.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", resp.Status)
	}
	if resp.Version != "test-version" {
		t.Errorf("Expected version 'test-version', got %s", resp.Version)
	}
}

func TestReadyEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"test": {
				Type:       "discord",
				WebhookURL: "http://example.com/webhook",
			},
		},
	}

	server := NewServer(cfg, "test")
	req := httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)
	w := httptest.NewRecorder()

	server.handleReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp models.ReadyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Ready {
		t.Error("Expected ready to be true")
	}
}

func TestStatusEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"test": {
				Type:       "discord",
				WebhookURL: "http://example.com/webhook",
			},
		},
	}

	server := NewServer(cfg, "test")
	req := httptest.NewRequest(http.MethodGet, "/status", http.NoBody)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp models.MetricsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.TotalNotifications != 0 {
		t.Errorf("Expected 0 total notifications, got %d", resp.TotalNotifications)
	}
}

func TestNotifyEndpointValidation(t *testing.T) {
	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"test": {
				Type:       "discord",
				WebhookURL: "http://example.com/webhook",
			},
		},
	}

	server := NewServer(cfg, "test")

	tests := []struct {
		name           string
		payload        any
		method         string
		expectedStatus int
	}{
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing title and body",
			method: http.MethodPost,
			payload: models.NotificationRequest{
				Priority: "high",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.payload != nil {
				if str, ok := tt.payload.(string); ok {
					body = bytes.NewBufferString(str)
				} else {
					data, _ := json.Marshal(tt.payload)
					body = bytes.NewBuffer(data)
				}
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, "/notify", body)
			w := httptest.NewRecorder()

			server.handleNotify(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestNotifyEndpointSuccess(t *testing.T) {
	// Create a test server for webhooks
	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"discord": {
				Type:       "discord",
				WebhookURL: webhookServer.URL,
			},
		},
		PriorityGroups: map[string][]string{
			"normal": {"discord"},
		},
	}

	server := NewServer(cfg, "test")

	notifReq := models.NotificationRequest{
		Title:    "Test",
		Body:     "Test body",
		Priority: "normal",
	}

	payload, _ := json.Marshal(notifReq)
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()

	server.handleNotify(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp models.NotificationResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestNotifyEndpointDuplicate(t *testing.T) {
	// Create a test server for webhooks
	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"discord": {
				Type:       "discord",
				WebhookURL: webhookServer.URL,
			},
		},
		PriorityGroups: map[string][]string{
			"normal": {"discord"},
		},
	}

	server := NewServer(cfg, "test")

	notifReq := models.NotificationRequest{
		Title:    "Test Duplicate",
		Body:     "Test body",
		Priority: "normal",
	}

	payload, _ := json.Marshal(notifReq)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(payload))
	w1 := httptest.NewRecorder()
	server.handleNotify(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", w1.Code)
	}

	// Second request (duplicate)
	req2 := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(payload))
	w2 := httptest.NewRecorder()
	server.handleNotify(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second request: Expected status 200, got %d", w2.Code)
	}

	var resp models.NotificationResponse
	if err := json.NewDecoder(w2.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Message != "Notification suppressed (duplicate)" {
		t.Errorf("Expected duplicate message, got: %s", resp.Message)
	}
}

func TestNotifyEndpointNoChannels(t *testing.T) {
	cfg := &config.Config{
		Port:           4545,
		LogLevel:       "info",
		DedupTTL:       5 * time.Minute,
		Channels:       map[string]config.ChannelConfig{},
		PriorityGroups: map[string][]string{},
	}

	server := NewServer(cfg, "test")

	notifReq := models.NotificationRequest{
		Title: "Test",
		Body:  "Test body",
	}

	payload, _ := json.Marshal(notifReq)
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(payload))
	w := httptest.NewRecorder()

	server.handleNotify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDetermineChannels(t *testing.T) {
	cfg := &config.Config{
		Port:     4545,
		LogLevel: "info",
		DedupTTL: 5 * time.Minute,
		Channels: map[string]config.ChannelConfig{
			"email": {
				Type: "smtp",
				Host: "smtp.example.com",
				Port: 587,
			},
			"discord": {
				Type:       "discord",
				WebhookURL: "http://example.com/webhook",
			},
			"slack": {
				Type:       "slack",
				WebhookURL: "http://example.com/webhook",
			},
		},
		PriorityGroups: map[string][]string{
			"high":   {"email", "discord"},
			"normal": {"discord"},
			"low":    {"slack"},
		},
	}

	server := NewServer(cfg, "test")

	tests := []struct {
		name     string
		req      models.NotificationRequest
		expected []string
	}{
		{
			name: "Explicit channels",
			req: models.NotificationRequest{
				Channels: []string{"email"},
			},
			expected: []string{"email"},
		},
		{
			name: "High priority",
			req: models.NotificationRequest{
				Priority: "high",
			},
			expected: []string{"email", "discord"},
		},
		{
			name: "Normal priority",
			req: models.NotificationRequest{
				Priority: "normal",
			},
			expected: []string{"discord"},
		},
		{
			name:     "Default priority (normal)",
			req:      models.NotificationRequest{},
			expected: []string{"discord"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels := server.determineChannels(&tt.req)
			if len(channels) != len(tt.expected) {
				t.Errorf("Expected %d channels, got %d", len(tt.expected), len(channels))
			}
		})
	}
}
