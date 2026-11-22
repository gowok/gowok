package retry

import (
	"testing"
	"time"

	"github.com/golang-must/must"
)

func TestFixed(t *testing.T) {
	t.Run("positive/constant delay", func(t *testing.T) {
		backoff := Fixed(50 * time.Millisecond)

		must.Equal(t, 50*time.Millisecond, backoff(1))
		must.Equal(t, 50*time.Millisecond, backoff(10))
	})

	t.Run("positive/negative delay clamps to zero", func(t *testing.T) {
		must.Equal(t, time.Duration(0), Fixed(-time.Second)(1))
	})
}

func TestLinear(t *testing.T) {
	t.Run("positive/grows by step", func(t *testing.T) {
		backoff := Linear(10*time.Millisecond, 20*time.Millisecond, 0)

		must.Equal(t, 10*time.Millisecond, backoff(1))
		must.Equal(t, 30*time.Millisecond, backoff(2))
		must.Equal(t, 50*time.Millisecond, backoff(3))
	})

	t.Run("positive/retry below one behaves like first retry", func(t *testing.T) {
		backoff := Linear(10*time.Millisecond, 20*time.Millisecond, 0)

		must.Equal(t, 10*time.Millisecond, backoff(0))
	})

	t.Run("positive/max caps the delay", func(t *testing.T) {
		backoff := Linear(10*time.Millisecond, 20*time.Millisecond, 40*time.Millisecond)

		must.Equal(t, 30*time.Millisecond, backoff(2))
		must.Equal(t, 40*time.Millisecond, backoff(3))
		must.Equal(t, 40*time.Millisecond, backoff(10))
	})

	t.Run("positive/negative inputs clamp", func(t *testing.T) {
		backoff := Linear(-time.Second, -time.Second, 0)

		must.Equal(t, time.Duration(0), backoff(1))
		must.Equal(t, time.Duration(0), backoff(3))
	})
}

func TestExponential(t *testing.T) {
	t.Run("positive/default doubles from initial", func(t *testing.T) {
		backoff := Exponential(100 * time.Millisecond)

		must.Equal(t, 100*time.Millisecond, backoff(1))
		must.Equal(t, 200*time.Millisecond, backoff(2))
		must.Equal(t, 400*time.Millisecond, backoff(3))
	})

	t.Run("positive/max caps the delay", func(t *testing.T) {
		backoff := Exponential(100*time.Millisecond, WithMaxDelay(500*time.Millisecond))

		must.Equal(t, 400*time.Millisecond, backoff(3))
		must.Equal(t, 500*time.Millisecond, backoff(4))
		must.Equal(t, 500*time.Millisecond, backoff(100))
	})

	t.Run("positive/custom multiplier", func(t *testing.T) {
		backoff := Exponential(100*time.Millisecond, WithMultiplier(3))

		must.Equal(t, 100*time.Millisecond, backoff(1))
		must.Equal(t, 300*time.Millisecond, backoff(2))
		must.Equal(t, 900*time.Millisecond, backoff(3))
	})

	t.Run("positive/multiplier below one clamps to constant", func(t *testing.T) {
		backoff := Exponential(100*time.Millisecond, WithMultiplier(0))

		must.Equal(t, 100*time.Millisecond, backoff(1))
		must.Equal(t, 100*time.Millisecond, backoff(5))
	})

	t.Run("positive/retry below one behaves like first retry", func(t *testing.T) {
		backoff := Exponential(100 * time.Millisecond)

		must.Equal(t, 100*time.Millisecond, backoff(0))
	})

	t.Run("positive/negative initial clamps to zero", func(t *testing.T) {
		must.Equal(t, time.Duration(0), Exponential(-time.Second)(1))
	})

	t.Run("positive/jitter stays within bounds", func(t *testing.T) {
		backoff := Exponential(100*time.Millisecond, WithJitter(0.2))

		for i := 0; i < 200; i++ {
			delay := backoff(1)
			must.True(t, delay >= 80*time.Millisecond)
			must.True(t, delay <= 120*time.Millisecond)
		}
	})

	t.Run("positive/jitter above one clamps to full range", func(t *testing.T) {
		backoff := Exponential(100*time.Millisecond, WithJitter(2))

		for i := 0; i < 200; i++ {
			delay := backoff(1)
			must.True(t, delay >= 0)
			must.True(t, delay <= 200*time.Millisecond)
		}
	})

	t.Run("positive/zero jitter is deterministic", func(t *testing.T) {
		backoff := Exponential(100*time.Millisecond, WithJitter(0))

		must.Equal(t, 100*time.Millisecond, backoff(1))
		must.Equal(t, 200*time.Millisecond, backoff(2))
	})
}
