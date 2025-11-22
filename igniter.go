package gowok

import (
	"context"
	"flag"
	"log"
	"log/slog"
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

	SQL.configure(Config.SQLs)
	if !Config.Forever {
		Config.Forever = Config.Web.Enabled || Config.Grpc.Enabled
	}

	_project(project)
	return project
}

func (p *Project) run() {
	Hooks.OnStarting()()

	if Config.Web.Enabled {
		go Web.configure()
		Health.Configure()
	}

	if Config.Grpc.Enabled {
		go GRPC.configure()
	}

	if Config.Net.Enabled {
		go Net.configure()
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
		if Config.Net.Enabled {
			slog.Info("stopping net")
			Net.Shutdown()
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
