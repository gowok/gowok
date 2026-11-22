package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-must/must"
)

func TestConfigFile(t *testing.T) {
	t.Run("positive/writes content", func(t *testing.T) {
		path := ConfigFile(t, "config.yaml", "web: {enabled: true}")

		data, err := os.ReadFile(path)
		must.Nil(t, err)
		must.Equal(t, "web: {enabled: true}", string(data))
	})

	t.Run("positive/creates nested directories", func(t *testing.T) {
		path := ConfigFile(t, "config/app.yml", "web: {enabled: true}")

		_, err := os.Stat(path)
		must.Nil(t, err)
		must.Equal(t, "app.yml", filepath.Base(path))
	})
}

func TestParseConfig(t *testing.T) {
	t.Run("positive/yaml", func(t *testing.T) {
		conf, err := ParseConfig("config.yaml", `
web:
  enabled: true
  host: ":8080"
sql:
  default:
    driver: postgres
    dsn: postgres://localhost
    max_open_conns: 10
`)
		must.Nil(t, err)
		must.True(t, conf.Web.Enabled)
		must.Equal(t, ":8080", conf.Web.Host)
		must.Equal(t, "postgres", conf.SQLs["default"].Driver)
		must.Equal(t, "postgres://localhost", conf.SQLs["default"].DSN)
		must.Equal(t, 10, conf.SQLs["default"].MaxOpenConns)
	})

	t.Run("positive/json", func(t *testing.T) {
		conf, err := ParseConfig("config.json", `{"web":{"enabled":true,"host":":9090"}}`)
		must.Nil(t, err)
		must.True(t, conf.Web.Enabled)
		must.Equal(t, ":9090", conf.Web.Host)
	})

	t.Run("positive/toml", func(t *testing.T) {
		conf, err := ParseConfig("config.toml", `
[web]
enabled = true
host = ":7070"
`)
		must.Nil(t, err)
		must.True(t, conf.Web.Enabled)
		must.Equal(t, ":7070", conf.Web.Host)
	})

	t.Run("positive/expands environment", func(t *testing.T) {
		t.Setenv("TESTUTIL_HOST", ":6060")

		conf, err := ParseConfig("config.yaml", "web:\n  host: ${TESTUTIL_HOST}\n")
		must.Nil(t, err)
		must.Equal(t, ":6060", conf.Web.Host)
	})

	t.Run("negative/unsupported file type", func(t *testing.T) {
		conf, err := ParseConfig("config.conf", "web: {}")

		must.NotNil(t, err)
		must.Nil(t, conf)
	})

	t.Run("negative/malformed content", func(t *testing.T) {
		conf, err := ParseConfig("config.yaml", "\tweb: enabled")

		must.NotNil(t, err)
		must.Nil(t, conf)
	})
}
