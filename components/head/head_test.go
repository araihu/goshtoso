package head

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
)

type testLoaderEntry struct {
	Name                string `json:"name"`
	PrimaryURL          string `json:"primary_url"`
	FallbackURL         string `json:"fallback_url,omitempty"`
	Integrity           string `json:"integrity,omitempty"`
	WaitForWindowLoaded bool   `json:"wait_for_window_loaded,omitempty"`
}

type testLoaderConfig struct {
	Dependencies []testLoaderEntry `json:"dependencies"`
}

var loaderConfigPattern = regexp.MustCompile(`data-goshtoso-dependencies="([^"]+)"`)

func render(t *testing.T, c interface {
	Render(context.Context, io.Writer) error
}) string {
	return renderWithContext(t, context.Background(), c)
}

func renderWithContext(t *testing.T, ctx context.Context, c interface {
	Render(context.Context, io.Writer) error
}) string {
	t.Helper()
	var sb strings.Builder
	if err := c.Render(ctx, &sb); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func parseLoaderConfig(t *testing.T, output string) testLoaderConfig {
	t.Helper()
	match := loaderConfigPattern.FindStringSubmatch(output)
	if len(match) != 2 {
		t.Fatalf("rendered dependencies missing loader configuration\n%s", output)
	}
	var cfg testLoaderConfig
	if err := json.Unmarshal([]byte(html.UnescapeString(match[1])), &cfg); err != nil {
		t.Fatalf("decode loader configuration: %v\n%s", err, output)
	}
	return cfg
}

func loaderEntry(t *testing.T, cfg testLoaderConfig, name string) testLoaderEntry {
	t.Helper()
	for _, entry := range cfg.Dependencies {
		if entry.Name == name {
			return entry
		}
	}
	t.Fatalf("loader configuration missing %q: %#v", name, cfg.Dependencies)
	return testLoaderEntry{}
}

func TestDependenciesUsesVersionedPaths(t *testing.T) {
	out := render(t, Dependencies())
	for _, want := range []string{
		"/assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
		"/assets/js/runtime/alpinejs-collapse/3.14.9/alpine-collapse.min.js",
		"/assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js",
		"/assets/js/runtime/htmx.org/2.0.8/htmx.min.js",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Dependencies() missing versioned path %q\n%s", want, out)
		}
	}
	// Hard cut: no /vendor/ segment. Go module zips omit directories named vendor.
	for _, bad := range []string{`/assets/js/vendor/`, `vendor/alpine.min.js`, `vendor/htmx.min.js`} {
		if strings.Contains(out, bad) {
			t.Errorf("Dependencies() still emits flat path %q", bad)
		}
	}
}

func TestDependenciesMinimalUsesVersionedPaths(t *testing.T) {
	out := render(t, DependenciesMinimal())
	for _, want := range []string{
		"/assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
		"/assets/js/runtime/htmx.org/2.0.8/htmx.min.js",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("DependenciesMinimal() missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "/assets/js/vendor/") {
		t.Errorf("DependenciesMinimal() must avoid /vendor/ paths: %s", out)
	}
}

func TestDependenciesDefaultsToPinnedCDNWithLocalFallback(t *testing.T) {
	out := render(t, Dependencies())

	if !strings.Contains(out, `defer src="/assets/js/dependency-loader.js"`) {
		t.Fatalf("Dependencies() must use the ordered dependency loader\n%s", out)
	}
	cfg := parseLoaderConfig(t, out)
	wantOrder := []string{"alpine-collapse", "alpine-focus", "alpine-mask", "alpine", "htmx", "combobox"}
	if len(cfg.Dependencies) != len(wantOrder) {
		t.Fatalf("loader dependencies = %#v, want %v", cfg.Dependencies, wantOrder)
	}
	for i, want := range wantOrder {
		if cfg.Dependencies[i].Name != want {
			t.Errorf("loader dependency %d = %q, want %q", i, cfg.Dependencies[i].Name, want)
		}
	}

	for _, tc := range []struct {
		name, primary, fallback string
	}{
		{"alpine-collapse", assets.AlpineCollapseCDNURL, assets.AlpineCollapseURL},
		{"alpine-focus", assets.AlpineFocusCDNURL, assets.AlpineFocusURL},
		{"alpine-mask", assets.AlpineMaskCDNURL, assets.AlpineMaskURL},
		{"alpine", assets.AlpineJSCDNURL, assets.AlpineJSURL},
		{"htmx", assets.HTMXCDNURL, assets.HTMXURL},
		{"combobox", "/assets/js/combobox.js", ""},
	} {
		entry := loaderEntry(t, cfg, tc.name)
		if entry.PrimaryURL != tc.primary || entry.FallbackURL != tc.fallback {
			t.Errorf("%s source = (%q, %q), want (%q, %q)", tc.name, entry.PrimaryURL, entry.FallbackURL, tc.primary, tc.fallback)
		}
		if tc.name != "combobox" && !strings.HasPrefix(entry.Integrity, "sha384-") {
			t.Errorf("%s missing SHA-384 subresource integrity: %#v", tc.name, entry)
		}
	}
	if !loaderEntry(t, cfg, "htmx").WaitForWindowLoaded {
		t.Error("HTMX must wait for window load when inserted dynamically so its bootstrap cannot miss DOMContentLoaded")
	}
}

func TestDependenciesRenderFromPublicManifestCopy(t *testing.T) {
	manifest := customRuntimeManifest()
	cfg := newConfigFromManifest(manifest, []Option{
		WithDependencyCDNURL(DependencyHTMX, "https://override.example/htmx.js"),
		WithDependencyIntegrity(DependencyHTMX, "sha384-override"),
	})

	manifest.Stylesheet.PrimaryURL = "/mutated/styles.css"
	manifest.Loader.PrimaryURL = "/mutated/loader.js"
	manifest.Dependencies[0].PrimaryURL = "/mutated/collapse.js"

	out := render(t, dependenciesTemplate(cfg))
	if !strings.Contains(out, `href="/contract/styles.css"`) {
		t.Fatalf("stylesheet did not come from copied manifest\n%s", out)
	}
	if !strings.Contains(out, `src="/contract/loader.js"`) {
		t.Fatalf("loader did not come from copied manifest\n%s", out)
	}

	loader := parseLoaderConfig(t, out)
	if got := loaderEntry(t, loader, "alpine-collapse"); got.PrimaryURL != "https://primary.example/alpine-collapse.js" || got.FallbackURL != "/contract/alpine-collapse.js" {
		t.Fatalf("collapse entry = %#v", got)
	}
	if got := loaderEntry(t, loader, "htmx"); got.PrimaryURL != "https://override.example/htmx.js" || got.FallbackURL != "/contract/htmx.js" || got.Integrity != "sha384-override" || !got.WaitForWindowLoaded {
		t.Fatalf("option-adjusted HTMX entry = %#v", got)
	}
}

func TestDependenciesMinimalFiltersPublicManifestOrder(t *testing.T) {
	manifest := customRuntimeManifest()
	manifest.Dependencies[0].IncludeInMinimal = true
	manifest.Dependencies[3].IncludeInMinimal = false
	cfg := newConfigFromManifest(manifest, nil)
	loader := parseLoaderConfig(t, render(t, dependenciesMinimalTemplate(cfg)))

	want := []string{"alpine-collapse", "htmx", "combobox"}
	got := make([]string, 0, len(loader.Dependencies))
	for _, dependency := range loader.Dependencies {
		got = append(got, dependency.Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("minimal order = %v, want %v", got, want)
	}
}

func TestDependenciesLocalRuntimeUsesPublicManifestOrderAndDefer(t *testing.T) {
	manifest := customRuntimeManifest()
	manifest.Stylesheet.PrimaryURL = "https://primary.example/styles.css"
	manifest.Loader.PrimaryURL = "https://primary.example/loader.js"
	cfg := newConfigFromManifest(manifest, []Option{WithLocalRuntime()})
	out := render(t, dependenciesTemplate(cfg))

	if !strings.Contains(out, `href="/contract/styles.css"`) || strings.Contains(out, manifest.Stylesheet.PrimaryURL) {
		t.Fatalf("local runtime must use manifest stylesheet local URL\n%s", out)
	}
	previous := -1
	for _, dependency := range manifest.Dependencies {
		index := strings.Index(out, `src="`+dependency.LocalURL+`"`)
		if index < 0 {
			t.Fatalf("local rendering missing %s at %s\n%s", dependency.Role, dependency.LocalURL, out)
		}
		if index <= previous {
			t.Fatalf("local rendering order changed at %s\n%s", dependency.Role, out)
		}
		previous = index
	}
	if strings.Contains(out, manifest.Loader.LocalURL) {
		t.Fatalf("local runtime must not execute bootstrap loader\n%s", out)
	}
	if !strings.Contains(out, `<script src="/contract/htmx.js"`) {
		t.Fatalf("HTMX must preserve non-deferred direct tag semantics\n%s", out)
	}
	if !strings.Contains(out, `<script defer src="/contract/alpine.js"`) {
		t.Fatalf("Alpine must preserve deferred direct tag semantics\n%s", out)
	}
}

func customRuntimeManifest() assets.RuntimeManifest {
	manifest := assets.DefaultRuntimeManifest()
	manifest.Stylesheet.PrimaryURL = "/contract/styles.css"
	manifest.Stylesheet.LocalURL = "/contract/styles.css"
	manifest.Loader.PrimaryURL = "/contract/loader.js"
	manifest.Loader.LocalURL = "/contract/loader.js"
	for index := range manifest.Dependencies {
		dependency := &manifest.Dependencies[index]
		dependency.PrimaryURL = fmt.Sprintf("https://primary.example/%s.js", dependency.Role)
		dependency.LocalURL = fmt.Sprintf("/contract/%s.js", dependency.Role)
		dependency.Integrity = fmt.Sprintf("sha384-%s", dependency.Role)
	}
	manifest.Dependencies[len(manifest.Dependencies)-1].PrimaryURL = assets.ComboboxURL
	manifest.Dependencies[len(manifest.Dependencies)-1].LocalURL = assets.ComboboxURL
	manifest.Dependencies[len(manifest.Dependencies)-1].Integrity = ""
	return manifest
}

func TestDependenciesFunctionalOptionsOverrideIndividualSources(t *testing.T) {
	out := render(t, Dependencies(
		WithDependencyCDNURL(DependencyHTMX, "https://cdn.example.test/htmx-2.0.8.js"),
		WithDependencyLocalURL(DependencyHTMX, "/static/vendor/htmx-2.0.8.js"),
		WithDependencyIntegrity(DependencyHTMX, "sha384-custom"),
		WithStylesheetURL("/static/goshtoso.css"),
		WithComboboxURL("/static/combobox.js"),
		WithLoaderURL("/static/dependency-loader.js"),
	))

	if !strings.Contains(out, `href="/static/goshtoso.css"`) {
		t.Errorf("custom Dependencies() missing stylesheet override\n%s", out)
	}
	if !strings.Contains(out, `src="/static/dependency-loader.js"`) {
		t.Errorf("custom Dependencies() missing loader override\n%s", out)
	}
	cfg := parseLoaderConfig(t, out)
	htmx := loaderEntry(t, cfg, "htmx")
	if htmx.PrimaryURL != "https://cdn.example.test/htmx-2.0.8.js" || htmx.FallbackURL != "/static/vendor/htmx-2.0.8.js" || htmx.Integrity != "sha384-custom" {
		t.Errorf("HTMX override not applied: %#v", htmx)
	}
	if combobox := loaderEntry(t, cfg, "combobox"); combobox.PrimaryURL != "/static/combobox.js" {
		t.Errorf("combobox override not applied: %#v", combobox)
	}
}

func TestDependenciesCanDisableIntegrityForCustomSource(t *testing.T) {
	out := render(t, Dependencies(WithDependencyIntegrity(DependencyHTMX, "")))
	if htmx := loaderEntry(t, parseLoaderConfig(t, out), "htmx"); htmx.Integrity != "" {
		t.Fatalf("empty integrity override must disable SRI for a custom source: %#v", htmx)
	}
}

func TestDependenciesCanDisableLocalFallback(t *testing.T) {
	out := render(t, Dependencies(WithoutLocalFallback()))

	cfg := parseLoaderConfig(t, out)
	for _, entry := range cfg.Dependencies {
		if entry.FallbackURL != "" {
			t.Errorf("WithoutLocalFallback() retained fallback for %s: %#v", entry.Name, entry)
		}
	}
	if loaderEntry(t, cfg, "alpine").PrimaryURL != assets.AlpineJSCDNURL || loaderEntry(t, cfg, "htmx").PrimaryURL != assets.HTMXCDNURL {
		t.Fatalf("WithoutLocalFallback() must keep CDN primaries: %#v", cfg.Dependencies)
	}
}

func TestDependenciesCanUseLocalRuntimeAsPrimary(t *testing.T) {
	out := render(t, Dependencies(WithLocalRuntime()))

	for _, want := range []string{
		assets.AlpineCollapseURL,
		assets.AlpineFocusURL,
		assets.AlpineMaskURL,
		assets.AlpineJSURL,
		assets.HTMXURL,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("WithLocalRuntime() missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "https://unpkg.com/") || strings.Contains(out, "dependency-loader.js") {
		t.Fatalf("WithLocalRuntime() must not request the CDN or fallback loader\n%s", out)
	}
}

func TestDependenciesCanOmitAnOwnedDependency(t *testing.T) {
	out := render(t, Dependencies(WithoutDependency(DependencyAlpineMask)))

	cfg := parseLoaderConfig(t, out)
	for _, entry := range cfg.Dependencies {
		if entry.Name == "alpine-mask" {
			t.Fatalf("WithoutDependency(DependencyAlpineMask) retained mask: %#v", cfg.Dependencies)
		}
	}
	if loaderEntry(t, cfg, "alpine").PrimaryURL != assets.AlpineJSCDNURL {
		t.Fatalf("omitting one dependency must keep the rest of the defaults: %#v", cfg.Dependencies)
	}
}

func TestDependenciesMinimalAcceptsTheSameOptions(t *testing.T) {
	out := render(t, DependenciesMinimal(
		WithDependencyCDNURL(DependencyAlpineJS, "https://cdn.example.test/alpine.js"),
		WithoutLocalFallback(),
	))

	cfg := parseLoaderConfig(t, out)
	if loaderEntry(t, cfg, "alpine").PrimaryURL != "https://cdn.example.test/alpine.js" {
		t.Fatalf("DependenciesMinimal() must apply functional options: %#v", cfg.Dependencies)
	}
	for _, entry := range cfg.Dependencies {
		if entry.Name == "alpine-collapse" || entry.Name == "alpine-focus" || entry.Name == "alpine-mask" {
			t.Fatalf("DependenciesMinimal() must stay minimal: %#v", cfg.Dependencies)
		}
	}
}

func TestDependenciesZeroValueKeepsStrongDefaults(t *testing.T) {
	var instance Instance
	cfg := parseLoaderConfig(t, render(t, instance))
	if loaderEntry(t, cfg, "alpine-mask").FallbackURL != assets.AlpineMaskURL {
		t.Fatalf("zero-value Instance lost default runtime: %#v", cfg.Dependencies)
	}
}

func TestDependenciesPropagatesTemplNonceToLoader(t *testing.T) {
	ctx := templ.WithNonce(context.Background(), "test-nonce-123")
	out := renderWithContext(t, ctx, Dependencies())
	if !strings.Contains(out, `nonce="test-nonce-123"`) {
		t.Fatalf("Dependencies() must propagate templ CSP nonce to the loader\n%s", out)
	}
}

func TestDependenciesPropagatesTemplNonceToEveryLocalScript(t *testing.T) {
	ctx := templ.WithNonce(context.Background(), "offline-nonce-456")
	out := renderWithContext(t, ctx, Dependencies(WithLocalRuntime()))
	if got, want := strings.Count(out, `nonce="offline-nonce-456"`), 6; got != want {
		t.Fatalf("WithLocalRuntime() nonce count = %d, want %d\n%s", got, want, out)
	}
}
