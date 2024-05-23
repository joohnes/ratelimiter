package ratelimiter

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBuf      = errors.New("buffor must be greater than 0")
	ErrInterval = errors.New("interval must be greater than 0")
)

type RateLimiter struct {
	buf      uint
	maxBuf   uint
	interval time.Duration
	ticker   *time.Ticker
}

func NewRateLimiter(ctx context.Context, Interval time.Duration) (*RateLimiter, error) {
	return NewRateLimiterWithBuffor(ctx, 1, Interval)
}

func NewRateLimiterWithBuffor(ctx context.Context, Buffor int, Interval time.Duration) (*RateLimiter, error) {
	if Buffor < 1 {
		return nil, ErrBuf
	}
	if Interval < 1 {
		return nil, ErrInterval
	}

	rl := &RateLimiter{buf: uint(Buffor), maxBuf: uint(Buffor), interval: Interval}
	rl.ticker = time.NewTicker(rl.interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				rl.ticker.Stop()
				return
			case <-rl.ticker.C:
				if rl.buf < rl.maxBuf {
					rl.buf += 1
				}
			}
		}
	}()

	return rl, nil
}

func (rl *RateLimiter) Use() bool {
	if rl.buf > 0 {
		rl.buf -= 1
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

func (rl *RateLimiter) Buffor() int {
	return int(rl.maxBuf)
}

func (rl *RateLimiter) SetBuffor(newMaxBuf int) error {
	if newMaxBuf < 1 {
		return ErrBuf
	}
	rl.maxBuf = uint(newMaxBuf)
	return nil
}

func (rl *RateLimiter) Interval() time.Duration {
	return rl.interval
}

func (rl *RateLimiter) SetInterval(newInterval time.Duration) error {
	if newInterval < 1 {
		return ErrInterval
	}
	rl.interval = newInterval
	rl.ticker.Reset(rl.interval)
	return nil
}
