//go:build !darwin && !linux && !windows && !freebsd && !openbsd

package iconpack

import "fmt"

func renameNoReplace(string, string) error {
	return fmt.Errorf("atomic no-replace directory publication is unsupported on this platform")
}
