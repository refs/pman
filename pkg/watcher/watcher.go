package watcher

import (
	"fmt"
	"log"
	"os"

	"github.com/refs/pacman/pkg/process"
)

// Watcher watches a process.
type Watcher struct {
}

// NewWatcher initializes a watcher.
func NewWatcher() Watcher {
	return Watcher{}
}

// Watch watches a process.
// on most operating systems, the Process must be a child
// of the current process or an error will be returned.
func (w *Watcher) Watch(pid int) (*os.ProcessState, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	return p.Wait()
}

// Follow blocks watching a process until it exits.
func (w *Watcher) Follow(pe process.ProcEntry) {
	state := make(chan *os.ProcessState, 1)

	fmt.Printf("watching [%v]...\n", pe.Pid)
	go func() {
		ps, err := w.Watch(pe.Pid)
		if err != nil {
			log.Fatal(err)
		}

		state <- ps
	}()

	go func() {
		select {
		case status := <-state:
			fmt.Printf("process `%v` exited with code: `%v`\n", pe.Pid, status.ExitCode())
		}
	}()
}
