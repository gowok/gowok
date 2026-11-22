package retry

import (
	"context"
	"errors"
	"time"
)

const (
	defaultMaxAttempts  = 3
	defaultInitialDelay = 100 * time.Millisecond
)

var errNilFunction = errors.New("retry: nil function")

type Option func(*config)

type config struct {
	maxAttempts int
	backoff     Backoff
	retryIf     func(error) bool
	onRetry     func(retry int, err error, delay time.Duration)
	sleep       func(context.Context, time.Duration) error
}

func newConfig(opts ...Option) *config {
	cfg := &config{
		maxAttempts: defaultMaxAttempts,
		backoff:     Exponential(defaultInitialDelay),
		retryIf:     func(error) bool { return true },
		sleep:       sleepContext,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	if cfg.maxAttempts < 1 {
		cfg.maxAttempts = 1
	}
	if cfg.backoff == nil {
		cfg.backoff = Exponential(defaultInitialDelay)
	}
	if cfg.retryIf == nil {
		cfg.retryIf = func(error) bool { return true }
	}
	if cfg.sleep == nil {
		cfg.sleep = sleepContext
	}

	return cfg
}

func WithMaxAttempts(attempts int) Option {
	return func(c *config) {
		c.maxAttempts = attempts
	}
}

func WithBackoff(backoff Backoff) Option {
	return func(c *config) {
		c.backoff = backoff
	}
}

func WithRetryIf(judge func(error) bool) Option {
	return func(c *config) {
		c.retryIf = judge
	}
}

func WithOnRetry(fn func(retry int, err error, delay time.Duration)) Option {
	return func(c *config) {
		c.onRetry = fn
	}
}

func Do(ctx context.Context, fn func(context.Context) error, opts ...Option) error {
	if fn == nil {
		return errNilFunction
	}

	_, err := DoValue(ctx, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	}, opts...)

	return err
}

func DoValue[T any](ctx context.Context, fn func(context.Context) (T, error), opts ...Option) (T, error) {
	var zero T
	if fn == nil {
		return zero, errNilFunction
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := newConfig(opts...)
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	var lastErr error
	for attempt := 1; attempt <= cfg.maxAttempts; attempt++ {
		if attempt > 1 {
			retry := attempt - 1
			delay := cfg.backoff(retry)
			if cfg.onRetry != nil {
				cfg.onRetry(retry, lastErr, delay)
			}
			if err := cfg.sleep(ctx, delay); err != nil {
				return zero, errors.Join(lastErr, err)
			}
		}

		if err := ctx.Err(); err != nil {
			return zero, errors.Join(lastErr, err)
		}

		value, err := fn(ctx)
		if err == nil {
			return value, nil
		}

		lastErr = err
		if IsPermanent(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return zero, err
		}
		if !cfg.retryIf(err) {
			return zero, err
		}
	}

	return zero, lastErr
}

func Permanent(err error) error {
	if err == nil {
		return nil
	}

	return permanentError{err: err}
}

func IsPermanent(err error) bool {
	var target permanentError
	return errors.As(err, &target)
}

type permanentError struct {
	err error
}

func (e permanentError) Error() string {
	return e.err.Error()
}

func (e permanentError) Unwrap() error {
	return e.err
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func withSleep(sleep func(context.Context, time.Duration) error) Option {
	return func(c *config) {
		c.sleep = sleep
	}
}
