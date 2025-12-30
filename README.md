# Messagarr

[![Workflow Status](https://github.com/eslutz/Messagarr/actions/workflows/release.yml/badge.svg)](https://github.com/eslutz/Messagarr/actions/workflows/release.yml)
[![Security Check](https://github.com/eslutz/Messagarr/actions/workflows/security.yml/badge.svg)](https://github.com/eslutz/Messagarr/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/eslutz/Messagarr)](https://goreportcard.com/report/github.com/eslutz/Messagarr)
[![License](https://img.shields.io/github/license/eslutz/Messagarr)](LICENSE)
[![Release](https://img.shields.io/github/v/release/eslutz/Messagarr?color=007ec6)](https://github.com/eslutz/Messagarr/releases/latest)

A lightweight, self-hosted notification aggregation service that routes messages to multiple channels (Email, Discord, Slack, Microsoft Teams) based on configurable priority groups. Built for the *arr stack with observability and reliability in mind.

## Features

- **Multi-Channel Support**: Email (SMTP), Discord, Slack, Microsoft Teams webhooks
- **Priority-Based Routing**: Route notifications to different channels based on priority (high/normal/low)
- **Config/Secret Separation**: YAML configuration with environment variable interpolation
- **Parallel Delivery**: Goroutine fan-out for fast, concurrent channel dispatch
- **Resilience & Reliability**: Exponential backoff retry, rate limiting, duplicate suppression
- **Full Observability**: Prometheus metrics for monitoring and alerting
- **Health & Readiness**: Kubernetes-compatible health check endpoints
- **Production Ready**: Non-root Docker container, graceful shutdown, comprehensive error handling

## Quick Start

### Docker Compose

See [docs/docker-compose.example.yml](docs/docker-compose.example.yml) for a complete setup with Prometheus and Grafana.

### Docker CLI

```bash
docker run -d \
  --name messagarr \
  -p 8080:8080 \
  -e SMTP_HOST=smtp.example.com \
  -e SMTP_USER=user@example.com \
  -e SMTP_PASSWORD=secret \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  -v $(pwd)/config:/app/config \
  --restart unless-stopped \
  ghcr.io/eslutz/messagarr:latest
```

## Configuration

All configuration is done via a YAML file with environment variable interpolation for secrets. An example configuration file and environment variables are available at [docs/.env.example](docs/.env.example).

Create `config/messagarr/config.yaml`:

```yaml
port: 8080
log_level: info
dedup_ttl: 5m

channels:
  email:
    type: smtp
    host: ${SMTP_HOST}
    port: 587
    user: ${SMTP_USER}
    password: ${SMTP_PASSWORD}
    from: ${SMTP_FROM}
    to: ${EMAIL_TO}
    
  discord-alerts:
    type: discord
    webhook_url: ${DISCORD_WEBHOOK_URL}
    
  slack-ops:
    type: slack
    webhook_url: ${SLACK_WEBHOOK_URL}
    
  teams-infra:
    type: teams
    webhook_url: ${TEAMS_WEBHOOK_URL}

priority_groups:
  high:
    - email
    - discord-alerts
  normal:
    - discord-alerts
    - slack-ops
  low:
    - slack-ops
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CONFIG_PATH` | `config/messagarr/config.yaml` | Path to configuration file |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `ENV` | `development` | Environment mode (set to `production` for JSON logs) |
| `SMTP_HOST` | *(required)* | SMTP server hostname |
| `SMTP_PORT` | `587` | SMTP server port |
| `SMTP_USER` | *(required)* | SMTP username |
| `SMTP_PASSWORD` | *(required)* | SMTP password |
| `SMTP_FROM` | *(required)* | Sender email address |
| `EMAIL_TO` | *(required)* | Recipient email address |
| `DISCORD_WEBHOOK_URL` | *(optional)* | Discord webhook URL |
| `SLACK_WEBHOOK_URL` | *(optional)* | Slack webhook URL |
| `TEAMS_WEBHOOK_URL` | *(optional)* | Microsoft Teams webhook URL |

## Architecture

```txt
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /notify
       ▼
┌─────────────────┐
│   HTTP Server   │
│   (API Layer)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  Deduplicator   │─────►│   Priority   │
│  (SHA256 Hash)  │      │   Router     │
└─────────────────┘      └──────┬───────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │  Channel Dispatcher   │
                    │  (Fan-out Goroutines) │
                    └───────┬───────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
   ┌────────┐         ┌──────────┐        ┌────────┐
   │ Email  │         │ Discord  │        │ Slack  │
   │ (SMTP) │         │(Webhook) │        │(Webhook)│
   └────────┘         └──────────┘        └────────┘
```

**How it works:**

1. Client sends a notification via `POST /notify` with title, body, and priority
2. Deduplicator checks for duplicate notifications within the configured TTL window
3. Priority router maps priority level to configured channel groups
4. Dispatcher sends to all channels in parallel using goroutines
5. Each channel implements retry with exponential backoff and rate limiting

## HTTP Endpoints

| Endpoint | Purpose | Response |
|----------|---------|----------|
| `POST /notify` | Send notification | JSON with results per channel |
| `GET /health` | Liveness probe | `200 OK` if running |
| `GET /ready` | Readiness probe | `200 OK` if configured and ready |
| `GET /status` | Diagnostics | JSON status snapshot |
| `GET /metrics` | Prometheus metrics | Metrics in OpenMetrics format |

### Endpoint Usage

- **/notify**: Send notifications with optional priority and channel override
- **/health**: Liveness probe (restart container if it fails)
- **/ready**: Readiness probe for Kubernetes/orchestrators
- **/status**: Manual debugging and monitoring (JSON format)
- **/metrics**: Prometheus scraper target

### Notify Request Example

```bash
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Deploy Complete",
    "body": "Application v1.2.3 deployed to production",
    "priority": "high",
    "service": "deployment-service",
    "event_type": "deploy.success",
    "metadata": {
      "version": "1.2.3",
      "environment": "production"
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Notification sent",
  "results": {
    "email": {"success": true},
    "discord-alerts": {"success": true}
  },
  "duration": "245ms"
}
```

## Prometheus Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `messagarr_notifications_total` | Counter | Total number of notifications sent |
| `messagarr_notifications_failed_total` | Counter | Total number of failed notifications |
| `messagarr_notifications_by_channel_total` | Counter | Notifications per channel (with `channel` and `success` labels) |
| `messagarr_notification_duration_seconds` | Histogram | Notification processing duration (with `priority` label) |
| `messagarr_http_requests_total` | Counter | HTTP requests (with `path`, `method`, `code` labels) |
| `messagarr_http_request_duration_seconds` | Histogram | HTTP request duration (with `path`, `method`, `code` labels) |
| `messagarr_uptime_seconds` | Gauge | Service uptime in seconds |

A Grafana dashboard is available at [docs/messagarr-grafana-dashboard.json](docs/messagarr-grafana-dashboard.json).

## Channel Setup

### Email (SMTP)

Supports STARTTLS for secure connections. For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833).

```yaml
channels:
  email:
    type: smtp
    host: smtp.gmail.com
    port: 587
    user: ${SMTP_USER}
    password: ${SMTP_PASSWORD}  # Use app password for Gmail
    from: alerts@example.com
    to: admin@example.com
```

### Discord

Create webhook: Server Settings → Integrations → Webhooks

```yaml
channels:
  discord:
    type: discord
    webhook_url: ${DISCORD_WEBHOOK_URL}
```

### Slack

Create webhook: Workspace Settings → Apps → Incoming Webhooks

```yaml
channels:
  slack:
    type: slack
    webhook_url: ${SLACK_WEBHOOK_URL}
```

### Microsoft Teams

Create webhook: Channel → Connectors → Incoming Webhook

```yaml
channels:
  teams:
    type: teams
    webhook_url: ${TEAMS_WEBHOOK_URL}
```

For detailed setup instructions, see [docs/channels.md](docs/channels.md).

## Development

### Building from Source

```bash
git clone https://github.com/eslutz/Messagarr.git
cd Messagarr
go build -o messagarr ./cmd/messagarr
```

### Running Tests

```bash
go test ./...                      # Run all tests
go test -v ./...                   # Run with verbose output
go test -cover ./...               # Run with coverage
```

### Docker Build

```bash
docker build -t messagarr:local .
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
