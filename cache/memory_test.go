package cache

import (
	"context"
	"testing"
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
