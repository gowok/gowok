package config

type App struct {
	Key  string
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
