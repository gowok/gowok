package storage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func newMem(t *testing.T) *memory {
	t.Helper()
	return Memory()
}

func TestMemoryRoundTrip(t *testing.T) {
	m := newMem(t)
	ctx := context.Background()
	key := "k"
	value := []byte("hello")

	if err := m.Put(ctx, key, value); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, err := m.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("Get = %q, want %q", got, value)
	}
}

func TestMemoryGetMissing(t *testing.T) {
	m := newMem(t)
	_, err := m.Get(context.Background(), "nope")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestMemoryEmptyKey(t *testing.T) {
	m := newMem(t)
	ctx := context.Background()
	for _, op := range []string{"get", "put"} {
		t.Run(op, func(t *testing.T) {
			var err error
			switch op {
			case "get":
				_, err = m.Get(ctx, "")
			case "put":
				err = m.Put(ctx, "", []byte("x"))
			}
			if err == nil {
				t.Fatalf("expected error for empty key on %s", op)
			}
		})
	}
}

func TestMemoryClone(t *testing.T) {
	m := newMem(t)
	ctx := context.Background()
	key := "k"
	original := []byte("hello")

	if err := m.Put(ctx, key, original); err != nil {
		t.Fatal(err)
	}
	// mutate original after Put
	original[0] = 'X'

	got, err := m.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("stored value changed by input mutation: %q", got)
	}

	// mutate returned slice
	got[0] = 'Y'
	got2, err := m.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got2) != "hello" {
		t.Fatalf("stored value changed by returned slice mutation: %q", got2)
	}
}

func TestMemoryConcurrency(t *testing.T) {
	m := newMem(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key"
			_ = m.Put(ctx, key, []byte("v"))
			_, _ = m.Get(ctx, key)
		}(i)
	}
	wg.Wait()
}

func TestMemoryContextCanceled(t *testing.T) {
	m := newMem(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := m.Get(ctx, "k"); err != context.Canceled {
		t.Fatalf("Get err = %v, want context.Canceled", err)
	}
	if err := m.Put(ctx, "k", []byte("v")); err != context.Canceled {
		t.Fatalf("Put err = %v, want context.Canceled", err)
	}
}

func TestMemoryContextDeadlineExceeded(t *testing.T) {
	m := newMem(t)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	if _, err := m.Get(ctx, "k"); err != context.DeadlineExceeded {
		t.Fatalf("Get err = %v, want context.DeadlineExceeded", err)
	}
	if err := m.Put(ctx, "k", []byte("v")); err != context.DeadlineExceeded {
		t.Fatalf("Put err = %v, want context.DeadlineExceeded", err)
	}
}
