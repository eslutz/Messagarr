# Next Steps for Messagarr

This document outlines the next steps for using and enhancing Messagarr now that the core implementation is complete.

## Immediate Usage

### 1. Local Testing

```bash
# Clone the repository
git clone https://github.com/eslutz/Messagarr.git
cd Messagarr

# Set up environment variables
cp .env.example .env
# Edit .env with your actual credentials

# Run locally
make run

# Or build and run
make build
./messagarr
```

### 2. Docker Deployment

```bash
# Using Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f messagarr

# Test endpoints
curl http://localhost:8080/health
```

### 3. Configure Channels

Follow the guides in `docs/channels.md` to set up each channel:
- **Email**: Use Gmail or custom SMTP server
- **Discord**: Create webhook in server settings
- **Slack**: Create incoming webhook app
- **Teams**: Add incoming webhook connector

### 4. Send Test Notifications

```bash
# Test with curl
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Notification",
    "body": "Testing Messagarr notification system",
    "priority": "high"
  }'

# Test specific channels
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Direct Test",
    "body": "Testing specific channel",
    "channels": ["discord-alerts"]
  }'
```

## Integration with Existing Services

### Torrent-Services Integration

1. **Add to docker-compose.yml** (in Torrent-Services repo):
   ```yaml
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
       # Add your secrets here
     networks:
       - torrent-network
   ```

2. **Update healthcheck_utils.sh** (if exists):
   ```bash
   # Add Messagarr health check
   check_service "messagarr" "http://messagarr:8080/health"
   ```

3. **Configure *arr Apps**:
   - Go to Settings → Connect → Add Connection
   - Choose "Webhook"
   - URL: `http://messagarr:8080/notify`
   - Method: POST
   - Add JSON payload:
     ```json
     {
       "title": "{Movie.Title} Downloaded",
       "body": "Quality: {Quality}, Size: {Size}",
       "priority": "normal",
       "service": "radarr",
       "event_type": "download.complete"
     }
     ```

### Custom Application Integration

```go
// Example Go client
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type Notification struct {
    Title    string            `json:"title"`
    Body     string            `json:"body"`
    Priority string            `json:"priority"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

func sendNotification(title, body string) error {
    notif := Notification{
        Title:    title,
        Body:     body,
        Priority: "normal",
    }
    
    data, _ := json.Marshal(notif)
    resp, err := http.Post(
        "http://messagarr:8080/notify",
        "application/json",
        bytes.NewBuffer(data),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}
```

```python
# Example Python client
import requests

def send_notification(title, body, priority="normal"):
    payload = {
        "title": title,
        "body": body,
        "priority": priority
    }
    response = requests.post(
        "http://messagarr:8080/notify",
        json=payload
    )
    return response.json()
```

## Enhancements & Future Development

### High Priority

1. **Additional Channels**
   - [ ] SMS via Twilio
   - [ ] Pushover
   - [ ] Telegram
   - [ ] Generic Webhook

2. **Security**
   - [ ] API authentication (Bearer tokens)
   - [ ] Rate limiting per client
   - [ ] TLS/HTTPS support

3. **Observability**
   - [ ] Prometheus metrics export
   - [ ] OpenTelemetry tracing
   - [ ] Grafana dashboards

### Medium Priority

4. **Reliability**
   - [ ] Database-backed queue (PostgreSQL/Redis)
   - [ ] Message persistence
   - [ ] Dead letter queue
   - [ ] Circuit breaker improvements

5. **Features**
   - [ ] Template system for formatting
   - [ ] Notification scheduling
   - [ ] Conditional routing rules
   - [ ] Message threading/grouping

6. **Testing**
   - [ ] E2E tests with real channels
   - [ ] Load testing suite
   - [ ] Chaos engineering tests

### Low Priority

7. **UI/UX**
   - [ ] Web UI for testing
   - [ ] Configuration UI
   - [ ] Real-time dashboard
   - [ ] Mobile app (future)

8. **Advanced**
   - [ ] Multi-tenancy support
   - [ ] Message replay/audit log
   - [ ] A/B testing for messages
   - [ ] Analytics dashboard

## Development Workflow

### Setting Up Development Environment

```bash
# Clone and setup
git clone https://github.com/eslutz/Messagarr.git
cd Messagarr

# Install dependencies
make install

# Run tests
make test

# Run with live reload (using entr or similar)
find . -name "*.go" | entr -r go run ./cmd/messagarr
```

### Making Changes

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes
# ... edit files ...

# Run tests
make test

# Run linter
make lint

# Commit and push
git add .
git commit -m "Add new feature"
git push origin feature/new-feature

# Create PR on GitHub
```

### Adding a New Channel

1. Create `internal/channels/newchannel.go`:
   ```go
   package channels
   
   type NewChannelDispatcher struct {
       name   string
       config config.ChannelConfig
   }
   
   func NewNewChannelDispatcher(name string, cfg config.ChannelConfig) *NewChannelDispatcher {
       return &NewChannelDispatcher{name: name, config: cfg}
   }
   
   func (n *NewChannelDispatcher) Name() string {
       return n.name
   }
   
   func (n *NewChannelDispatcher) Send(req *models.NotificationRequest) error {
       // Implementation here
       return nil
   }
   ```

2. Update `internal/channels/dispatcher.go`:
   ```go
   case "newchannel":
       dispatcher = NewNewChannelDispatcher(name, channelCfg)
   ```

3. Update `internal/config/config.go` validation

4. Add tests in `internal/channels/newchannel_test.go`

5. Update documentation in `docs/channels.md`

## Monitoring & Operations

### Health Checks

```bash
# Liveness probe (load balancer)
curl http://localhost:8080/health

# Readiness probe (orchestrator)
curl http://localhost:8080/ready

# Metrics
curl http://localhost:8080/metrics
```

### Logs

```bash
# Docker Compose
docker-compose logs -f messagarr

# Kubernetes
kubectl logs -f deployment/messagarr

# Local
# Logs go to stdout, pipe to file if needed
./messagarr 2>&1 | tee messagarr.log
```

### Troubleshooting

Common issues and solutions:

1. **Notifications not sending**
   - Check logs for errors
   - Verify webhook URLs
   - Test webhooks manually
   - Check firewall rules

2. **High latency**
   - Check external service status
   - Review retry settings
   - Monitor rate limits
   - Consider adding caching

3. **Memory issues**
   - Review deduplication cache size
   - Check for goroutine leaks
   - Monitor with pprof

## Production Checklist

Before deploying to production:

- [ ] Configure all required channels
- [ ] Set appropriate log levels (info or warn)
- [ ] Enable JSON logging (ENV=production)
- [ ] Configure proper secrets management
- [ ] Set up monitoring/alerting
- [ ] Configure backup strategy (if using DB)
- [ ] Set resource limits (CPU/memory)
- [ ] Configure TLS certificates
- [ ] Set up log aggregation
- [ ] Document runbooks for operations

## Support & Resources

- **Documentation**: `/docs` directory
- **Issues**: https://github.com/eslutz/Messagarr/issues
- **Discussions**: GitHub Discussions
- **Examples**: Check `/examples` if added

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Update documentation
6. Submit a pull request

See `README.md` for detailed contribution guidelines.

---

**Current Status**: ✅ Production Ready

The core implementation is complete and ready for use. All features from the problem statement have been implemented and tested.
