package controller

import (
	"github.com/refs/pman/pkg/process"
	"github.com/refs/pman/pkg/storage"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type janitor struct {
	// interval at which db is cleared.
	interval time.Duration

	store storage.Storage
}

func (j *janitor) run() {
	ticker := time.NewTicker(j.interval)
	work := make(chan os.Signal, 1)
	signal.Notify(work, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	for {
		select {
		case <-work:
			return
		case <-ticker.C:
			j.cleanup()
		}
	}
}

// cleanup removes orphaned extension + pid that were killed via SIGKILL given the nature of is being un-catchable,
// the only way to update pman's database is by polling.
func (j *janitor) cleanup() {
	for name, pid := range j.store.LoadAll() {
		// On unix like systems (linux, freebsd, etc) os.FindProcess will never return an error
		if p, err := os.FindProcess(pid); err == nil {
			if err := p.Signal(syscall.Signal(0)); err != nil {
				j.store.Delete(process.ProcEntry{
					Pid:       pid,
					Extension: name,
				})
			}
		}
	}
}
