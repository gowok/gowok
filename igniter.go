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
	"github.com/gowok/gowok/runtime"
	"github.com/gowok/gowok/singleton"
	"github.com/gowok/gowok/some"
	"github.com/gowok/gowok/web"
	"github.com/spf13/cobra"
)

type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	ConfigMap  map[string]any
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

func Get(config ...Config) *Project {
	pp := _project()
	if *pp != nil {
		return *pp
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

	project.Config = conf
	project.ConfigMap = confRaw
	project.runtime = runtime.New(
		runtime.WithRLimitEnabled(),
		runtime.WithGracefulStopFunc(project.stop(Hooks())),
	)

	SQL.Configure(project.Config.SQLs)
	Web.Configure(&project.Config.Web)
	if project.Config.Web.Enabled {
		Health.Configure()
	}
	if !project.Config.Forever {
		project.Config.Forever = project.Config.Web.Enabled || project.Config.Grpc.Enabled
	}

	pp = _project(project)
	return *pp
}

func run(project *Project) {
	Hooks().OnStarting()()

	if project.Config != nil {
		if project.Config.Web.Enabled {
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

		if project.Config.Grpc.Enabled {
			go func() {
				slog.Info("starting GRPC")
				listen, err := net.Listen("tcp", project.Config.Grpc.Host)
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
		if p.Config.Grpc.Enabled {
			slog.Info("stopping GRPC")
			GRPC.GracefulStop()
		}
		if p.Config.Web.Enabled {
			slog.Info("stopping web")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_ = web.Server().Shutdown(ctx)
		}
		hooks.OnStopped()()
	}
}

func (p *Project) Run() {
	p.runtime.AddRunFunc(func() {
		run(p)
	})
	p.runtime.Run(p.Config.Forever)
}

func Run(config ...Config) {
	p := Get(config...)
	p.runtime.AddRunFunc(func() {
		run(p)
	})
	p.runtime.Run(p.Config.Forever)
}

func (p *Project) Configures(configures ...ConfigureFunc) *Project {
	p.configures = append(p.configures, configures...)
	for _, configure := range configures {
		configure(p)
	}
	return p
}

func Configures(configures ...ConfigureFunc) *Project {
	p := Get()
	p.configures = append(p.configures, configures...)
	for _, configure := range configures {
		configure(p)
	}
	return p
}

var cmd = singleton.New(func() *cobra.Command {
	return &cobra.Command{}
})

func CMD(cmds ...*cobra.Command) *cobra.Command {
	return *cmd(cmds...)
}

func AddCMD(cmd ...*cobra.Command) {
	for _, c := range cmd {
		CMD().AddCommand(c)
	}
}

func WrapCMD(c *cobra.Command) *cobra.Command {
	flagParse()
	c.Flags().AddGoFlagSet(flag.CommandLine)
	return c
}
