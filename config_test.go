package gowok

import (
	"os"
	"testing"
	"time"

	"github.com/golang-must/must"
	"github.com/gowok/fp/maps"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/json"
	"github.com/gowok/gowok/some"
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
