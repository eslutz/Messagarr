package channels

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

func TestDiscordDispatcher(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.ChannelConfig{
		Type:       "discord",
		WebhookURL: server.URL,
	}

	dispatcher := NewDiscordDispatcher("test-discord", cfg)

	req := &models.NotificationRequest{
		Title:    "Test Notification",
		Body:     "This is a test",
		Priority: "high",
		Metadata: map[string]string{
			"key1": "value1",
		},
	}

	err := dispatcher.Send(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSlackDispatcher(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.ChannelConfig{
		Type:       "slack",
		WebhookURL: server.URL,
	}

	dispatcher := NewSlackDispatcher("test-slack", cfg)

	req := &models.NotificationRequest{
		Title:    "Test Notification",
		Body:     "This is a test",
		Service:  "test-service",
		Priority: "normal",
	}

	err := dispatcher.Send(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestTeamsDispatcher(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.ChannelConfig{
		Type:       "teams",
		WebhookURL: server.URL,
	}

	dispatcher := NewTeamsDispatcher("test-teams", cfg)

	req := &models.NotificationRequest{
		Title:     "Test Notification",
		Body:      "This is a test",
		EventType: "test.event",
		Priority:  "low",
	}

	err := dispatcher.Send(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestChannelDispatcher(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{
			"discord": {
				Type:       "discord",
				WebhookURL: server.URL,
			},
			"slack": {
				Type:       "slack",
				WebhookURL: server.URL,
			},
		},
	}

	dispatcher := NewChannelDispatcher(cfg)

	req := &models.NotificationRequest{
		Title: "Test",
		Body:  "Test body",
	}

	results := dispatcher.Dispatch(req, []string{"discord", "slack"})

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for channel, result := range results {
		if !result.Success {
			t.Errorf("Channel %s failed: %s", channel, result.Error)
		}
	}
}

func TestChannelDispatcherNonExistentChannel(t *testing.T) {
	cfg := &config.Config{
		Channels: map[string]config.ChannelConfig{},
	}

	dispatcher := NewChannelDispatcher(cfg)

	req := &models.NotificationRequest{
		Title: "Test",
		Body:  "Test body",
	}

	results := dispatcher.Dispatch(req, []string{"nonexistent"})

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	result := results["nonexistent"]
	if result.Success {
		t.Error("Expected failure for non-existent channel")
	}
	if result.Error != "Channel not found" {
		t.Errorf("Expected 'Channel not found' error, got %s", result.Error)
	}
}

func TestEmailDispatcherName(t *testing.T) {
	cfg := config.ChannelConfig{
		Type: "smtp",
		Host: "smtp.example.com",
		Port: 587,
		From: "sender@example.com",
		To:   "receiver@example.com",
	}

	dispatcher := NewEmailDispatcher("test-email", cfg)

	if dispatcher.Name() != "test-email" {
		t.Errorf("Expected name 'test-email', got %s", dispatcher.Name())
	}
}

func TestDiscordGetColor(t *testing.T) {
	cfg := config.ChannelConfig{
		Type:       "discord",
		WebhookURL: "http://example.com/webhook",
	}

	dispatcher := NewDiscordDispatcher("test-discord", cfg)

	tests := []struct {
		priority string
		expected int
	}{
		{"high", 0xFF0000},
		{"normal", 0x00FF00},
		{"low", 0x0000FF},
		{"unknown", 0x808080},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			color := dispatcher.getColor(tt.priority)
			if color != tt.expected {
				t.Errorf("Expected color %x, got %x", tt.expected, color)
			}
		})
	}
}

