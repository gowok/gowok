package logger

import (
	"io"
	"os"
)

type FileWriterOption func(*fileWriterConfig)

type fileWriterConfig struct {
	rotation   bool
	maxSize    int64
	maxBackups int
}

func WithFileWriterRotation(maxSize int64, maxBackups ...int) FileWriterOption {
	backups := 3
	if len(maxBackups) > 0 {
		backups = maxBackups[0]
	}

	return func(c *fileWriterConfig) {
		c.rotation = maxSize > 0
		c.maxSize = maxSize
		c.maxBackups = backups
	}
}

func NewFileWriter(filepath string, opts ...FileWriterOption) io.Writer {
	cfg := &fileWriterConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.rotation {
		writer, err := newRotatingFileWriter(filepath, cfg.maxSize, cfg.maxBackups)
		if err != nil {
			return os.Stdout
		}

		return writer
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return os.Stdout
	}

	return f
}
