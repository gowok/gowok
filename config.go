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
	SQLs     map[string]config.SQL  `yaml:"sql"`
	Http     map[string]config.Http `yaml:"http"`
	Smtp     map[string]config.Smtp `yaml:"smtp"`
	Others   map[string]string      `yaml:"others"`

	IsTesting bool `yaml:"is_testing"`
	// Environtment string `yaml:"environtment"`
	EnvFile string `yaml:"env_file"`
}

func NewConfig(pathConfig string) (*Config, map[string]any, error) {
	fiConfig, err := os.OpenFile(pathConfig, os.O_RDONLY, 0600)
	if err != nil {
		return nil, nil, err
	}

	fiContent, err := io.ReadAll(fiConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read config file: %w", err)
	}

	confRaw := map[string]any{}
	err = yaml.Unmarshal(fiContent, confRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	envFile := maps.Get(confRaw, "env_file", "")
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, nil, fmt.Errorf("can't load .env file: %w", err)
		}
	}

	cfgS := os.ExpandEnv(string(fiContent))
	fiContent = []byte(cfgS)

	conf := &Config{}
	err = yaml.Unmarshal(fiContent, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, confRaw, nil
}
