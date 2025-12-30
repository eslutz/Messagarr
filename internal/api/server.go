package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/eslutz/Messagarr/internal/channels"
	"github.com/eslutz/Messagarr/internal/config"
	"github.com/eslutz/Messagarr/internal/models"
	"github.com/eslutz/Messagarr/internal/resilience"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	dispatcher *channels.ChannelDispatcher
	deduper    *resilience.Deduplicator
	metrics    *Metrics
	startTime  time.Time
}

// Metrics holds server metrics
type Metrics struct {
	mu                  sync.RWMutex
	totalNotifications  int64
	failedNotifications int64
	channelStats        map[string]int64
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config:     cfg,
		dispatcher: channels.NewChannelDispatcher(cfg),
		deduper:    resilience.NewDeduplicator(cfg.DedupTTL),
		metrics: &Metrics{
			channelStats: make(map[string]int64),
		},
		startTime: time.Now(),
	}
}

// Router returns the HTTP router
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/notify", s.handleNotify)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	
	return s.loggingMiddleware(mux)
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		
		next.ServeHTTP(w, r)
		
		slog.Debug("HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

// handleNotify handles POST /notify
func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	var req models.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" && req.Body == "" {
		http.Error(w, "Title or body is required", http.StatusBadRequest)
		return
	}

	// Check for duplicate
	if s.deduper.IsDuplicate(&req) {
		slog.Info("Duplicate notification suppressed", "title", req.Title)
		resp := models.NotificationResponse{
			Success:  true,
			Message:  "Notification suppressed (duplicate)",
			Results:  make(map[string]models.Result),
			Duration: time.Since(start).String(),
		}
		s.sendJSON(w, http.StatusOK, resp)
		return
	}

	// Determine channels to use
	channelNames := s.determineChannels(&req)
	if len(channelNames) == 0 {
		http.Error(w, "No channels available for notification", http.StatusBadRequest)
		return
	}

	// Dispatch to channels
	results := s.dispatcher.Dispatch(&req, channelNames)

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.totalNotifications++
	allSuccess := true
	for channelName, result := range results {
		if result.Success {
			s.metrics.channelStats[channelName]++
		} else {
			allSuccess = false
		}
	}
	if !allSuccess {
		s.metrics.failedNotifications++
	}
	s.metrics.mu.Unlock()

	// Build response
	resp := models.NotificationResponse{
		Success:  allSuccess,
		Message:  "Notification sent",
		Results:  results,
		Duration: time.Since(start).String(),
	}
	if !allSuccess {
		resp.Message = "Notification partially failed"
	}

	statusCode := http.StatusOK
	if !allSuccess {
		statusCode = http.StatusMultiStatus
	}

	s.sendJSON(w, statusCode, resp)
}

// handleHealth handles GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	s.sendJSON(w, http.StatusOK, resp)
}

// handleReady handles GET /ready
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checks := []models.Check{
		{
			Name:   "config",
			Status: "ok",
		},
		{
			Name:   "channels",
			Status: "ok",
		},
	}

	resp := models.ReadyResponse{
		Ready:     true,
		Timestamp: time.Now(),
		Checks:    checks,
	}
	s.sendJSON(w, http.StatusOK, resp)
}

// handleMetrics handles GET /metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metrics.mu.RLock()
	channelStats := make(map[string]int64)
	for k, v := range s.metrics.channelStats {
		channelStats[k] = v
	}
	totalNotifications := s.metrics.totalNotifications
	failedNotifications := s.metrics.failedNotifications
	s.metrics.mu.RUnlock()

	resp := models.MetricsResponse{
		TotalNotifications:  totalNotifications,
		FailedNotifications: failedNotifications,
		ChannelStats:        channelStats,
		Uptime:              time.Since(s.startTime).String(),
	}
	s.sendJSON(w, http.StatusOK, resp)
}

// determineChannels determines which channels to use for a notification
func (s *Server) determineChannels(req *models.NotificationRequest) []string {
	// If channels explicitly specified, use those
	if len(req.Channels) > 0 {
		return req.Channels
	}

	// Otherwise use priority groups
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	if channels, exists := s.config.PriorityGroups[priority]; exists {
		return channels
	}

	// Fallback to all channels
	allChannels := make([]string, 0, len(s.config.Channels))
	for name := range s.config.Channels {
		allChannels = append(allChannels, name)
	}
	return allChannels
}

// sendJSON sends a JSON response
func (s *Server) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}
