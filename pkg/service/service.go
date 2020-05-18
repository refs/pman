package service

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/refs/pacman/pkg/controller"
	"github.com/refs/pacman/pkg/process"
)

// Service represents a RPC service. It wraps a Controller.
type Service struct {
	Controller controller.Controller
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
func Start(opts ...controller.Option) error {
	c := controller.NewController()
	s := &Service{
		Controller: c,
	}

	if err := rpc.Register(s); err != nil {
		log.Fatal(err)
	}
	rpc.HandleHTTP()

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func(c controller.Controller) {
		_ = <-sigs
		fmt.Println("gracefully terminating children")
		c.Shutdown(done)
		os.Exit(0)
	}(c)

	// Publish Controller port onto a runtime file? Or use a preconfigured port?
	// Running with preconfigured port for the time being

	l, e := net.Listen("tcp", ":10666")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	return http.Serve(l, nil)
}
