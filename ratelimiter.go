package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrBurst    = errors.New("buffor must be greater than 0")
	ErrInterval = errors.New("interval must be greater than 0")
)

type RateLimiter struct {
	mu sync.Mutex

	burst         uint
	maxBurst      uint
	burstInterval time.Duration

	expiry   time.Time
	interval time.Duration
	ticker   *time.Ticker
}

// RateLimiterOptions is a struct that holds the options for the RateLimiter
//
// # BurstAmount is the amount of uses that can be used in a burst
//
// # BurstInterval is the minimum time between each use in a burst
//
// # Interval is the time to wait for the burst to refill by one
type RateLimiterOptions struct {
	BurstAmount   int
	BurstInterval time.Duration
	Interval      time.Duration
}

func NewRateLimiter(ctx context.Context, interval time.Duration) (*RateLimiter, error) {
	opts := RateLimiterOptions{
		BurstAmount:   1,
		BurstInterval: interval,
		Interval:      interval,
	}
	return NewRateLimiterWithBurst(ctx, opts)
}

func NewRateLimiterWithBurst(ctx context.Context, opts RateLimiterOptions) (*RateLimiter, error) {
	if opts.BurstAmount < 1 {
		return nil, ErrBurst
	}
	if opts.Interval < 1 {
		return nil, ErrInterval
	}

	rl := &RateLimiter{
		burst:         uint(opts.BurstAmount),
		maxBurst:      uint(opts.BurstAmount),
		interval:      opts.Interval,
		burstInterval: opts.BurstInterval,
		expiry:        time.Now(),
	}
	rl.ticker = time.NewTicker(rl.interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				rl.ticker.Stop()
				return
			case <-rl.ticker.C:
				if rl.burst < rl.maxBurst {
					rl.burst += 1
				}
			}
		}
	}()

	return rl, nil
}

func (rl *RateLimiter) Use() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.burst > 0 && rl.expiry.Before(time.Now()) {
		rl.expiry = time.Now().Add(rl.burstInterval)
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

func (rl *RateLimiter) SetBurst(newMaxBurst int) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if newMaxBurst < 1 {
		return ErrBurst
	}

	rl.maxBurst = uint(newMaxBurst)
	return nil
}

func (rl *RateLimiter) Interval() time.Duration {
	return rl.interval
}

func (rl *RateLimiter) SetInterval(newInterval time.Duration) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if newInterval < 1 {
		return ErrInterval
	}

	rl.interval = newInterval
	rl.ticker.Reset(rl.interval)
	return nil
}
