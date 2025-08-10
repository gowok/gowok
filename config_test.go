package gowok

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/gowok/fp/maps"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/some"
	"github.com/gowok/should"
)

func TestNewConfig(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		expectedC := &Config{
			App: config.App{
				Web: config.Web{
					Enabled: true,
					Host:    ":8080",
					Cors: some.Of(config.WebCors{
						Enabled: true,
					}),
				},
			},
		}

		yy, err := json.Marshal(expectedC)
		should.Nil(t, err)

		cMap, err := newConfigRaw(string(yy))
		should.Nil(t, err)

		var c *Config
		err = maps.ToStruct(cMap, &c)
		should.Nil(t, err)
		should.NotNil(t, c)
		should.NotNil(t, cMap)

		should.Equal(t, c.App.Web, expectedC.App.Web)
	})

	t.Run("negative not found file", func(t *testing.T) {
		c, cMap, err := newConfig(time.Now().Format(time.RFC3339), "")
		should.NotNil(t, err)
		should.Nil(t, c)
		should.Nil(t, cMap)
	})

	t.Run("negative invalid config format", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "TestNewConfig*.yaml")
		should.Nil(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write([]byte("app: *"))
		should.Nil(t, err)
		defer tempFile.Close()

		c, cMap, err := newConfig(tempFile.Name(), "")
		should.NotNil(t, err)
		should.Nil(t, c)
		should.Nil(t, cMap)
	})
}
