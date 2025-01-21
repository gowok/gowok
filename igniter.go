package gowok

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gowok/gowok/async"
	"github.com/gowok/gowok/grpc"
	"github.com/gowok/gowok/health"
	"github.com/gowok/gowok/router"
	"github.com/gowok/gowok/runner"
	"github.com/gowok/gowok/singleton"
	"github.com/gowok/gowok/sql"
)

type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	ConfigMap  map[string]any
	Hooks      *Hooks
	configures []ConfigureFunc
	runner     *runner.Runner
}

var flags = struct {
	Config  string
	EnvFile string
}{}

func flagParse() {
	flag.StringVar(&flags.Config, "config", "config.yaml", "configuration file location (yaml)")
	flag.StringVar(&flags.EnvFile, "env-file", ".env", "env file location (.env)")
	flag.Parse()
}

var project = singleton.New(func() *Project {
	flagParse()

	conf, confRaw, err := newConfig(flags.Config, flags.EnvFile)
	if err != nil {
		panic(err)
	}

	hooks := &Hooks{}
	running := runner.New(
		runner.WithRLimitEnabled(),
		runner.WithGracefulStopFunc(stop(conf, hooks)),
	)

	project := &Project{
		Config:     conf,
		ConfigMap:  confRaw,
		runner:     running,
		Hooks:      hooks,
		configures: make([]ConfigureFunc, 0),
	}
	sql.Configure(project.Config.SQLs)
	if project.Config.App.Web.Enabled {
		router.Configure(&project.Config.App.Web)
		health.Configure()
	}

	return project
})

func Get() *Project {
	pp := project()
	return *pp
}

func run(project *Project) {
	project.Hooks.onStarting.IfPresent(func(f func()) {
		f()
	})

	if project.Config.App.Web.Enabled {
		go func() {
			slog.Info("starting web")
			err := router.Server().ListenAndServe()
			if err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return
				}
				panic("failed to start web, because " + err.Error())
			}
		}()
	}

	if project.Config.App.Grpc.Enabled {
		go func() {
			slog.Info("starting GRPC")
			listen, err := net.Listen("tcp", project.Config.App.Grpc.Host)
			if err != nil {
				panic("failed to start GRPC, because: " + err.Error())
			}

			err = grpc.Server().Serve(listen)
			if err != nil {
				panic("failed to start GRPC, because: " + err.Error())
			}
		}()
	}

	project.Hooks.onStarted.IfPresent(func(f func()) {
		f()
	})
}

func stop(conf *Config, hooks *Hooks) func() {
	return func() {
		println()
		if conf.App.Grpc.Enabled {
			slog.Info("stopping GRPC")
			grpc.Server().GracefulStop()
		}
		if conf.App.Web.Enabled {
			slog.Info("stopping web")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_ = router.Server().Shutdown(ctx)
		}
		hooks.onStopped.IfPresent(func(f func()) {
			f()
		})
	}
}

func (p *Project) Run(forever ...bool) {
	p.runner.AddRunFunc(func() {
		run(p)
	})
	if p.Config.App.Web.Enabled || p.Config.App.Grpc.Enabled {
		forever = append([]bool{true}, forever...)
	}
	p.runner.Run(forever...)
}

func (p *Project) Configures(configures ...ConfigureFunc) *Project {
	p.configures = append(p.configures, configures...)
	tasks := make([]func() (any, error), len(configures))
	for i, configure := range configures {
		tasks[i] = func() (any, error) {
			configure(p)
			return struct{}{}, nil
		}
	}
	_, err := async.All(tasks...)
	if err != nil {
		panic(err)
	}
	return p
}
