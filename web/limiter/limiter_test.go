package limiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	tests := []struct {
		name           string
		maxTokens      float64
		refillRate     float64
		interval       time.Duration
		requests       int
		sleepIntervals []time.Duration
		wantResults    []bool
	}{
		{
			name:           "Single request with filled bucket",
			maxTokens:      5,
			refillRate:     1,
			interval:       time.Second,
			requests:       1,
			sleepIntervals: []time.Duration{0},
			wantResults:    []bool{true},
		},
		{
			name:           "Multiple requests with filled bucket",
			maxTokens:      5,
			refillRate:     1,
			interval:       time.Second,
			requests:       6,
			sleepIntervals: []time.Duration{0, 0, 0, 0, 0, 0},
			wantResults:    []bool{true, true, true, true, true, false},
		},
		{
			name:           "Refill after depletion",
			maxTokens:      2,
			refillRate:     1,
			interval:       time.Millisecond * 50,
			requests:       4,
			sleepIntervals: []time.Duration{0, 0, time.Millisecond * 100, 0},
			wantResults:    []bool{true, true, true, true},
		},
		{
			name:           "Partial refill after depletion",
			maxTokens:      3,
			refillRate:     1,
			interval:       time.Millisecond * 50,
			requests:       5,
			sleepIntervals: []time.Duration{0, 0, 0, time.Millisecond * 10, time.Millisecond * 50},
			wantResults:    []bool{true, true, true, false, true},
		},
		{
			name:           "Refill up to max capacity",
			maxTokens:      2,
			refillRate:     1,
			interval:       time.Millisecond * 50,
			requests:       3,
			sleepIntervals: []time.Duration{0, 0, time.Millisecond * 50},
			wantResults:    []bool{true, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := NewTokenBucket(tt.maxTokens, tt.refillRate, tt.interval)

			for i := 0; i < tt.requests; i++ {
				if i > 0 && tt.sleepIntervals[i] > 0 {
					time.Sleep(tt.sleepIntervals[i])
				}

				got := tb.Allow()
				if want := tt.wantResults[i]; got != want {
					t.Errorf("request %d: got = %v, want %v", i, got, want)
				}
			}
		})
	}
}

func TestRateLimiter_GetBucket(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		burst    float64
		interval time.Duration
		items    []int32
		wantSame []bool // whether the same bucket should be returned for consecutive calls with the same item
		excluded map[int32]struct{}
	}{
		{
			name:     "Single item",
			rate:     1,
			burst:    5,
			interval: time.Second,
			items:    []int32{1, 1},
			wantSame: []bool{true},
		},
		{
			name:     "Multiple items",
			rate:     1,
			burst:    5,
			interval: time.Second,
			items:    []int32{1, 2, 1, 1},
			wantSame: []bool{false, false, true},
		},
		{
			name:     "Multiple itemss with exclusions",
			rate:     1,
			burst:    5,
			interval: time.Second,
			items:    []int32{1, 2, 3},
			wantSame: []bool{true, false},
			excluded: map[int32]struct{}{1: {}, 2: {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimiter := NewRateLimiter(tt.rate, tt.burst, tt.interval, tt.excluded)
			size := len(tt.items)
			buckets := make([]*TokenBucket, 0, size)

			for _, item := range tt.items {
				bucket := rateLimiter.GetBucket(item)

				if _, ok := tt.excluded[item]; ok {
					if bucket != nil {
						t.Errorf("expected nil bucket for excluded item %d, but got %v", item, bucket)
					}
				}

				buckets = append(buckets, rateLimiter.GetBucket(item))
			}

			for i := 1; i < size; i++ {
				sameBucket := buckets[i] == buckets[i-1]

				if sameBucket != tt.wantSame[i-1] {
					t.Errorf(
						"request %d and %d: got same bucket = %v, want same = %v for items %d and %d",
						i, i-1, sameBucket, tt.wantSame[i-1], tt.items[i-1], tt.items[i],
					)
				}
			}
		})
	}
}

func TestRateLimiter_RateLimiting(t *testing.T) {
	tests := []struct {
		name           string
		rate           float64
		burst          float64
		interval       time.Duration
		items          []int32
		sleepIntervals []time.Duration
		wantResults    []bool
		excluded       map[int32]struct{}
	}{
		{
			name:           "Single item within limit",
			rate:           1,
			burst:          2,
			interval:       time.Second,
			items:          []int32{1, 1},
			sleepIntervals: []time.Duration{0, 0},
			wantResults:    []bool{true, true},
		},
		{
			name:           "Single item exceeding limit",
			rate:           1,
			burst:          2,
			interval:       time.Second,
			items:          []int32{1, 1, 1},
			sleepIntervals: []time.Duration{0, 0, 0},
			wantResults:    []bool{true, true, false},
		},
		{
			name:           "Multiple items separate limits",
			rate:           1,
			burst:          1,
			interval:       time.Second,
			items:          []int32{1, 2, 1, 2},
			sleepIntervals: []time.Duration{0, 0, 0, 0},
			wantResults:    []bool{true, true, false, false},
		},
		{
			name:           "Item refill after time",
			rate:           1,
			burst:          1,
			interval:       time.Millisecond * 50,
			items:          []int32{1, 1, 1},
			sleepIntervals: []time.Duration{0, 0, time.Millisecond * 60},
			wantResults:    []bool{true, false, true},
		},
		{
			name:           "Excluded item",
			rate:           1,
			burst:          1,
			interval:       time.Second,
			items:          []int32{1, 2, 1, 2},
			sleepIntervals: []time.Duration{0, 0, 0, 0},
			wantResults:    []bool{true, true, true, true},
			excluded:       map[int32]struct{}{1: {}, 2: {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimit := NewRateLimiter(tt.rate, tt.burst, tt.interval, tt.excluded)

			for i, item := range tt.items {
				if i > 0 && tt.sleepIntervals[i] > 0 {
					time.Sleep(tt.sleepIntervals[i])
				}

				bucket := rateLimit.GetBucket(item)
				got := bucket == nil || bucket.Allow()

				if want := tt.wantResults[i]; got != want {
					t.Errorf("request %d for item %d: got = %v, want %v", i+1, item, got, want)
				}
			}
		})
	}
}

func TestRateLimiter_CleanupBuckets(t *testing.T) {
	tests := []struct {
		name            string
		items           []int32
		sleepDurations  []time.Duration
		interval        time.Duration
		cleanupInterval time.Duration
		wantRemoved     uint64
		wantRemaining   uint64
	}{
		{
			name:            "No buckets to clean",
			items:           []int32{},
			interval:        time.Millisecond * 100,
			cleanupInterval: time.Millisecond,
		},
		{
			name:            "No idle buckets",
			items:           []int32{1, 2},
			sleepDurations:  []time.Duration{time.Millisecond, time.Millisecond},
			interval:        time.Millisecond * 100,
			cleanupInterval: time.Millisecond * 50,
			wantRemoved:     0,
			wantRemaining:   2,
		},
		{
			name:            "Some idle buckets",
			items:           []int32{1, 2, 3},
			sleepDurations:  []time.Duration{time.Millisecond * 25, time.Millisecond, time.Millisecond},
			interval:        time.Millisecond * 100,
			cleanupInterval: time.Millisecond * 20,
			wantRemoved:     1,
			wantRemaining:   2,
		},
		{
			name:            "All buckets idle",
			items:           []int32{1, 2},
			sleepDurations:  []time.Duration{time.Millisecond * 10, time.Millisecond * 20},
			interval:        time.Millisecond * 100,
			cleanupInterval: time.Millisecond * 10,
			wantRemoved:     2,
		},
		{
			name:            "Multiple identical items",
			items:           []int32{1, 2, 1},
			sleepDurations:  []time.Duration{time.Millisecond * 5, time.Millisecond * 25, time.Millisecond},
			interval:        time.Millisecond * 100,
			cleanupInterval: time.Millisecond * 20,
			wantRemoved:     1,
			wantRemaining:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimiter := NewRateLimiter[int32](10.0, 20.0, tt.interval, nil)

			for i, item := range tt.items {
				bucket := rateLimiter.GetBucket(item)
				_ = bucket != nil && bucket.Allow()

				time.Sleep(tt.sleepDurations[i])
			}

			got := rateLimiter.cleanupBuckets(tt.cleanupInterval)
			if got != tt.wantRemoved {
				t.Errorf("cleanupBuckets() = %v, want %v", got, tt.wantRemoved)
			}

			rateLimiter.RLock()
			got = uint64(len(rateLimiter.buckets))
			rateLimiter.RUnlock()

			if got != tt.wantRemaining {
				t.Errorf("remaining buckets = %v, want %v", got, tt.wantRemaining)
			}
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	interval := time.Millisecond * 500
	rateLimiter := NewRateLimiter[int32](10.0, 20.0, interval, nil)

	items := []int32{1, 2, 3}
	for _, item := range items {
		bucket := rateLimiter.GetBucket(item)
		_ = bucket.Allow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*150)
	defer cancel()

	cleanupInterval := 100 * time.Millisecond
	done := rateLimiter.Cleanup(ctx, cleanupInterval)

	time.Sleep(80 * time.Millisecond)
	bucket := rateLimiter.GetBucket(items[2]) // remaining bucket
	_ = bucket.Allow()

	// wait for cleanup to finish
	<-done

	rateLimiter.RLock()
	remainingCount := len(rateLimiter.buckets)
	rateLimiter.RUnlock()

	if remainingCount != 1 {
		t.Errorf("after cleanup goroutine: bucket count = %v, want %v", remainingCount, 1)
	}
}
