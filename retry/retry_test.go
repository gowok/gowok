package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-must/must"
)

func recordSleep(delays *[]time.Duration) Option {
	return withSleep(func(_ context.Context, delay time.Duration) error {
		*delays = append(*delays, delay)
		return nil
	})
}

func TestDo(t *testing.T) {
	t.Run("positive/succeeds on first attempt", func(t *testing.T) {
		var delays []time.Duration
		calls := 0

		err := Do(context.Background(), func(context.Context) error {
			calls++
			return nil
		}, recordSleep(&delays))

		must.Nil(t, err)
		must.Equal(t, 1, calls)
		must.Equal(t, 0, len(delays))
	})

	t.Run("positive/succeeds after retries", func(t *testing.T) {
		var delays []time.Duration
		calls := 0

		err := Do(context.Background(), func(context.Context) error {
			calls++
			if calls < 3 {
				return errors.New("transient")
			}
			return nil
		},
			WithMaxAttempts(5),
			WithBackoff(Fixed(10*time.Millisecond)),
			recordSleep(&delays),
		)

		must.Nil(t, err)
		must.Equal(t, 3, calls)
		must.Equal(t, 2, len(delays))
		must.Equal(t, 10*time.Millisecond, delays[0])
		must.Equal(t, 10*time.Millisecond, delays[1])
	})

	t.Run("negative/exhausts attempts and keeps the last error", func(t *testing.T) {
		var delays []time.Duration
		calls := 0
		first := errors.New("first")
		last := errors.New("last")

		err := Do(context.Background(), func(context.Context) error {
			calls++
			if calls == 1 {
				return first
			}
			return last
		}, WithMaxAttempts(3), recordSleep(&delays))

		must.Equal(t, 3, calls)
		must.Equal(t, 2, len(delays))
		must.True(t, errors.Is(err, last))
		must.False(t, errors.Is(err, first))
	})

	t.Run("negative/permanent error stops immediately", func(t *testing.T) {
		calls := 0
		sentinel := errors.New("permanent")

		err := Do(context.Background(), func(context.Context) error {
			calls++
			return Permanent(sentinel)
		})

		must.Equal(t, 1, calls)
		must.True(t, IsPermanent(err))
		must.True(t, errors.Is(err, sentinel))
	})

	t.Run("negative/retry_if rejects the error", func(t *testing.T) {
		calls := 0
		stop := errors.New("stop")

		err := Do(context.Background(), func(context.Context) error {
			calls++
			return stop
		}, WithRetryIf(func(err error) bool {
			return !errors.Is(err, stop)
		}))

		must.Equal(t, 1, calls)
		must.True(t, errors.Is(err, stop))
	})

	t.Run("negative/context cancelled before the first attempt", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		calls := 0
		err := Do(ctx, func(context.Context) error {
			calls++
			return nil
		})

		must.Equal(t, 0, calls)
		must.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("negative/context cancelled while waiting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		calls := 0
		transient := errors.New("transient")
		err := Do(ctx, func(context.Context) error {
			calls++
			return transient
		},
			WithMaxAttempts(3),
			WithOnRetry(func(int, error, time.Duration) { cancel() }),
			withSleep(func(ctx context.Context, _ time.Duration) error {
				return ctx.Err()
			}),
		)

		must.Equal(t, 1, calls)
		must.True(t, errors.Is(err, transient))
		must.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("negative/context error from the operation stops retrying", func(t *testing.T) {
		calls := 0

		err := Do(context.Background(), func(context.Context) error {
			calls++
			return context.DeadlineExceeded
		}, WithMaxAttempts(3))

		must.Equal(t, 1, calls)
		must.True(t, errors.Is(err, context.DeadlineExceeded))
	})

	t.Run("positive/on_retry receives retry number and delay", func(t *testing.T) {
		type retryCall struct {
			retry int
			delay time.Duration
		}

		var delays []time.Duration
		var retries []retryCall
		err := Do(context.Background(), func(context.Context) error {
			return errors.New("transient")
		},
			WithMaxAttempts(3),
			WithBackoff(Linear(10*time.Millisecond, 10*time.Millisecond, 0)),
			recordSleep(&delays),
			WithOnRetry(func(retry int, err error, delay time.Duration) {
				retries = append(retries, retryCall{retry, delay})
			}),
		)

		must.NotNil(t, err)
		must.Equal(t, 2, len(retries))
		must.Equal(t, 1, retries[0].retry)
		must.Equal(t, 10*time.Millisecond, retries[0].delay)
		must.Equal(t, 2, retries[1].retry)
		must.Equal(t, 20*time.Millisecond, retries[1].delay)
	})

	t.Run("positive/max_attempts below one clamps to one", func(t *testing.T) {
		calls := 0

		_ = Do(context.Background(), func(context.Context) error {
			calls++
			return errors.New("transient")
		}, WithMaxAttempts(0))

		must.Equal(t, 1, calls)
	})

	t.Run("positive/nil context uses background", func(t *testing.T) {
		//nolint:staticcheck
		err := Do(nil, func(context.Context) error {
			return nil
		})

		must.Nil(t, err)
	})

	t.Run("negative/nil function", func(t *testing.T) {
		must.NotNil(t, Do(context.Background(), nil))
	})
}

func TestDoValue(t *testing.T) {
	t.Run("positive/returns the value", func(t *testing.T) {
		value, err := DoValue(context.Background(), func(context.Context) (int, error) {
			return 42, nil
		})

		must.Nil(t, err)
		must.Equal(t, 42, value)
	})

	t.Run("positive/retries until success", func(t *testing.T) {
		var delays []time.Duration
		calls := 0

		value, err := DoValue(context.Background(), func(context.Context) (string, error) {
			calls++
			if calls < 2 {
				return "", errors.New("transient")
			}
			return "done", nil
		}, recordSleep(&delays))

		must.Nil(t, err)
		must.Equal(t, "done", value)
		must.Equal(t, 2, calls)
		must.Equal(t, 1, len(delays))
	})

	t.Run("negative/returns zero value on failure", func(t *testing.T) {
		sentinel := errors.New("fail")

		value, err := DoValue(context.Background(), func(context.Context) (int, error) {
			return 42, sentinel
		}, WithMaxAttempts(1))

		must.True(t, errors.Is(err, sentinel))
		must.Equal(t, 0, value)
	})

	t.Run("negative/nil function", func(t *testing.T) {
		value, err := DoValue[int](context.Background(), nil)

		must.NotNil(t, err)
		must.Equal(t, 0, value)
	})
}

func TestPermanent(t *testing.T) {
	t.Run("positive/nil stays nil", func(t *testing.T) {
		must.Nil(t, Permanent(nil))
	})

	t.Run("positive/marks the error", func(t *testing.T) {
		sentinel := errors.New("sentinel")

		err := Permanent(sentinel)

		must.True(t, IsPermanent(err))
		must.True(t, errors.Is(err, sentinel))
	})

	t.Run("positive/wrapped error stays permanent", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		err := fmt.Errorf("outer: %w", Permanent(sentinel))

		must.True(t, IsPermanent(err))
		must.True(t, errors.Is(err, sentinel))
	})

	t.Run("negative/plain error is not permanent", func(t *testing.T) {
		must.False(t, IsPermanent(errors.New("plain")))
	})
}
