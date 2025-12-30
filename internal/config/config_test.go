package config

import (
	"os"
	"testing"
	"time"
)

func TestInterpolateEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "Replace ${VAR}",
			input:    "host: ${TEST_HOST}",
			envVars:  map[string]string{"TEST_HOST": "localhost"},
			expected: "host: localhost",
		},
		{
			name:     "Replace $VAR",
			input:    "port: $TEST_PORT",
			envVars:  map[string]string{"TEST_PORT": "4545"},
			expected: "port: 4545",
		},
		{
			name:     "No replacement if not found",
			input:    "value: ${NOT_SET}",
			envVars:  map[string]string{},
			expected: "value: ${NOT_SET}",
		},
		{
			name:     "Multiple replacements",
			input:    "url: ${PROTO}://${HOST}:${PORT}",
			envVars:  map[string]string{"PROTO": "http", "HOST": "localhost", "PORT": "4545"},
			expected: "url: http://localhost:4545",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := interpolateEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateChannel(t *testing.T) {
	tests := []struct {
		name      string
		channel   ChannelConfig
		wantError bool
	}{
		{
			name: "Valid SMTP channel",
			channel: ChannelConfig{
				Type: "smtp",
				Host: "smtp.example.com",
				Port: 587,
				From: "sender@example.com",
				To:   "receiver@example.com",
			},
			wantError: false,
		},
		{
			name: "Invalid SMTP - missing host",
			channel: ChannelConfig{
				Type: "smtp",
				Port: 587,
				From: "sender@example.com",
				To:   "receiver@example.com",
			},
			wantError: true,
		},
		{
			name: "Valid Discord channel",
			channel: ChannelConfig{
				Type:       "discord",
				WebhookURL: "https://discord.com/api/webhooks/123/abc",
			},
			wantError: false,
		},
		{
			name: "Invalid Discord - missing webhook",
			channel: ChannelConfig{
				Type: "discord",
			},
			wantError: true,
		},
		{
			name: "Unsupported channel type",
			channel: ChannelConfig{
				Type: "unsupported",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChannel("test", tt.channel)
			if (err != nil) != tt.wantError {
				t.Errorf("validateChannel() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
port: 9090
log_level: debug
dedup_ttl: 10m

channels:
  - type: discord
    name: test-discord
    webhook_url: ${TEST_WEBHOOK_URL}

priority_groups:
  high:
    - test-discord
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Set environment variables
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	os.Setenv("TEST_WEBHOOK_URL", "https://example.com/webhook")
	defer os.Unsetenv("CONFIG_PATH")
	defer os.Unsetenv("TEST_WEBHOOK_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got %s", cfg.LogLevel)
	}
	if cfg.Channels["test-discord"].WebhookURL != "https://example.com/webhook" {
		t.Errorf("Expected interpolated webhook URL, got %s", cfg.Channels["test-discord"].WebhookURL)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create a minimal config file
	configContent := `
channels:
  - type: discord
    webhook_url: https://example.com/webhook
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if cfg.Port != 4545 {
		t.Errorf("Expected default port 4545, got %d", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.LogLevel)
	}
	if cfg.DedupTTL != 5*time.Minute {
		t.Errorf("Expected default dedup TTL 5m, got %v", cfg.DedupTTL)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "Valid config",
			config: Config{
				LogLevel: "info",
				DedupTTL: 5 * time.Minute,
				Channels: map[string]ChannelConfig{
					"email": {
						Type: "smtp",
						Host: "smtp.example.com",
						Port: 587,
						From: "sender@example.com",
						To:   "receiver@example.com",
					},
				},
				PriorityGroups: map[string][]string{
					"high": {"email"},
				},
			},
			wantError: false,
		},
		{
			name: "Invalid log level",
			config: Config{
				LogLevel: "invalid",
				Channels: map[string]ChannelConfig{
					"email": {
						Type: "smtp",
						Host: "smtp.example.com",
						Port: 587,
						From: "sender@example.com",
						To:   "receiver@example.com",
					},
				},
			},
			wantError: true,
		},
		{
			name: "No channels",
			config: Config{
				LogLevel: "info",
				Channels: map[string]ChannelConfig{},
			},
			wantError: true,
		},
		{
			name: "Priority group references non-existent channel",
			config: Config{
				LogLevel: "info",
				Channels: map[string]ChannelConfig{
					"email": {
						Type: "smtp",
						Host: "smtp.example.com",
						Port: 587,
						From: "sender@example.com",
						To:   "receiver@example.com",
					},
				},
				PriorityGroups: map[string][]string{
					"high": {"nonexistent"},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Config.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
