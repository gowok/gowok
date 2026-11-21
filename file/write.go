package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const writeChunkSize = 64 << 10

func Write(path string, data []byte, perm ...os.FileMode) error {
	return WriteContext(context.Background(), path, data, perm...)
}

func WriteContext(ctx context.Context, path string, data []byte, perm ...os.FileMode) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("file: write: %w", err)
	}

	mode := os.FileMode(0644)
	if len(perm) > 0 && perm[0] != 0 {
		mode = perm[0]
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("file: create directory: %w", err)
		}
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("file: write: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("file: create temp: %w", err)
	}

	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(mode); err != nil {
		return fmt.Errorf("file: chmod temp: %w", err)
	}

	for offset := 0; offset < len(data); {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("file: write: %w", err)
		}

		end := offset + writeChunkSize
		if end > len(data) {
			end = len(data)
		}

		n, err := tmp.Write(data[offset:end])
		if err != nil {
			return fmt.Errorf("file: write temp: %w", err)
		}
		if n == 0 {
			return fmt.Errorf("file: write temp: %w", io.ErrShortWrite)
		}
		offset += n
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("file: write: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("file: sync temp: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("file: close temp: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("file: write: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("file: rename temp: %w", err)
	}

	cleanup = false
	return nil
}
