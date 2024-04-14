package gowok

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gowok/gowok/optional"
	"github.com/gowok/gowok/runner"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type getterByName[T any] func(name ...string) optional.Optional[T]
type ConfigureFunc func(*Project)

type Project struct {
	Config     *Config
	Runner     *runner.Runner
	Hooks      *Hooks
	SQL        getterByName[*gorm.DB]
	MongoDB    getterByName[*mongo.Client]
	Redis      getterByName[*redis.Client]
	Validator  *Validator
	Web        *fiber.App
	GRPC       *grpc.Server
	configures []ConfigureFunc
}

var project *Project

func Ignite() (*Project, error) {
	var pathConfig string
	flag.StringVar(&pathConfig, "config", "config.yaml", "configuration file location (yaml)")
	flag.Parse()

	conf, err := NewConfig(pathConfig)
	if err != nil {
		return nil, err
	}
	dbSQL, err := NewSQL(conf.Databases)
	if err != nil {
		return nil, err
	}

	dbMongo, err := NewMongoDB(conf.Databases)
	if err != nil {
		return nil, err
	}

	dbRedis, err := NewRedis(conf.Databases)
	if err != nil {
		return nil, err
	}

	validator := NewValidator()

	web := NewHTTP(&conf.App.Web)
	GRPC := grpc.NewServer()

	hooks := &Hooks{}
	running := runner.New(
		runner.WithRLimitEnable(true),
		runner.WithGracefulStopFunc(func() {
			println()
			if conf.App.Grpc.Enabled {
				println("project: stopping GRPC")
				GRPC.GracefulStop()
			}
			if conf.App.Web.Enabled {
				println("project: stopping web")
				web.ShutdownWithTimeout(10 * time.Second)
			}
			if hooks.onStopped != nil {
				hooks.onStopped()
			}
		}),
		runner.WithRunFunc(run),
	)
	project = &Project{
		Config:    conf,
		Runner:    running,
		Hooks:     hooks,
		SQL:       dbSQL.Get,
		MongoDB:   dbMongo.Get,
		Redis:     dbRedis.Get,
		Validator: validator,
		Web:       web,
		GRPC:      GRPC,

		configures: make([]ConfigureFunc, 0),
	}
	return project, nil
}

func Get() *Project {
	if project != nil {
		return project
	}

	Must(Ignite())
	return project
}

func run() {
	project := Get()

	for _, configure := range project.configures {
		configure(project)
	}

	if project.Hooks.onStarting != nil {
		project.Hooks.onStarting()
	}

	go func() {
		if !project.Config.App.Web.Enabled {
			return
		}

		println("project: starting web")
		err := project.Web.Listen(project.Config.App.Web.Host)
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

		err = project.GRPC.Serve(listen)
		if err != nil {
			log.Fatalf("GRPC can't start, because: %v", err)
		}
	}()

	if project.Hooks.onStarted != nil {
		project.Hooks.onStarted()
	}
}

func (p *Project) Run() {
	Get().Runner.Run()
}

func (p *Project) Configures(configures ...ConfigureFunc) {
	p.configures = make([]ConfigureFunc, len(configures))
	copy(p.configures, configures)
}
