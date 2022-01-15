package gowok

import (
	"net/http"

	"github.com/ngamux/ngamux"
)

type App struct {
	Config      *Config
	Controllers Controllers
	Models      Models

	mux *ngamux.Ngamux
}

func New() *App {
	return &App{
		Config:      new(Config),
		Controllers: make(Controllers),
		Models:      make(Models),
		mux:         ngamux.NewNgamux(),
	}
}

func (app *App) buildRoute() {
	for _, controller := range app.Controllers {
		controller.Route(app.mux)
	}
}

func (app *App) Start() error {
	app.buildRoute()
	return http.ListenAndServe(":8080", app.mux)
}
