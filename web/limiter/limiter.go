// Package limiter provides rate limiting functionality for the application.
package limiter

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// TokenBucket is a rate limiter that uses the token bucket algorithm.
type TokenBucket struct {
	lastRefillTime time.Time
	interval       time.Duration
	tokens         float64
	maxTokens      float64
	refillRate     float64 // per interval
	sync.RWMutex
}

// NewTokenBucket creates a new TokenBucket with the specified max tokens and refill rate.
func NewTokenBucket(maxTokens, refillRate float64, interval time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		interval:       interval,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request is allowed based on the token bucket algorithm.
func (tb *TokenBucket) Allow() bool {
	const step float64 = 1.0

	tb.Lock()
	defer tb.Unlock()

	now := time.Now()
	elapsed := float64(now.Sub(tb.lastRefillTime)) / float64(tb.interval)
	tb.lastRefillTime = now

	// add tokens by rate from last refill time
	tb.tokens = min(tb.tokens+elapsed*tb.refillRate, tb.maxTokens)

	if tb.tokens < step {
		return false
	}

	tb.tokens -= step
	return true
}

// RateLimiter is a rate limiter that limits requests based on a key.
type RateLimiter[T comparable] struct {
	buckets  map[T]*TokenBucket
	excluded map[T]struct{}
	interval time.Duration
	rate     float64
	burst    float64
	sync.RWMutex
}

// NewRateLimiter creates a new RateLimiter with the specified rate and burst.
func NewRateLimiter[T comparable](rate, burst float64, interval time.Duration, excluded map[T]struct{}) *RateLimiter[T] {
	return &RateLimiter[T]{
		buckets:  make(map[T]*TokenBucket),
		rate:     rate,
		burst:    burst,
		interval: interval,
		excluded: excluded,
	}
}

// GetBucket returns the TokenBucket for the given key.
func (irl *RateLimiter[T]) GetBucket(key T) *TokenBucket {
	if _, ok := irl.excluded[key]; ok {
		return nil
	}

	irl.RLock()
	bucket, ok := irl.buckets[key]
	irl.RUnlock()

	if !ok {
		return irl.getOrCreateBucket(key)
	}

	return bucket
}

// Cleanup starts a goroutine that periodically cleans up buckets that have not been used for a specified duration.
// It returns a channel that can be used to wait the cleanup process and stop it using the context.
func (irl *RateLimiter[T]) Cleanup(ctx context.Context, interval time.Duration) chan struct{} {
	var (
		ticker = time.NewTicker(interval)
		done   = make(chan struct{})
		count  uint64
	)

	go func() {
		defer func() {
			ticker.Stop()
			close(done)
		}()

		slog.Info("starting rate limit cleanup", "interval", interval)
		for {
			select {
			case <-ticker.C:
				count = irl.cleanupBuckets(interval)
				slog.Info("cleanup rate limit buckets", "count", count)
			case <-ctx.Done():
				slog.Info("stopping cleanup of rate limit buckets")
				return
			}
		}
	}()

	return done
}

// getOrCreateBucket returns the TokenBucket for the given key.
// It uses privileged mode to check if the limiter was created before.
func (irl *RateLimiter[T]) getOrCreateBucket(key T) *TokenBucket {
	irl.Lock()
	bucket, ok := irl.buckets[key]

	if !ok {
		bucket = NewTokenBucket(irl.burst, irl.rate, irl.interval)
		irl.buckets[key] = bucket
	}

	irl.Unlock()
	return bucket
}

// cleanupBuckets removes buckets that have not been used for a specified duration.
func (irl *RateLimiter[T]) cleanupBuckets(cleanupInterval time.Duration) uint64 {
	var (
		count uint64
		now   = time.Now()
	)
	irl.Lock()

	for key, bucket := range irl.buckets {
		bucket.RLock()
		lastUsed := bucket.lastRefillTime

		if now.Sub(lastUsed) > cleanupInterval {
			delete(irl.buckets, key)
			count++
		}
		bucket.RUnlock()
	}

	irl.Unlock()
	return count
}
