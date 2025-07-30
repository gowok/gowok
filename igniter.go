package gowok

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gowok/gowok/grpc"
	"github.com/gowok/gowok/health"
	"github.com/gowok/gowok/router"
	"github.com/gowok/gowok/runner"
	"github.com/gowok/gowok/singleton"
	"github.com/gowok/gowok/some"
	"github.com/gowok/gowok/sql"
)

type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	ConfigMap  map[string]any
	configures []ConfigureFunc
	runner     *runner.Runner
}

type flags struct {
	Config  string
	EnvFile string
	Help    bool
}

var _flags = singleton.New(func() *flags {
	return &flags{}
})

func Flags() *flags {
	return *_flags()
}

func FlagParse() {
	flag.StringVar(&Flags().Config, "config", "config.yaml", "configuration file location (yaml, toml)")
	flag.StringVar(&Flags().EnvFile, "env-file", "", "env file location")
}

var hooks = singleton.New(func() *runner.Hooks {
	return &runner.Hooks{
		Init: some.Empty[func()](),
	}
})

func Hooks() *runner.Hooks {
	return *hooks()
}

var project = singleton.New(func() *Project {
	Hooks().Init.OrElse(func() {
		FlagParse()
		flag.Parse()
	})()

	project := &Project{
		configures: make([]ConfigureFunc, 0),
		runner:     runner.New(),
	}

	conf, confRaw, err := newConfig(Flags().Config, Flags().EnvFile)
	if err != nil {
		log.Fatalln(err)
	}
	project.Config = conf
	project.ConfigMap = confRaw
	project.runner = runner.New(
		runner.WithRLimitEnabled(),
		runner.WithGracefulStopFunc(stop(conf, Hooks())),
	)

	sql.Configure(project.Config.SQLs)
	if project.Config.App.Web.Enabled {
		router.Configure(&project.Config.App.Web)
		health.Configure()
	}

	return project
})

func Get() *Project {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	pp := project()
	return *pp
}

func run(project *Project) {
	Hooks().OnStarting()()

	if project.Config != nil {
		if project.Config.App.Web.Enabled {
			go func() {
				slog.Info("starting web")
				err := router.Server().ListenAndServe()
				if err != nil {
					if errors.Is(err, http.ErrServerClosed) {
						return
					}
					log.Fatalln("web: failed to start: " + err.Error())
				}
			}()
		}

		if project.Config.App.Grpc.Enabled {
			go func() {
				slog.Info("starting GRPC")
				listen, err := net.Listen("tcp", project.Config.App.Grpc.Host)
				if err != nil {
					log.Fatalln("grpc: failed to start: " + err.Error())
				}

				err = grpc.Server().Serve(listen)
				if err != nil {
					log.Fatalln("grpc: failed to start: " + err.Error())
				}
			}()
		}
	}

	Hooks().OnStarted()()
}

func stop(conf *Config, hooks *runner.Hooks) func() {
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
		hooks.OnStopped()()
	}
}

func (p *Project) Run(forever ...bool) {
	p.runner.AddRunFunc(func() {
		run(p)
	})
	if p.Config != nil {
		if p.Config.App.Web.Enabled || p.Config.App.Grpc.Enabled {
			forever = append([]bool{true}, forever...)
		}
	}
	p.runner.Run(forever...)
}

func (p *Project) Configures(configures ...ConfigureFunc) *Project {
	p.configures = append(p.configures, configures...)
	for _, configure := range configures {
		configure(p)
	}
	return p
}
