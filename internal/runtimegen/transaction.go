package runtimegen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var (
	renameFile = os.Rename
	removeFile = os.Remove
)

type fileUpdate struct {
	path     string
	contents []byte
	mode     os.FileMode
}

type stagedFile struct {
	update fileUpdate
	stage  string
	backup string
	hadOld bool
}

func commitFileUpdates(updates []fileUpdate) error {
	ordered := append([]fileUpdate(nil), updates...)
	sort.Slice(ordered, func(i, j int) bool { return filepath.Clean(ordered[i].path) < filepath.Clean(ordered[j].path) })
	for index := 1; index < len(ordered); index++ {
		if filepath.Clean(ordered[index-1].path) == filepath.Clean(ordered[index].path) {
			return fmt.Errorf("duplicate destination %s", ordered[index].path)
		}
	}

	staged, err := stageFileUpdates(ordered)
	if err != nil {
		return err
	}
	defer cleanupStageFiles(staged)
	committed, err := applyStagedFiles(staged)
	if err != nil {
		return rollbackStagedFiles(staged, committed, err)
	}
	return cleanupBackupFiles(staged)
}

func stageFileUpdates(updates []fileUpdate) ([]stagedFile, error) {
	staged := make([]stagedFile, 0, len(updates))
	for _, update := range updates {
		if err := os.MkdirAll(filepath.Dir(update.path), 0o755); err != nil {
			cleanupStageFiles(staged)
			return nil, err
		}
		file, err := os.CreateTemp(filepath.Dir(update.path), ".runtimegen-stage-*")
		if err != nil {
			cleanupStageFiles(staged)
			return nil, err
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
			_ = removeFile(stage)
			cleanupStageFiles(staged)
			return nil, err
		}
		staged = append(staged, stagedFile{update: update, stage: stage})
	}
	return staged, nil
}

func applyStagedFiles(staged []stagedFile) (int, error) {
	for index := range staged {
		file := &staged[index]
		if _, err := os.Stat(file.update.path); err == nil {
			backup, createErr := os.CreateTemp(filepath.Dir(file.update.path), ".runtimegen-backup-*")
			if createErr != nil {
				return index, createErr
			}
			file.backup = backup.Name()
			if err := backup.Close(); err != nil {
				return index, err
			}
			if err := removeFile(file.backup); err != nil {
				return index, err
			}
			if err := renameFile(file.update.path, file.backup); err != nil {
				return index, err
			}
			file.hadOld = true
		} else if !os.IsNotExist(err) {
			return index, err
		}
		if err := renameFile(file.stage, file.update.path); err != nil {
			restoreErr := restoreStagedFile(file)
			return index, errors.Join(fmt.Errorf("replace %s: %w", file.update.path, err), restoreErr)
		}
		file.stage = ""
	}
	return len(staged), nil
}

func restoreStagedFile(file *stagedFile) error {
	if err := removeFile(file.update.path); err != nil && !os.IsNotExist(err) {
		if !file.hadOld {
			return fmt.Errorf("removing new file %s failed and no original existed: %w", file.update.path, err)
		}
		return fmt.Errorf("preserved recovery artifact %s after removing replacement %s failed: %w", file.backup, file.update.path, err)
	}
	if !file.hadOld {
		return nil
	}
	if err := renameFile(file.backup, file.update.path); err != nil {
		return fmt.Errorf("preserved recovery artifact %s after restoring %s failed: %w", file.backup, file.update.path, err)
	}
	file.backup = ""
	return nil
}

func rollbackStagedFiles(staged []stagedFile, committed int, cause error) error {
	rollbackErr := cause
	for index := committed - 1; index >= 0; index-- {
		rollbackErr = errors.Join(rollbackErr, restoreStagedFile(&staged[index]))
	}
	return rollbackErr
}

func cleanupStageFiles(staged []stagedFile) {
	for _, file := range staged {
		if file.stage != "" {
			_ = removeFile(file.stage)
		}
	}
}

func cleanupBackupFiles(staged []stagedFile) error {
	var cleanupErr error
	for index := range staged {
		file := &staged[index]
		if file.backup == "" {
			continue
		}
		if err := removeFile(file.backup); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("committed %s but preserved recovery artifact %s: %w", file.update.path, file.backup, err))
			continue
		}
		file.backup = ""
	}
	return cleanupErr
}
