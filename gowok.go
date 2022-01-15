package gowok

import (
	"net/http"

	"github.com/ngamux/ngamux"
	"gorm.io/gorm"
)

type App struct {
	Config      *Config
	Controllers Controllers
	Models      Models

	mux *ngamux.Ngamux
	db  *gorm.DB
}

func New() *App {
	return &App{
		Config:      new(Config),
		Controllers: make(Controllers),
		Models:      make(Models),
		mux:         ngamux.NewNgamux(),
	}
}

func (app *App) AddControllers(controllers ...Controller) {
	for _, controller := range controllers {
		controller.SetConfig(app.Config)
		controller.SetModels(app.Models)
		app.Controllers.Add(controller)
	}
}

func (app *App) AddModels(models ...Model) {
	for _, model := range models {
		app.Models.Add(model)
	}
}

func (app *App) GetController(controller Controller) (Controller, bool) {
	return app.Controllers.Get(controller)
}

func (app *App) GetModel(model Model) (Model, bool) {
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
