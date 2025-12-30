package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

// DiscordDispatcher sends notifications to Discord via webhooks
type DiscordDispatcher struct {
	config *config.ChannelConfig
	name   string
}

// NewDiscordDispatcher creates a new Discord dispatcher
func NewDiscordDispatcher(name string, cfg *config.ChannelConfig) *DiscordDispatcher {
	return &DiscordDispatcher{
		config: cfg,
		name:   name,
	}
}

// Name returns the dispatcher name
func (d *DiscordDispatcher) Name() string {
	return d.name
}

// DiscordWebhook represents a Discord webhook payload
type DiscordWebhook struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Color       int                 `json:"color,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Send sends a Discord notification
func (d *DiscordDispatcher) Send(req *models.NotificationRequest) error {
	embed := DiscordEmbed{
		Title:       req.Title,
		Description: req.Body,
		Color:       d.getColor(req.Priority),
	}

	// Add metadata as fields
	if len(req.Metadata) > 0 {
		for k, v := range req.Metadata {
			embed.Fields = append(embed.Fields, DiscordEmbedField{
				Name:   k,
				Value:  v,
				Inline: true,
			})
		}
	}

	// Add service and event type if present
	if req.Service != "" {
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:   "Service",
			Value:  req.Service,
			Inline: true,
		})
	}
	if req.EventType != "" {
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:   "Event Type",
			Value:  req.EventType,
			Inline: true,
		})
	}

	webhook := DiscordWebhook{
		Embeds: []DiscordEmbed{embed},
	}

	payload, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord webhook: %w", err)
	}

	resp, err := http.Post(d.config.WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// getColor returns a color based on priority
func (d *DiscordDispatcher) getColor(priority string) int {
	switch priority {
	case "high":
		return 0xFF0000 // Red
	case "normal":
		return 0x00FF00 // Green
	case "low":
		return 0x0000FF // Blue
	default:
		return 0x808080 // Gray
	}
}
