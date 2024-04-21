package config

type SQL struct {
	Driver string
	DSN    string
	With   map[string]string
}

type MongoDB struct {
	DSN  string
	With map[string]string
}

type Cache struct {
	Driver string
	DSN    string
	With   map[string]string
}
