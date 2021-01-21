package controller

import (
	"fmt"
	"github.com/refs/pman/pkg/config"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
	cfg *config.Config
	// Bin is the OCIS single binary name.
	Bin string
	// BinPath is the OCIS single binary path withing the host machine.
	// The Controller needs to know the binary location in order to spawn new extensions.
	BinPath string
	// Terminated facilitates communication from Watcher <-> Controller. Writes to this
	// channel WILL always attempt to restart the crashed process.
	Terminated chan process.ProcEntry
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
		m: &sync.RWMutex{},
		options: *opts,
		log: log.NewLogger(
			log.WithPretty(true),
		),
		cfg: opts.Config,
		Bin:  "ocis",
		Terminated: make(chan process.ProcEntry),
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

	if _, err := os.Stat(opts.Config.File); err != nil {
		c.log.Debug().Str("package", "controller").Msgf("setting up db")
		ioutil.WriteFile(opts.Config.File, []byte("{}"), 0644)
	}

	return c
}

// Start and watches a process.
func (c *Controller) Start(pe process.ProcEntry) error {
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
	w.Follow(pe, c.Terminated, c.options.Config.KeepAlive)

	once.Do(func() {
		j := janitor{
			c.m,
			c.cfg.File,
			time.Second,
		}

		go j.run()
		go detach(c)
	})
	return nil
}

// Kill a managed process.
// TODO(refs) this interface MUST also work with PIDs.
// Should a process managed by the runtime be allowed to be killed if the runtime is configured not to?
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
	// We cannot be sure when was the last write before a shutdown routine begins without a plan. "Plan" in the sense of
	// an argument with every extension that must be started. This way we could block access to the io.Writer the "db"
	// uses, and either halt and prevent main from forking more children, or let them all run and once the reader gets
	// freed, start shutting them all down, a.k.a reverse the process. For the time being a simple Sleep would ensure
	// that all children are spawned and the last writer has been executed. Alternatively proper synchronization can
	// be ensured with the combination of a set of extensions that must run and a wait group.
	time.Sleep(1 * time.Second)

	entries, err := loadDB(c.cfg.File)
	if err != nil {
		return err
	}

	for cmd, pid := range entries {
		c.log.Info().Str("package", "watcher").Msgf("gracefully terminating %v", cmd)
		p, _ := os.FindProcess(pid)
		if err := p.Kill(); err != nil {
			return err
		}
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
	entries, err := loadDB(c.cfg.File)
	if err != nil {
		c.log.Err(err).Msg(fmt.Sprintf("error loading file: %s", c.cfg.File))
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
	return os.Remove(c.cfg.File)
}
