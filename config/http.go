package config

type Http struct {
	URL  string            `yaml:"url"`
	With map[string]string `yaml:"with"`
}
