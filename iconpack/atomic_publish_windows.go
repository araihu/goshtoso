//go:build windows

package iconpack

import "golang.org/x/sys/windows"

func renameNoReplace(from, to string) error {
	oldPath, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	newPath, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	// Omitting MOVEFILE_REPLACE_EXISTING makes an existing destination fail.
	return windows.MoveFileEx(oldPath, newPath, windows.MOVEFILE_WRITE_THROUGH)
}
