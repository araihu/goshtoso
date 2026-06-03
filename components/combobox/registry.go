package combobox

import (
	"fmt"
	"sync"
)

var (
	registryMu  sync.Mutex
	registryIDs = make(map[string]struct{})
)

// registerID panics if id has already been registered. Callers pass
// the Config.ID at Handler construction to catch duplicate-id bugs at
// boot rather than in the browser.
func registerID(id string) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registryIDs[id]; dup {
		panic(fmt.Sprintf("combobox: duplicate Config.ID %q — IDs must be globally unique per process", id))
	}
	registryIDs[id] = struct{}{}
}

// resetRegistry is test-only.
func resetRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registryIDs = make(map[string]struct{})
}
