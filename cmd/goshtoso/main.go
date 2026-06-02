// Command goshtoso extracts embedded Goshtoso assets and reports versions.
//
// Usage:
//
//	goshtoso -out=css/goshtoso-base.css   # extract compiled styles.css
//	goshtoso -theme -out=goshtoso-theme.css  # extract theme SOURCE for your own tailwind build
//	goshtoso -version                     # print goshtoso + tailwind versions
//	goshtoso -source-path                 # print the components dir to @source
package main

import (
	"bytes"
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

// moduleDir resolves the installed goshtoso module directory via `go list`.
func moduleDir() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/araihu/goshtoso").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(bytes.TrimSpace(ee.Stderr)) > 0 {
			return "", fmt.Errorf("go list (run from a module that requires goshtoso): %s", bytes.TrimSpace(ee.Stderr))
		}
		return "", fmt.Errorf("go list (run from a module that requires goshtoso): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	out := flag.String("out", "goshtoso-base.css", "output path for extracted CSS")
	theme := flag.Bool("theme", false, "extract the theme SOURCE (for your own Tailwind build) instead of compiled CSS")
	version := flag.Bool("version", false, "print goshtoso and tailwind versions, then exit")
	srcPath := flag.Bool("source-path", false, "print the components dir to feed Tailwind @source, then exit")
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
