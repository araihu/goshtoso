// worker-assets exports the exact HTTP bytes and cache policy used by the Go
// server, so the Workers ASSETS binding can serve files without invoking WASM.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"

	chartassets "github.com/araihu/goshtoso-charts/assets"
	libraryassets "github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/site/internal/server"
)

func main() {
	output := flag.String("out", "dist/public", "Worker static assets directory")
	flag.Parse()
	if err := export(*output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func export(output string) error {
	root, err := filepath.Abs("..")
	if err != nil {
		return err
	}
	paths, err := assetPaths(root)
	if err != nil {
		return err
	}
	handler := server.New(root)
	var headers strings.Builder
	rules := 0
	for _, path := range paths {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			return fmt.Errorf("export %s: HTTP %d", path, response.Code)
		}
		target := filepath.Join(output, filepath.FromSlash(strings.TrimPrefix(path, "/")))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, response.Body.Bytes(), 0o644); err != nil {
			return err
		}
		// Workers assets already revalidate aliases by default. Emit only policy
		// overrides; keeping rules below 100 avoids silent truncation by Cloudflare.
		cache := response.Header().Get("Cache-Control")
		if cache != "" && cache != libraryassets.RevalidateCacheControl {
			fmt.Fprintf(&headers, "%s\n  Cache-Control: %s\n", path, cache)
			rules++
		}
		if strings.HasSuffix(path, ".webmanifest") {
			fmt.Fprintf(&headers, "%s\n  Content-Type: application/manifest+json\n", path)
			rules++
		}

	}
	if rules > 100 {
		return fmt.Errorf("static export needs %d header rules; Cloudflare permits 100", rules)
	}
	return os.WriteFile(filepath.Join(output, "_headers"), []byte(headers.String()), 0o644)
}

func assetPaths(root string) ([]string, error) {
	paths := []string{
		"/favicon.ico", "/favicon.svg", "/favicon-96x96.png", "/apple-touch-icon.png",
		"/site.webmanifest", "/web-app-manifest-192x192.png", "/web-app-manifest-512x512.png",
		"/assets/icons/bootstrapicons/sprite.svg", "/assets/icons/heroicons/sprite.svg",
		chartassets.RuntimeURL, chartassets.RuntimeLicenseURL, chartassets.RuntimeNoticeURL, chartassets.RuntimeD3LicenseURL,
		chartassets.WordCloudRuntimeURL, chartassets.WordCloudLicenseURL,
		chartassets.LiquidRuntimeURL, chartassets.LiquidLicenseURL,
		chartassets.ThreeDRuntimeURL, chartassets.ThreeDLicenseURL,
		chartassets.BrazilMapURL, chartassets.SaoPauloMapURL, chartassets.BrazilMapLicenseURL,
		chartassets.ChartIconsSpriteURL,
	}
	for version := 1; version <= 6; version++ {
		paths = append(paths, fmt.Sprintf("%sjs/controls/%d/controls.js", chartassets.Prefix, version))
	}
	for _, prefix := range []string{"componentdocshell", "consoleshell", "landingshell"} {
		names := []string{"shell.css", "shell.js"}
		if prefix == "componentdocshell" {
			names = append(names, "araihu.css", "goshtoso-logo.svg", "goshtoso-mark.svg", "goshtoso-mark-reverse.svg", "goshtoso-favicon.svg")
		}
		for _, name := range names {
			paths = append(paths, "/"+prefix+"/assets/"+name)
		}
	}
	if err := appendFiles(&paths, libraryassets.FS(), "/assets/", false); err != nil {
		return nil, err
	}
	if err := appendFiles(&paths, os.DirFS(filepath.Join(root, "site/assets/js")), "/site-assets/js/", true); err != nil {
		return nil, err
	}
	if err := appendFiles(&paths, os.DirFS(filepath.Join(root, "examples/getting-started/assets/dogs")), "/dog-images/", false); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func appendFiles(paths *[]string, files fs.FS, prefix string, minifiedOnly bool) error {
	return fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() && (!minifiedOnly || strings.HasSuffix(name, ".min.js")) {
			*paths = append(*paths, prefix+name)
		}
		return nil
	})
}
