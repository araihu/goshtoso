//go:build !darwin && !linux && !windows

package iconpack

import (
	"fmt"
	"runtime"
)

func renameNoReplace(from, to string) error {
	return fmt.Errorf("atomic no-replace output publication unsupported on %s", runtime.GOOS)
}
