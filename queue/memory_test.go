package queue

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryFIFO(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	if err := client.Enqueue(ctx, "topic", Message{ID: "first", Payload: []byte("one")}); err != nil {
		t.Fatal(err)
	}
	if err := client.Enqueue(ctx, "topic", Message{ID: "second", Payload: []byte("two")}); err != nil {
		t.Fatal(err)
	}

	first, err := client.Receive(ctx, "topic")
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.Receive(ctx, "topic")
	if err != nil {
		t.Fatal(err)
	}

	if first.Message.ID != "first" || string(first.Message.Payload) != "one" {
		t.Fatalf("first = %#v", first)
	}
	if second.Message.ID != "second" || string(second.Message.Payload) != "two" {
		t.Fatalf("second = %#v", second)
	}
	if first.Receipt == "" || second.Receipt == "" || first.Receipt == second.Receipt {
		t.Fatalf("receipts = %q, %q; want distinct non-empty values", first.Receipt, second.Receipt)
	}

	if err := client.Ack(ctx, first.Receipt); err != nil {
		t.Fatal(err)
	}
	if err := client.Ack(ctx, second.Receipt); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryNackRequeues(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	if err := client.Enqueue(ctx, "topic", Message{ID: "message"}); err != nil {
		t.Fatal(err)
	}

	first, err := client.Receive(ctx, "topic")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Nack(ctx, first.Receipt); err != nil {
		t.Fatal(err)
	}

	second, err := client.Receive(ctx, "topic")
	if err != nil {
		t.Fatal(err)
	}
	if second.Message.ID != first.Message.ID {
		t.Fatalf("message ID = %q, want %q", second.Message.ID, first.Message.ID)
	}
	if second.Receipt == first.Receipt {
		t.Fatalf("receipt = %q; want a new receipt", second.Receipt)
	}
	if err := client.Ack(ctx, second.Receipt); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryUnknownReceipt(t *testing.T) {
	client := Memory()
	ctx := context.Background()

	if err := client.Ack(ctx, "missing"); !errors.Is(err, ErrUnknownReceipt) {
		t.Fatalf("Ack error = %v, want ErrUnknownReceipt", err)
	}
	if err := client.Nack(ctx, "missing"); !errors.Is(err, ErrUnknownReceipt) {
		t.Fatalf("Nack error = %v, want ErrUnknownReceipt", err)
	}
}

func TestMemoryReceiveContextCancellation(t *testing.T) {
	client := Memory()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Receive(ctx, "topic"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Receive error = %v, want context.Canceled", err)
	}
}

func TestMemoryCloseUnblocksReceive(t *testing.T) {
	client := Memory()
	done := make(chan error, 1)

	go func() {
		_, err := client.Receive(context.Background(), "topic")
		done <- err
	}()

	if err := client.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, ErrClosed) {
		t.Fatalf("Receive error = %v, want ErrClosed", err)
	}

	if _, err := client.Receive(context.Background(), "topic"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Receive error = %v, want ErrClosed", err)
	}
	if err := client.Enqueue(context.Background(), "topic", Message{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("Enqueue error = %v, want ErrClosed", err)
	}
}

func TestMemoryCloseIsIdempotent(t *testing.T) {
	client := Memory()

	if err := client.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
