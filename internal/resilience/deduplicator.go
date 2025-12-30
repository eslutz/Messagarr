package resilience

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/eslutz/Messagarr/internal/models"
)

// Deduplicator prevents duplicate notifications within a time window
type Deduplicator struct {
	mu      sync.RWMutex
	seen    map[string]time.Time
	ttl     time.Duration
	cleanup *time.Ticker
}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator(ttl time.Duration) *Deduplicator {
	d := &Deduplicator{
		seen:    make(map[string]time.Time),
		ttl:     ttl,
		cleanup: time.NewTicker(ttl),
	}

	// Background cleanup goroutine
	go d.cleanupLoop()

	return d
}

// IsDuplicate checks if a notification is a duplicate
func (d *Deduplicator) IsDuplicate(req *models.NotificationRequest) bool {
	hash := d.hash(req)

	d.mu.RLock()
	lastSeen, exists := d.seen[hash]
	d.mu.RUnlock()

	if exists && time.Since(lastSeen) < d.ttl {
		return true
	}

	d.mu.Lock()
	d.seen[hash] = time.Now()
	d.mu.Unlock()

	return false
}

// hash generates a SHA256 hash of the notification
func (d *Deduplicator) hash(req *models.NotificationRequest) string {
	// Create a normalized representation for hashing
	data, _ := json.Marshal(map[string]interface{}{
		"title":      req.Title,
		"body":       req.Body,
		"priority":   req.Priority,
		"service":    req.Service,
		"event_type": req.EventType,
	})

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// cleanupLoop periodically removes old entries
func (d *Deduplicator) cleanupLoop() {
	for range d.cleanup.C {
		d.mu.Lock()
		now := time.Now()
		for hash, lastSeen := range d.seen {
			if now.Sub(lastSeen) > d.ttl {
				delete(d.seen, hash)
			}
		}
		d.mu.Unlock()
	}
}

// Stop stops the deduplicator cleanup loop
func (d *Deduplicator) Stop() {
	d.cleanup.Stop()
}
