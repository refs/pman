// +build windows

package process

import (
	"os"
)

// ProcEntry is an entry in the File db.
type ProcEntry struct {
	Args      []string
	Pid       int
	Extension string
}

// NewProcEntry returns a new ProcEntry.
func NewProcEntry(extension string, args ...string) ProcEntry {
	return ProcEntry{
		Extension: extension,
		Args:      args,
	}
}

// Start a process.
func (e *ProcEntry) Start(binPath string) error {
	var argv = []string{binPath}
	argv = append(argv, e.Args...)

	p, err := os.StartProcess(binPath, argv, &os.ProcAttr{
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
		// TODO for security reasons we might want to set PGID on Windows.
	})
	if err != nil {
		return err
	}

	e.Pid = p.Pid

	return nil
}

// Kill the wrapped process.
func (e *ProcEntry) Kill() error {
	p, err := os.FindProcess(e.Pid)
	if err != nil {
		return err
	}

	return p.Kill()
}
