package gowok

import "github.com/gowok/gowok/base"

type Option func(*App)

func WithConfig(config *base.Config) Option {
	return func(app *App) {
		app.Config = config
	}
}
