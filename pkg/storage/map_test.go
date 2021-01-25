package storage

import (
	"fmt"
	"github.com/refs/pman/pkg/process"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	loadStore()
	os.Exit(m.Run())
}

var (
	store = NewMapStorage()
)

func loadStore() {
	for i := 0; i < 20; i++ {
		store.Store(process.ProcEntry{
			Pid:       rand.Int(),
			Extension: fmt.Sprintf("extension-%s", strconv.Itoa(i)),
		})
	}
}

func TestLoadAll(t *testing.T) {
	all := store.LoadAll()
	assert.NotNil(t, all["extension-1"])
}

func TestDelete(t *testing.T) {
	store.Delete(process.ProcEntry{
		Extension: "extension-1",
	})
	all := store.LoadAll()
	assert.Zero(t, all["extension-1"])
}
