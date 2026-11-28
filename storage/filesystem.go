package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type filesystem struct {
	root string
	mu   sync.RWMutex
}

func Filesystem(root string) *filesystem {
	if root == "" {
		panic("storage: root is empty")
	}
	return &filesystem{root: root}
}

var _ Storage = (*filesystem)(nil)

func (f *filesystem) resolve(key string) (string, error) {
	if key == "" {
		return "", errors.New("storage: empty key")
	}
	if strings.ContainsRune(key, '\x00') {
		return "", errors.New("storage: invalid key")
	}
	if filepath.IsAbs(key) {
		return "", errors.New("storage: absolute key")
	}

	cleaned := filepath.Clean(key)
	if cleaned == "." || cleaned == "" {
		return "", errors.New("storage: empty key")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", errors.New("storage: path traversal")
	}

	full := filepath.Join(f.root, cleaned)

	rel, err := filepath.Rel(f.root, full)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", errors.New("storage: path traversal")
	}

	return full, nil
}

func (f *filesystem) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path, err := f.resolve(key)
	if err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, ErrNotFound
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return append([]byte(nil), data...), nil
}

func (f *filesystem) Put(ctx context.Context, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path, err := f.resolve(key)
	if err != nil {
		return err
	}

	data := append([]byte(nil), value...)

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("storage: mkdir: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".gowok-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return err
	}

	return nil
}
