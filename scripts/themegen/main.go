// Command themegen generates assets/goshtoso-theme.css — the self-contained
// Goshtoso theme source a consumer imports into their own Tailwind v4 build.
// Run from the repo root: go run ./scripts/themegen
package main

import (
	"fmt"
	"os"
)

func main() {
	mainCSS, err := os.ReadFile("css/main.css")
	must(err)
	allThemes, err := os.ReadFile("all-themes.css")
	must(err)
	codeblock, err := os.ReadFile("css/codeblock.css")
	must(err)

	out := generateTheme(string(mainCSS), map[string]string{
		"all-themes.css": string(allThemes),
		"codeblock.css":  string(codeblock),
	})

	must(os.WriteFile("assets/goshtoso-theme.css", []byte(out), 0644))
	fmt.Printf("themegen: wrote assets/goshtoso-theme.css (%d bytes)\n", len(out))
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "themegen: %v\n", err)
		os.Exit(1)
	}
}
