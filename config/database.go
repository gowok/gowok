package config

type SQL struct {
	Driver  string
	DSN     string
	Enabled bool
	With    map[string]string
}
