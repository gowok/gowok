package config

type SQL struct {
	Driver  string
	DSN     string
	Enabled bool
	With    map[string]string
}

type MongoDB struct {
	DSN     string
	Enabled bool
	With    map[string]string
}

type Cache struct {
	Driver  string
	DSN     string
	Enabled bool
	With    map[string]string
}
