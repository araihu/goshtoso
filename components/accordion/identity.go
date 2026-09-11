package accordion

import (
	"fmt"
	"sync/atomic"
)

var identitySequence atomic.Uint64

// Implicit identities are render-local; addressable accordions should supply ID.
func implicitID() string { return fmt.Sprintf("goshtoso-accordion-%d", identitySequence.Add(1)) }
