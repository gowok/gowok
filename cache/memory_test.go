package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemory(t *testing.T) {
	client := &memory{data: make(map[string]_value)}
	ctx := context.Background()

	got, err := client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("missing key = %#v, want nil", got)
	}

	if err := client.Set(ctx, "key", "first"); err != nil {
		t.Fatal(err)
	}
	got, err = client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if got != "first" {
		t.Fatalf("value = %#v, want %q", got, "first")
	}

	if err := client.Set(ctx, "key", "second"); err != nil {
		t.Fatal(err)
	}
	got, err = client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if got != "second" {
		t.Fatalf("overwritten value = %#v, want %q", got, "second")
	}

	if err := client.Delete(ctx, "key"); err != nil {
		t.Fatal(err)
	}
	got, err = client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("deleted key = %#v, want nil", got)
	}
}

func TestMemoryTTL(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	if err := client.Set(ctx, "key", "value", WithTTL(time.Hour)); err != nil {
		t.Fatal(err)
	}
	got, err := client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if got != "value" {
		t.Fatalf("value = %#v, want %q", got, "value")
	}

	if err := client.Set(ctx, "key", "value", WithTTL(time.Nanosecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	got, err = client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expired value = %#v, want nil", got)
	}
}

func TestMemoryFallback(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	got, err := client.Get(ctx, "missing", WithFallback("fallback"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "fallback" {
		t.Fatalf("missing value = %#v, want %q", got, "fallback")
	}

	if err := client.Set(ctx, "key", nil); err != nil {
		t.Fatal(err)
	}
	got, err = client.Get(ctx, "key", WithFallback("fallback"))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("stored nil = %#v, want nil", got)
	}
}

func TestMemoryTTLNonPositiveDoesNotExpire(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	for _, ttl := range []time.Duration{0, -time.Second} {
		if err := client.Set(ctx, "key", "value", WithTTL(ttl)); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
		got, err := client.Get(ctx, "key")
		if err != nil {
			t.Fatal(err)
		}
		if got != "value" {
			t.Fatalf("value with TTL %v = %#v, want %q", ttl, got, "value")
		}
	}
}

func TestMemoryExpiredRemovedOnGet(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	if err := client.Set(ctx, "key", "value", WithTTL(time.Nanosecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)

	got, err := client.Get(ctx, "key", WithFallback("fallback"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "fallback" {
		t.Fatalf("expired value = %#v, want %q", got, "fallback")
	}

	client.mu.Lock()
	n := len(client.data)
	client.mu.Unlock()
	if n != 0 {
		t.Fatalf("len data = %d, want 0", n)
	}

	got, err = client.Get(ctx, "key", WithFallback("fallback"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "fallback" {
		t.Fatalf("subsequent value = %#v, want %q", got, "fallback")
	}
}

func TestMemoryRace(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if err := client.Set(ctx, "key", "value", WithTTL(time.Millisecond)); err != nil {
					t.Error(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if _, err := client.Get(ctx, "key", WithFallback("fallback")); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
}
