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
	mu       sync.Mutex
	burst    uint
	maxBurst uint
	interval time.Duration
	ticker   *time.Ticker
}

func NewRateLimiter(ctx context.Context, Interval time.Duration) (*RateLimiter, error) {
	return NewRateLimiterWithBurst(ctx, 1, Interval)
}

func NewRateLimiterWithBurst(ctx context.Context, burst int, interval time.Duration) (*RateLimiter, error) {
	if burst < 1 {
		return nil, ErrBurst
	}
	if interval < 1 {
		return nil, ErrInterval
	}

	rl := &RateLimiter{burst: uint(burst), maxBurst: uint(burst), interval: interval}
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

	if rl.burst > 0 {
		rl.burst -= 1
		rl.ticker.Reset(rl.interval)
		return true
	}

	return false
}

func (rl *RateLimiter) Wait() {
	allow := make(chan struct{})
	defer close(allow)
	go func() {
		for {
			if rl.Use() {
				allow <- struct{}{}
				return
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
