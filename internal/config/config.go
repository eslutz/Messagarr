package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Port           int                      `yaml:"port"`
	LogLevel       string                   `yaml:"log_level"`
	DedupTTL       time.Duration            `yaml:"dedup_ttl"`
	ChannelsList   []ChannelConfig          `yaml:"channels"`
	Channels       map[string]ChannelConfig `yaml:"-"` // Built from ChannelsList
	PriorityGroups map[string][]string      `yaml:"priority_groups"`
}

// ChannelConfig represents a notification channel configuration
type ChannelConfig struct {
	Type       string            `yaml:"type"`
	Name       string            `yaml:"name,omitempty"` // Optional: defaults to Type
	Host       string            `yaml:"host,omitempty"`
	Port       int               `yaml:"port,omitempty"`
	User       string            `yaml:"user,omitempty"`
	Password   string            `yaml:"password,omitempty"`
	From       string            `yaml:"from,omitempty"`
	To         string            `yaml:"to,omitempty"`
	WebhookURL string            `yaml:"webhook_url,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty"`
}

// GetName returns the channel's name (defaults to type if not specified)
func (c *ChannelConfig) GetName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Type
}

// Load reads and parses the configuration file
func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/messagarr/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Perform environment variable interpolation
	configStr := interpolateEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(configStr), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Port == 0 {
		cfg.Port = 4545
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.DedupTTL == 0 {
		cfg.DedupTTL = 5 * time.Minute
	}

	// Build channels map from list (name defaults to type)
	cfg.Channels = make(map[string]ChannelConfig)
	for _, channel := range cfg.ChannelsList {
		name := channel.GetName()
		if _, exists := cfg.Channels[name]; exists {
			return nil, fmt.Errorf("duplicate channel name: %s (use 'name' field to differentiate)", name)
		}
		cfg.Channels[name] = channel
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	// Validate channels
	if len(c.Channels) == 0 {
		return fmt.Errorf("no channels configured")
	}

	for name, channel := range c.Channels {
		if err := validateChannel(name, channel); err != nil {
			return err
		}
	}

	// Validate priority groups reference existing channels
	for priority, channels := range c.PriorityGroups {
		for _, channelName := range channels {
			if _, exists := c.Channels[channelName]; !exists {
				return fmt.Errorf("priority group %s references non-existent channel: %s", priority, channelName)
			}
		}
	}

	return nil
}

// validateChannel validates a single channel configuration
func validateChannel(name string, channel ChannelConfig) error {
	if channel.Type == "" {
		return fmt.Errorf("channel %s: type is required", name)
	}

	switch channel.Type {
	case "smtp":
		if channel.Host == "" {
			return fmt.Errorf("channel %s: smtp host is required", name)
		}
		if channel.Port == 0 {
			return fmt.Errorf("channel %s: smtp port is required", name)
		}
		if channel.From == "" {
			return fmt.Errorf("channel %s: smtp from address is required", name)
		}
		if channel.To == "" {
			return fmt.Errorf("channel %s: smtp to address is required", name)
		}
	case "discord", "slack", "teams":
		if channel.WebhookURL == "" {
			return fmt.Errorf("channel %s: webhook_url is required for %s", name, channel.Type)
		}
	default:
		return fmt.Errorf("channel %s: unsupported channel type: %s", name, channel.Type)
	}
	return nil
}

// interpolateEnvVars replaces ${VAR} or $VAR with environment variable values
func interpolateEnvVars(s string) string {
	// Match ${VAR} or $VAR patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name
		varName := strings.TrimPrefix(match, "$")
		varName = strings.TrimPrefix(varName, "{")
		varName = strings.TrimSuffix(varName, "}")

		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return original if not found
		return match
	})
}
