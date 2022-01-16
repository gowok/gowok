package base

import "fmt"

type DBDriver string

const (
	DriverMySQL      = "DriverMySQL"
	DriverPostgreSQL = "DriverPostgreSQL"
)

type AppConfig struct {
	Name string
	Port uint
	Host string
}

type DBConfig struct {
	Driver   string
	Host     string
	Port     int
	Username string
	Password string
	Database string

	Options map[string]string
}

func (db DBConfig) DSN() string {
	dsn := ""
	if db.Driver == DriverPostgreSQL {
		options := ""
		for key, val := range db.Options {
			options = fmt.Sprintf("%s %s=%s", options, key, val)
		}
		dsn = fmt.Sprintf(
			"host=%s port=%d username=%s password=%s database=%s %s",
			db.Host,
			db.Port,
			db.Username,
			db.Password,
			db.Database,
			options,
		)
	} else if db.Driver == DriverMySQL {
		options := ""
		for key, val := range db.Options {
			options = fmt.Sprintf("%s&%s=%s", options, key, val)
		}
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?%s",
			db.Username,
			db.Password,
			db.Host,
			db.Port,
			db.Database,
			options,
		)
	}

	return dsn
}

type Config struct {
	App *AppConfig
	DB  *DBConfig
}

func NewConfig() *Config {
	return &Config{
		App: &AppConfig{},
		DB:  &DBConfig{},
	}
}
