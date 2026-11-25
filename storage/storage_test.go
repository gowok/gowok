package storage

import (
	"context"
	"testing"
)

type testStorage struct{}

func (*testStorage) Get(context.Context, string) ([]byte, error) {
	return nil, nil
}

func (*testStorage) Put(context.Context, string, []byte) error {
	return nil
}

func TestClient(t *testing.T) {
	defaultClient := &testStorage{}
	namedClient := &testStorage{}
	Register("default", defaultClient)
	Register("test-named", namedClient)

	testCases := []struct {
		name string
		args []string
		want Storage
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
		want := `storage: "missing" is not registered`
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()

	Client("missing")
}

func TestRegister(t *testing.T) {
	first := &testStorage{}
	second := &testStorage{}

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
		client     Storage
		want       string
	}{
		{
			name:   "empty name",
			client: &testStorage{},
			want:   "storage: name is empty",
		},
		{
			name:       "nil client",
			clientName: "test-nil",
			want:       "storage: client is nil",
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
