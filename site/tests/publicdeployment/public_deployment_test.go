package publicdeployment

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

var candidateEnvironmentKeys = []string{
	"MODULE_CANDIDATE_PROXY",
	"MODULE_CANDIDATE_VERSION",
	"MODULE_CANDIDATE_COMMIT",
	"MODULE_CANDIDATE_TREE",
	"MODULE_CANDIDATE_CSS_SHA256",
	"MODULE_CANDIDATE_JS_SHA256",
	"MODULE_CANDIDATE_SPRITE_SHA256",
	"MODULE_CANDIDATE_LOGO_SHA256",
}

func exactCandidateEnvironment(getenv func(string) string) error {
	for _, key := range candidateEnvironmentKeys {
		if strings.TrimSpace(getenv(key)) == "" {
			return fmt.Errorf("required candidate environment is missing: %s", key)
		}
	}
	if !strings.HasPrefix(getenv("MODULE_CANDIDATE_PROXY"), "file://") {
		return fmt.Errorf("MODULE_CANDIDATE_PROXY must use file:// with no fallback")
	}
	return nil
}

func TestPublicDeploymentDownloadsExactCandidate(t *testing.T) {
	if os.Getenv("MODULE_CANDIDATE_PROXY") == "" {
		t.Skip("set exact MODULE_CANDIDATE_* environment to run candidate download proof")
	}
	if err := exactCandidateEnvironment(os.Getenv); err != nil {
		t.Fatal(err)
	}

	version := os.Getenv("MODULE_CANDIDATE_VERSION")
	consumer := t.TempDir()
	goMod := fmt.Sprintf("module example.com/goshtoso-public-deployment\n\ngo 1.26.5\n\nrequire github.com/araihu/goshtoso %s\n", version)
	if strings.Contains(goMod, "replace ") {
		t.Fatal("candidate consumer go.mod must not contain replace")
	}
	if err := os.WriteFile(filepath.Join(consumer, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	moduleCacheRoot := t.TempDir()
	moduleCache := filepath.Join(moduleCacheRoot, "modcache")
	t.Cleanup(func() {
		_ = filepath.Walk(moduleCacheRoot, func(path string, _ os.FileInfo, walkErr error) error {
			if walkErr == nil {
				_ = os.Chmod(path, 0o700)
			}
			return nil
		})
	})
	command := exec.Command("go", "mod", "download", "-json", "github.com/araihu/goshtoso@"+version)
	command.Dir = consumer
	command.Env = candidateCommandEnvironment(
		os.Getenv("MODULE_CANDIDATE_PROXY"),
		moduleCache,
		filepath.Join(t.TempDir(), "buildcache"),
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("download exact candidate: %v\n%s", err, output)
	}
	var downloaded struct {
		Version  string
		Sum      string
		GoModSum string
		Dir      string
		Error    string
	}
	if err := json.Unmarshal(output, &downloaded); err != nil {
		t.Fatalf("decode go mod download result: %v\n%s", err, output)
	}
	if downloaded.Error != "" {
		t.Fatalf("go mod download error: %s", downloaded.Error)
	}
	if downloaded.Version != version {
		t.Fatalf("downloaded version = %q, want %q", downloaded.Version, version)
	}
	if downloaded.Sum == "" || downloaded.GoModSum == "" {
		t.Fatalf("candidate sums are incomplete: Sum=%q GoModSum=%q", downloaded.Sum, downloaded.GoModSum)
	}
	for _, asset := range []struct {
		path   string
		digest string
	}{
		{"assets/styles.css", os.Getenv("MODULE_CANDIDATE_CSS_SHA256")},
		{"assets/js/goshtoso.min.js", os.Getenv("MODULE_CANDIDATE_JS_SHA256")},
		{"assets/icons/heroicons.svg", os.Getenv("MODULE_CANDIDATE_SPRITE_SHA256")},
		{"assets/images/goshtoso-logo.svg", os.Getenv("MODULE_CANDIDATE_LOGO_SHA256")},
	} {
		data, err := os.ReadFile(filepath.Join(downloaded.Dir, filepath.FromSlash(asset.path)))
		if err != nil {
			t.Fatalf("read downloaded candidate %s: %v", asset.path, err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != asset.digest {
			t.Fatalf("downloaded candidate %s sha256 = %s, want %s", asset.path, got, asset.digest)
		}
	}

	consumerTest := fmt.Sprintf(`package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/icon/heroicons"
)

func TestExactCandidateRendersPublicComponents(t *testing.T) {
	resolved := %q

	var dependencies bytes.Buffer
	if err := head.Dependencies(head.WithLocalRuntime()).Render(context.Background(), &dependencies); err != nil { t.Fatal(err) }
	if !strings.Contains(dependencies.String(), "/assets/styles.css") || !strings.Contains(dependencies.String(), "/assets/js/goshtoso.min.js") {
		t.Fatalf("representative head omitted candidate assets: %%s", dependencies.String())
	}
	var glyph bytes.Buffer
	if err := icon.Icon(icon.Config{SpriteURL: heroicons.SpriteURL, Symbol: heroicons.Icon16SolidCheck, Label: "candidate ready"}).Render(context.Background(), &glyph); err != nil { t.Fatal(err) }
	if !strings.Contains(glyph.String(), heroicons.SpriteURL+"#"+string(heroicons.Icon16SolidCheck)) { t.Fatalf("representative icon = %%s", glyph.String()) }

	proof := fmt.Sprintf("<output data-candidate-version=%%q data-candidate-commit=%%q data-candidate-tree=%%q>%%s%%s</output>", resolved, %q, %q, dependencies.String(), glyph.String())
	for _, identity := range []string{%q, %q, %q} {
		if !strings.Contains(proof, identity) { t.Fatalf("rendered candidate proof omitted %%s", identity) }
	}
}
`, version, os.Getenv("MODULE_CANDIDATE_COMMIT"), os.Getenv("MODULE_CANDIDATE_TREE"), version, os.Getenv("MODULE_CANDIDATE_COMMIT"), os.Getenv("MODULE_CANDIDATE_TREE"))
	if err := os.WriteFile(filepath.Join(consumer, "candidate_test.go"), []byte(consumerTest), 0o600); err != nil {
		t.Fatal(err)
	}
	consumerMain := fmt.Sprintf(`package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/icon"
	"github.com/araihu/goshtoso/components/icon/heroicons"
)

func main() {
	var dependencies, glyph bytes.Buffer
	if err := head.Dependencies(head.WithLocalRuntime()).Render(context.Background(), &dependencies); err != nil { panic(err) }
	if err := icon.Icon(icon.Config{SpriteURL: heroicons.SpriteURL, Symbol: heroicons.Icon16SolidCheck, Label: "candidate ready"}).Render(context.Background(), &glyph); err != nil { panic(err) }
	page := "<!doctype html><html><head>" + dependencies.String() + "</head><body data-requested=\"" + template.HTMLEscapeString(os.Getenv("CANDIDATE_REQUESTED")) + "\" data-resolved=\"" + template.HTMLEscapeString(os.Getenv("CANDIDATE_RESOLVED")) + "\"><main>" + glyph.String() + "<button id=\"proof-action\" type=\"button\" onclick=\"this.dataset.clicked='true'\">Verify candidate</button></main></body></html>"
	mux := http.NewServeMux()
	mux.Handle("/assets/", assets.Handler())
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) { _, _ = writer.Write([]byte(page)) })
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { panic(err) }
	fmt.Println(listener.Addr().String())
	if err := http.Serve(listener, mux); err != nil { panic(err) }
}
`)
	if err := os.WriteFile(filepath.Join(consumer, "main.go"), []byte(consumerMain), 0o600); err != nil {
		t.Fatal(err)
	}
	graphCommand := exec.Command("go", "mod", "download", "all")
	graphCommand.Dir = consumer
	graphCommand.Env = command.Env
	if output, err := graphCommand.CombinedOutput(); err != nil {
		t.Fatalf("download exact candidate graph: %v\n%s", err, output)
	}
	tidyCommand := exec.Command("go", "mod", "tidy")
	tidyCommand.Dir = consumer
	tidyCommand.Env = command.Env
	if output, err := tidyCommand.CombinedOutput(); err != nil {
		t.Fatalf("tidy exact candidate consumer: %v\n%s", err, output)
	}
	finalGoMod, err := os.ReadFile(filepath.Join(consumer, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(finalGoMod), "replace ") || !strings.Contains(string(finalGoMod), "github.com/araihu/goshtoso "+version) {
		t.Fatalf("consumer module identity drifted:\n%s", finalGoMod)
	}
	listCommand := exec.Command("go", "list", "-mod=readonly", "-m", "-json", "all")
	listCommand.Dir = consumer
	listCommand.Env = command.Env
	listOutput, err := listCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("list exact candidate graph: %v\n%s", err, listOutput)
	}
	decoder := json.NewDecoder(bytes.NewReader(listOutput))
	foundCandidate := false
	for decoder.More() {
		var moduleValue struct {
			Path    string
			Version string
			Main    bool
			Replace *json.RawMessage
		}
		if err := decoder.Decode(&moduleValue); err != nil {
			t.Fatalf("decode candidate module graph: %v", err)
		}
		if moduleValue.Path != "github.com/araihu/goshtoso" {
			continue
		}
		foundCandidate = true
		if moduleValue.Version != version || moduleValue.Main || moduleValue.Replace != nil {
			t.Fatalf("resolved candidate graph identity = %+v, want version %s, Main=false, Replace=nil", moduleValue, version)
		}
	}
	if !foundCandidate {
		t.Fatal("resolved candidate is absent from go list -m -json all")
	}
	testCommand := exec.Command("go", "test", "-mod=readonly", "./...")
	testCommand.Dir = consumer
	testCommand.Env = command.Env
	if output, err := testCommand.CombinedOutput(); err != nil {
		t.Fatalf("test exact candidate consumer: %v\n%s", err, output)
	}
	binaryPath := filepath.Join(t.TempDir(), "candidate-server")
	buildCommand := exec.Command("go", "build", "-mod=readonly", "-o", binaryPath, ".")
	buildCommand.Dir = consumer
	buildCommand.Env = command.Env
	if output, err := buildCommand.CombinedOutput(); err != nil {
		t.Fatalf("build exact candidate consumer server: %v\n%s", err, output)
	}
	serverCommand := exec.Command(binaryPath)
	serverCommand.Dir = consumer
	serverCommand.Env = append(command.Env, "GOPROXY=off", "CANDIDATE_REQUESTED="+version, "CANDIDATE_RESOLVED="+version)
	serverOutput, err := serverCommand.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var serverErrors bytes.Buffer
	serverCommand.Stderr = &serverErrors
	if err := serverCommand.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if serverCommand.Process != nil {
			_ = serverCommand.Process.Kill()
			_, _ = serverCommand.Process.Wait()
		}
	})
	address, err := bufio.NewReader(serverOutput).ReadString('\n')
	if err != nil {
		t.Fatalf("read candidate server address: %v: %s", err, serverErrors.String())
	}
	address = strings.TrimSpace(address)

	pw, err := playwright.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: new(true)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = browser.Close() })
	page, err := browser.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := page.Goto("http://" + address); err != nil {
		t.Fatal(err)
	}
	requested, err := page.Locator("body").GetAttribute("data-requested")
	if err != nil || requested != version {
		t.Fatalf("browser requested identity = %q, %v; want %q", requested, err, version)
	}
	resolved, err := page.Locator("body").GetAttribute("data-resolved")
	if err != nil || resolved != version {
		t.Fatalf("browser resolved identity = %q, %v; want %q", resolved, err, version)
	}
	digestsValue, err := page.Evaluate(`async () => {
		const paths = ["/assets/styles.css", "/assets/js/goshtoso.min.js", "/assets/icons/heroicons.svg", "/assets/images/goshtoso-logo.svg"];
		const result = {};
		for (const path of paths) {
			const response = await fetch(path);
			if (!response.ok) throw new Error(path + ": " + response.status);
			const digest = await crypto.subtle.digest("SHA-256", await response.arrayBuffer());
			result[path] = Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, "0")).join("");
		}
		return result;
	}`)
	if err != nil {
		t.Fatal(err)
	}
	digests, ok := digestsValue.(map[string]any)
	if !ok {
		t.Fatalf("browser asset digests type = %T", digestsValue)
	}
	for path, want := range map[string]string{
		"/assets/styles.css":               os.Getenv("MODULE_CANDIDATE_CSS_SHA256"),
		"/assets/js/goshtoso.min.js":       os.Getenv("MODULE_CANDIDATE_JS_SHA256"),
		"/assets/icons/heroicons.svg":      os.Getenv("MODULE_CANDIDATE_SPRITE_SHA256"),
		"/assets/images/goshtoso-logo.svg": os.Getenv("MODULE_CANDIDATE_LOGO_SHA256"),
	} {
		if got := fmt.Sprint(digests[path]); got != want {
			t.Fatalf("browser %s sha256 = %s, want %s", path, got, want)
		}
	}
	button := page.Locator("#proof-action")
	if err := button.Click(); err != nil {
		t.Fatal(err)
	}
	clicked, err := button.GetAttribute("data-clicked")
	if err != nil || clicked != "true" {
		t.Fatalf("candidate consumer interaction data-clicked = %q, %v", clicked, err)
	}
}

func candidateCommandEnvironment(proxy, moduleCache, buildCache string) []string {
	environment := make([]string, 0, len(os.Environ())+9)
	for _, value := range os.Environ() {
		key := strings.SplitN(value, "=", 2)[0]
		switch key {
		case "GOWORK", "GOPROXY", "GOSUMDB", "GONOSUMDB", "GONOPROXY", "GOPRIVATE", "GOVCS", "GOMODCACHE", "GOCACHE":
			continue
		}
		environment = append(environment, value)
	}
	return append(environment,
		"GOWORK=off",
		"GOPROXY="+proxy,
		"GOSUMDB=off",
		"GONOSUMDB=*",
		"GONOPROXY=none",
		"GOPRIVATE=",
		"GOVCS=*:off",
		"GOMODCACHE="+moduleCache,
		"GOCACHE="+buildCache,
	)
}

func TestPublicDeploymentRejectsCandidateProxyDrift(t *testing.T) {
	if os.Getenv("MODULE_CANDIDATE_PROXY") == "" {
		t.Skip("set exact MODULE_CANDIDATE_* environment to run candidate proxy negatives")
	}
	if err := exactCandidateEnvironment(os.Getenv); err != nil {
		t.Fatal(err)
	}
	version := os.Getenv("MODULE_CANDIDATE_VERSION")
	sourceRoot := filepath.FromSlash(strings.TrimPrefix(os.Getenv("MODULE_CANDIDATE_PROXY"), "file://"))
	moduleRelative := filepath.Join("github.com", "araihu", "goshtoso", "@v")

	for _, test := range []struct {
		name             string
		requestedVersion string
		mutate           func(*testing.T, string)
	}{
		{
			name:             "missing candidate zip",
			requestedVersion: version,
			mutate: func(t *testing.T, versionDir string) {
				t.Helper()
				if err := os.Remove(filepath.Join(versionDir, version+".zip")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:             "tampered candidate zip",
			requestedVersion: version,
			mutate: func(t *testing.T, versionDir string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(versionDir, version+".zip"), []byte("not a module zip\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:             "requested version mismatch",
			requestedVersion: "v0.0.0-20260812022327-000000000000",
			mutate:           func(*testing.T, string) {},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			proxy := t.TempDir()
			versionDir := filepath.Join(proxy, moduleRelative)
			if err := os.MkdirAll(versionDir, 0o700); err != nil {
				t.Fatal(err)
			}
			for _, extension := range []string{"info", "mod", "zip"} {
				name := version + "." + extension
				data, err := os.ReadFile(filepath.Join(sourceRoot, moduleRelative, name))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(versionDir, name), data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			test.mutate(t, versionDir)

			moduleCacheRoot := t.TempDir()
			moduleCache := filepath.Join(moduleCacheRoot, "modcache")
			t.Cleanup(func() {
				_ = filepath.Walk(moduleCacheRoot, func(path string, _ os.FileInfo, walkErr error) error {
					if walkErr == nil {
						_ = os.Chmod(path, 0o700)
					}
					return nil
				})
			})
			command := exec.Command("go", "mod", "download", "-json", "github.com/araihu/goshtoso@"+test.requestedVersion)
			command.Env = candidateCommandEnvironment("file://"+filepath.ToSlash(proxy), moduleCache, filepath.Join(t.TempDir(), "buildcache"))
			output, commandErr := command.CombinedOutput()
			var result struct{ Error string }
			decodeErr := json.Unmarshal(output, &result)
			if commandErr == nil && decodeErr == nil && result.Error == "" {
				t.Fatalf("candidate proxy drift unexpectedly downloaded:\n%s", output)
			}
		})
	}
}

func TestPublicDeploymentRequiresExactCandidateEnv(t *testing.T) {
	valid := map[string]string{
		"MODULE_CANDIDATE_PROXY":         "file:///tmp/candidate-proxy",
		"MODULE_CANDIDATE_VERSION":       "v0.0.0-20260812022327-cef353e27f86",
		"MODULE_CANDIDATE_COMMIT":        "cef353e27f86ac2fdfb2b6b07426c27bebe1470b",
		"MODULE_CANDIDATE_TREE":          "0ebc5ebc208903505aa04104d7e078a6518dd940",
		"MODULE_CANDIDATE_CSS_SHA256":    "c5c3e7b355821169dd829638b444817f2b40d9fdfb4f38b3feb6ba8dc51dc581",
		"MODULE_CANDIDATE_JS_SHA256":     "6026d692f0789ef6a496c89af5dafbbc5247f12005958155e84084aae094bd51",
		"MODULE_CANDIDATE_SPRITE_SHA256": "65cdb814125787460b548428dd49edd8e29250ee9eba5e6f27f4eb1b746fc3ca",
		"MODULE_CANDIDATE_LOGO_SHA256":   "5801b31fc6b1f54cde98b1a3f3f5e57553f6e67aa3fa0318879e5e2603cd540e",
	}

	t.Run("complete file proxy identity", func(t *testing.T) {
		if err := exactCandidateEnvironment(func(key string) string { return valid[key] }); err != nil {
			t.Fatalf("exact candidate environment: %v", err)
		}
	})

	for _, missing := range candidateEnvironmentKeys {
		missing := missing
		t.Run("missing "+missing, func(t *testing.T) {
			err := exactCandidateEnvironment(func(key string) string {
				if key == missing {
					return ""
				}
				return valid[key]
			})
			if err == nil || !strings.Contains(err.Error(), missing) {
				t.Fatalf("missing %s error = %v", missing, err)
			}
		})
	}

	t.Run("network proxy rejected", func(t *testing.T) {
		err := exactCandidateEnvironment(func(key string) string {
			if key == "MODULE_CANDIDATE_PROXY" {
				return "https://proxy.golang.org"
			}
			return valid[key]
		})
		if err == nil || !strings.Contains(err.Error(), "file://") {
			t.Fatalf("network proxy error = %v", err)
		}
	})
}
