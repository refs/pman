package watcher

import (
	"fmt"
	"log"
	"os"

	"github.com/refs/pman/pkg/process"
)

// Watcher watches a process.
type Watcher struct {
}

// NewWatcher initializes a watcher.
func NewWatcher() Watcher {
	return Watcher{}
}

// Follow a process until it dies.
func (w *Watcher) Follow(pe process.ProcEntry) {
	state := make(chan *os.ProcessState, 1)

	fmt.Printf("watching [%v]...\n", pe.Pid)
	go func() {
		ps, err := watch(pe.Pid)
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

// watch a process by its pid. This operation blocks.
func watch(pid int) (*os.ProcessState, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	return p.Wait()
}
