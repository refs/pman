package service

import (
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/refs/pman/pkg/controller"
	"github.com/refs/pman/pkg/log"
	"github.com/refs/pman/pkg/process"
	"github.com/rs/zerolog"
)

// Service represents a RPC service.
// The controller manager the service's state. When an action on a service is required,
// this will read the PID from the DB file for the given extensions and act upon its PID.
type Service struct {
	Controller controller.Controller
	Log        zerolog.Logger
}

// NewService returns a configured service with a controller and a default logger.
func NewService(options ...log.Option) *Service {
	return &Service{
		Controller: controller.NewController(),
		Log:        log.NewLogger(options...),
	}
}

// Start a process
func (s *Service) Start(args process.ProcEntry, reply *int) error {
	if err := s.Controller.Start(args); err != nil {
		*reply = 1
		return err
	}

	*reply = 0
	return nil
}

// List running processes for the controller.
func (s *Service) List(args struct{}, reply *string) error {
	*reply = s.Controller.List()
	return nil
}

// Kill a process
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
