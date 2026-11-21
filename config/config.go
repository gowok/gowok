package config

import "time"

type Config struct {
	ENV      string          `json:"env"`
	Net      Net             `json:"net"`
	Web      Web             `json:"web"`
	Security Security        `json:"security"`
	SQLs     map[string]SQL  `json:"sql,omitempty"`
	Smtp     map[string]Smtp `json:"smtp,omitempty"`
	Others   map[string]any  `json:"others,omitempty"`

	EnvFile   string `json:"env_file,omitempty"`
	IsTesting bool   `json:"is_testing,omitempty"`
	Forever   bool   `json:"-"`
}

var ConfigMap = map[string]any{}

func (c *Config) Map() map[string]any {
	return ConfigMap
}

func parseDuration(value string) time.Duration {
	if value == "" {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0
	}

	return duration
}
