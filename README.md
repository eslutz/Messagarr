# Messagarr

[![CI](https://github.com/eslutz/Messagarr/workflows/CI/badge.svg)](https://github.com/eslutz/Messagarr/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/eslutz/Messagarr)](https://goreportcard.com/report/github.com/eslutz/Messagarr)
[![License](https://img.shields.io/github/license/eslutz/Messagarr)](LICENSE)

A lightweight, self-hosted notification aggregation service that routes messages to multiple channels based on priority. Built in Go with security and reliability in mind.

## Features

- 🔒 **Config/Secret Separation** - Configuration via YAML with environment variable interpolation
- 🎯 **Priority-Based Routing** - Route notifications to different channels based on priority (high/normal/low)
- 📢 **Multi-Channel Support** - Email (SMTP), Discord, Slack, Microsoft Teams
- ⚡ **Parallel Delivery** - Goroutine fan-out for fast, concurrent channel dispatch
- 🔄 **Resilience** - Exponential backoff retry, rate limiting, and duplicate suppression
- 📊 **Observability** - Health checks, readiness probes, and metrics endpoints
- 🧪 **Well Tested** - Comprehensive test suite with 67%+ coverage
- 🐳 **Docker Ready** - Multi-stage Dockerfile with non-root user

## Quick Start

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  -e SMTP_HOST=smtp.example.com \
  -e SMTP_PORT=587 \
  -e SMTP_USER=user@example.com \
  -e SMTP_PASSWORD=secret \
  -e DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... \
  eslutz/messagarr:latest
```

### From Source

```bash
# Clone repository
git clone https://github.com/eslutz/Messagarr.git
cd Messagarr

# Build
go build -o messagarr ./cmd/messagarr

# Run
./messagarr
```

## Configuration

Create `config/messagarr/config.yaml`:

```yaml
port: 8080
log_level: info
dedup_ttl: 5m  # Duplicate suppression window

channels:
  email:
    type: smtp
    host: ${SMTP_HOST}
    port: ${SMTP_PORT}
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

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_PATH` | Path to config file | `config/messagarr/config.yaml` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `ENV` | Environment mode (production for JSON logs) | `development` |
| `DEDUP_TTL` | Duplicate suppression TTL | From config or `5m` |

## API Reference

### POST /notify

Send a notification to configured channels.

**Request:**
```json
{
  "title": "Deploy Complete",
  "body": "Application deployed successfully to production",
  "priority": "normal",
  "service": "deployment-service",
  "event_type": "deploy.success",
  "metadata": {
    "version": "1.2.3",
    "environment": "production"
  },
  "channels": ["discord-alerts"]  // Optional: override priority groups
}
```

**Response:**
```json
{
  "success": true,
  "message": "Notification sent",
  "results": {
    "discord-alerts": {
      "success": true
    }
  },
  "duration": "245ms"
}
```

### GET /health

Health check endpoint for load balancers.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-12-30T01:00:00Z",
  "version": "1.0.0"
}
```

### GET /ready

Readiness probe for orchestrators.

**Response:**
```json
{
  "ready": true,
  "timestamp": "2025-12-30T01:00:00Z",
  "checks": [
    {"name": "config", "status": "ok"},
    {"name": "channels", "status": "ok"}
  ]
}
```

### GET /metrics

Prometheus metrics endpoint for monitoring and alerting.

**Response:** Prometheus text format with the following metrics:
- `messagarr_notifications_total` - Total notifications sent
- `messagarr_notifications_failed_total` - Total failed notifications
- `messagarr_notifications_by_channel_total{channel, success}` - Notifications per channel
- `messagarr_notification_duration_seconds{priority}` - Notification processing duration histogram
- `messagarr_http_requests_total{path, method, code}` - HTTP request counter
- `messagarr_http_request_duration_seconds{path, method, code}` - HTTP request duration histogram
- `messagarr_uptime_seconds` - Service uptime

See [docs/messagarr-grafana-dashboard.json](docs/messagarr-grafana-dashboard.json) for a ready-to-use Grafana dashboard.

### GET /status

JSON status endpoint (legacy).

**Response:**
```json
{
  "total_notifications": 1234,
  "failed_notifications": 5,
  "channel_stats": {
    "discord-alerts": 800,
    "email": 400,
    "slack-ops": 34
  },
  "uptime": "24h30m15s"
}
```

## Monitoring

### Prometheus + Grafana

Messagarr exposes Prometheus metrics on the `/metrics` endpoint. See [docs/docker-compose.example.yml](docs/docker-compose.example.yml) for a complete monitoring stack setup with Prometheus and Grafana.

**Quick Start:**
```bash
cd docs
cp docker-compose.example.yml docker-compose.yml
cp .env.example .env
# Edit .env with your values
docker-compose up -d
```

Access Grafana at http://localhost:3000 (default credentials: admin/admin) and import the dashboard from `docs/messagarr-grafana-dashboard.json`.

**Available Metrics:**
- Request rates and durations
- Notification success/failure rates
- Per-channel statistics
- Service uptime

## Channel Configuration

### Email (SMTP)

```yaml
channels:
  email:
    type: smtp
    host: smtp.gmail.com
    port: 587
    user: ${SMTP_USER}
    password: ${SMTP_PASSWORD}
    from: alerts@example.com
    to: admin@example.com
```

Supports STARTTLS for secure connections.

### Discord

```yaml
channels:
  discord:
    type: discord
    webhook_url: ${DISCORD_WEBHOOK_URL}
```

Create webhook: Server Settings → Integrations → Webhooks

### Slack

```yaml
channels:
  slack:
    type: slack
    webhook_url: ${SLACK_WEBHOOK_URL}
```

Create webhook: Workspace Settings → Apps → Incoming Webhooks

### Microsoft Teams

```yaml
channels:
  teams:
    type: teams
    webhook_url: ${TEAMS_WEBHOOK_URL}
```

Create webhook: Channel → Connectors → Incoming Webhook

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /notify
       v
┌─────────────────┐
│   HTTP Server   │
│   (API Layer)   │
└────────┬────────┘
         │
         v
┌─────────────────┐      ┌──────────────┐
│  Deduplicator   │─────>│ Priority     │
│  (SHA256 Hash)  │      │ Router       │
└─────────────────┘      └──────┬───────┘
                                │
                                v
                    ┌───────────────────────┐
                    │  Channel Dispatcher   │
                    │  (Fan-out Goroutines) │
                    └───────┬───────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        v                   v                   v
   ┌────────┐         ┌──────────┐        ┌────────┐
   │ Email  │         │ Discord  │        │ Slack  │
   │ (SMTP) │         │(Webhook) │        │(Webhook)│
   └────────┘         └──────────┘        └────────┘
```

### Key Components

- **Config Layer** - YAML parsing with environment variable interpolation
- **API Layer** - HTTP server with JSON endpoints
- **Dispatcher** - Parallel notification delivery with goroutines
- **Resilience** - Retry logic, rate limiting, deduplication
- **Channels** - Pluggable notification channel implementations

## Development

### Prerequisites

- Go 1.24+
- Docker (optional)

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./internal/channels
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Building

```bash
# Build binary
go build -o messagarr ./cmd/messagarr

# Build Docker image
docker build -t messagarr:local .

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o messagarr-linux-amd64 ./cmd/messagarr
GOOS=darwin GOARCH=arm64 go build -o messagarr-darwin-arm64 ./cmd/messagarr
```

## Integration with Torrent-Services

Add to your `docker-compose.yml`:

```yaml
services:
  messagarr:
    image: eslutz/messagarr:latest
    container_name: messagarr
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config/messagarr:/app/config/messagarr
    environment:
      - LOG_LEVEL=info
      - ENV=production
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT}
      - SMTP_USER=${SMTP_USER}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
      - SMTP_FROM=${SMTP_FROM}
      - EMAIL_TO=${EMAIL_TO}
      - DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL}
      - SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
    networks:
      - torrent-network
```

Then use from other services:

```bash
# Send notification
curl -X POST http://messagarr:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Download Complete",
    "body": "Movie.2025.1080p.WEB.mkv has finished downloading",
    "priority": "normal",
    "service": "radarr",
    "event_type": "download.complete"
  }'
```

## Roadmap

- [ ] SMS support (Twilio, AWS SNS)
- [ ] Pushover, Pushbullet support
- [ ] Template system for notification formatting
- [ ] Web UI for testing and monitoring
- [ ] Prometheus metrics export
- [ ] Database-backed queue for reliability
- [ ] Authentication/API keys

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for new functionality
4. Ensure tests pass (`go test ./...`)
5. Commit changes (`git commit -am 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [Notifiarr](https://github.com/Notifiarr/notifiarr) and [Apprise](https://github.com/caronc/apprise)
- Built to integrate with the *arr ecosystem (Sonarr, Radarr, etc.)
- Uses patterns from [Forwardarr](https://github.com/eslutz/Forwardarr) and [Torarr](https://github.com/eslutz/Torarr)
