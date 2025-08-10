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

	"github.com/gowok/fp/maps"
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
	if flag.Lookup("config") == nil {
		flag.StringVar(&Flags().Config, "config", "", "configuration file location (yaml, toml)")
	}
	if flag.Lookup("env-file") == nil {
		flag.StringVar(&Flags().EnvFile, "env-file", "", "env file location")
	}
}

var hooks = singleton.New(func() *runner.Hooks {
	return &runner.Hooks{
		Init: some.Empty[func()](),
	}
})

func Hooks() *runner.Hooks {
	return *hooks()
}

var _project = singleton.New(func() *Project {
	return nil
})

func Get(config ...Config) *Project {
	pp := _project()
	if *pp != nil {
		return *pp
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	Hooks().Init.OrElse(func() {
		FlagParse()
		flag.Parse()
	})()

	project := &Project{
		configures: make([]ConfigureFunc, 0),
		runner:     runner.New(),
	}

	var conf *Config
	confRaw := map[string]any{}
	if Flags().Config != "" {
		_conf, _confRaw, err := newConfig(Flags().Config, Flags().EnvFile)
		if err != nil {
			log.Fatalln(err)
		}
		conf = _conf
		confRaw = _confRaw
	} else if len(config) > 0 {
		conf = &config[0]
		confRaw = maps.FromStruct(conf)
	}

	if conf == nil {
		conf, confRaw = newConfigEmpty()
	}

	project.Config = conf
	project.ConfigMap = confRaw
	project.runner = runner.New(
		runner.WithRLimitEnabled(),
		runner.WithGracefulStopFunc(project.stop(Hooks())),
	)

	sql.Configure(project.Config.SQLs)
	if project.Config.App.Web.Enabled {
		router.Configure(&project.Config.App.Web)
		health.Configure()
	}
	pp = _project(project)
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

func (p Project) stop(hooks *runner.Hooks) func() {
	return func() {
		println()
		if p.Config.App.Grpc.Enabled {
			slog.Info("stopping GRPC")
			grpc.Server().GracefulStop()
		}
		if p.Config.App.Web.Enabled {
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
