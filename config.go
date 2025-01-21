package gowok

import (
	"fmt"
	"io"
	"os"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/maps"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App      config.App
	Security config.Security
	SQLs     map[string]config.SQL  `json:"sql"`
	Http     map[string]config.Http `json:"http"`
	Smtp     map[string]config.Smtp `json:"smtp"`
	Others   map[string]string      `json:"others"`

	EnvFile   string `json:"env_file"`
	IsTesting bool   `json:"is_testing"`
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
	if err != nil {
		return nil, fmt.Errorf("can't decode config file: %w", err)
	}

	return confRaw, nil
}
