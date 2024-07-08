package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	mu sync.Mutex

	burst         uint
	maxBurst      uint
	burstInterval time.Duration

	burstCooldown time.Time
	interval      time.Duration
	ticker        *time.Ticker
}

// RateLimiterOptions is a struct that holds the options for the RateLimiter
//
// # BurstAmount is the amount of uses that can be used in a burst
//
// # BurstInterval is the minimum time between each use in a burst
//
// # Interval is the time to wait for the burst to refill by one
type Options struct {
	BurstAmount   int
	BurstInterval time.Duration
	Interval      time.Duration
}

func NewRateLimiter(ctx context.Context, interval time.Duration) *RateLimiter {
	opts := Options{
		BurstAmount:   1,
		BurstInterval: interval,
		Interval:      interval,
	}
	return NewRateLimiterWithBurst(ctx, opts)
}

func NewRateLimiterWithBurst(ctx context.Context, opts Options) *RateLimiter {
	if opts.BurstAmount < 1 {
		opts.BurstAmount = 1
	}
	if opts.Interval < 1 {
		opts.Interval = time.Second
	}

	rl := &RateLimiter{
		burst:         uint(opts.BurstAmount),
		maxBurst:      uint(opts.BurstAmount),
		interval:      opts.Interval,
		burstInterval: opts.BurstInterval,
		burstCooldown: time.Now(),
	}
	rl.ticker = time.NewTicker(rl.interval)
	defer rl.ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-rl.ticker.C:
				if rl.burst < rl.maxBurst {
					rl.mu.Lock()
					rl.burst += 1
					rl.mu.Unlock()
				}
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Use() bool {
	if rl.burst > 0 && rl.burstCooldown.Before(time.Now()) {
		rl.mu.Lock()
		defer rl.mu.Unlock()

		rl.burstCooldown = time.Now().Add(rl.burstInterval)
		rl.burst -= 1
		rl.ticker.Reset(rl.interval)
		return true
	}

	return false
}

func (rl *RateLimiter) Wait(ctx context.Context) {
	allow := make(chan struct{})
	defer close(allow)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if rl.Use() {
					allow <- struct{}{}
					return
				}
			}
		}
	}()
	<-allow
}

func (rl *RateLimiter) MaxBurst() int {
	return int(rl.maxBurst)
}

func (rl *RateLimiter) CurrentBurst() int {
	return int(rl.burst)
}

func (rl *RateLimiter) SetBurst(newMaxBurst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if newMaxBurst < 1 {
		newMaxBurst = 1
	}

	rl.maxBurst = uint(newMaxBurst)
}

func (rl *RateLimiter) ResetBurst() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.burst = rl.maxBurst
}

func (rl *RateLimiter) BurstInterval() time.Duration {
	return rl.burstInterval
}

func (rl *RateLimiter) SetBurstInterval(newBurstInterval time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if newBurstInterval < 1 {
		newBurstInterval = time.Second
	}

	rl.burstInterval = newBurstInterval
}

func (rl *RateLimiter) Interval() time.Duration {
	return rl.interval
}

func (rl *RateLimiter) SetInterval(newInterval time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if newInterval < 1 {
		newInterval = time.Second
	}

	rl.interval = newInterval
	rl.ticker.Reset(rl.interval)
}
