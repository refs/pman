package watcher

import (
	golog "log"
	"os"

	"github.com/refs/pman/pkg/log"
	"github.com/refs/pman/pkg/process"
	"github.com/rs/zerolog"
)

// Watcher watches a process.
type Watcher struct {
	log zerolog.Logger
}

// NewWatcher initializes a watcher.
func NewWatcher() Watcher {
	return Watcher{
		log: log.NewLogger(log.WithPretty(true)),
	}
}

// Follow a process until it dies.
func (w *Watcher) Follow(pe process.ProcEntry) {
	state := make(chan *os.ProcessState, 1)

	w.log.Debug().Str("package", "watcher").Msgf("watching %v", pe.Extension)
	go func() {
		ps, err := watch(pe.Pid)
		if err != nil {
			golog.Fatal(err)
		}

		state <- ps
	}()

	go func() {
		select {
		case status := <-state:
			w.log.Info().Str("package", "watcher").Msgf("%v exited with code: %v", pe.Extension, status.ExitCode())
		}
	}()
}

// watch a process by its pid. This operation blocks.
func watch(pid int) (*os.ProcessState, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	return p.Wait()
}
