package file

import (
	"os"
	"testing"

	"github.com/golang-must/must"
)

func TestExtentionToMime(t *testing.T) {
	testCases := []struct {
		extension string
		expected  string
	}{
		{".jpg", "image/jpeg"},
		{".png", "image/png"},
		{".pdf", "application/pdf"},
		{".invalid", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.extension, func(t *testing.T) {
			must.Equal(t, tc.expected, ExtentionToMime(tc.extension))
		})
	}
}

func TestMimeToExtension(t *testing.T) {
	testCases := []struct {
		mime     string
		expected string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"application/pdf", ".pdf"},
		{"invalid/mime", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.mime, func(t *testing.T) {
			result := MimeToExtension(tc.mime)
			if tc.mime == "image/jpeg" {
				if result != ".jpg" && result != ".jpeg" {
					t.Errorf("MimeToExtension() = %v, want .jpg or .jpeg", result)
				}
			} else {
				must.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetTypeFromBase64(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"positive/jpeg", "data:image/jpeg;base64,...", "jpeg"},
		{"positive/png", "data:image/png;base64,...", "png"},
		{"positive/pdf", "data:application/pdf;base64,...", "pdf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			must.Equal(t, tc.expected, GetTypeFromBase64(tc.input))
		})
	}
}

func TestGetMimeFromBase64(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"positive/jpeg", "data:image/jpeg;base64,...", "image/jpeg"},
		{"positive/png", "data:image/png;base64,...", "image/png"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			must.Equal(t, tc.expected, GetMimeFromBase64(tc.input))
		})
	}
}

func TestSaveBase64StringToFile(t *testing.T) {
	base64String := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="

	testCases := []struct {
		name     string
		path     string
		fileName string
		encoded  string
		wantErr  bool
		setup    func()
	}{
		{
			name:     "positive/save png",
			path:     "tmp_test",
			fileName: "test_image",
			encoded:  base64String,
			wantErr:  false,
		},
		{
			name:     "positive/save in nested dir",
			path:     "tmp_test/nested/dir",
			fileName: "test_nested",
			encoded:  base64String,
			wantErr:  false,
		},
		{
			name:     "negative/invalid path (file exists)",
			path:     "tmp_test/file_exists",
			fileName: "test_fail",
			encoded:  base64String,
			wantErr:  true,
			setup: func() {
				must.Nil(t, os.MkdirAll("tmp_test", 0755))
				must.Nil(t, os.WriteFile("tmp_test/file_exists", []byte("not a dir"), 0644))
			},
		},
		{
			name:     "negative/create file fails (dir exists)",
			path:     "tmp_test",
			fileName: "is_a_dir",
			encoded:  base64String, // will try to create tmp_test/is_a_dir.png
			wantErr:  true,
			setup: func() {
				must.Nil(t, os.MkdirAll("tmp_test/is_a_dir.png", 0755))
			},
		},
		{
			name:     "negative/invalid base64 content",
			path:     "tmp_test",
			fileName: "invalid_content",
			encoded:  "data:image/png;base64,invalid!base64",
			wantErr:  true,
		},
		{
			name:     "negative/mkdir failure",
			path:     "tmp_test/a_file/new_dir",
			fileName: "test",
			encoded:  base64String,
			wantErr:  true,
			setup: func() {
				must.Nil(t, os.MkdirAll("tmp_test", 0755))
				must.Nil(t, os.WriteFile("tmp_test/a_file", []byte("not a dir"), 0644))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.RemoveAll("tmp_test")
			defer os.RemoveAll("tmp_test")

			if tc.setup != nil {
				tc.setup()
			}

			filePath, err := SaveBase64StringToFile(tc.path, tc.fileName, tc.encoded)
			if tc.wantErr {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
				must.NotNil(t, filePath)

				// Verify file exists
				_, err = os.Stat(filePath)
				must.Nil(t, err)

				// Verify extension
				must.True(t, len(filePath) > 4)
				must.Equal(t, ".png", filePath[len(filePath)-4:])
			}
		})
	}
}
