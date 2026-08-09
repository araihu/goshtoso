//go:build freebsd || openbsd

package iconpack

import "fmt"

// These targets expose renameat but no portable atomic no-replace directory
// operation through x/sys. Refusing publication is safer than a check-then-
// rename fallback that could replace a concurrent destination.
func renameNoReplace(string, string) error {
	return fmt.Errorf("atomic no-replace directory publication is unsupported on this platform")
}
