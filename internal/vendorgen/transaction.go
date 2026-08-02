package vendorgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type fileUpdate struct {
	path     string
	contents []byte
	mode     os.FileMode
}

// commitFileUpdates stages every byte before replacing any destination. It
// applies paths deterministically and restores all originals if a replacement
// fails. There is no portable multi-file atomic rename, so this provides a
// transactional all-old-or-all-new result at process completion.
func commitFileUpdates(updates []fileUpdate) error {
	ordered := append([]fileUpdate(nil), updates...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].path < ordered[j].path })
	type stagedFile struct {
		update fileUpdate
		stage  string
		backup string
		hadOld bool
	}
	staged := make([]stagedFile, 0, len(ordered))
	cleanup := func() {
		for _, file := range staged {
			_ = os.Remove(file.stage)
			_ = os.Remove(file.backup)
		}
	}
	defer cleanup()

	for _, update := range ordered {
		if err := os.MkdirAll(filepath.Dir(update.path), 0o755); err != nil {
			return err
		}
		file, err := os.CreateTemp(filepath.Dir(update.path), ".vendorgen-stage-*")
		if err != nil {
			return err
		}
		stage := file.Name()
		if _, err = file.Write(update.contents); err == nil {
			err = file.Chmod(update.mode)
		}
		if err == nil {
			err = file.Sync()
		}
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(stage)
			return err
		}
		staged = append(staged, stagedFile{update: update, stage: stage})
	}

	committed := 0
	rollback := func(cause error) error {
		var rollbackErr error
		for index := committed - 1; index >= 0; index-- {
			file := &staged[index]
			_ = os.Remove(file.update.path)
			if file.hadOld {
				rollbackErr = errors.Join(rollbackErr, os.Rename(file.backup, file.update.path))
			}
		}
		return errors.Join(cause, rollbackErr)
	}
	for index := range staged {
		file := &staged[index]
		if _, err := os.Stat(file.update.path); err == nil {
			backup, err := os.CreateTemp(filepath.Dir(file.update.path), ".vendorgen-backup-*")
			if err != nil {
				return rollback(err)
			}
			file.backup = backup.Name()
			if err := backup.Close(); err != nil {
				return rollback(err)
			}
			if err := os.Remove(file.backup); err != nil {
				return rollback(err)
			}
			if err := os.Rename(file.update.path, file.backup); err != nil {
				return rollback(err)
			}
			file.hadOld = true
		} else if !os.IsNotExist(err) {
			return rollback(err)
		}
		if err := os.Rename(file.stage, file.update.path); err != nil {
			if file.hadOld {
				_ = os.Rename(file.backup, file.update.path)
			}
			return rollback(fmt.Errorf("replace %s: %w", file.update.path, err))
		}
		committed++
	}
	return nil
}
