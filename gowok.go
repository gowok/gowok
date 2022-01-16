package gowok

import (
	"net/http"

	"github.com/gowok/gowok/base"
	"github.com/ngamux/ngamux"
	"gorm.io/gorm"
)

var Config = &base.Config{}
var Models = base.Models{}

type App struct {
	Config      *base.Config
	Controllers base.Controllers
	Models      base.Models

	mux *ngamux.Ngamux
	db  *gorm.DB
}

func New() *App {
	app := &App{
		Config:      new(base.Config),
		Controllers: make(base.Controllers),
		Models:      make(base.Models),
		mux:         ngamux.NewNgamux(),
	}

	Config = app.Config

	return app
}

func (app *App) AddControllers(controllers ...base.Controller) {
	for _, controller := range controllers {
		app.Controllers.Add(controller)
	}

}

func (app *App) AddModels(models ...base.Model) {
	for _, model := range models {
		app.Models.Add(model)
	}

	Models = app.Models
}

func (app *App) GetController(controller base.Controller) (base.Controller, bool) {
	return app.Controllers.Get(controller)
}

func (app *App) GetModel(model base.Model) (base.Model, bool) {
	return app.Models.Get(model)
}

func (app *App) buildRoute() {
	for _, controller := range app.Controllers {
		controller.Route(app.mux)
	}
}

func (app *App) Start() error {
	app.dbConnect()
	app.buildRoute()
	return http.ListenAndServe(":8080", app.mux)
}
