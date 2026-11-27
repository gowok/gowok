package cache

import (
	"context"
	"testing"
)

type testCache struct{}

func (*testCache) Get(context.Context, string, ...GetOptionFunc) (any, error) { return nil, nil }
func (*testCache) Set(context.Context, string, any, ...SetOptionFunc) error   { return nil }
func (*testCache) Delete(context.Context, string) error                       { return nil }
func (*testCache) Close() error                                               { return nil }

func TestClient(t *testing.T) {
	defaultClient := &testCache{}
	namedClient := &testCache{}
	Register("default", defaultClient)
	Register("test-named", namedClient)

	testCases := []struct {
		name string
		args []string
		want Cache
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
			if got := Client(tc.args...); got != tc.want {
				t.Fatalf("Client() = %p, want %p", got, tc.want)
			}
		})
	}
}

func TestClientUnknownPanics(t *testing.T) {
	defer func() {
		want := `cache: "missing" is not registered`
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()

	Client("missing")
}

func TestRegister(t *testing.T) {
	first := &testCache{}
	second := &testCache{}

	Register("test-replace", first)
	Register("test-replace", second)

	if got := Client("test-replace"); got != second {
		t.Fatalf("Client() = %p, want replacement %p", got, second)
	}
}

func TestRegisterPanics(t *testing.T) {
	testCases := []struct {
		name       string
		clientName string
		client     Cache
		want       string
	}{
		{
			name:   "empty name",
			client: &testCache{},
			want:   "cache: name is empty",
		},
		{
			name:       "nil client",
			clientName: "test-nil",
			want:       "cache: client is nil",
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
