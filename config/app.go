package config

type App struct {
	Rest Rest
}

type Rest struct {
	Enabled bool
	Host    string
}
