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
	"github.com/gowok/gowok/some"
)

type ConfigureFunc func(*Project)

type Project struct {
	configures []ConfigureFunc
	runtime    *runtime.Runtime
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

func flagParse() {
	if flag.Lookup("config") == nil {
		flag.StringVar(&Flags().Config, "config", "", "configuration file location (yaml, toml)")
	}
	if flag.Lookup("env-file") == nil {
		flag.StringVar(&Flags().EnvFile, "env-file", "", "env file location")
	}
}

var hooks = singleton.New(func() *runtime.Hooks {
	return &runtime.Hooks{
		Init: some.Empty[func()](),
	}
})

func Hooks() *runtime.Hooks {
	return *hooks()
}

var _project = singleton.New(func() *Project {
	return nil
})

func Configure(config ...config.Config) {
	p := *_project()
	if p != nil {
		return
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	Hooks().Init.OrElse(func() {
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
	if len(config) > 0 {
		conf = &config[0]
		confRaw = maps.FromStruct(conf)
	}

	Config = conf
	ConfigMap = confRaw
	project.runtime = runtime.New(
		runtime.WithRLimitEnabled(),
		runtime.WithGracefulStopFunc(project.stop(Hooks())),
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
}

func run(project *Project) {
	Hooks().OnStarting()()

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

	Hooks().OnStarted()()
}

func (p Project) stop(hooks *runtime.Hooks) func() {
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
		hooks.OnStopped()()
	}
}

func Run(config ...config.Config) {
	Configure(config...)
	p := *_project()
	p.runtime.AddRunFunc(func() {
		run(p)
	})
	p.runtime.Run(Config.Forever)
}

func Configures(configures ...ConfigureFunc) *Project {
	Configure()
	p := *_project()
	p.configures = append(p.configures, configures...)
	for _, configure := range configures {
		configure(p)
	}
	return p
}
