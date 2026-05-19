// Command goshtoso extracts embedded Goshtoso assets to disk.
//
// Usage:
//
//	go tool goshtoso -out=css/goshtoso-base.css
//	go run github.com/araihu/goshtoso/cmd/goshtoso@latest -out=goshtoso.css
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/araihu/goshtoso/assets"
)

func main() {
	out := flag.String("out", "goshtoso-base.css", "output path for extracted CSS")
	flag.Parse()

	data, err := assets.StylesCSS()
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
