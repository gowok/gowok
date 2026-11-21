package gowok

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-must/must"
	"github.com/gowok/fp/maps"
	"github.com/gowok/fp/some"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/json"
)

func TestNewConfig(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		expectedC := &config.Config{
			Web: config.Web{
				Enabled: true,
				Host:    ":8080",
				Cors: some.Of(config.WebCors{
					Enabled: true,
				}),
			},
		}

		yy, err := json.Marshal(expectedC)
		must.Nil(t, err)

		cMap, err := newConfigRaw(".json", string(yy))
		must.Nil(t, err)

		var c *config.Config
		err = maps.ToStruct(cMap, &c)
		must.Nil(t, err)
		must.NotNil(t, c)
		must.NotNil(t, cMap)

		must.Equal(t, c.Web, expectedC.Web)
	})

	t.Run("negative not found file", func(t *testing.T) {
		c, cMap, err := newConfig(time.Now().Format(time.RFC3339), "")
		must.NotNil(t, err)
		must.Nil(t, c)
		must.Nil(t, cMap)
	})

	t.Run("negative invalid config format", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "TestNewConfig*.yaml")
		must.Nil(t, err)
		defer func() {
			_ = os.Remove(tempFile.Name())
		}()

		_, err = tempFile.Write([]byte("app: *"))
		must.Nil(t, err)
		defer func() {
			_ = tempFile.Close()
		}()

		c, cMap, err := newConfig(tempFile.Name(), "")
		must.NotNil(t, err)
		must.Nil(t, c)
		must.Nil(t, cMap)
	})
}

func TestNewConfigRaw(t *testing.T) {
	tests := []struct {
		name      string
		filetype  string
		content   string
		wantErr   bool
		wantKey   string
		wantValue any
	}{
		{
			name:      "positive/json",
			filetype:  ".json",
			content:   `{"key": "value"}`,
			wantKey:   "key",
			wantValue: "value",
		},
		{
			name:      "positive/yaml",
			filetype:  ".yaml",
			content:   "key: value",
			wantKey:   "key",
			wantValue: "value",
		},
		{
			name:      "positive/yml",
			filetype:  ".yml",
			content:   "key: value",
			wantKey:   "key",
			wantValue: "value",
		},
		{
			name:      "positive/toml",
			filetype:  ".toml",
			content:   `key = "value"`,
			wantKey:   "key",
			wantValue: "value",
		},
		{
			name:     "negative/invalid json",
			filetype: ".json",
			content:  `{key: value}`,
			wantErr:  true,
		},
		{
			name:     "negative/invalid yaml",
			filetype: ".yaml",
			content:  "\tkey: value",
			wantErr:  true,
		},
		{
			name:     "negative/invalid toml",
			filetype: ".toml",
			content:  `key = value`,
			wantErr:  true,
		},
		{
			name:     "negative/unsupported filetype",
			filetype: ".conf",
			content:  "key: value",
			wantErr:  true,
		},
		{
			name:     "negative/empty filetype",
			filetype: "",
			content:  "key: value",
			wantErr:  true,
		},
		{
			name:      "positive/uppercase filetype",
			filetype:  ".YAML",
			content:   "key: value",
			wantKey:   "key",
			wantValue: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newConfigRaw(tt.filetype, tt.content)
			if tt.wantErr {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
				must.NotNil(t, got)
				must.Equal(t, tt.wantValue, got[tt.wantKey])
			}
		})
	}
}

func TestDiscoverConfig(t *testing.T) {
	oldFiles := configFiles
	defer func() { configFiles = oldFiles }()

	write := func(t *testing.T, dir, name string) string {
		t.Helper()
		filePath := filepath.Join(dir, name)
		must.Nil(t, os.MkdirAll(filepath.Dir(filePath), 0755))
		must.Nil(t, os.WriteFile(filePath, []byte(""), 0644))
		return filePath
	}

	t.Run("negative/no candidates", func(t *testing.T) {
		ConfigFiles()
		must.Equal(t, "", configFilesDiscover(t.TempDir()))
	})

	t.Run("negative/no matching candidate", func(t *testing.T) {
		ConfigFiles("gowok.yaml")
		must.Equal(t, "", configFilesDiscover(t.TempDir()))
	})

	t.Run("positive/custom candidate", func(t *testing.T) {
		dir := t.TempDir()
		expected := write(t, dir, "app.yaml")
		ConfigFiles("app.yaml")

		must.Equal(t, expected, configFilesDiscover(dir))
	})

	t.Run("positive/uses candidate order", func(t *testing.T) {
		dir := t.TempDir()
		expected := write(t, dir, "app.local.yaml")
		write(t, dir, "app.yaml")
		ConfigFiles("app.local.yaml", "app.yaml")

		must.Equal(t, expected, configFilesDiscover(dir))
	})

	t.Run("positive/supports nested path", func(t *testing.T) {
		dir := t.TempDir()
		expected := write(t, dir, "config/app.yaml")
		ConfigFiles("config/app.yaml")

		must.Equal(t, expected, configFilesDiscover(dir))
	})

	t.Run("positive/supports absolute path", func(t *testing.T) {
		dir := t.TempDir()
		expected := write(t, dir, "app.yaml")
		ConfigFiles(expected)

		must.Equal(t, expected, configFilesDiscover(t.TempDir()))
	})

	t.Run("negative/ignores directories", func(t *testing.T) {
		dir := t.TempDir()
		must.Nil(t, os.Mkdir(filepath.Join(dir, "app.yaml"), 0755))
		ConfigFiles("app.yaml")

		must.Equal(t, "", configFilesDiscover(dir))
	})
}

func TestRunningUnderTest(t *testing.T) {
	must.True(t, runningUnderTest())
}
