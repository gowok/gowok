package gowok

import (
	"fmt"
	"io"

	"github.com/gowok/gowok/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App       config.App
	Databases []config.Database
	Messaging config.Messaging
	Security  config.Security
	Http      map[string]config.Http
	Smtp      map[string]config.Smtp

	IsTesting bool `yaml:"is_testing"`
}

type configFile interface {
	io.Reader
	io.Closer
}

func Configure(fi configFile, err error) (*Config, error) {
	if err != nil {
		return nil, fmt.Errorf("can't open config file: %w", err)
	}

	defer fi.Close()

	fiContent, err := io.ReadAll(fi)
	if err != nil {
		return nil, fmt.Errorf("can't read config file: %w", err)
	}

	conf := &Config{}
	err = yaml.Unmarshal(fiContent, conf)
	if err != nil {
		return conf, fmt.Errorf("can't decode config file: %w", err)
	}

	return conf, nil
}
