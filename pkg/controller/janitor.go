package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type janitor struct {
	m *sync.RWMutex
	db string
	interval time.Duration
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
			cleanup(j.db, j.m)
		}
	}
}

// cleanup removes orphaned extension + pid that were killed via SIGKILL given the nature of is being un-catchable,
// the only way to update pman's database is by polling.
func cleanup(f string, m *sync.RWMutex) {
	m.Lock()
	entries, err := loadDB(f)
	if err != nil {
		//return err
	}
	m.Unlock()

	m.RLock()
	for name, pid := range entries {
		// On unix like systems (linux, freebsd, etc) os.FindProcess will never return an error
		if p, err := os.FindProcess(pid); err == nil {
			if err := p.Signal(syscall.Signal(0)); err != nil {
				fmt.Printf("\ndeleting orphaned process: %v\n", pid)
				delete(entries, name)
			}
		}
	}

	bytes, err := json.Marshal(entries)
	if err != nil {
		//return err
	}

	_ = ioutil.WriteFile(f, bytes, 0644)
	m.RUnlock()
}