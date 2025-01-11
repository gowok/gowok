package gowok

import (
	"os"
	"testing"
	"time"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/some"
	"github.com/gowok/should"
	"gopkg.in/yaml.v3"
)

func TestNewConfig(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "TestNewConfig*.yaml")
		should.Nil(t, err)
		defer os.Remove(tempFile.Name())

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

		yy, err := yaml.Marshal(expectedC)
		_, err = tempFile.Write(yy)
		should.Nil(t, err)
		defer tempFile.Close()

		c, cMap, err := NewConfig(tempFile.Name())
		should.Nil(t, err)
		should.NotNil(t, c)
		should.NotNil(t, cMap)

		should.Equal(t, c.App.Web, expectedC.App.Web)
	})

	t.Run("negative not found file", func(t *testing.T) {
		c, cMap, err := NewConfig(time.Now().Format(time.RFC3339))
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

		c, cMap, err := NewConfig(tempFile.Name())
		should.NotNil(t, err)
		should.Nil(t, c)
		should.Nil(t, cMap)
	})
}
