package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

// SlackDispatcher sends notifications to Slack via webhooks
type SlackDispatcher struct {
	config *config.ChannelConfig
	name   string
}

// NewSlackDispatcher creates a new Slack dispatcher
func NewSlackDispatcher(name string, cfg *config.ChannelConfig) *SlackDispatcher {
	return &SlackDispatcher{
		config: cfg,
		name:   name,
	}
}

// Name returns the dispatcher name
func (s *SlackDispatcher) Name() string {
	return s.name
}

// SlackWebhook represents a Slack webhook payload
type SlackWebhook struct {
	Text        string       `json:"text,omitempty"`
	Blocks      []SlackBlock `json:"blocks,omitempty"`
	Attachments []any        `json:"attachments,omitempty"`
}

// SlackBlock represents a Slack block
type SlackBlock struct {
	Extra  map[string]any `json:"-"`
	Text   *SlackText     `json:"text,omitempty"`
	Type   string         `json:"type"`
	Fields []SlackText    `json:"fields,omitempty"`
}

// SlackText represents Slack text
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Send sends a Slack notification
func (s *SlackDispatcher) Send(req *models.NotificationRequest) error {
	blocks := []SlackBlock{}

	// Add title as header
	if req.Title != "" {
		blocks = append(blocks, SlackBlock{
			Type: "header",
			Text: &SlackText{
				Type: "plain_text",
				Text: req.Title,
			},
		})
	}

	// Add body as section
	if req.Body != "" {
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: req.Body,
			},
		})
	}

	// Add metadata as fields
	if len(req.Metadata) > 0 || req.Service != "" || req.EventType != "" {
		fields := []SlackText{}

		if req.Service != "" {
			fields = append(fields, SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Service:*\n%s", req.Service),
			})
		}
		if req.EventType != "" {
			fields = append(fields, SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Event Type:*\n%s", req.EventType),
			})
		}

		for k, v := range req.Metadata {
			fields = append(fields, SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*%s:*\n%s", k, v),
			})
		}

		if len(fields) > 0 {
			blocks = append(blocks, SlackBlock{
				Type:   "section",
				Fields: fields,
			})
		}
	}

	webhook := SlackWebhook{
		Blocks: blocks,
	}

	payload, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack webhook: %w", err)
	}

	resp, err := http.Post(s.config.WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
