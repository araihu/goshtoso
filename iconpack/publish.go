package iconpack

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gofrs/flock"
)

func publishOutput(ctx context.Context, requested string, files map[string][]byte, check bool) (bool, string, error) {
	return publishFileOutput(ctx, requested, memoryFiles(files), check)
}

func memoryFiles(files map[string][]byte) map[string]fileData {
	result := make(map[string]fileData, len(files))
	for name, data := range files {
		result[name] = memoryFile(data)
	}
	return result
}

func publishFileOutput(ctx context.Context, requested string, files map[string]fileData, check bool) (bool, string, error) {
	location, err := resolveOutputLocation(requested)
	if err != nil {
		return false, "", err
	}
	lock, err := acquireOutputLock(ctx, location)
	if err != nil {
		return false, "", err
	}
	defer func() {
		_ = lock.Unlock()
		_ = lock.Close()
	}()

	identical, exists, err := compareFileOutput(ctx, location.output, files)
	if err != nil {
		return false, "", err
	}
	if check {
		if !exists || !identical {
			return false, "", fmt.Errorf("iconpack output %s is stale or absent", location.output)
		}
		return false, location.output, nil
	}
	if exists {
		if identical {
			return false, location.output, nil
		}
		return false, "", fmt.Errorf("output %s already exists with different or unrelated files; refusing to overwrite", location.output)
	}
	if err := stageFilesAndPublish(ctx, location, files, nil); err != nil {
		return false, "", err
	}
	return true, location.output, nil
}

type outputLocation struct {
	output string
	parent string
	base   string
}

func resolveOutputLocation(requested string) (outputLocation, error) {
	absOutput, err := filepath.Abs(requested)
	if err != nil {
		return outputLocation{}, fmt.Errorf("resolve output directory: %w", err)
	}
	location := outputLocation{output: absOutput, parent: filepath.Dir(absOutput), base: filepath.Base(absOutput)}
	if location.base == "." || location.base == string(os.PathSeparator) || location.base == "" {
		return outputLocation{}, fmt.Errorf("output directory must name one dedicated child directory")
	}
	parentInfo, err := os.Lstat(location.parent)
	if err != nil {
		return outputLocation{}, fmt.Errorf("inspect output parent: %w", err)
	}
	if !parentInfo.IsDir() {
		return outputLocation{}, fmt.Errorf("output parent must be an existing directory")
	}
	return location, nil
}

func acquireOutputLock(ctx context.Context, location outputLocation) (*flock.Flock, error) {
	lockPath := filepath.Join(location.parent, "."+location.base+".goshtoso-iconpack.lock")
	lock := flock.New(lockPath, flock.SetPermissions(0o600))
	lockContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	locked, err := lock.TryLockContext(lockContext, 100*time.Millisecond)
	if err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("acquire output lock: %w", err)
	}
	if !locked {
		_ = lock.Close()
		return nil, fmt.Errorf("output lock was not acquired")
	}
	return lock, nil
}

func stageAndPublishWithHook(location outputLocation, files map[string][]byte, beforeFinalize func() error) error {
	return stageFilesAndPublish(context.Background(), location, memoryFiles(files), beforeFinalize)
}

func stageFilesAndPublish(ctx context.Context, location outputLocation, files map[string]fileData, beforeFinalize func() error) error {
	staging, err := os.MkdirTemp(location.parent, "."+location.base+".tmp-*")
	if err != nil {
		return fmt.Errorf("create staged output directory: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(staging)
		}
	}()
	paths := make([]string, 0, len(files))
	for relative := range files {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	for _, relative := range paths {
		if err := writeStagedData(ctx, staging, relative, files[relative]); err != nil {
			return err
		}
	}
	if err := syncDirectory(staging); err != nil {
		return err
	}
	if beforeFinalize != nil {
		if err := beforeFinalize(); err != nil {
			return fmt.Errorf("prepare final output publication: %w", err)
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := renameNoReplace(staging, location.output); err != nil {
		return fmt.Errorf("atomically publish output directory without replacement: %w", err)
	}
	committed = true
	return nil
}

func writeStagedData(ctx context.Context, staging, relative string, contents fileData) error {
	if err := safeRelativePath(relative); err != nil {
		return fmt.Errorf("generated output path: %w", err)
	}
	target := filepath.Join(staging, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create staged output parent: %w", err)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create staged output %q: %w", relative, err)
	}
	if err := contents.copy(ctx, file); err != nil {
		_ = file.Close()
		return fmt.Errorf("write staged output %q: %w", relative, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync staged output %q: %w", relative, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close staged output %q: %w", relative, err)
	}
	return nil
}

func compareFileOutput(ctx context.Context, output string, expected map[string]fileData) (bool, bool, error) {
	info, err := os.Lstat(output)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("inspect output: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, true, fmt.Errorf("output exists and is not an owned real directory")
	}
	if err := validateOwnedMarker(output); err != nil {
		return false, true, err
	}
	count, identical := 0, true
	err = filepath.WalkDir(output, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == output {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("output contains symbolic link %q", path)
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("output contains non-regular file %q", path)
		}
		relative, err := filepath.Rel(output, path)
		if err != nil {
			return err
		}
		want, ok := expected[filepath.ToSlash(relative)]
		count++
		if !ok || info.Size() != want.size {
			identical = false
			return nil
		}
		matches, err := matchesFile(ctx, path, want)
		if err != nil {
			return err
		}
		if !matches {
			identical = false
		}
		return nil
	})
	if err != nil {
		return false, true, fmt.Errorf("inspect owned output: %w", err)
	}
	return identical && count == len(expected), true, nil
}

func validateOwnedMarker(output string) error {
	manifestBytes, err := os.ReadFile(filepath.Join(output, "manifest.json"))
	if err != nil {
		return fmt.Errorf("output exists without readable iconpack manifest; refusing to overwrite")
	}
	var marker struct {
		SchemaVersion int    `json:"schemaVersion"`
		Tool          string `json:"tool"`
	}
	if err := json.Unmarshal(manifestBytes, &marker); err != nil || marker.SchemaVersion != OutputSchemaVersion || marker.Tool != toolName {
		return fmt.Errorf("output manifest is not owned by %s; refusing to overwrite", toolName)
	}
	return nil
}

func matchesFile(ctx context.Context, path string, want fileData) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = file.Close() }()
	h := sha256.New()
	size, err := io.Copy(h, cancelReader{ctx, io.LimitReader(file, want.size+1)})
	if err != nil {
		return false, err
	}
	return size == want.size && fmt.Sprintf("%x", h.Sum(nil)) == want.hash, nil
}
