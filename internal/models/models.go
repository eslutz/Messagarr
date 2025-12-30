package models

import "time"

// NotificationRequest represents an incoming notification request
// @Description Notification request payload
type NotificationRequest struct {
	// Notification title/subject
	Title string `json:"title" example:"Deployment Complete"`
	// Notification body/message
	Body string `json:"body" example:"Application v1.2.3 deployed to production"`
	// Priority level (determines which channels receive the notification)
	Priority string `json:"priority" example:"high" enums:"high,normal,low"`
	// Name of the service sending the notification
	Service string `json:"service,omitempty" example:"deployment-service"`
	// Type of event triggering the notification
	EventType string `json:"event_type,omitempty" example:"deploy.success"`
	// Additional key-value metadata
	Metadata map[string]string `json:"metadata,omitempty"`
	// Explicit list of channels to send to (overrides priority-based routing)
	Channels []string `json:"channels,omitempty" example:"discord,email"`
}

// NotificationResponse represents the response to a notification request
// @Description Notification response with results per channel
type NotificationResponse struct {
	// Results per channel
	Results map[string]Result `json:"results"`
	// Summary message
	Message string `json:"message" example:"Notification sent"`
	// Total time taken to process the notification
	Duration string `json:"duration" example:"245ms"`
	// True if all channels succeeded
	Success bool `json:"success" example:"true"`
}

// Result represents the result of sending to a specific channel
// @Description Result of sending to a single channel
type Result struct {
	// Error message if failed
	Error string `json:"error,omitempty" example:"failed to connect to SMTP server"`
	// True if this channel succeeded
	Success bool `json:"success" example:"true"`
}

// HealthResponse represents the health check response
// @Description Health check response
type HealthResponse struct {
	// Health status
	Status string `json:"status" example:"ok"`
	// Current server time
	Timestamp time.Time `json:"timestamp" example:"2025-12-30T01:00:00Z"`
	// Application version
	Version string `json:"version" example:"1.0.0-dev.1"`
}

// ReadyResponse represents the readiness check response
// @Description Readiness check response
type ReadyResponse struct {
	// Current server time
	Timestamp time.Time `json:"timestamp" example:"2025-12-30T01:00:00Z"`
	// Individual readiness checks
	Checks []Check `json:"checks"`
	// True if service is ready
	Ready bool `json:"ready" example:"true"`
}

// Check represents an individual readiness check
// @Description Individual readiness check result
type Check struct {
	// Name of the check
	Name string `json:"name" example:"config"`
	// Status of the check
	Status string `json:"status" example:"ok" enums:"ok,error"`
	// Additional message (usually for errors)
	Message string `json:"message,omitempty" example:"Configuration loaded successfully"`
}

// MetricsResponse represents basic metrics
// @Description JSON metrics response
type MetricsResponse struct {
	// Number of successful notifications per channel
	ChannelStats map[string]int64 `json:"channel_stats"`
	// Service uptime
	Uptime string `json:"uptime" example:"24h30m15s"`
	// Total number of notifications sent
	TotalNotifications int64 `json:"total_notifications" example:"1234"`
	// Total number of failed notifications
	FailedNotifications int64 `json:"failed_notifications" example:"5"`
}
