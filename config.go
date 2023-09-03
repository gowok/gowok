package gowok

import (
	"fmt"
	"io"

	"github.com/gowok/gowok/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App            config.App
	Security       config.Security
	MessageBrokers map[string]config.MessageBroker `yaml:"message_brokers"`
	Databases      map[string]config.Database      `yaml:"databases"`
	Http           map[string]config.Http          `yaml:"http"`
	Smtp           map[string]config.Smtp          `yaml:"smtp"`

	IsTesting bool `yaml:"is_testing"`
}

type configFile interface {
	io.Reader
	io.Closer
}

func NewConfig(fi configFile, err error) (*Config, error) {
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
