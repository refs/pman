package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	golog "log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/refs/pman/pkg/log"
	"github.com/refs/pman/pkg/process"
	"github.com/refs/pman/pkg/watcher"
	"github.com/rs/zerolog"

	"github.com/olekukonko/tablewriter"
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

	log zerolog.Logger
}

var (
	defaultFile = "/var/tmp/.pman"
)

// NewController initializes a new controller.
func NewController(o ...Option) Controller {
	c := Controller{
		Bin:  "ocis",
		File: defaultFile,
		log: log.NewLogger(
			log.WithPretty(true),
		),
	}

	opts := &Options{}

	for _, f := range o {
		f(opts)
	}

	if opts.Bin != "" {
		c.Bin = opts.Bin
	}

	// Get binary location from $PATH lookup. If not present, it uses arg[0] as entry point.
	path, err := exec.LookPath(c.Bin)
	if err != nil {
		golog.Print("oCIS binary not present on `$PATH`")
		path = os.Args[0]
	}

	c.BinPath = path

	if _, err := os.Stat(defaultFile); err != nil {
		c.log.Info().Str("package", "watcher").Msgf("db file doesn't exist, creating one with contents: `{}`")
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
	return c.writeEntries(entries)
}

// Start and watches a process.
func (c *Controller) Start(pe process.ProcEntry) error {
	w := watcher.NewWatcher()
	if err := pe.Start(c.BinPath); err != nil {
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
	pid, err := c.pidFromName(*ext)
	if err != nil {
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	c.log.Info().Str("package", "watcher").Msgf("terminating %v", *ext)
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
		c.log.Info().Str("package", "watcher").Msgf("gracefully terminating %v", cmd)
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
func (c *Controller) List() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"Extension", "PID"})
	fd, err := ioutil.ReadFile(c.File)
	if err != nil {
		c.log.Fatal().Err(err)
	}

	entries := make(map[string]int)
	json.Unmarshal(fd, &entries)

	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, v := range keys {
		table.Append([]string{v, strconv.Itoa(entries[v])})
	}

	table.Render()
	return tableString.String()
}

// Reset clears the db file.
func (c *Controller) Reset() error {
	return ioutil.WriteFile(defaultFile, []byte("{}"), 0644)
}

// pidFromName reads from controller's db for the extension name, and returns it's pid for the running process.
func (c *Controller) pidFromName(name string) (int, error) {
	fd, err := ioutil.ReadFile(c.File)
	if err != nil {
		return 0, err
	}

	entries := make(map[string]int)
	json.Unmarshal(fd, &entries)

	pid, ok := entries[name]
	if !ok {
		return 0, fmt.Errorf("pid for extension `%v` not found", name)
	}

	delete(entries, name)
	c.writeEntries(entries)

	return pid, nil
}

func (c *Controller) writeEntries(e map[string]int) error {
	bytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.File, bytes, 0644)
}
