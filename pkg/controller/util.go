package controller

import (
	"encoding/json"
	"github.com/refs/pman/pkg/process"
	"io/ioutil"
	"fmt"
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

// loadDB loads pman db file from disk.
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

// delete removes a managed process from db.
func (c *Controller) delete(name string) error {
	c.m.Lock()
	entries, err := loadDB(c.cfg.File)
	if err != nil {
		return err
	}
	c.m.Unlock()

	_, ok := entries[name]
	if !ok {
		return fmt.Errorf("pid not found for extension: %v", name)
	}

	delete(entries, name)

	c.m.RLock()
	defer c.m.RUnlock()
	return c.writeEntries(entries)
}

// storedPID reads from controller's db for the extension name, and returns it's pid for the running process.
func (c *Controller) storedPID(name string) (int, error) {
	c.m.Lock()
	entries, err := loadDB(c.cfg.File)
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

func (c *Controller) writeEntries(e map[string]int) error {
	c.m.RLock()
	defer c.m.RUnlock()

	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.cfg.File, bytes, 0644)
}

// write a new entry to File.
func (c *Controller) write(pe process.ProcEntry) error {
	c.m.RLock()
	defer c.m.RUnlock()

	entries, err := loadDB(c.cfg.File)
	if err != nil {
		return err
	}

	entries[pe.Extension] = pe.Pid
	return c.writeEntries(entries)
}