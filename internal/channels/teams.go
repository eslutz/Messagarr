package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

// TeamsDispatcher sends notifications to Microsoft Teams via webhooks
type TeamsDispatcher struct {
	name   string
	config config.ChannelConfig
}

// NewTeamsDispatcher creates a new Teams dispatcher
func NewTeamsDispatcher(name string, cfg config.ChannelConfig) *TeamsDispatcher {
	return &TeamsDispatcher{
		name:   name,
		config: cfg,
	}
}

// Name returns the dispatcher name
func (t *TeamsDispatcher) Name() string {
	return t.name
}

// TeamsWebhook represents a Microsoft Teams webhook payload (Adaptive Card)
type TeamsWebhook struct {
	Type        string                   `json:"type"`
	Attachments []TeamsAttachment        `json:"attachments"`
}

// TeamsAttachment represents a Teams attachment
type TeamsAttachment struct {
	ContentType string      `json:"contentType"`
	Content     TeamsCard   `json:"content"`
}

// TeamsCard represents an adaptive card
type TeamsCard struct {
	Schema  string           `json:"$schema"`
	Type    string           `json:"type"`
	Version string           `json:"version"`
	Body    []map[string]any `json:"body"`
}

// Send sends a Teams notification
func (t *TeamsDispatcher) Send(req *models.NotificationRequest) error {
	body := []map[string]any{}

	// Add title
	if req.Title != "" {
		body = append(body, map[string]any{
			"type":   "TextBlock",
			"text":   req.Title,
			"weight": "Bolder",
			"size":   "Large",
		})
	}

	// Add body text
	if req.Body != "" {
		body = append(body, map[string]any{
			"type": "TextBlock",
			"text": req.Body,
			"wrap": true,
		})
	}

	// Add metadata as facts
	facts := []map[string]string{}
	if req.Service != "" {
		facts = append(facts, map[string]string{
			"title": "Service",
			"value": req.Service,
		})
	}
	if req.EventType != "" {
		facts = append(facts, map[string]string{
			"title": "Event Type",
			"value": req.EventType,
		})
	}
	if req.Priority != "" {
		facts = append(facts, map[string]string{
			"title": "Priority",
			"value": req.Priority,
		})
	}
	for k, v := range req.Metadata {
		facts = append(facts, map[string]string{
			"title": k,
			"value": v,
		})
	}

	if len(facts) > 0 {
		body = append(body, map[string]any{
			"type":  "FactSet",
			"facts": facts,
		})
	}

	card := TeamsCard{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.2",
		Body:    body,
	}

	webhook := TeamsWebhook{
		Type: "message",
		Attachments: []TeamsAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     card,
			},
		},
	}

	payload, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams webhook: %w", err)
	}

	resp, err := http.Post(t.config.WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Teams webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Teams webhook returned status %d", resp.StatusCode)
	}

	return nil
}
