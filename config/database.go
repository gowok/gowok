package config

import "time"

type SQL struct {
	Driver  string
	DSN     string
	Enabled bool
	With    map[string]string

	MaxOpenConns    int    `json:"max_open_conns,omitempty"`
	MaxIdleConns    int    `json:"max_idle_conns,omitempty"`
	ConnMaxLifetime string `json:"conn_max_lifetime,omitempty"`
	ConnMaxIdleTime string `json:"conn_max_idle_time,omitempty"`
}

func (s SQL) GetConnMaxLifetime() time.Duration {
	return parseDuration(s.ConnMaxLifetime)
}

func (s SQL) GetConnMaxIdleTime() time.Duration {
	return parseDuration(s.ConnMaxIdleTime)
}
