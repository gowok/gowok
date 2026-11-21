package logger

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/golang-must/must"
)

func TestNewFileWriterRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	writer := NewFileWriter(path, WithFileWriterRotation(10, 2))
	closer, ok := writer.(io.Closer)
	must.True(t, ok)
	defer func() { _ = closer.Close() }()

	write := func(value string) {
		t.Helper()
		_, err := writer.Write([]byte(value))
		must.Nil(t, err)
	}

	write("AAAA")
	write("BBBB")
	write("CCCC")
	write("DDDD")
	write("EEEE")

	content := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dir, name))
		must.Nil(t, err)
		return string(data)
	}

	must.Equal(t, "EEEE", content("app.log"))
	must.Equal(t, "CCCCDDDD", content("app.log.1"))
	must.Equal(t, "AAAABBBB", content("app.log.2"))
}

func TestNewFileWriterRotationNoBackups(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	writer := NewFileWriter(path, WithFileWriterRotation(10, 0))
	closer, ok := writer.(io.Closer)
	must.True(t, ok)
	defer func() { _ = closer.Close() }()

	_, err := writer.Write([]byte("AAAA"))
	must.Nil(t, err)
	_, err = writer.Write([]byte("BBBB"))
	must.Nil(t, err)
	_, err = writer.Write([]byte("CCCC"))
	must.Nil(t, err)

	data, err := os.ReadFile(path)
	must.Nil(t, err)
	must.Equal(t, "CCCC", string(data))

	_, err = os.Stat(path + ".1")
	must.True(t, os.IsNotExist(err))
}

func TestNewFileWriterRotationDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	writer := NewFileWriter(path, WithFileWriterRotation(0))
	defer func() {
		if closer, ok := writer.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	_, err := writer.Write([]byte("012345678901234567890123456789"))
	must.Nil(t, err)

	info, err := os.Stat(path)
	must.Nil(t, err)
	must.Equal(t, int64(30), info.Size())

	_, err = os.Stat(path + ".1")
	must.True(t, os.IsNotExist(err))
}

func TestNewFileWriterRotationFallback(t *testing.T) {
	dir := t.TempDir()

	writer := NewFileWriter(dir, WithFileWriterRotation(10))
	must.Equal(t, os.Stdout, writer)
}

func TestNewFileWriterRotationClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	writer := NewFileWriter(path, WithFileWriterRotation(10))
	closer := writer.(io.Closer)

	must.Nil(t, closer.Close())
	must.Nil(t, closer.Close())

	_, err := writer.Write([]byte("after close"))
	must.True(t, errors.Is(err, os.ErrClosed))
}

func TestNewFileWriterRotationConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")

	writer := NewFileWriter(path, WithFileWriterRotation(100, 100))
	closer := writer.(io.Closer)
	defer func() { _ = closer.Close() }()

	const goroutines = 8
	const writesPerGoroutine = 50

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range writesPerGoroutine {
				if _, err := writer.Write([]byte("x")); err != nil {
					errs <- err
					return
				}
			}
			errs <- nil
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		must.Nil(t, err)
	}

	files, err := filepath.Glob(path + "*")
	must.Nil(t, err)
	must.Equal(t, 4, len(files))

	var total int64
	for _, file := range files {
		info, err := os.Stat(file)
		must.Nil(t, err)
		total += info.Size()
	}

	must.Equal(t, int64(goroutines*writesPerGoroutine), total)
}
