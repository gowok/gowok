package queue

import (
	"context"
	"testing"
)

type testQueue struct{}

func (*testQueue) Enqueue(context.Context, string, Message) error { return nil }
func (*testQueue) Receive(context.Context, string) (Delivery, error) {
	return Delivery{}, nil
}
func (*testQueue) Ack(context.Context, string) error  { return nil }
func (*testQueue) Nack(context.Context, string) error { return nil }
func (*testQueue) Close(context.Context) error        { return nil }

func TestClient(t *testing.T) {
	defaultClient := &testQueue{}
	namedClient := &testQueue{}
	Register("default", defaultClient)
	Register("test-named", namedClient)

	testCases := []struct {
		name string
		args []string
		want Queue
	}{
		{
			name: "default",
			want: defaultClient,
		},
		{
			name: "empty name uses default",
			args: []string{""},
			want: defaultClient,
		},
		{
			name: "named",
			args: []string{"test-named"},
			want: namedClient,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Client(tc.args...).OrElse(nil); got != tc.want {
				t.Fatalf("Client() = %p, want %p", got, tc.want)
			}
		})
	}
}

func TestClientUnknownReturnsEmpty(t *testing.T) {
	if got := Client("missing"); got.IsPresent() {
		t.Fatalf("Client() = %#v, want empty", got)
	}
}

func TestRegister(t *testing.T) {
	first := &testQueue{}
	second := &testQueue{}

	Register("test-replace", first)
	Register("test-replace", second)

	if got := Client("test-replace").OrElse(nil); got != second {
		t.Fatalf("Client() = %p, want replacement %p", got, second)
	}
}

func TestRegisterPanics(t *testing.T) {
	testCases := []struct {
		name       string
		clientName string
		client     Queue
		want       string
	}{
		{
			name:   "empty name",
			client: &testQueue{},
			want:   "queue: name is empty",
		},
		{
			name:       "nil client",
			clientName: "test-nil",
			want:       "queue: client is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tc.want {
					t.Fatalf("panic = %v, want %q", got, tc.want)
				}
			}()

			Register(tc.clientName, tc.client)
		})
	}
}
