//go:build e2e && (full || iconpack)

package e2e

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type e2eIconpackRelease struct {
	Root              string
	Release           string
	CatalogSHA256     string
	ReleaseJSONSHA256 string
	ChecksumsSHA256   string
	Names             []string
}

func TestIconpackGeneratedConsumerBrowserProof(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	consumerURL := startIconpackConsumer(t)
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	t.Cleanup(func() {
		waitForPageSettled(t, page)
		failures.RequireEmpty(t)
	})

	_, err := page.Goto(consumerURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	root := page.Locator("[data-testid='generated-iconpack']")
	require.NoError(t, root.WaitFor())
	assert.Equal(t, "brand-developer-icons-tRPC", mustAttribute(t, root, "data-canonical-name"))
	assert.Equal(t, "devicon-trpc", mustAttribute(t, root, "data-symbol"))

	labelled := root.Locator("[data-testid='iconpack-labelled'] svg")
	assertConsumerSVGAttributes(t, labelled, "img", "tRPC", "")
	assert.Equal(t, "/assets/icons/appicons/sprite.svg#devicon-trpc", mustAttribute(t, labelled.Locator("use"), "href"))

	decorative := root.Locator("[data-testid='iconpack-decorative'] svg")
	assertConsumerSVGAttributes(t, decorative, "", "", "true")
	assert.Equal(t, "/assets/icons/appicons/sprite.svg#hi-16-solid-check", mustAttribute(t, decorative.Locator("use"), "href"))

	currentColor := root.Locator("[data-testid='iconpack-current-color'] svg")
	assertConsumerSVGAttributes(t, currentColor, "img", "current color check", "")
	assert.Equal(t, "/assets/icons/appicons/sprite.svg#hi-16-solid-check", mustAttribute(t, currentColor.Locator("use"), "href"))

	missing := root.Locator("[data-testid='iconpack-missing'] svg")
	assertConsumerSVGAttributes(t, missing, "img", "missing symbol", "")
	assert.Equal(t, "/assets/icons/appicons/sprite.svg#missing-symbol", mustAttribute(t, missing.Locator("use"), "href"))

	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
		_ = session.Detach()
	})
	cells := 0
	for _, theme := range []string{"araihu", "goshtoso", "minimal"} {
		for _, dark := range []bool{false, true} {
			for _, zoom := range []float64{1, 2} {
				name := fmt.Sprintf("%s/dark=%t/zoom=%.0f", theme, dark, zoom)
				t.Run(name, func(t *testing.T) {
					setIconpackAppearance(t, page, session, theme, dark, zoom)
					assertConsumerSVGGeometry(t, labelled)
					assertConsumerSVGGeometry(t, decorative)
					assertConsumerSVGGeometry(t, currentColor)
					color, err := currentColor.Evaluate(`el => getComputedStyle(el).color`, nil)
					require.NoError(t, err)
					require.NotEmpty(t, color, "%s currentColor must resolve", name)
					assertConsumerMissingSymbol(t, missing)
					cells++
				})
			}
		}
	}
	require.Equal(t, 12, cells)

	response, err := http.Get(consumerURL + "/assets/icons/appicons/sprite.svg")
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "image/svg+xml", strings.Split(response.Header.Get("Content-Type"), ";")[0])
	sprite, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Contains(t, string(sprite), `id="devicon-trpc"`)
	assert.Contains(t, string(sprite), `id="hi-16-solid-check"`)
}

func TestIconpackConfigGeneratedConsumerBrowserProof(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	consumerURL := startIconpackConfigConsumer(t)
	page := newPage(t, sharedBrowser)
	failures := watchPageFailures(page)
	t.Cleanup(func() {
		waitForPageSettled(t, page)
		failures.RequireEmpty(t)
	})
	_, err := page.Goto(consumerURL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)

	root := page.Locator("[data-testid='generated-iconpack']")
	require.NoError(t, root.WaitFor())
	assert.Equal(t, "demo-pack-16-solid-check", mustAttribute(t, root, "data-canonical-name"))
	assert.Equal(t, "demo-pack-16-solid-check", mustAttribute(t, root, "data-symbol"))
	icon := root.Locator("svg").First()
	assertConsumerSVGAttributes(t, icon, "img", "demo check", "")
	assert.Equal(t, "/assets/icons/appicons/sprite.svg#demo-pack-16-solid-check", mustAttribute(t, icon.Locator("use"), "href"))
	assertConsumerSVGGeometry(t, icon)

	response, err := http.Get(consumerURL + "/assets/icons/appicons/sprite.svg")
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	sprite, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Contains(t, string(sprite), `id="demo-pack-16-solid-check"`)
}

func startIconpackConfigConsumer(t *testing.T) string {
	t.Helper()
	archive := e2eIconpackArchive(t)
	archiveServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/icons.tar.gz" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(archive)
	}))
	t.Cleanup(archiveServer.Close)

	root := t.TempDir()
	configPath := filepath.Join(root, ".iconpack.yaml")
	config := fmt.Sprintf(`schemaVersion: 1
sources:
  - id: demo-source
    kind: archive
    url: %s/icons.tar.gz
    packName: demo-pack
    stripComponents: 1
    paths:
      - 16/solid/check.svg
    license: MIT
    licensePath: LICENSE
`, archiveServer.URL)
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0o644))
	output := filepath.Join(root, "consumer", "appicons")
	require.NoError(t, os.MkdirAll(filepath.Dir(output), 0o755))
	repoRoot := e2eRepositoryRoot(t)
	runIconpack(t, repoRoot,
		"run", "./cmd/iconpack",
		"-config", configPath, "-trust", "-allow-http",
		"-out", output, "-package", "appicons", "-const-prefix", "Icon",
		"-sprite-url", "/assets/icons/appicons/sprite.svg",
	)

	consumerDir := filepath.Dir(output)
	goMod := fmt.Sprintf(`module example.com/goshtoso-iconpack-config-e2e

go 1.26.5

require github.com/araihu/goshtoso v0.0.0

replace github.com/araihu/goshtoso => %s
`, repoRoot)
	require.NoError(t, os.WriteFile(filepath.Join(consumerDir, "go.mod"), []byte(goMod), 0o644))
	mainSource := `package main

import (
    "bytes"
    "context"
    "fmt"
    "net"
    "net/http"
    "os"

    "github.com/araihu/goshtoso/components/icon"
    "example.com/goshtoso-iconpack-config-e2e/appicons"
)

func main() {
    sprite, err := os.ReadFile("appicons/sprite.svg")
    if err != nil { panic(err) }
    glyph, ok := appicons.Lookup(appicons.NameDemoPack16SolidCheck)
    if !ok || glyph.CanonicalName != "demo-pack-16-solid-check" || glyph.Symbol != appicons.IconDemoPack16SolidCheck { panic("generated config lookup failed") }
    mux := http.NewServeMux()
    mux.HandleFunc("GET /assets/icons/appicons/sprite.svg", func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "image/svg+xml")
        _, _ = w.Write(sprite)
    })
    mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
        var page bytes.Buffer
        page.WriteString("<main data-testid=\"generated-iconpack\" data-canonical-name=\"" + glyph.CanonicalName + "\" data-symbol=\"" + string(glyph.Symbol) + "\">")
        _ = appicons.Icon(appicons.Config{Symbol: glyph.Symbol, Label: "demo check", Size: icon.SizeLG}).Render(context.Background(), &page)
        page.WriteString("</main>")
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        _, _ = w.Write(page.Bytes())
    })
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { panic(err) }
    fmt.Printf("READY http://%s\n", listener.Addr())
    if err := http.Serve(listener, mux); err != nil { panic(err) }
}
`
	require.NoError(t, os.WriteFile(filepath.Join(consumerDir, "main.go"), []byte(mainSource), 0o644))
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = consumerDir
	tidy.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-p=2")
	outputBytes, err := tidy.CombinedOutput()
	require.NoErrorf(t, err, "tidy generated config consumer: %s", outputBytes)

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = consumerDir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-p=2")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { stopIconpackConsumer(cmd) })
	ready := make(chan string, 1)
	readDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "READY ") {
				ready <- strings.TrimSpace(strings.TrimPrefix(line, "READY "))
				return
			}
		}
		if err := scanner.Err(); err != nil {
			readDone <- err
			return
		}
		readDone <- fmt.Errorf("config consumer exited before READY: %s", strings.TrimSpace(stderr.String()))
	}()
	select {
	case url := <-ready:
		return url
	case err := <-readDone:
		t.Fatalf("start generated config consumer: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatalf("timed out waiting for generated config consumer: %s", strings.TrimSpace(stderr.String()))
	}
	return ""
}

func e2eIconpackArchive(t *testing.T) []byte {
	t.Helper()
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	files := map[string][]byte{
		"iconpack/16/solid/check.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path d="M1 1h14v14H1z"/></svg>`),
		"iconpack/LICENSE":            []byte("MIT\n"),
	}
	for name, contents := range files {
		require.NoError(t, tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(contents))}))
		_, err := tarWriter.Write(contents)
		require.NoError(t, err)
	}
	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())
	return output.Bytes()
}

func startIconpackConsumer(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	release := writeE2EIconpackRelease(t, filepath.Join(root, "release"))
	output := filepath.Join(root, "consumer", "appicons")
	require.NoError(t, os.MkdirAll(filepath.Dir(output), 0o755))
	repoRoot := e2eRepositoryRoot(t)
	styles, err := os.ReadFile(filepath.Join(repoRoot, "assets", "styles.css"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(filepath.Dir(output), "styles.css"), styles, 0o644))
	args := []string{
		"run", "./cmd/iconpack", "-release-root", release.Root,
		"-release", release.Release, "-catalog-sha256", release.CatalogSHA256,
		"-release-json-sha256", release.ReleaseJSONSHA256,
		"-checksums-sha256", release.ChecksumsSHA256,
		"-out", output, "-package", "appicons", "-const-prefix", "Icon",
		"-sprite-url", "/assets/icons/appicons/sprite.svg",
	}
	for _, name := range release.Names {
		args = append(args, "-name", name)
	}
	runIconpack(t, repoRoot, args...)

	consumerDir := filepath.Dir(output)
	goMod := fmt.Sprintf(`module example.com/goshtoso-iconpack-e2e

go 1.26.5

require github.com/araihu/goshtoso v0.0.0

replace github.com/araihu/goshtoso => %s
`, repoRoot)
	require.NoError(t, os.WriteFile(filepath.Join(consumerDir, "go.mod"), []byte(goMod), 0o644))
	mainSource := `package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/araihu/goshtoso/components/icon"
	"example.com/goshtoso-iconpack-e2e/appicons"
)

func main() {
	sprite, err := os.ReadFile("appicons/sprite.svg")
	if err != nil {
		panic(err)
	}
	glyph, ok := appicons.Lookup(appicons.NameBrandDeveloperIconsTRPC)
	if !ok || glyph.CanonicalName != "brand-developer-icons-tRPC" || glyph.Symbol != appicons.IconBrandDeveloperIconsTRPC {
		panic("generated canonical lookup failed")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		http.ServeFile(w, r, "styles.css")
	})
	mux.HandleFunc("GET /assets/icons/appicons/sprite.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write(sprite)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		var page bytes.Buffer
		page.WriteString("<!doctype html><html data-theme=\"araihu\"><head><link rel=\"stylesheet\" href=\"/assets/styles.css\"></head><body>")
		page.WriteString("<main data-testid=\"generated-iconpack\" data-canonical-name=\"" + glyph.CanonicalName + "\" data-symbol=\"" + string(glyph.Symbol) + "\">")
		page.WriteString("<span data-testid=\"iconpack-labelled\">")
		_ = appicons.Icon(appicons.Config{Symbol: appicons.IconBrandDeveloperIconsTRPC, Label: "tRPC", Size: icon.SizeLG}).Render(contextBackground(), &page)
		page.WriteString("</span><span data-testid=\"iconpack-decorative\">")
		_ = appicons.Icon(appicons.Config{Symbol: appicons.IconUiHi16SolidCheck, Decorative: true}).Render(contextBackground(), &page)
		page.WriteString("</span><span data-testid=\"iconpack-current-color\">")
		_ = appicons.Icon(appicons.Config{Symbol: appicons.IconUiHi16SolidCheck, Label: "current color check", RootClass: "text-primary"}).Render(contextBackground(), &page)
		page.WriteString("</span><span data-testid=\"iconpack-missing\">")
		_ = icon.Icon(icon.Config{Symbol: icon.Symbol("missing-symbol"), SpriteURL: "/assets/icons/appicons/sprite.svg", Label: "missing symbol"}).Render(contextBackground(), &page)
		page.WriteString("</span></main></body></html>")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page.Bytes())
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	fmt.Printf("READY http://%s\n", listener.Addr())
	if err := http.Serve(listener, mux); err != nil {
		panic(err)
	}
}

func contextBackground() context.Context { return context.Background() }
`
	require.NoError(t, os.WriteFile(filepath.Join(consumerDir, "main.go"), []byte(mainSource), 0o644))
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = consumerDir
	tidy.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-p=2")
	outputBytes, err := tidy.CombinedOutput()
	require.NoErrorf(t, err, "tidy generated consumer: %s", outputBytes)

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = consumerDir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-p=2")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() { stopIconpackConsumer(cmd) })

	ready := make(chan string, 1)
	readDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "READY ") {
				ready <- strings.TrimSpace(strings.TrimPrefix(line, "READY "))
				return
			}
		}
		if err := scanner.Err(); err != nil {
			readDone <- err
			return
		}
		readDone <- fmt.Errorf("consumer exited before READY: %s", strings.TrimSpace(stderr.String()))
	}()

	select {
	case url := <-ready:
		return url
	case err := <-readDone:
		t.Fatalf("start generated consumer: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatalf("timed out waiting for generated consumer: %s", strings.TrimSpace(stderr.String()))
	}
	return ""
}

func stopIconpackConsumer(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil || cmd.ProcessState != nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
	}
}

func writeE2EIconpackRelease(t *testing.T, root string) e2eIconpackRelease {
	t.Helper()
	devAsset := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path fill="#398CCB" d="M0 0h100v100H0z"/></svg>`)
	heroAsset := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill="currentColor" d="M2 8l4 4 8-8"/></svg>`)
	devSprite := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><symbol id="devicon-trpc" viewBox="0 0 100 100"><path fill="#398CCB" d="M0 0h100v100H0z"/></symbol></svg>`)
	heroSprite := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><symbol id="hi-16-solid-check" viewBox="0 0 16 16"><path fill="currentColor" d="M2 8l4 4 8-8"/></symbol></svg>`)
	files := map[string][]byte{
		"NOTICE":                                      []byte("E2E iconpack notice\n"),
		"campaigns.json":                              []byte("{}\n"),
		"themes.json":                                 []byte("{}\n"),
		"icons/brand/developer-icons/tRPC.svg":        devAsset,
		"icons/brand/developer-icons/sprite.svg":      devSprite,
		"icons/brand/developer-icons/provenance.json": []byte("{}\n"),
		"licenses/developer-icons-MIT.txt":            []byte("Developer Icons MIT\n"),
		"icons/ui/heroicons/16-solid-check.svg":       heroAsset,
		"icons/ui/sprite.svg":                         heroSprite,
		"icons/ui/heroicons/provenance.json":          []byte("{}\n"),
		"licenses/heroicons-MIT.txt":                  []byte("Heroicons MIT\n"),
	}
	assets := []map[string]any{
		{
			"canonicalName": "brand-developer-icons-tRPC", "namespace": "brand", "path": "icons/brand/developer-icons/tRPC.svg", "product": "developer-icons", "artwork": "icon", "appearance": "default", "surface": "transparent", "framing": "optical", "format": "svg", "dimensions": map[string]any{"viewBox": "0 0 100 100"}, "spriteSymbol": "devicon-trpc", "colorBehavior": "protected", "license": "MIT", "source": "developer-icons@v7.0.1:icons/tRPC.svg", "sha256": digest(devAsset),
		},
		{
			"canonicalName": "ui-hi-16-solid-check", "namespace": "ui", "path": "icons/ui/heroicons/16-solid-check.svg", "product": "heroicons", "artwork": "icon", "appearance": "default", "surface": "transparent", "framing": "optical", "format": "svg", "dimensions": map[string]any{"viewBox": "0 0 16 16"}, "spriteSymbol": "hi-16-solid-check", "colorBehavior": "monochrome", "license": "MIT", "source": "heroicons@v2.2.0:16/solid/check.svg", "sha256": digest(heroAsset),
		},
	}
	catalog := map[string]any{"schemaVersion": 2, "release": "v0.2.0-e2e", "identityRevision": 11, "assets": assets}
	catalogBytes := mustJSON(t, catalog)
	files["catalog.json"] = catalogBytes

	paths := sortedFilePaths(files)
	releasePaths := []string{"catalog.json", "themes.json", "campaigns.json"}
	for _, path := range paths {
		if path != "catalog.json" && path != "themes.json" && path != "campaigns.json" {
			releasePaths = append(releasePaths, path)
		}
	}
	releaseFiles := make([]map[string]any, 0, len(releasePaths))
	for _, path := range releasePaths {
		releaseFiles = append(releaseFiles, map[string]any{"path": path, "sha256": digest(files[path]), "size": len(files[path])})
	}
	release := map[string]any{
		"schemaVersion": 1, "release": "v0.2.0-e2e", "identityRevision": 11, "runtimeVersion": 1,
		"catalogSha256": digest(catalogBytes), "themesSha256": digest(files["themes.json"]), "campaignsSha256": digest(files["campaigns.json"]), "files": releaseFiles,
	}
	releaseBytes := mustJSON(t, release)
	files["release.json"] = releaseBytes
	paths = sortedFilePaths(files)
	var checksums strings.Builder
	for _, path := range paths {
		fmt.Fprintf(&checksums, "%s  %s\n", digest(files[path]), path)
	}
	checksumsBytes := []byte(checksums.String())
	writeE2EFiles(t, root, files)
	require.NoError(t, os.WriteFile(filepath.Join(root, "checksums.txt"), checksumsBytes, 0o644))
	return e2eIconpackRelease{
		Root: root, Release: "v0.2.0-e2e", CatalogSHA256: digest(catalogBytes), ReleaseJSONSHA256: digest(releaseBytes), ChecksumsSHA256: digest(checksumsBytes),
		Names: []string{"brand-developer-icons-tRPC", "ui-hi-16-solid-check"},
	}
}

func runIconpack(t *testing.T, repoRoot string, args ...string) {
	t.Helper()
	command := exec.Command("go", args...)
	command.Dir = repoRoot
	command.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-p=2")
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "iconpack CLI: %s", output)
}

func writeE2EFiles(t *testing.T, root string, files map[string][]byte) {
	t.Helper()
	for path, contents := range files {
		filename := filepath.Join(root, filepath.FromSlash(path))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
		require.NoError(t, os.WriteFile(filename, contents, 0o644))
	}
}

func sortedFilePaths(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	contents, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	return append(contents, '\n')
}

func assertConsumerSVGAttributes(t *testing.T, svg playwright.Locator, role, label, hidden string) {
	t.Helper()
	actualRole, err := svg.GetAttribute("role")
	require.NoError(t, err)
	assert.Equal(t, role, actualRole)
	actualLabel, err := svg.GetAttribute("aria-label")
	require.NoError(t, err)
	assert.Equal(t, label, actualLabel)
	actualHidden, err := svg.GetAttribute("aria-hidden")
	require.NoError(t, err)
	assert.Equal(t, hidden, actualHidden)
}

func assertConsumerSVGGeometry(t *testing.T, svg playwright.Locator) {
	t.Helper()
	value, err := svg.Evaluate(`el => {
		const use = el.querySelector('use');
		const glyph = use && use.getBBox();
		return {
			width: el.getBoundingClientRect().width,
			height: el.getBoundingClientRect().height,
			glyphWidth: glyph ? glyph.width : 0,
			glyphHeight: glyph ? glyph.height : 0,
		};
	}`, nil)
	require.NoError(t, err)
	geometry, ok := value.(map[string]interface{})
	require.True(t, ok, "expected SVG geometry object, got %T", value)
	for _, key := range []string{"width", "height", "glyphWidth", "glyphHeight"} {
		measurement := 0.0
		switch value := geometry[key].(type) {
		case float64:
			measurement = value
		case int:
			measurement = float64(value)
		default:
			t.Fatalf("expected %s to be numeric, got %T", key, geometry[key])
		}
		assert.Greater(t, measurement, float64(0), key)
	}
}

func setIconpackAppearance(t *testing.T, page playwright.Page, session playwright.CDPSession, theme string, dark bool, zoom float64) {
	t.Helper()
	_, err := session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": zoom})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`zoom => Math.abs((window.visualViewport?.scale || 1) - zoom) < 0.05`, zoom)
	require.NoError(t, err)
	_, err = page.Evaluate(`([theme, dark]) => {
		document.documentElement.dataset.theme = theme;
		document.documentElement.classList.toggle('dark', dark);
	}`, []any{theme, dark})
	require.NoError(t, err)
}

func assertConsumerMissingSymbol(t *testing.T, svg playwright.Locator) {
	t.Helper()
	value, err := svg.Evaluate(`el => {
		const use = el.querySelector('use');
		const glyph = use && use.getBBox();
		return {
			width: el.getBoundingClientRect().width,
			height: el.getBoundingClientRect().height,
			glyphWidth: glyph ? glyph.width : 0,
			glyphHeight: glyph ? glyph.height : 0,
		};
	}`, nil)
	require.NoError(t, err)
	geometry := value.(map[string]interface{})
	assert.Greater(t, consumerMeasurement(t, geometry["width"]), float64(0))
	assert.Greater(t, consumerMeasurement(t, geometry["height"]), float64(0))
	assert.Zero(t, consumerMeasurement(t, geometry["glyphWidth"]))
	assert.Zero(t, consumerMeasurement(t, geometry["glyphHeight"]))
}

func consumerMeasurement(t *testing.T, value interface{}) float64 {
	t.Helper()
	switch value := value.(type) {
	case float64:
		return value
	case int:
		return float64(value)
	default:
		t.Fatalf("expected numeric browser measurement, got %T", value)
		return 0
	}
}

func digest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

func e2eRepositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	require.NoError(t, err)
	return root
}
