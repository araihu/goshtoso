// Command themegen generates assets/goshtoso-theme.css — the self-contained
// Goshtoso theme source a consumer imports into their own Tailwind v4 build.
// Run from the repo root: go run ./cmd/themegen
package themegen

import (
	"fmt"
	"io"
	"os"
)

// Run regenerates assets/goshtoso-theme.css. It must be called from the repo root.
func Run(stdout io.Writer) error {
	mainCSS, err := os.ReadFile("css/main.css")
	if err != nil {
		return err
	}
	allThemes, err := os.ReadFile("all-themes.css")
	if err != nil {
		return err
	}
	codeblock, err := os.ReadFile("css/codeblock.css")
	if err != nil {
		return err
	}

	out := generateTheme(string(mainCSS), map[string]string{
		"all-themes.css": string(allThemes),
		"codeblock.css":  string(codeblock),
	})

	if err := os.WriteFile("assets/goshtoso-theme.css", []byte(out), 0o644); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "themegen: wrote assets/goshtoso-theme.css (%d bytes)\n", len(out))
	return err
}
