# Messagarr Architecture

## Overview

Messagarr is designed as a lightweight notification aggregation service that separates configuration from secrets, routes notifications based on priority, and delivers messages to multiple channels in parallel.

## System Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        External Services                      │
│  ┌──────┐  ┌─────────┐  ┌───────┐  ┌────────┐              │
│  │ SMTP │  │ Discord │  │ Slack │  │ Teams  │              │
│  └───┬──┘  └────┬────┘  └───┬───┘  └───┬────┘              │
└──────┼──────────┼───────────┼──────────┼────────────────────┘
       │          │            │          │
       │          │            │          │
┌──────┼──────────┼───────────┼──────────┼────────────────────┐
│      │          │            │          │                    │
│   ┌──▼──────────▼────────────▼──────────▼─────┐             │
│   │          Channel Dispatchers              │             │
│   │  ┌──────┐ ┌────────┐ ┌──────┐ ┌────────┐│             │
│   │  │Email │ │Discord │ │Slack │ │ Teams  ││             │
│   │  └──────┘ └────────┘ └──────┘ └────────┘│             │
│   └──────────────────┬─────────────────────────┘             │
│                      │                                       │
│   ┌──────────────────▼─────────────────────────┐             │
│   │         Channel Dispatcher Manager         │             │
│   │  - Goroutine fan-out                       │             │
│   │  - Retry logic (exponential backoff)       │             │
│   │  - Rate limiting (token bucket)            │             │
│   │  - Circuit breakers                        │             │
│   └──────────────────┬─────────────────────────┘             │
│                      │                                       │
│   ┌──────────────────▼─────────────────────────┐             │
│   │         Request Processing Layer           │             │
│   │  - Deduplication (SHA256, 5min TTL)        │             │
│   │  - Priority routing                        │             │
│   │  - Channel selection                       │             │
│   │  - Metrics collection                      │             │
│   └──────────────────┬─────────────────────────┘             │
│                      │                                       │
│   ┌──────────────────▼─────────────────────────┐             │
│   │            HTTP API Layer                  │             │
│   │  POST /notify    - Send notification       │             │
│   │  GET  /health    - Health check            │             │
│   │  GET  /ready     - Readiness probe         │             │
│   │  GET  /metrics   - Metrics endpoint        │             │
│   └──────────────────┬─────────────────────────┘             │
│                      │                                       │
└──────────────────────┼───────────────────────────────────────┘
                       │
                  ┌────▼─────┐
                  │  Clients │
                  └──────────┘
```

## Component Details

### 1. HTTP API Layer (`internal/api`)

**Purpose**: Expose REST endpoints for notification submission and service monitoring.

**Components**:
- `Server` - HTTP server with routing and middleware
- Request validation
- JSON serialization/deserialization
- Logging middleware

**Endpoints**:
- `POST /notify` - Accept notification requests
- `GET /health` - Liveness probe for load balancers
- `GET /ready` - Readiness probe for orchestrators
- `GET /metrics` - Prometheus-compatible metrics

**Design Patterns**:
- Middleware pattern for logging
- Dependency injection for testability
- Graceful shutdown support

### 2. Request Processing Layer (`internal/api`, `internal/resilience`)

**Purpose**: Process notifications, apply deduplication, and route to appropriate channels.

**Components**:
- **Deduplicator** - Prevents duplicate notifications within a time window
  - Uses SHA256 hashing of notification content
  - In-memory cache with TTL cleanup
  - Configurable time window (default 5 minutes)
  
- **Priority Router** - Maps priority levels to channel groups
  - `high` → critical channels (email, SMS)
  - `normal` → standard channels (Discord, Slack)
  - `low` → logging channels
  
- **Metrics Collector** - Tracks notification statistics
  - Total notifications sent
  - Failed notifications
  - Per-channel success rates
  - Uptime tracking

### 3. Channel Dispatcher (`internal/channels`)

**Purpose**: Deliver notifications to external services with resilience.

**Components**:
- **Dispatcher Interface** - Common contract for all channels
  ```go
  type Dispatcher interface {
      Send(req *NotificationRequest) error
      Name() string
  }
  ```

- **Channel Implementations**:
  - **Email** - SMTP with STARTTLS support
  - **Discord** - Webhook API with rich embeds
  - **Slack** - Webhook API with block kit
  - **Teams** - Webhook API with adaptive cards

- **Parallel Delivery**:
  - Goroutine fan-out pattern
  - WaitGroup for synchronization
  - Mutex for result aggregation

### 4. Resilience Layer (`internal/resilience`)

**Purpose**: Ensure reliable delivery with retry, rate limiting, and circuit breaking.

**Components**:

#### Retrier (Exponential Backoff)
```go
attempts:  1, 2, 3, 4
delays:    1s, 2s, 4s, 8s (capped at 30s)
```

- Maximum 3 retries
- Base delay: 1 second
- Exponential growth: 2^attempt
- Maximum delay cap: 30 seconds

#### Rate Limiter (Token Bucket)
- Per-channel rate limiting
- Default: 10 requests/second per channel
- Token refill based on elapsed time
- Waiting mechanism when bucket is empty

#### Circuit Breaker (Future)
- Open: Stop requests after N failures
- Half-Open: Allow test requests
- Closed: Normal operation
- Configurable failure threshold

### 5. Configuration Layer (`internal/config`)

**Purpose**: Load and validate configuration with secret separation.

**Features**:
- YAML configuration file
- Environment variable interpolation (`${VAR}`)
- Default value handling
- Comprehensive validation
- Hot reload capability (future)

**Configuration Schema**:
```yaml
port: 8080
log_level: info
dedup_ttl: 5m

channels:
  <name>:
    type: smtp|discord|slack|teams
    # Type-specific fields

priority_groups:
  high: [<channel-names>]
  normal: [<channel-names>]
  low: [<channel-names>]
```

### 6. Logging Layer (`internal/logger`)

**Purpose**: Structured logging with configurable levels and formats.

**Features**:
- Uses `log/slog` (Go 1.21+)
- JSON format in production
- Text format in development
- Context-aware logging
- Log level: debug, info, warn, error

**Logged Information**:
- HTTP requests/responses
- Channel dispatch attempts
- Errors and failures
- Performance metrics

## Data Flow

### 1. Notification Request
```
Client → POST /notify → API Server
```

### 2. Validation
```
API Server → Validate JSON → Check required fields
```

### 3. Deduplication
```
Request → Hash → Check cache → Mark as seen
```

### 4. Channel Selection
```
Priority → Priority Group → Channel List
OR
Explicit channels → Use provided list
```

### 5. Parallel Dispatch
```
For each channel:
  ├─ Rate Limit Check
  ├─ Retry Loop (exponential backoff)
  │   └─ Channel.Send()
  └─ Collect Result
```

### 6. Response
```
Aggregate Results → JSON Response → Client
```

## Security Considerations

1. **Secret Management**
   - No secrets in config files
   - Environment variable interpolation
   - Secrets passed at runtime

2. **Input Validation**
   - JSON schema validation
   - Required field checks
   - Content sanitization

3. **Rate Limiting**
   - Per-channel rate limits
   - Prevents resource exhaustion
   - Protects external services

4. **Error Handling**
   - No sensitive data in errors
   - Generic error messages to clients
   - Detailed logging for debugging

5. **Docker Security**
   - Non-root user
   - Minimal base image (Alpine)
   - No unnecessary capabilities

## Performance Characteristics

### Throughput
- **Parallel Dispatch**: O(1) time complexity per notification
- **Concurrent Requests**: Limited by Go runtime and memory
- **Expected**: 100-1000 notifications/second on modest hardware

### Latency
- **Deduplication**: O(1) hash lookup
- **Priority Routing**: O(1) map lookup
- **Channel Dispatch**: Network-bound, parallelized
- **Typical P95**: < 500ms (depends on external services)

### Memory
- **Deduplication Cache**: O(N) where N = notifications in TTL window
- **Configuration**: O(1) after load
- **Per Request**: ~1KB per notification
- **Expected**: < 100MB for typical workloads

### Scalability
- **Vertical**: Single instance handles hundreds of req/sec
- **Horizontal**: Stateless design allows multiple instances
- **Bottleneck**: External service rate limits

## Testing Strategy

### Unit Tests
- Configuration parsing and validation
- Channel dispatcher mocking
- Resilience mechanisms (retry, rate limit)
- Priority routing logic

### Integration Tests
- HTTP API endpoints
- Full request/response cycle
- Mock external services (httptest)
- Error scenarios

### Load Tests (Future)
- Concurrent request handling
- Rate limit behavior
- Memory usage under load
- Latency percentiles

## Monitoring & Observability

### Metrics
- Total notifications sent
- Failed notifications
- Per-channel statistics
- Uptime

### Health Checks
- `/health` - Always returns OK if server is running
- `/ready` - Checks configuration and channels

### Logging
- Structured JSON logs in production
- Request IDs for tracing
- Timing information
- Error details

## Future Enhancements

1. **Additional Channels**
   - SMS (Twilio, AWS SNS)
   - Push notifications (Pushover, Pushbullet)
   - Webhook (generic HTTP POST)

2. **Advanced Features**
   - Template system for formatting
   - Notification scheduling
   - Delivery confirmation
   - Retry queue with persistence

3. **Observability**
   - Prometheus metrics export
   - OpenTelemetry tracing
   - ELK/Grafana dashboards

4. **Reliability**
   - Database-backed queue
   - Event replay
   - Dead letter queue

5. **Security**
   - API authentication
   - Rate limiting per client
   - TLS/mTLS support
