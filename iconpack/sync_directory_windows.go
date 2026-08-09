//go:build windows

package iconpack

// Windows does not support syncing a directory through os.File. Every staged
// file is flushed before publication, and the no-replace rename is flushed by
// MoveFileEx.
func syncDirectory(string) error { return nil }
