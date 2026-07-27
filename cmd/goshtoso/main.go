// Command goshtoso extracts embedded Goshtoso assets and reports versions.
//
// Usage:
//
//	goshtoso -out=css/goshtoso-base.css   # extract compiled styles.css
//	goshtoso -theme -out=goshtoso-theme.css  # extract theme SOURCE for your own tailwind build
//	goshtoso -version                     # print goshtoso + tailwind versions
//	goshtoso -source-path                 # print the components dir to @source
//	goshtoso -init-brand-site=./my-site   # copy the static brand-site starter
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/araihu/goshtoso/assets"
)

func versionString(info *debug.BuildInfo, tailwind string) string {
	v := "(devel)"
	if info != nil && info.Main.Version != "" {
		v = info.Main.Version
	}
	return fmt.Sprintf("goshtoso %s (tailwindcss %s)", v, tailwind)
}

func sourcePath(moduleDir string) string {
	return filepath.Join(moduleDir, "components")
}

func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

func initBrandSite(module, destination string) error {
	source := filepath.Join(module, "examples", "brand-site")
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("brand-site template: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("brand-site template is not a directory: %s", source)
	}
	if _, err := os.Stat(destination); err == nil {
		entries, readErr := os.ReadDir(destination)
		if readErr != nil {
			return readErr
		}
		if len(entries) > 0 {
			return fmt.Errorf("destination is not empty: %s", destination)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return copyTree(source, destination)
}

// moduleDir resolves the installed goshtoso module directory via `go list`.
func moduleDir() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/araihu/goshtoso").Output()
	if err == nil {
		if dir := strings.TrimSpace(string(out)); dir != "" {
			return dir, nil
		}
	}

	// `go run github.com/araihu/goshtoso/cmd/goshtoso@latest` runs outside a
	// consumer module. Resolve its cached module copy so starter creation still
	// works before the consumer has added Goshtoso as a dependency.
	type download struct {
		Dir string
	}
	downloaded, downloadErr := exec.Command("go", "mod", "download", "-json", "github.com/araihu/goshtoso@latest").Output()
	if downloadErr == nil {
		var result download
		if json.Unmarshal(downloaded, &result) == nil && result.Dir != "" {
			return result.Dir, nil
		}
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(bytes.TrimSpace(ee.Stderr)) > 0 {
		return "", fmt.Errorf("resolve goshtoso module: %s", bytes.TrimSpace(ee.Stderr))
	}
	return "", fmt.Errorf("resolve goshtoso module: %w", err)
}

func main() {
	out := flag.String("out", "goshtoso-base.css", "output path for extracted CSS")
	theme := flag.Bool("theme", false, "extract the theme SOURCE (for your own Tailwind build) instead of compiled CSS")
	version := flag.Bool("version", false, "print goshtoso and tailwind versions, then exit")
	srcPath := flag.Bool("source-path", false, "print the components dir to feed Tailwind @source, then exit")
	initBrand := flag.String("init-brand-site", "", "copy the static brand-site starter into an empty directory")
	flag.Parse()

	if *version {
		info, _ := debug.ReadBuildInfo()
		fmt.Println(versionString(info, assets.TailwindVersion()))
		return
	}

	if *srcPath {
		dir, err := moduleDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "goshtoso: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(sourcePath(dir))
		return
	}

	if *initBrand != "" {
		dir, err := moduleDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "goshtoso: %v\n", err)
			os.Exit(1)
		}
		if err := initBrandSite(dir, *initBrand); err != nil {
			fmt.Fprintf(os.Stderr, "goshtoso: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("goshtoso: copied static brand-site starter to %s\n", *initBrand)
		return
	}

	var (
		data []byte
		err  error
	)
	if *theme {
		data, err = assets.ThemeCSS()
	} else {
		data, err = assets.StylesCSS()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "goshtoso: %v\n", err)
		os.Exit(1)
	}

	if dir := filepath.Dir(*out); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "goshtoso: mkdir %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
	if err := os.WriteFile(*out, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "goshtoso: write %s: %v\n", *out, err)
		os.Exit(1)
	}
	fmt.Printf("goshtoso: wrote %s (%d bytes)\n", *out, len(data))
}
