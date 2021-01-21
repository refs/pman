package controller

import (
	"encoding/json"
	"fmt"
	"github.com/refs/pman/pkg/process"
	"io/ioutil"
)

// detach will try to restart processes on failures.
func detach(c *Controller) {
	func(c *Controller) {
		for {
			select {
			case proc := <- c.Terminated:
				if err := c.Start(proc); err != nil {
					c.log.Err(err)
				}
			}
		}
	}(c)
}

// loadDB loads pman db file from disk. It is not thread safe, and callers must synchronize access when calling.
func loadDB(file string) (map[string]int, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]int)
	if err := json.Unmarshal(contents, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// storedPID retrieves a managed process PID by its name.
func (c *Controller) storedPID(name string) (int, error) {
	c.m.Lock()
	entries, err := loadDB(c.Config.File)
	if err != nil {
		return 0, err
	}
	c.m.Unlock()

	pid, ok := entries[name]
	if !ok {
		return 0, nil
	}

	return pid, nil
}

// === DB Lifecycle Functions ===

// write a new entry to File.
func (c *Controller) write(pe process.ProcEntry) error {
	c.m.RLock()
	defer c.m.RUnlock()

	entries, err := loadDB(c.Config.File)
	if err != nil {
		return err
	}

	entries[pe.Extension] = pe.Pid

	bytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.Config.File, bytes, 0644)
}

// delete removes a managed process from db.
func (c *Controller) delete(name string) error {
	c.m.Lock()
	entries, err := loadDB(c.Config.File)
	if err != nil {
		return err
	}
	c.m.Unlock()

	_, ok := entries[name]
	if !ok {
		return fmt.Errorf("pid not found for extension: %v", name)
	}

	delete(entries, name)

	bytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.Config.File, bytes, 0644)
}