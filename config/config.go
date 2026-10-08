package config

type Config struct {
	Key      string          `json:"key,omitempty"`
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
