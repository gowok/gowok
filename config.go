package gowok

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/gowok/fp/maps"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/json"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var Config = &config.Config{}

func newConfigEmpty() (*config.Config, map[string]any) {
	conf := &config.Config{
		SQLs:   make(map[string]config.SQL),
		Smtp:   make(map[string]config.Smtp),
		Others: make(map[string]any),
	}

	confRaw := maps.FromStruct(conf)
	return conf, confRaw
}

func newConfig(pathConfig string, envFile string) (*config.Config, map[string]any, error) {
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

	confRaw, err := newConfigRaw(path.Ext(fiConfig.Name()), string(fiContent))
	if err != nil {
		return nil, nil, err
	}

	conf := &config.Config{}
	err = maps.ToStruct(confRaw, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, confRaw, nil
}

func newConfigRaw(filetype string, configString string) (map[string]any, error) {
	confRaw := map[string]any{}
	var err error

	switch filetype {
	case ".json":
		err = json.Unmarshal([]byte(configString), &confRaw)
	case ".yaml", ".yml":
		err = yaml.Unmarshal([]byte(configString), &confRaw)
	case ".toml":
		err = toml.Unmarshal([]byte(configString), &confRaw)
	}
	if err != nil {
		return nil, err
	}

	return confRaw, nil
}
