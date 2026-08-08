//go:build !windows

package iconpack

import (
	"fmt"
	"os"
)

func syncDirectory(directory string) error {
	stagingDir, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open staged output directory: %w", err)
	}
	syncErr := stagingDir.Sync()
	closeErr := stagingDir.Close()
	if syncErr != nil {
		return fmt.Errorf("sync staged output directory: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close staged output directory: %w", closeErr)
	}
	return nil
}
