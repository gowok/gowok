package gowok

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gowok/gowok/must"
	"github.com/gowok/gowok/router"
	"github.com/gowok/gowok/runner"
	"github.com/gowok/gowok/singleton"
	"github.com/gowok/gowok/some"
	"github.com/gowok/gowok/sql"
	"google.golang.org/grpc"
)

type getterByName[T any] func(name ...string) some.Some[T]
type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	ConfigMap  map[string]any
	Runner     *runner.Runner
	Hooks      *Hooks
	Validator  *Validator
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

	conf, confRaw, err := NewConfig(pathConfig)
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

	GRPC := singleton.New(func() *grpc.Server {
		return grpc.NewServer()
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

				_ = router.Server().Shutdown(ctx)
			}
			hooks.onStopped.IfPresent(func(f Hook) {
				f()
			})
		}),
	)
	project = &Project{
		Config:     conf,
		ConfigMap:  confRaw,
		Runner:     running,
		Hooks:      hooks,
		Validator:  validator,
		grpc:       GRPC,
		configures: make([]ConfigureFunc, 0),
	}
	return project, nil
}

var projectSingleton = singleton.New(func() *Project {
	return must.Must(ignite())
})

func Get() *Project {
	pp := projectSingleton()
	return *pp
}

func run(project *Project) {
	sql.Configure(project.Config.SQLs)
	router.Configure(&project.Config.App.Web)
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
		err := router.Server().ListenAndServe()
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

func (p *Project) GRPC() *grpc.Server {
	g := p.grpc()
	return *g
}

func (p *Project) Run(forever ...bool) {
	p.Runner.AddRunFunc(func() {
		run(p)
	})
	if p.Config.App.Web.Enabled || p.Config.App.Grpc.Enabled {
		forever = append([]bool{true}, forever...)
	}
	p.Runner.Run(forever...)
}

func (p *Project) Configures(configures ...ConfigureFunc) *Project {
	p.configures = make([]ConfigureFunc, len(configures))
	copy(p.configures, configures)
	return p
}

func (p *Project) Reload() {
	must.Must(ignite())
}
