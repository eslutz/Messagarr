package channels

import (
	"log/slog"
	"sync"

	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
	"github.com/eslutz/Messagarr/internal/resilience"
)

// Dispatcher is the interface for notification channels
type Dispatcher interface {
	Send(req *models.NotificationRequest) error
	Name() string
}

// ChannelDispatcher manages multiple notification channels
type ChannelDispatcher struct {
	channels  map[string]Dispatcher
	retrier   *resilience.Retrier
	rateLimit *resilience.RateLimiter
}

// NewChannelDispatcher creates a new channel dispatcher
func NewChannelDispatcher(cfg *config.Config) *ChannelDispatcher {
	cd := &ChannelDispatcher{
		channels:  make(map[string]Dispatcher),
		retrier:   resilience.NewRetrier(),
		rateLimit: resilience.NewRateLimiter(10, 10), // 10 requests per second per channel
	}

	// Initialize channels based on config
	for name, channelCfg := range cfg.Channels {
		var dispatcher Dispatcher
		switch channelCfg.Type {
		case "smtp":
			dispatcher = NewEmailDispatcher(name, channelCfg)
		case "discord":
			dispatcher = NewDiscordDispatcher(name, channelCfg)
		case "slack":
			dispatcher = NewSlackDispatcher(name, channelCfg)
		case "teams":
			dispatcher = NewTeamsDispatcher(name, channelCfg)
		default:
			slog.Warn("Unknown channel type", "channel", name, "type", channelCfg.Type)
			continue
		}
		cd.channels[name] = dispatcher
	}

	return cd
}

// Dispatch sends a notification to multiple channels in parallel
func (cd *ChannelDispatcher) Dispatch(req *models.NotificationRequest, channelNames []string) map[string]models.Result {
	results := make(map[string]models.Result)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, channelName := range channelNames {
		channel, exists := cd.channels[channelName]
		if !exists {
			mu.Lock()
			results[channelName] = models.Result{
				Success: false,
				Error:   "Channel not found",
			}
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(name string, ch Dispatcher) {
			defer wg.Done()

			// Rate limiting
			cd.rateLimit.Wait(name)

			// Retry logic
			err := cd.retrier.Do(func() error {
				return ch.Send(req)
			})

			mu.Lock()
			if err != nil {
				slog.Error("Failed to send notification", "channel", name, "error", err)
				results[name] = models.Result{
					Success: false,
					Error:   err.Error(),
				}
			} else {
				slog.Info("Notification sent successfully", "channel", name)
				results[name] = models.Result{
					Success: true,
				}
			}
			mu.Unlock()
		}(channelName, channel)
	}

	wg.Wait()
	return results
}
