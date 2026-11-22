package retry

import (
	"math/rand/v2"
	"time"
)

const defaultMaxDelay = 30 * time.Second

type Backoff func(retry int) time.Duration

type BackoffOption func(*exponentialConfig)

type exponentialConfig struct {
	multiplier float64
	maxDelay   time.Duration
	jitter     float64
}

func Fixed(delay time.Duration) Backoff {
	if delay < 0 {
		delay = 0
	}

	return func(int) time.Duration {
		return delay
	}
}

func Linear(initial, step, max time.Duration) Backoff {
	if initial < 0 {
		initial = 0
	}
	if step < 0 {
		step = 0
	}

	return func(retry int) time.Duration {
		if retry < 1 {
			retry = 1
		}

		delay := initial + step*time.Duration(retry-1)
		if max > 0 && delay > max {
			delay = max
		}

		return delay
	}
}

func Exponential(initial time.Duration, opts ...BackoffOption) Backoff {
	cfg := exponentialConfig{
		multiplier: 2,
		maxDelay:   defaultMaxDelay,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if initial < 0 {
		initial = 0
	}
	if cfg.multiplier < 1 {
		cfg.multiplier = 1
	}
	if cfg.maxDelay <= 0 {
		cfg.maxDelay = defaultMaxDelay
	}
	if cfg.jitter < 0 {
		cfg.jitter = 0
	}
	if cfg.jitter > 1 {
		cfg.jitter = 1
	}

	return func(retry int) time.Duration {
		if retry < 1 {
			retry = 1
		}

		delay := float64(initial)
		maxDelay := float64(cfg.maxDelay)
		for i := 1; i < retry; i++ {
			delay *= cfg.multiplier
			if delay >= maxDelay {
				delay = maxDelay
				break
			}
		}

		delay = applyJitter(delay, cfg.jitter)
		if delay > maxDelay {
			delay = maxDelay
		}

		return time.Duration(delay)
	}
}

func WithMultiplier(multiplier float64) BackoffOption {
	return func(c *exponentialConfig) {
		c.multiplier = multiplier
	}
}

func WithMaxDelay(maxDelay time.Duration) BackoffOption {
	return func(c *exponentialConfig) {
		c.maxDelay = maxDelay
	}
}

func WithJitter(factor float64) BackoffOption {
	return func(c *exponentialConfig) {
		c.jitter = factor
	}
}

func applyJitter(delay float64, factor float64) float64 {
	if factor <= 0 || delay <= 0 {
		return delay
	}

	delta := delay * factor
	return delay - delta + rand.Float64()*2*delta
}
