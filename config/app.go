package config

type App struct {
	Rest Rest
	Grpc Grpc
}

type Rest struct {
	Enabled bool
	Host    string
}

type Grpc struct {
	Enabled bool
	Host    string
}
