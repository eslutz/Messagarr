package channels

import (
	"crypto/tls"
	"fmt"
	"html"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

// multiSpaceRegex matches multiple consecutive spaces
var multiSpaceRegex = regexp.MustCompile(`\s+`)

// sanitize cleans a string for safe use in email headers and body
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")        // Remove null bytes
	s = strings.ReplaceAll(s, "\r", " ")         // Replace carriage returns with space
	s = strings.ReplaceAll(s, "\n", " ")         // Replace newlines with space
	s = multiSpaceRegex.ReplaceAllString(s, " ") // Collapse multiple spaces into one
	s = strings.TrimSpace(s)                     // Remove leading and trailing whitespace
	return s
}

// sanitizeAndEscapeHTML sanitizes for header safety and escapes HTML metacharacters
func sanitizeAndEscapeHTML(s string) string {
	return html.EscapeString(sanitize(s))
}

// EmailDispatcher sends notifications via SMTP
type EmailDispatcher struct {
	config *config.ChannelConfig
	name   string
}

// NewEmailDispatcher creates a new email dispatcher
func NewEmailDispatcher(name string, cfg *config.ChannelConfig) *EmailDispatcher {
	return &EmailDispatcher{
		config: cfg,
		name:   name,
	}
}

// Name returns the dispatcher name
func (e *EmailDispatcher) Name() string {
	return e.name
}

// Send sends an email notification
func (e *EmailDispatcher) Send(req *models.NotificationRequest) error {
	// Build email message
	subject := sanitize(req.Title)
	if subject == "" {
		subject = "Notification from Messagarr"
	}

	var message strings.Builder
	fmt.Fprintf(&message, "From: %s\r\n", e.config.From)
	fmt.Fprintf(&message, "To: %s\r\n", e.config.To)
	fmt.Fprintf(&message, "Subject: %s\r\n", subject)
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(sanitizeAndEscapeHTML(req.Body))

	// Add metadata if present
	if len(req.Metadata) > 0 {
		message.WriteString("\n\n--- Metadata ---\n")
		for k, v := range req.Metadata {
			fmt.Fprintf(&message, "%s: %s\n", sanitize(k), sanitizeAndEscapeHTML(v))
		}
	}

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	// Attempt TLS connection
	tlsConfig := &tls.Config{
		ServerName: e.config.Host,
	}

	// Try STARTTLS
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func() { _ = client.Close() }()

	// STARTTLS if available
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate if credentials provided
	if e.config.User != "" && e.config.Password != "" {
		auth := smtp.PlainAuth("", e.config.User, e.config.Password, e.config.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Send email
	if err = client.Mail(e.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(e.config.To); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write([]byte(message.String()))
	if err != nil {
		_ = w.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}
