package models

import "time"

// NotificationRequest represents an incoming notification request
type NotificationRequest struct {
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Priority  string            `json:"priority"`
	Service   string            `json:"service,omitempty"`
	EventType string            `json:"event_type,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Channels  []string          `json:"channels,omitempty"`
}

// NotificationResponse represents the response to a notification request
type NotificationResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Results  map[string]Result `json:"results"`
	Duration string            `json:"duration"`
}

// Result represents the result of sending to a specific channel
type Result struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ReadyResponse represents the readiness check response
type ReadyResponse struct {
	Ready     bool      `json:"ready"`
	Timestamp time.Time `json:"timestamp"`
	Checks    []Check   `json:"checks"`
}

// Check represents an individual readiness check
type Check struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// MetricsResponse represents basic metrics
type MetricsResponse struct {
	TotalNotifications int64             `json:"total_notifications"`
	FailedNotifications int64            `json:"failed_notifications"`
	ChannelStats       map[string]int64  `json:"channel_stats"`
	Uptime             string            `json:"uptime"`
}
