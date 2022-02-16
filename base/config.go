package base

import (
	"fmt"

	"github.com/spf13/viper"
)

type DBDriver string

const (
	DriverMySQL      DBDriver = "mysql"
	DriverPostgreSQL DBDriver = "postgresql"
)

type AppConfig struct {
	Name string
	Port uint
	Host string
}

type DBConfig struct {
	Driver   DBDriver
	Host     string
	Port     uint
	Username string
	Password string
	Name     string

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
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable %s",
			db.Host,
			db.Port,
			db.Username,
			db.Password,
			db.Name,
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
			db.Name,
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

func ConfigFromFile(configLocation string) *Config {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configLocation)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	config := &Config{
		App: &AppConfig{
			Name: viper.GetString("app.name"),
			Host: viper.GetString("app.host"),
			Port: viper.GetUint("app.port"),
		},
		DB: &DBConfig{
			Driver:   DBDriver(viper.GetString("db.driver")),
			Host:     viper.GetString("db.host"),
			Port:     viper.GetUint("db.port"),
			Username: viper.GetString("db.username"),
			Password: viper.GetString("db.password"),
			Name:     viper.GetString("db.name"),
			Options:  viper.GetStringMapString("db.options"),
		},
	}

	if config.App.Name == "" {
		config.App.Name = "Gowok"
	}
	if config.App.Host == "" {
		config.App.Host = "0.0.0.0"
	}
	if config.App.Port == 0 {
		config.App.Port = 8080
	}

	if config.DB.Driver == "" {
		panic("config: \"db.driver\" doesn't have value")
	}
	if config.DB.Host == "" {
		config.DB.Host = "localhost"
	}
	if config.DB.Port == 0 {
		if config.DB.Driver == DriverPostgreSQL {
			config.DB.Port = 5432
		} else if config.DB.Driver == DriverMySQL {
			config.DB.Port = 3306
		}
	}
	if config.DB.Username == "" {
		if config.DB.Driver == DriverPostgreSQL {
			config.DB.Username = "postgres"
		} else if config.DB.Driver == DriverMySQL {
			config.DB.Username = "root"
		}
	}
	if config.DB.Name == "" {
		panic("config: \"db.name\" doesn't have value")
	}

	return config
}
