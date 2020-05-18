package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/refs/pacman/pkg/process"
	"github.com/refs/pacman/pkg/watcher"
)

// Controller writes the current managed processes onto a file, or any ReadWrite.
type Controller struct {
	// File refers to the Controller database, where we keep the controller's status. It formats as json.
	File string

	// Bin is the ocis single binary name.
	Bin string

	// BinPath is the ocis single binary path withing the host machine.
	// The Controller needs to know the binary location in order to spawn new extensions.
	BinPath string
}

var (
	defaultFile = "/var/tmp/.pacman"
)

// NewController initializes a new controller.
func NewController(o ...Option) Controller {
	c := Controller{
		Bin:  "ocis",
		File: defaultFile,
	}

	opts := Options{}

	for _, f := range o {
		f(opts)
	}

	if opts.Bin != "" {
		c.Bin = opts.Bin
	}

	// Get binary location from $PATH lookup
	path, err := exec.LookPath(c.Bin)
	if err != nil {
		log.Fatal("oCIS binary not present on `$PATH`")
	}

	c.BinPath = path

	if _, err := os.Stat(defaultFile); err != nil {
		fmt.Printf("db file doesn't exist, creating one with contents: `{}`\n")
		ioutil.WriteFile(defaultFile, []byte("{}"), 0644)
	}

	return c
}

// Write a new entry to File.
func (c *Controller) Write(pe process.ProcEntry) error {
	fd, err := ioutil.ReadFile(c.File)
	if err != nil {
		return err
	}

	entries := make(map[string]int)
	json.Unmarshal(fd, &entries)

	entries[pe.Extension] = pe.Pid
	if err := c.writeEntries(entries); err != nil {
		return err
	}

	return nil
}

// Start and watches a process.
func (c *Controller) Start(pe process.ProcEntry) error {
	w := watcher.NewWatcher()
	if err := pe.Start(c.BinPath); err != nil {
		print("erroring")
		return err
	}

	if err := c.Write(pe); err != nil {
		return err
	}

	w.Follow(pe)

	return nil
}

// Kill a managed process.
func (c *Controller) Kill(ext *string) error {
	pid, err := c.pidFromName(ext)
	if err != nil {
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	fmt.Printf("killing `%v`...\n", pid)
	return p.Kill()
}

// Shutdown a running runtime.
func (c *Controller) Shutdown(ch chan struct{}) error {
	fd, err := ioutil.ReadFile(c.File)
	if err != nil {
		return err
	}

	entries := make(map[string]int)
	json.Unmarshal(fd, &entries)

	for cmd, pid := range entries {
		fmt.Printf("gracefully shutting down process `%v` with pid `[%v]`...\n", cmd, pid)
		// swallow errors
		p, _ := os.FindProcess(pid)
		p.Kill()
	}

	if err := c.Reset(); err != nil {
		return err
	}

	ch <- struct{}{}

	return nil
}

// List managed processes.
func (c *Controller) List() error {
	return nil
}

// Reset clears the db file.
func (c *Controller) Reset() error {
	return ioutil.WriteFile(defaultFile, []byte("{}"), 0644)
}

// pidFromName reads from controller's db for the extension name, and returns it's pid for the running process.
func (c *Controller) pidFromName(name *string) (int, error) {
	fd, err := ioutil.ReadFile(c.File)
	if err != nil {
		return 0, err
	}

	entries := make(map[string]int)
	json.Unmarshal(fd, &entries)

	pid, ok := entries[*name]
	if !ok {
		return 0, fmt.Errorf("pid for extension `%v` not found", *name)
	}

	delete(entries, *name)
	c.writeEntries(entries)

	return pid, nil
}

func (c *Controller) writeEntries(e map[string]int) error {
	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(c.File, bytes, 0644); err != nil {
		return err
	}

	return nil
}
