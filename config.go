package gowok

import (
	"fmt"
	"io"
	"os"

	"github.com/gowok/gowok/config"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App      config.App
	Security config.Security
	ENV      config.Env             `yaml:"env"`
	SQLs     map[string]config.SQL  `yaml:"sql"`
	Http     map[string]config.Http `yaml:"http"`
	Smtp     map[string]config.Smtp `yaml:"smtp"`
	Others   map[string]string      `yaml:"others"`

	IsTesting    bool   `yaml:"is_testing"`
	Environtment string `yaml:"environtment"`
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

	conf := &Config{}
	err = yaml.Unmarshal(fiContent, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	if conf.ENV.ImportFromDotEnv != "" {
		if err := godotenv.Load(conf.ENV.ImportFromDotEnv); err != nil {
			return nil, nil, fmt.Errorf("can't load .env file: %w", err)
		}
		conf.ENV.UseOSEnv = true
	}

	if conf.ENV.UseOSEnv {

		cfgS := os.ExpandEnv(string(fiContent))
		fiContent = []byte(cfgS)
		err = yaml.Unmarshal(fiContent, conf)
		if err != nil {
			return nil, nil, fmt.Errorf("can't decode config file: %w", err)
		}

	}

	confRaw := map[string]any{}
	err = yaml.Unmarshal(fiContent, confRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, confRaw, nil
}
