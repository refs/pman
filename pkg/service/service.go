package service

import (
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
	"syscall"
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

	_ = viper.BindEnv("keep-alive", "RUNTIME_KEEP_ALIVE")
	_ = viper.BindEnv("file", "RUNTIME_DB_FILE")

	cfg.KeepAlive = viper.GetBool("keep-alive")

	if viper.GetString("file") != "" {
		// remove tmp dir before overwriting to avoid stale tmp files.
		if err := os.Remove(cfg.File); err != nil {
			golog.Fatal(err)
		}

		cfg.File = viper.GetString("file")
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

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// shutdown controller if interrupted.
	go func(c controller.Controller) {
		rec := <-sigs
		s.Log.Debug().Str("service", "runtime service").Msgf("signal [%v] received. gracefully terminating children", rec.String())
		c.Shutdown(done)
		os.Exit(0)
	}(s.Controller)

	l, err := net.Listen("tcp", ":10666")
	if err != nil {
		s.Log.Fatal().Err(err)
	}
	s.Log.Info().Str("service", "runtime").Msg("runtime ready on localhost:10666")
	return http.Serve(l, nil)
}
