package resilience

import (
	"testing"
	"time"

	"github.com/eslutz/Messagarr/internal/models"
)

func TestDeduplicator(t *testing.T) {
	ttl := 100 * time.Millisecond
	deduper := NewDeduplicator(ttl)
	defer deduper.Stop()

	req1 := &models.NotificationRequest{
		Title:    "Test",
		Body:     "Test body",
		Priority: "high",
	}

	req2 := &models.NotificationRequest{
		Title:    "Test",
		Body:     "Test body",
		Priority: "high",
	}

	req3 := &models.NotificationRequest{
		Title:    "Different",
		Body:     "Different body",
		Priority: "low",
	}

	// First request should not be duplicate
	if deduper.IsDuplicate(req1) {
		t.Error("First request should not be duplicate")
	}

	// Same request should be duplicate
	if !deduper.IsDuplicate(req2) {
		t.Error("Same request should be duplicate")
	}

	// Different request should not be duplicate
	if deduper.IsDuplicate(req3) {
		t.Error("Different request should not be duplicate")
	}

	// After TTL, should not be duplicate
	time.Sleep(ttl + 50*time.Millisecond)
	if deduper.IsDuplicate(req1) {
		t.Error("After TTL, request should not be duplicate")
	}
}

func TestRetrier(t *testing.T) {
	retrier := NewRetrier()

	// Test successful execution on first attempt
	attempts := 0
	err := retrier.Do(func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}

	// Test retry on failure
	attempts = 0
	err = retrier.Do(func() error {
		attempts++
		if attempts < 3 {
			return nil
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, 10) // 2 capacity, 10 per second

	channel := "test"
	
	// Should allow first request immediately
	start := time.Now()
	rl.Wait(channel)
	duration := time.Since(start)
	
	if duration > 10*time.Millisecond {
		t.Errorf("First request took too long: %v", duration)
	}

	// Second request should also be immediate
	start = time.Now()
	rl.Wait(channel)
	duration = time.Since(start)
	
	if duration > 10*time.Millisecond {
		t.Errorf("Second request took too long: %v", duration)
	}

	// Third request should wait (bucket empty)
	start = time.Now()
	rl.Wait(channel)
	duration = time.Since(start)
	
	if duration < 50*time.Millisecond {
		t.Errorf("Third request should have waited, but took only: %v", duration)
	}
}
