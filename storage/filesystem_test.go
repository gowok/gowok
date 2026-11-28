package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newFS(t *testing.T) *filesystem {
	t.Helper()
	return Filesystem(t.TempDir())
}

func TestFilesystemRoundTrip(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	key := "some/key.txt"
	value := []byte("hello world")

	if err := f.Put(ctx, key, value); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := f.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("Get = %q, want %q", got, value)
	}
}

func TestFilesystemGetMissing(t *testing.T) {
	f := newFS(t)
	_, err := f.Get(context.Background(), "does/not/exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestFilesystemEmptyKey(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	for _, op := range []string{"get", "put"} {
		t.Run(op, func(t *testing.T) {
			var err error
			switch op {
			case "get":
				_, err = f.Get(ctx, "")
			case "put":
				err = f.Put(ctx, "", []byte("x"))
			}
			if err == nil {
				t.Fatalf("expected error for empty key on %s", op)
			}
		})
	}
}

func TestFilesystemPathTraversal(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	badKeys := []string{
		"../escape",
		"../../etc/passwd",
		"/absolute/path",
		"..",
		"a/../../escape",
	}
	for _, key := range badKeys {
		t.Run("get/"+key, func(t *testing.T) {
			_, err := f.Get(ctx, key)
			if err == nil {
				t.Fatalf("expected error for key %q on Get", key)
			}
		})
		t.Run("put/"+key, func(t *testing.T) {
			err := f.Put(ctx, key, []byte("x"))
			if err == nil {
				t.Fatalf("expected error for key %q on Put", key)
			}
		})
	}
}

func TestFilesystemCreatesParentDirs(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	key := "a/b/c/deep.txt"
	if err := f.Put(ctx, key, []byte("deep")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := f.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "deep" {
		t.Fatalf("Get = %q, want %q", got, "deep")
	}
	// ensure parent dir exists on disk
	info, err := os.Stat(f.root + "/a/b/c")
	if err != nil || !info.IsDir() {
		t.Fatalf("parent dir not created: %v", err)
	}
}

func TestFilesystemClone(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	key := "k"
	original := []byte("hello")

	if err := f.Put(ctx, key, original); err != nil {
		t.Fatal(err)
	}
	original[0] = 'X'

	got, err := f.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("stored value changed by input mutation: %q", got)
	}

	got[0] = 'Y'
	got2, err := f.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got2) != "hello" {
		t.Fatalf("stored value changed by returned slice mutation: %q", got2)
	}
}

func TestFilesystemPersistsAcrossInstances(t *testing.T) {
	root := t.TempDir()
	f1 := Filesystem(root)
	ctx := context.Background()
	key := "persist.txt"
	value := []byte("persisted")

	if err := f1.Put(ctx, key, value); err != nil {
		t.Fatalf("Put: %v", err)
	}

	f2 := Filesystem(root)
	got, err := f2.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(value) {
		t.Fatalf("Get = %q, want %q", got, value)
	}
}

func TestFilesystemConcurrency(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "key"
			_ = f.Put(ctx, key, []byte("v"))
			_, _ = f.Get(ctx, key)
		}()
	}
	wg.Wait()
}

func TestFilesystemContextCanceled(t *testing.T) {
	f := newFS(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := f.Get(ctx, "k"); err != context.Canceled {
		t.Fatalf("Get err = %v, want context.Canceled", err)
	}
	if err := f.Put(ctx, "k", []byte("v")); err != context.Canceled {
		t.Fatalf("Put err = %v, want context.Canceled", err)
	}
}


func TestFilesystemEmptyRootPanics(t *testing.T) {
	defer func() {
		if got := recover(); got != "storage: root is empty" {
			t.Fatalf("panic = %v, want %q", got, "storage: root is empty")
		}
	}()
	Filesystem("")
}

func TestFilesystemResolveBranches(t *testing.T) {
	f := newFS(t)
	cases := []struct {
		key     string
		wantErr bool
	}{
		{"", true},
		{"a\x00b", true},
		{"/absolute", true},
		{".", true},
		{"./", true},
		{"..", true},
		{"../x", true},
		{"a/../../x", true},
		{"ok.txt", false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			_, err := f.resolve(tc.key)
			if tc.wantErr && err == nil {
				t.Fatalf("resolve(%q) expected error", tc.key)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("resolve(%q) unexpected error: %v", tc.key, err)
			}
		})
	}
}

func TestFilesystemGetDirectoryReturnsNotFound(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	if err := os.MkdirAll(filepath.Join(f.root, "adir"), 0755); err != nil {
		t.Fatal(err)
	}
	_, err := f.Get(ctx, "adir")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestFilesystemGetStatPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root")
	}
	f := newFS(t)
	ctx := context.Background()
	dir := filepath.Join(f.root, "secret")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	_, err := f.Get(ctx, "secret/f.txt")
	if err == nil {
		t.Fatal("expected permission error")
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want permission error not ErrNotFound", err)
	}
}

func TestFilesystemGetReadPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root")
	}
	f := newFS(t)
	ctx := context.Background()
	path := filepath.Join(f.root, "locked.txt")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0644) })

	_, err := f.Get(ctx, "locked.txt")
	if err == nil {
		t.Fatal("expected permission error")
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want permission error not ErrNotFound", err)
	}
}

func TestFilesystemPutMkdirAllFails(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	if err := os.WriteFile(filepath.Join(f.root, "block"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	err := f.Put(ctx, "block/sub.txt", []byte("v"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFilesystemPutCreateTempFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root")
	}
	f := newFS(t)
	ctx := context.Background()
	dir := filepath.Join(f.root, "rodir")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	err := f.Put(ctx, "rodir/file.txt", []byte("v"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFilesystemPutRenameFails(t *testing.T) {
	f := newFS(t)
	ctx := context.Background()
	if err := os.MkdirAll(filepath.Join(f.root, "targetdir"), 0755); err != nil {
		t.Fatal(err)
	}
	err := f.Put(ctx, "targetdir", []byte("v"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFilesystemContextDeadlineExceeded(t *testing.T) {
	f := newFS(t)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	if _, err := f.Get(ctx, "k"); err != context.DeadlineExceeded {
		t.Fatalf("Get err = %v, want context.DeadlineExceeded", err)
	}
	if err := f.Put(ctx, "k", []byte("v")); err != context.DeadlineExceeded {
		t.Fatalf("Put err = %v, want context.DeadlineExceeded", err)
	}
}
