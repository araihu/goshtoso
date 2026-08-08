//go:build e2e && (full || iconpack)

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	goshtosoiconpack "github.com/araihu/goshtoso/iconpack"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

const (
	iconpackCanonicalName = "brand-developer-icons-tRPC"
	iconpackSpriteSymbol  = "devicon-trpc"
)

func TestIconpackGeneratedConsumerBrowserProof(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	repositoryRoot, err := filepath.Abs("../../..")
	require.NoError(t, err)
	consumerRoot := t.TempDir()
	options := synthesizeIconpackRelease(t)
	options.Names = []string{iconpackCanonicalName}
	options.OutputDir = filepath.Join(consumerRoot, "appicons")
	options.Package = "appicons"
	options.ConstPrefix = "Icon"
	options.SpriteURL = "/assets/icons/app.svg"

	result, err := goshtosoiconpack.Generate(context.Background(), options)
	require.NoError(t, err)
	require.True(t, result.Published)
	require.Equal(t, 1, result.SelectedCount)
	require.Equal(t, options.Release, result.Release)

	writeIconpackConsumerModule(t, consumerRoot, repositoryRoot)
	consumerURL := startIconpackConsumer(t, consumerRoot)

	page := newPage(t, sharedBrowser)
	response, err := page.Goto(consumerURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, 200, response.Status())
	assertIconpackMediaType(t, response, "text/html")

	identity := page.Locator("[data-testid='iconpack-identity']")
	require.NoError(t, identity.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}))
	canonicalName, err := identity.GetAttribute("data-canonical-name")
	require.NoError(t, err)
	require.Equal(t, iconpackCanonicalName, canonicalName)
	spriteSymbol, err := identity.GetAttribute("data-sprite-symbol")
	require.NoError(t, err)
	require.Equal(t, iconpackSpriteSymbol, spriteSymbol)

	labelled := page.Locator("[data-testid='iconpack-labelled'] svg")
	require.NoError(t, labelled.WaitFor())
	require.Equal(t, "img", iconpackAttribute(t, labelled, "role"))
	require.Equal(t, "tRPC", iconpackAttribute(t, labelled, "aria-label"))
	require.Empty(t, iconpackAttribute(t, labelled, "aria-hidden"))
	assertIconpackExternalUse(t, labelled)

	decorative := page.Locator("[data-testid='iconpack-decorative'] svg")
	require.NoError(t, decorative.WaitFor())
	require.Equal(t, "true", iconpackAttribute(t, decorative, "aria-hidden"))
	require.Empty(t, iconpackAttribute(t, decorative, "role"))
	require.Empty(t, iconpackAttribute(t, decorative, "aria-label"))
	assertIconpackExternalUse(t, decorative)

	_, err = page.WaitForFunction(`() => {
		const use = document.querySelector("[data-testid='iconpack-labelled'] use");
		if (!use) return false;
		const bounds = use.getBBox();
		return bounds.width > 0 && bounds.height > 0;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)})
	require.NoError(t, err, "generated external symbol must paint positive geometry")

	spriteBytes, err := os.ReadFile(filepath.Join(options.OutputDir, "sprite.svg"))
	require.NoError(t, err)
	spritePage := newPage(t, sharedBrowser)
	spriteResponse, err := spritePage.Goto(consumerURL + options.SpriteURL)
	require.NoError(t, err)
	require.NotNil(t, spriteResponse)
	require.Equal(t, 200, spriteResponse.Status())
	assertIconpackMediaType(t, spriteResponse, "image/svg+xml")
	servedSprite, err := spriteResponse.Body()
	require.NoError(t, err)
	require.Equal(t, spriteBytes, servedSprite)
}

func assertIconpackExternalUse(t *testing.T, svg playwright.Locator) {
	t.Helper()
	href, err := svg.Locator("use").GetAttribute("href")
	require.NoError(t, err)
	require.Equal(t, "/assets/icons/app.svg#"+iconpackSpriteSymbol, href)
}

func assertIconpackMediaType(t *testing.T, response playwright.Response, want string) {
	t.Helper()
	contentType, err := response.HeaderValue("content-type")
	require.NoError(t, err)
	mediaType, _, err := mime.ParseMediaType(contentType)
	require.NoError(t, err)
	require.Equal(t, want, mediaType)
}

func iconpackAttribute(t *testing.T, locator playwright.Locator, name string) string {
	t.Helper()
	value, err := locator.GetAttribute(name)
	require.NoError(t, err)
	return value
}

type iconpackReleaseDocument struct {
	SchemaVersion    int                   `json:"schemaVersion"`
	Release          string                `json:"release"`
	IdentityRevision int                   `json:"identityRevision"`
	RuntimeVersion   int                   `json:"runtimeVersion"`
	CatalogSHA256    string                `json:"catalogSha256"`
	ThemesSHA256     string                `json:"themesSha256"`
	CampaignsSHA256  string                `json:"campaignsSha256"`
	Files            []iconpackReleaseFile `json:"files"`
}

type iconpackReleaseFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

func synthesizeIconpackRelease(t *testing.T) goshtosoiconpack.Options {
	t.Helper()
	releaseRoot := t.TempDir()
	asset := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path fill="#398CCB" d="M0 0h100v100H0z"/></svg>`)
	sprite := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><symbol id="devicon-trpc" viewBox="0 0 100 100"><path fill="#398CCB" d="M0 0h100v100H0z"/></symbol></svg>`)
	release := "v0.2.0-test"
	catalog := fmt.Sprintf(`{
  "schemaVersion": 2,
  "release": %q,
  "identityRevision": 11,
  "assets": [
    {
      "canonicalName": %q,
      "namespace": "brand",
      "path": "icons/brand/developer-icons/tRPC.svg",
      "product": "developer-icons",
      "artwork": "icon",
      "appearance": "default",
      "surface": "transparent",
      "framing": "optical",
      "format": "svg",
      "dimensions": {"viewBox": "0 0 100 100"},
      "spriteSymbol": %q,
      "colorBehavior": "protected",
      "license": "MIT",
      "source": "developer-icons@example:icons/tRPC.svg",
      "sha256": %q
    }
  ]
}
`, release, iconpackCanonicalName, iconpackSpriteSymbol, iconpackSHA256(asset))
	files := map[string][]byte{
		"NOTICE":                                      []byte("Synthetic consumer proof notice\n"),
		"campaigns.json":                              []byte("{}\n"),
		"catalog.json":                                []byte(catalog),
		"icons/brand/developer-icons/tRPC.svg":        asset,
		"icons/brand/developer-icons/sprite.svg":      sprite,
		"icons/brand/developer-icons/provenance.json": []byte("{}\n"),
		"licenses/developer-icons-MIT.txt":            []byte("Synthetic MIT license\n"),
		"themes.json":                                 []byte("{}\n"),
	}
	for relative, contents := range files {
		writeIconpackFixtureFile(t, releaseRoot, relative, contents)
	}

	relatives := make([]string, 0, len(files))
	for relative := range files {
		relatives = append(relatives, relative)
	}
	sort.Strings(relatives)
	document := iconpackReleaseDocument{
		SchemaVersion:    1,
		Release:          release,
		IdentityRevision: 11,
		RuntimeVersion:   1,
		CatalogSHA256:    iconpackSHA256(files["catalog.json"]),
		ThemesSHA256:     iconpackSHA256(files["themes.json"]),
		CampaignsSHA256:  iconpackSHA256(files["campaigns.json"]),
	}
	releaseOrder := []string{"catalog.json", "themes.json", "campaigns.json"}
	for _, relative := range relatives {
		if relative != "catalog.json" && relative != "themes.json" && relative != "campaigns.json" {
			releaseOrder = append(releaseOrder, relative)
		}
	}
	for _, relative := range releaseOrder {
		document.Files = append(document.Files, iconpackReleaseFile{
			Path: relative, SHA256: iconpackSHA256(files[relative]), Size: int64(len(files[relative])),
		})
	}
	releaseJSON, err := json.MarshalIndent(document, "", "  ")
	require.NoError(t, err)
	releaseJSON = append(releaseJSON, '\n')
	files["release.json"] = releaseJSON
	writeIconpackFixtureFile(t, releaseRoot, "release.json", releaseJSON)

	relatives = append(relatives, "release.json")
	sort.Strings(relatives)
	var checksums strings.Builder
	for _, relative := range relatives {
		fmt.Fprintf(&checksums, "%s  %s\n", iconpackSHA256(files[relative]), relative)
	}
	checksumsBytes := []byte(checksums.String())
	writeIconpackFixtureFile(t, releaseRoot, "checksums.txt", checksumsBytes)

	return goshtosoiconpack.Options{
		ReleaseRoot:       releaseRoot,
		Release:           release,
		CatalogSHA256:     iconpackSHA256(files["catalog.json"]),
		ReleaseJSONSHA256: iconpackSHA256(releaseJSON),
		ChecksumsSHA256:   iconpackSHA256(checksumsBytes),
	}
}

func iconpackSHA256(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}

func writeIconpackFixtureFile(t *testing.T, root, relative string, contents []byte) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(relative))
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, contents, 0o644))
}

func writeIconpackConsumerModule(t *testing.T, root, repositoryRoot string) {
	t.Helper()
	goMod := fmt.Sprintf(`module example.com/goshtoso-iconpack-consumer

go 1.26.5

require github.com/araihu/goshtoso v0.0.0

replace github.com/araihu/goshtoso => %s
`, filepath.ToSlash(repositoryRoot))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte(iconpackConsumerMain), 0o644))

	runIconpackConsumerCommand(t, root, "go", "mod", "tidy")
	runIconpackConsumerCommand(t, root, "go", "build", "-o", filepath.Join(root, "consumer"), ".")
}

func runIconpackConsumerCommand(t *testing.T, directory, name string, args ...string) {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "%s %s\n%s", name, strings.Join(args, " "), output)
}

func startIconpackConsumer(t *testing.T, root string) string {
	t.Helper()
	command := exec.Command(filepath.Join(root, "consumer"))
	command.Dir = root
	stdout, err := command.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	require.NoError(t, command.Start())
	waited := make(chan error, 1)
	go func() { waited <- command.Wait() }()
	t.Cleanup(func() {
		if command.Process != nil {
			_ = command.Process.Kill()
		}
		select {
		case <-waited:
		case <-time.After(5 * time.Second):
			t.Errorf("consumer process did not stop")
		}
	})

	address := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			address <- strings.TrimSpace(scanner.Text())
			return
		}
		address <- ""
	}()
	select {
	case url := <-address:
		require.NotEmpty(t, url, "consumer exited before reporting address: %s", stderr.String())
		require.True(t, strings.HasPrefix(url, "http://127.0.0.1:"), "consumer address = %q", url)
		return url
	case err := <-waited:
		t.Fatalf("consumer exited before reporting address: %v\n%s", err, stderr.String())
	case <-time.After(10 * time.Second):
		t.Fatal("consumer did not report its 127.0.0.1:0 listener")
	}
	return ""
}

const iconpackConsumerMain = `package main

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"net"
	"net/http"
	"os"

	"example.com/goshtoso-iconpack-consumer/appicons"
)

func main() {
	sprite, err := os.ReadFile("appicons/sprite.svg")
	if err != nil {
		panic(err)
	}
	glyph, ok := appicons.Lookup(appicons.NameBrandDeveloperIconsTRPC)
	if !ok || glyph.CanonicalName != "brand-developer-icons-tRPC" || glyph.Symbol != appicons.IconBrandDeveloperIconsTRPC {
		panic("generated canonical-name lookup mismatch")
	}
	page, err := renderPage(glyph.CanonicalName, string(glyph.Symbol))
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/icons/app.svg", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/svg+xml")
		_, _ = writer.Write(sprite)
	})
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write(page)
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	fmt.Println("http://" + listener.Addr().String())
	if err := http.Serve(listener, mux); err != nil {
		panic(err)
	}
}

func renderPage(canonicalName, spriteSymbol string) ([]byte, error) {
	var output bytes.Buffer
	fmt.Fprintf(&output, ` + "`" + `<!doctype html><html><head><style>svg{width:64px;height:64px}</style></head><body><main><div data-testid="iconpack-identity" data-canonical-name="%s" data-sprite-symbol="%s"></div><div data-testid="iconpack-labelled">` + "`" + `,
		html.EscapeString(canonicalName), html.EscapeString(spriteSymbol))
	labelled := appicons.Icon(appicons.Config{Symbol: appicons.IconBrandDeveloperIconsTRPC, Label: "tRPC"})
	if err := labelled.Render(context.Background(), &output); err != nil {
		return nil, err
	}
	output.WriteString(` + "`" + `</div><div data-testid="iconpack-decorative">` + "`" + `)
	decorative := appicons.Icon(appicons.Config{
		Symbol: appicons.IconBrandDeveloperIconsTRPC, Label: "must be hidden", Decorative: true,
	})
	if err := decorative.Render(context.Background(), &output); err != nil {
		return nil, err
	}
	output.WriteString(` + "`" + `</div></main></body></html>` + "`" + `)
	return output.Bytes(), nil
}
`
