package gowok

import (
	"fmt"
	"io"
	"os"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/maps"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App      config.App
	Security config.Security        `json:"security,omitempty"`
	SQLs     map[string]config.SQL  `json:"sql,omitempty"`
	Http     map[string]config.Http `json:"http,omitempty"`
	Smtp     map[string]config.Smtp `json:"smtp,omitempty"`
	Others   map[string]any         `json:"others,omitempty"`

	EnvFile   string `json:"env_file,omitempty"`
	IsTesting bool   `json:"is_testing,omitempty"`
}

func newConfigEmpty() (*Config, map[string]any) {
	conf := &Config{
		config.App{},
		config.Security{},
		make(map[string]config.SQL),
		make(map[string]config.Http),
		make(map[string]config.Smtp),
		make(map[string]any),
		"", false,
	}

	confRaw, err := maps.FromStruct(conf)
	if err != nil {
		return conf, map[string]any{}
	}

	return conf, confRaw
}

func newConfig(pathConfig string, envFile string) (*Config, map[string]any, error) {
	fiConfig, err := os.OpenFile(pathConfig, os.O_RDONLY, 0600)
	if err != nil {
		return nil, nil, err
	}

	fiContent, err := io.ReadAll(fiConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read config file: %w", err)
	}

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, nil, fmt.Errorf("can't load .env file: %w", err)
		}
	}

	cfgS := os.ExpandEnv(string(fiContent))
	fiContent = []byte(cfgS)

	confRaw, err := newConfigRaw(string(fiContent))
	if err != nil {
		return nil, nil, err
	}

	conf := &Config{}
	err = maps.ToStruct(confRaw, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, confRaw, nil
}

func newConfigRaw(configString string) (map[string]any, error) {
	confRaw := map[string]any{}
	err := yaml.Unmarshal([]byte(configString), confRaw)
	if err == nil {
		return confRaw, nil
	}

	err = toml.Unmarshal([]byte(configString), &confRaw)
	if err == nil {
		return confRaw, nil
	}

	return nil, fmt.Errorf("can't decode config file: %w", err)
}
