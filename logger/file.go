package logger

import (
	"io"
	"os"
)

func NewFileWriter(filepath string) io.Writer {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return os.Stdout
	}

	return f
}
