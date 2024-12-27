package gowok

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gowok/gowok/driver"
	"github.com/gowok/gowok/must"
	"github.com/gowok/gowok/runner"
	"github.com/gowok/gowok/some"
	"github.com/ngamux/ngamux"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type getterByName[T any] func(name ...string) some.Some[T]
type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	Runner     *runner.Runner
	Hooks      *Hooks
	SQL        getterByName[*sql.DB]
	MongoDB    getterByName[*mongo.Client]
	Cache      getterByName[*cache.Cache[any]]
	Validator  *Validator
	web        func(...*HttpMux) **HttpMux
	grpc       func(...*grpc.Server) **grpc.Server
	configures []ConfigureFunc
}

var project *Project

func ignite() (*Project, error) {
	var pathConfig string
	if flag.Lookup("config") == nil {
		flag.StringVar(&pathConfig, "config", "config.yaml", "configuration file location (yaml)")
	} else {
		pathConfig = flag.Lookup("config").Value.String()
	}
	flag.Parse()

	conf, err := NewConfig(pathConfig)
	if err != nil {
		return nil, err
	}

	dbSQL, err := driver.NewSQL(conf.SQLs)
	if err != nil {
		return nil, err
	}

	dbMongo, err := driver.NewMongoDB(conf.MongoDBs)
	if err != nil {
		return nil, err
	}

	dbCache, err := driver.NewCache(conf.Caches)
	if err != nil {
		return nil, err
	}

	validator := NewValidator()
	en := en.New()
	uni := ut.New(en, en)
	trans, ok := uni.GetTranslator("en")
	if !ok {
		return nil, fmt.Errorf("validator: %w", ut.ErrUnknowTranslation)
	}

	err = en_translations.RegisterDefaultTranslations(validator.validate, trans)
	if err != nil {
		return nil, err
	}
	validator.trans = trans

	GRPC := Singleton(func() *grpc.Server {
		return grpc.NewServer()
	})
	web := Singleton(func() *HttpMux {
		return NewHTTP(&conf.App.Web)
	})

	hooks := &Hooks{}
	running := runner.New(
		runner.WithRLimitEnable(true),
		runner.WithGracefulStopFunc(func() {
			println()
			if conf.App.Grpc.Enabled {
				println("project: stopping GRPC")
				(*GRPC()).GracefulStop()
			}
			if conf.App.Web.Enabled {
				println("project: stopping web")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				_ = (*web()).Server.Shutdown(ctx)
			}
			hooks.onStopped.IfPresent(func(f Hook) {
				f()
			})
		}),
	)
	project = &Project{
		Config:     conf,
		Runner:     running,
		Hooks:      hooks,
		SQL:        dbSQL.Get,
		MongoDB:    dbMongo.Get,
		Cache:      dbCache.Get,
		Validator:  validator,
		web:        web,
		grpc:       GRPC,
		configures: make([]ConfigureFunc, 0),
	}
	return project, nil
}

var projectSingleton = Singleton(func() *Project {
	return must.Must(ignite())
})

func Get() *Project {
	pp := projectSingleton()
	return *pp
}

func run(project *Project) {
	for _, configure := range project.configures {
		configure(project)
	}

	project.Hooks.onStarting.IfPresent(func(f Hook) {
		f()
	})

	go func() {
		if !project.Config.App.Web.Enabled {
			return
		}

		println("project: starting web")
		err := (*project.web()).Server.ListenAndServe()
		if err != nil {
			log.Fatalf("web can't start, because: %v", err)
		}
	}()

	go func() {
		if !project.Config.App.Grpc.Enabled {
			return
		}

		println("project: starting GRPC")
		listen, err := net.Listen("tcp", project.Config.App.Grpc.Host)
		if err != nil {
			log.Fatalf("GRPC can't start, because: %v", err)
		}

		err = project.GRPC().Serve(listen)
		if err != nil {
			log.Fatalf("GRPC can't start, because: %v", err)
		}
	}()

	project.Hooks.onStarted.IfPresent(func(f Hook) {
		f()
	})
}

func (p *Project) Web() *ngamux.HttpServeMux {
	w := *p.web()
	return w.Mux
}
func (p *Project) GRPC() *grpc.Server {
	g := p.grpc()
	return *g
}

func (p *Project) Run() {
	p.Runner.AddRunFunc(func() {
		run(p)
	})
	p.Runner.Run()
}

func (p *Project) Configures(configures ...ConfigureFunc) {
	p.configures = make([]ConfigureFunc, len(configures))
	copy(p.configures, configures)
}

func (p *Project) Reload() {
	must.Must(ignite())
}
