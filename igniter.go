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
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/runtime"
	"github.com/gowok/gowok/singleton"
)

type ConfigureFunc func()

type Project struct {
	configures []ConfigureFunc
	runtime    *runtime.Runtime
}

var _project = singleton.New(func() *Project {
	return nil
})

func configure(configs ...config.Config) *Project {
	p := *_project()
	if p != nil {
		return p
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	Hooks.Init.OrElse(func() {
		flagParse()
		flag.Parse()
	})()

	project := &Project{
		configures: make([]ConfigureFunc, 0),
		runtime:    runtime.New(),
	}

	conf, confRaw := newConfigEmpty()
	if Flags().Config != "" {
		_conf, _confRaw, err := newConfig(Flags().Config, Flags().EnvFile)
		if err != nil {
			log.Fatalln(err)
		}
		conf = _conf
		confRaw = _confRaw
	}
	if len(configs) > 0 {
		conf = &configs[0]
		confRaw = maps.FromStruct(conf)
	}

	Config = conf
	config.ConfigMap = confRaw
	project.runtime = runtime.New(
		runtime.WithRLimitEnabled(),
		runtime.WithGracefulStopFunc(stop()),
	)

	SQL.Configure(Config.SQLs)
	Web.Configure(&Config.Web)
	if Config.Web.Enabled {
		Health.Configure()
	}
	if !Config.Forever {
		Config.Forever = Config.Web.Enabled || Config.Grpc.Enabled
	}

	_project(project)
	return project
}

func (p *Project) run() {
	Hooks.OnStarting()()

	if Config != nil {
		if Config.Web.Enabled {
			go func() {
				slog.Info("starting web")
				err := Web.Server.ListenAndServe()
				if err != nil {
					if errors.Is(err, http.ErrServerClosed) {
						return
					}
					log.Fatalln("web: failed to start: " + err.Error())
				}
			}()
		}

		if Config.Grpc.Enabled {
			go func() {
				slog.Info("starting GRPC")
				listen, err := net.Listen("tcp", Config.Grpc.Host)
				if err != nil {
					log.Fatalln("grpc: failed to start: " + err.Error())
				}

				err = GRPC.Serve(listen)
				if err != nil {
					log.Fatalln("grpc: failed to start: " + err.Error())
				}
			}()
		}
	}

	Hooks.OnStarted()()
}

func stop() func() {
	return func() {
		println()
		if Config.Grpc.Enabled {
			slog.Info("stopping GRPC")
			GRPC.GracefulStop()
		}
		if Config.Web.Enabled {
			slog.Info("stopping web")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_ = Web.Server.Shutdown(ctx)
		}
		Hooks.OnStopped()()
	}
}

func (p *Project) Run(config ...config.Config) {
	p.runtime.AddRunFunc(p.run)
	p.runtime.Run(Config.Forever)
}

func Run(config ...config.Config) {
	p := configure(config...)
	p.Run()
}

func Configures(configures ...ConfigureFunc) *Project {
	configure()
	p := *_project()
	p.configures = append(p.configures, configures...)
	for _, configure := range configures {
		configure()
	}
	return p
}
