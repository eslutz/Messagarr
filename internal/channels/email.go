package channels

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
)

// EmailDispatcher sends notifications via SMTP
type EmailDispatcher struct {
	name   string
	config config.ChannelConfig
}

// NewEmailDispatcher creates a new email dispatcher
func NewEmailDispatcher(name string, cfg config.ChannelConfig) *EmailDispatcher {
	return &EmailDispatcher{
		name:   name,
		config: cfg,
	}
}

// Name returns the dispatcher name
func (e *EmailDispatcher) Name() string {
	return e.name
}

// Send sends an email notification
func (e *EmailDispatcher) Send(req *models.NotificationRequest) error {
	// Build email message
	subject := req.Title
	if subject == "" {
		subject = "Notification from Messagarr"
	}

	message := fmt.Sprintf("From: %s\r\n", e.config.From)
	message += fmt.Sprintf("To: %s\r\n", e.config.To)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "Content-Type: text/plain; charset=utf-8\r\n"
	message += "\r\n"
	message += req.Body

	// Add metadata if present
	if len(req.Metadata) > 0 {
		message += "\n\n--- Metadata ---\n"
		for k, v := range req.Metadata {
			message += fmt.Sprintf("%s: %s\n", k, v)
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
	defer client.Close()

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

	_, err = w.Write([]byte(message))
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}
