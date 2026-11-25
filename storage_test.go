package gowok

import (
	"context"
	"testing"

	"github.com/gowok/gowok/storage"
)

type testStorage struct{}

func (*testStorage) Get(context.Context, string) ([]byte, error) {
	return nil, nil
}

func (*testStorage) Put(context.Context, string, []byte) error {
	return nil
}

func TestStorage(t *testing.T) {
	defaultClient := &testStorage{}
	namedClient := &testStorage{}
	storage.Register("default", defaultClient)
	storage.Register("test-named", namedClient)

	testCases := []struct {
		name string
		args []string
		want storage.Storage
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
			if got := Storage(tc.args...); got != tc.want {
				t.Fatalf("Storage() = %p, want %p", got, tc.want)
			}
		})
	}
}

func TestStorageUnknownPanics(t *testing.T) {
	defer func() {
		want := `storage: "missing" is not registered`
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()

	Storage("missing")
}
