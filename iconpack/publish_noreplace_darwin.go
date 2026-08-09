//go:build darwin

package iconpack

import "golang.org/x/sys/unix"

func renameNoReplace(from, to string) error {
	return unix.RenamexNp(from, to, unix.RENAME_EXCL)
}
