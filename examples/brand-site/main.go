// Brand-site generates static HTML. Replace the content and CSS; retain the
// small build boundary so the site works on any static host.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/araihu/goshtoso/assets"
)

func build(out string) error {
	if err := os.MkdirAll(filepath.Join(out, "assets"), 0755); err != nil {
		return err
	}
	css, err := assets.StylesCSS()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "assets", "goshtoso.css"), css, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "assets", "brand.css"), []byte(BrandCSS()), 0644); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(out, "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()
	return Home().Render(context.Background(), file)
}

func main() {
	if err := build("public"); err != nil {
		fmt.Fprintf(os.Stderr, "brand-site: %v\n", err)
		os.Exit(1)
	}
}
