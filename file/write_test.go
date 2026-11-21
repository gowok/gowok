package file

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/golang-must/must"
)

func TestWrite(t *testing.T) {
	t.Run("positive/new file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		must.Nil(t, Write(path, []byte(`{"ok":true}`)))

		data, err := os.ReadFile(path)
		must.Nil(t, err)
		must.Equal(t, `{"ok":true}`, string(data))
	})

	t.Run("positive/overwrites existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")
		must.Nil(t, os.WriteFile(path, []byte("old content that is longer"), 0644))

		must.Nil(t, Write(path, []byte("new")))

		data, err := os.ReadFile(path)
		must.Nil(t, err)
		must.Equal(t, "new", string(data))
	})

	t.Run("positive/creates parent directories", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "nested", "deeper", "state.json")

		must.Nil(t, Write(path, []byte("content")))

		data, err := os.ReadFile(path)
		must.Nil(t, err)
		must.Equal(t, "content", string(data))
	})

	t.Run("positive/default permission", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		must.Nil(t, Write(path, []byte("content")))

		info, err := os.Stat(path)
		must.Nil(t, err)
		must.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})

	t.Run("positive/custom permission", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		must.Nil(t, Write(path, []byte("content"), 0600))

		info, err := os.Stat(path)
		must.Nil(t, err)
		must.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("positive/empty content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")

		must.Nil(t, Write(path, nil))

		info, err := os.Stat(path)
		must.Nil(t, err)
		must.Equal(t, int64(0), info.Size())
	})

	t.Run("negative/parent path is a file", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "blocker")
		must.Nil(t, os.WriteFile(blocker, []byte("file"), 0644))

		err := Write(filepath.Join(blocker, "state.json"), []byte("content"))
		must.NotNil(t, err)
	})

	t.Run("negative/destination is a directory", func(t *testing.T) {
		dir := t.TempDir()
		destination := filepath.Join(dir, "destination")
		must.Nil(t, os.Mkdir(destination, 0755))

		err := Write(destination, []byte("content"))
		must.NotNil(t, err)

		entries, err := os.ReadDir(dir)
		must.Nil(t, err)
		for _, entry := range entries {
			must.False(t, strings.HasPrefix(entry.Name(), ".destination.tmp-"))
		}
	})

	t.Run("positive/no temporary files left", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		must.Nil(t, Write(path, []byte("content")))

		entries, err := os.ReadDir(dir)
		must.Nil(t, err)
		must.Equal(t, 1, len(entries))
		must.Equal(t, "state.json", entries[0].Name())
	})
}

func TestWriteContext(t *testing.T) {
	t.Run("negative/cancelled before writing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := WriteContext(ctx, path, []byte("content"))
		must.True(t, errors.Is(err, context.Canceled))

		_, statErr := os.Stat(path)
		must.True(t, os.IsNotExist(statErr))

		entries, readErr := os.ReadDir(dir)
		must.Nil(t, readErr)
		must.Equal(t, 0, len(entries))
	})

	t.Run("negative/cancelled during writing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		remaining := int32(3)
		ctx := cancelAfterCtx{
			Context:   context.Background(),
			remaining: &remaining,
		}

		data := make([]byte, writeChunkSize*2)
		err := WriteContext(ctx, path, data)
		must.True(t, errors.Is(err, context.Canceled))

		_, statErr := os.Stat(path)
		must.True(t, os.IsNotExist(statErr))

		entries, readErr := os.ReadDir(dir)
		must.Nil(t, readErr)
		must.Equal(t, 0, len(entries))
	})

	t.Run("positive/nil context", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")

		must.Nil(t, WriteContext(nil, path, []byte("content")))

		data, err := os.ReadFile(path)
		must.Nil(t, err)
		must.Equal(t, "content", string(data))
	})

	t.Run("positive/large content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")
		data := make([]byte, writeChunkSize*2+17)
		for i := range data {
			data[i] = byte(i % 251)
		}

		must.Nil(t, WriteContext(context.Background(), path, data))

		read, err := os.ReadFile(path)
		must.Nil(t, err)
		must.Equal(t, data, read)
	})
}

type cancelAfterCtx struct {
	context.Context
	remaining *int32
}

func (c cancelAfterCtx) Err() error {
	if atomic.AddInt32(c.remaining, -1) < 0 {
		return context.Canceled
	}

	return nil
}
