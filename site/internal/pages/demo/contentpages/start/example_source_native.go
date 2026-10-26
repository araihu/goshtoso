//go:build !js

package startpages

import (
	"os"
	"path/filepath"
	"runtime"
)

// readExampleFile reads a source file from examples/getting-started/
func readExampleFile(name string) string {
	// Try relative to working directory first
	paths := []string{
		filepath.Join("examples", "getting-started", name),
	}

	// Also try relative to this source file
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
		paths = append(paths, filepath.Join(root, "examples", "getting-started", name))
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data)
		}
	}

	return "// Source file not found: " + name
}
