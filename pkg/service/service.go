package service

import (
	"fmt"
	"github.com/refs/pman/pkg/config"
	"github.com/refs/pman/pkg/controller"
	"github.com/refs/pman/pkg/log"
	"github.com/refs/pman/pkg/process"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	golog "log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	halt = make(chan os.Signal, 1)
	done = make(chan struct{}, 1)
	finished = make(chan struct{}, 1)
)

// Service represents a RPC service.
// The controller manager the service's state. When an action on a service is required,
// this will read the PID from the DB file for the given extensions and act upon its PID.
// This package would act as a root command of sorts, since PMAN having 2 operational modes, as a
// cli tool and a library.
type Service struct {
	Controller controller.Controller
	Log        zerolog.Logger
}

// loadFromEnv would set cmd global variables. This is a workaround spf13/viper since pman used as a library does not
// parse flags.
func loadFromEnv() *config.Config {
	cfg := config.NewConfig()
	viper.AutomaticEnv()

	viper.BindEnv("keep-alive", "RUNTIME_KEEP_ALIVE")
	viper.BindEnv("file", "RUNTIME_DB_FILE")
	viper.BindEnv("port", "RUNTIME_PORT")

	cfg.KeepAlive = viper.GetBool("keep-alive")

	if viper.GetString("file") != "" {
		// remove tmp dir before overwriting to avoid stale tmp files.
		if err := os.Remove(cfg.File); err != nil {
			golog.Fatal(err)
		}

		cfg.File = viper.GetString("file")
	}

	if viper.GetString("port") != "" {
		cfg.Port = viper.GetString("port")
	}

	return cfg
}

// NewService returns a configured service with a controller and a default logger.
// When used as a library, flags are not parsed, and in order to avoid introducing a global state with init functions
// calls are done explicitly to loadFromEnv().
// Since this is the public constructor, options need to be added, at the moment only logging options
// are supported in order to match the running OwnCloud services structured log.
func NewService(options ...log.Option) *Service {
	cfg := loadFromEnv()

	return &Service{
		Controller: controller.NewController(
			controller.WithConfig(cfg),
		),
		Log:        log.NewLogger(options...),
	}
}

// Start indicates the Service Controller to start a new supervised service as an OS thread.
func (s *Service) Start(args process.ProcEntry, reply *int) error {
	s.Log.Info().Str("service", args.Extension).Msgf("%v", "started")
	if err := s.Controller.Start(args); err != nil {
		*reply = 1
		return err
	}

	*reply = 0
	return nil
}

// List running processes for the Service Controller.
func (s *Service) List(args struct{}, reply *string) error {
	*reply = s.Controller.List()
	return nil
}

// Kill a supervised process by subcommand name.
// TODO this API is rather simple and prone to failure. Terminate a process by PID MUST be allowed.
func (s *Service) Kill(args *string, reply *int) error {
	if err := s.Controller.Kill(args); err != nil {
		*reply = 1
		return err
	}

	*reply = 0
	return nil
}

// Start an rpc service with a registered configurable Controller process.
func Start() error {
	s := NewService(log.WithPretty(true))

	if err := rpc.Register(s); err != nil {
		s.Log.Fatal().Err(err)
	}
	rpc.HandleHTTP()

	signal.Notify(halt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", s.Controller.Config.Hostname, s.Controller.Config.Port))
	if err != nil {
		s.Log.Fatal().Err(err)
	}

	// handle panic within the Service scope.
	defer func() {
		if r := recover(); r != nil {
			reason := strings.Builder{}
			// small root cause analysis
			if _, err := net.Dial("localhost", s.Controller.Config.Port); err != nil {
				reason.WriteString("runtime address already in use")
			}

			fmt.Println(reason.String())
		}
	}()

	go func() error {
		return http.Serve(l, nil)
	}()

	// block until all processes end
	for {
		select {
			case _ = <- finished:
				println("done!")
				return nil
			case _ = <- halt:
				s.Log.Debug().
					Str("service", "runtime service").
					Msgf("terminating with signal: %v", s)
				if err := s.Controller.Shutdown(done); err != nil {
					s.Log.Err(err)
				}
				finished <- struct{}{}
				os.Exit(0)
				return nil
		}
	}
}
