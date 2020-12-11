package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/refs/pman/pkg/log"
	"github.com/refs/pman/pkg/process"
	"github.com/refs/pman/pkg/watcher"
	"github.com/rs/zerolog"

	"github.com/olekukonko/tablewriter"
)

// Controller writes the current managed processes onto a file, or any ReadWrite.
type Controller struct {
	m *sync.RWMutex
	options Options
	log zerolog.Logger
	// File refers to the Controller database, where we keep the controller's status. It formats as json.
	File string
	// Bin is the ocis single binary name.
	Bin string
	// BinPath is the ocis single binary path withing the host machine.
	// The Controller needs to know the binary location in order to spawn new extensions.
	BinPath string
	// Terminated is a bidirectional channel that tallows communication from Watcher <-> Controller. Writes to this
	// channel will attempt to restart the crashed process.
	Terminated chan process.ProcEntry
	// restarted keeps an account of how many times a process has been restarted.
	restarted map[string]int
}

var (
	once = sync.Once{}
)

// NewController initializes a new controller.
func NewController(o ...Option) Controller {
	opts := &Options{}

	for _, f := range o {
		f(opts)
	}

	c := Controller{
		Bin:  "ocis",
		File: opts.File,
		Terminated: make(chan process.ProcEntry),
		log: log.NewLogger(
			log.WithPretty(true),
		),
		options: *opts,
		restarted: map[string]int{},
		m: &sync.RWMutex{},
	}

	if opts.Bin != "" {
		c.Bin = opts.Bin
	}

	// Get binary location from $PATH lookup. If not present, it uses arg[0] as entry point.
	path, err := exec.LookPath(c.Bin)
	if err != nil {
		c.log.Debug().Msg("OCIS binary not present in PATH, using Args[0]")
		path = os.Args[0]
	}

	c.BinPath = path

	if _, err := os.Stat(opts.File); err != nil {
		c.log.Debug().Str("package", "controller").Msgf("setting up db")
		ioutil.WriteFile(opts.File, []byte("{}"), 0644)
	}

	return c
}

// write a new entry to File.
func (c *Controller) write(pe process.ProcEntry) error {
	c.m.RLock()
	defer c.m.RUnlock()

	entries, err := loadDB(c.File)
	if err != nil {
		return err
	}

	entries[pe.Extension] = pe.Pid
	return c.writeEntries(entries)
}

// Start and watches a process.
func (c *Controller) Start(pe process.ProcEntry) error {
	// TODO add support for the same process running on different ports. a.k.a db entries as []string.
	var err error
	var pid int

	if pid, err = c.storedPID(pe.Extension); pid != 0 {
		c.log.Debug().Msg(fmt.Sprintf("extension already running: %s", pe.Extension))
		return nil
	}
	if err != nil {
		return err
	}

	w := watcher.NewWatcher()
	if err := pe.Start(c.BinPath); err != nil {
		return err
	}

	if err := c.write(pe); err != nil {
		return err
	}

	w.Follow(pe, c.Terminated, c.options.Restart)

	once.Do(func() {
		go detach(c)
	})

	return nil
}

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

// Kill a managed process.
func (c *Controller) Kill(ext *string) error {
	pid, err := c.storedPID(*ext)
	if err != nil {
		return err
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := c.delete(*ext); err != nil {
		return err
	}
	c.log.Info().Str("package", "watcher").Msgf("terminating %v", *ext)
	return p.Kill()
}

// Shutdown a running runtime.
func (c *Controller) Shutdown(ch chan struct{}) error {
	c.m.Lock()
	entries, err := loadDB(c.File)
	if err != nil {
		return err
	}
	c.m.Unlock()

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

	c.m.Lock()
	entries, err := loadDB(c.File)
	if err != nil {
		c.log.Err(err).Msg(fmt.Sprintf("error loading file: %s", c.File))
	}
	c.m.Unlock()

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
	c.m.RLock()
	defer c.m.RUnlock()
	return ioutil.WriteFile(c.File, []byte("{}"), 0644)
}

// delete removes a managed process from db.
func (c *Controller) delete(name string) error {
	c.m.Lock()
	entries, err := loadDB(c.File)
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
	entries, err := loadDB(c.File)
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

	return ioutil.WriteFile(c.File, bytes, 0644)
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