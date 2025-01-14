package config

type Env struct {
	ImportFromDotEnv string `yaml:"importFromDotEnv"`
	UseOSEnv         bool   `yaml:"useOsEnv"`
}
