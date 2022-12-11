package gowok

import (
	"io/ioutil"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/exception"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App       config.App
	Databases []config.Database
	Messaging config.Messaging
	Security  config.Security
}

func Configure(filename ...string) (Config, error) {
	file := "gowok.yaml"
	if len(filename) > 0 {
		file = filename[0]
	}

	conf := &Config{}

	confFile, e := ioutil.ReadFile(file)
	if e != nil {
		return *conf, exception.ConfigNotFound
	}

	e = yaml.Unmarshal(confFile, conf)
	if e != nil {
		return *conf, exception.ConfigDecoding(e)
	}

	return *conf, nil
}
