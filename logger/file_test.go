package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-must/must"
)

func TestNewFileWriter(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name       string
		filepath   string
		setup      func(path string)
		wantStdout bool
	}{
		{
			name:       "valid path",
			filepath:   filepath.Join(tempDir, "test.log"),
			wantStdout: false,
		},
		{
			name:       "path is a directory",
			filepath:   tempDir,
			wantStdout: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(tc.filepath)
			}

			w := NewFileWriter(tc.filepath)
			must.NotNil(t, w)

			if tc.wantStdout {
				must.Equal(t, os.Stdout, w)
			} else {
				must.NotEqual(t, os.Stdout, w)
				// Clean up if it opened a file
				if f, ok := w.(*os.File); ok {
					must.Nil(t, f.Close())
				}
			}
		})
	}
}
