package gowok

import (
	"io/ioutil"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/err"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App       config.App
	Database  config.Database
	Cache     config.Cache
	Messaging config.Messaging
}

func Configure(filename ...string) (Config, error) {
	file := "gowok.yaml"
	if len(filename) > 0 {
		file = filename[0]
	}

	conf := &Config{}

	confFile, e := ioutil.ReadFile(file)
	if e != nil {
		return *conf, err.ErrConfigNotFound
	}

	e = yaml.Unmarshal(confFile, conf)
	if e != nil {
		return *conf, err.ErrConfigDecoding(e)
	}

	return *conf, nil
}
