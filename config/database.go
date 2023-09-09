package config

type Database struct {
	Driver string
	DSN    string
	With   map[string]string
}
