package gowok

import (
	"fmt"
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
type OnStart func(appConfig base.AppConfig)

func New(opts ...Option) *App {
	app := &App{
		Controllers: make(base.Controllers),
		Models:      make(base.Models),
		mux:         ngamux.NewNgamux(),
	}

	for _, opt := range opts {
		opt(app)
	}

	if app.Config == nil {
		app.Config = base.NewConfig()
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

func (app *App) GetController(controller base.Controller) (*base.Controller, bool) {
	return app.Controllers.Get(controller)
}

func (app *App) GetModel(model base.Model) (*base.Model, bool) {
	return app.Models.Get(model)
}

func (app *App) buildRoute() {
	for _, controller := range app.Controllers {
		controller.Route(app.mux)
	}
}

func (app *App) Start(onStarts ...OnStart) error {
	app.dbConnect()
	app.buildRoute()

	if len(onStarts) > 0 {
		for _, onStart := range onStarts {
			onStart(*app.Config.App)
		}
	} else {
		fmt.Printf("%s started at %s:%d\n", app.Config.App.Name, app.Config.App.Host, app.Config.App.Port)
	}

	return http.ListenAndServe(":8080", app.mux)
}
