package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gowok/fp/maps"
	"github.com/gowok/gowok/config"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func ConfigFile(t testing.TB, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("testutil: create config directory: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("testutil: write config file: %v", err)
	}

	return path
}

func ParseConfig(name string, content string) (*config.Config, error) {
	raw, err := configRaw(filepath.Ext(name), os.ExpandEnv(content))
	if err != nil {
		return nil, err
	}

	conf := &config.Config{}
	if err := maps.ToStruct(raw, conf); err != nil {
		return nil, fmt.Errorf("testutil: decode config: %w", err)
	}

	return conf, nil
}

func configRaw(filetype string, content string) (map[string]any, error) {
	raw := map[string]any{}
	var err error

	switch strings.ToLower(filetype) {
	case ".json":
		err = json.Unmarshal([]byte(content), &raw)
	case ".yaml", ".yml":
		err = yaml.Unmarshal([]byte(content), &raw)
	case ".toml":
		err = toml.Unmarshal([]byte(content), &raw)
	default:
		return nil, fmt.Errorf("testutil: unsupported config file type: %q", filetype)
	}
	if err != nil {
		return nil, fmt.Errorf("testutil: decode config: %w", err)
	}

	return raw, nil
}
