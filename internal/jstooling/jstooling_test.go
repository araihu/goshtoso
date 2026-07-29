package jstooling

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectInlineJavaScriptFindsExtractionCandidates(t *testing.T) {
	t.Parallel()

	source := `package fixture

func clientScript() string {
	return strings.Join([]string{
		"document.addEventListener('load', function () {",
		"window.example = true;",
		"});",
	}, "")
}

templ fixture() {
	<div x-data="{
		open: false,
	}"></div>
	<button x-on:click="open = !open">Allowed</button>
	<button x-on:click="open = true; window.dispatchEvent(new CustomEvent('fixture')); cleanup()">Too large</button>
	<script>
		window.fixture = true
	</script>
	@templ.Raw("<script>window.rawFixture = true</script>")
	<script type="application/ld+json">{"name":"data, not executable JavaScript"}</script>
}
`

	findings := DetectInlineJavaScript("fixture.templ", []byte(source))
	wantKinds := map[string]bool{
		"event-expression": false,
		"js-builder":       false,
		"script-body":      false,
		"templ-raw-script": false,
	}
	for _, finding := range findings {
		if _, ok := wantKinds[finding.Kind]; ok {
			wantKinds[finding.Kind] = true
		}
		if finding.Line <= 0 || finding.Path != "fixture.templ" {
			t.Fatalf("invalid finding location: %#v", finding)
		}
	}
	for kind, found := range wantKinds {
		if !found {
			t.Errorf("missing %s finding: %#v", kind, findings)
		}
	}
	for _, finding := range findings {
		if strings.Contains(finding.Summary, "open = !open") {
			t.Fatalf("small one-line expression must be allowed: %#v", finding)
		}
		if strings.Contains(finding.Summary, "application/ld+json") {
			t.Fatalf("non-executable JSON script must be ignored: %#v", finding)
		}
	}
}

func TestSimilarityReportFindsStructurallySimilarFunctions(t *testing.T) {
	t.Parallel()

	sources := map[string][]byte{
		"a.js": []byte(`function first(root) { const items = root.querySelectorAll("button"); items.forEach(function (item) { item.hidden = true; }); return items.length; }`),
		"b.js": []byte(`function second(node) { const controls = node.querySelectorAll("button"); controls.forEach(function (control) { control.hidden = true; }); return controls.length; }`),
		"c.js": []byte(`function unrelated(value) { return String(value).toUpperCase(); }`),
	}

	pairs := SimilarFunctions(sources, 0.80)
	if len(pairs) == 0 {
		t.Fatal("expected structurally similar function pair")
	}
	if pairs[0].Left.Path != "a.js" || pairs[0].Right.Path != "b.js" {
		t.Fatalf("first pair = %#v, want a.js and b.js", pairs[0])
	}
	if pairs[0].Score < 0.80 || pairs[0].Score > 1 {
		t.Fatalf("score = %f, want [0.80, 1]", pairs[0].Score)
	}
}

func TestBuildWritesDeterministicMinifiedArtifactsAndCheckDetectsDrift(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sources := map[string]string{
		"assets/js/src/combobox.js":                    `(() => { window.comboboxFixture = true })();`,
		"assets/js/src/action-group.js":                `(() => { window.actionGroupFixture = true })();`,
		"assets/js/src/components/structured-input.js": `(() => { window.structuredInputFixture = true })();`,
		"assets/js/src/components/tooltip.js":          `(() => { window.tooltipFixture = true })();`,
		"assets/js/src/components/data.js":             `(() => { window.dataFixture = true })();`,
		"assets/js/src/components/carousel.js":         `(() => { window.carouselFixture = true })();`,
		"assets/js/src/components/dropdown.js":         `(() => { window.dropdownFixture = true })();`,
		"assets/js/src/components/palette.js":          `(() => { window.paletteFixture = true })();`,
		"assets/js/src/components/select.js":           `(() => { window.selectFixture = true })();`,
		"assets/js/src/components/tabs.js":             `(() => { window.tabsFixture = true })();`,
		"site/assets/js/src/action-group.js":           `(() => { window.actionGroupDemoFixture = true })();`,
		"site/assets/js/src/avatar-showcase.js":        `(() => { window.avatarShowcaseFixture = true })();`,
		"site/assets/js/src/log-feed.js":               `(() => { window.logFeedFixture = true })();`,
		"site/assets/js/src/chat.js":                   `(() => { window.chatFixture = true })();`,
		"site/assets/js/src/profile-images.js":         `(() => { window.profileImagesFixture = true })();`,
		"site/assets/js/src/ticker-pane.js":            `(() => { window.tickerPaneFixture = true })();`,
		"site/assets/js/src/theme-page.js":             `(() => { window.themePageFixture = true })();`,
		"assets/js/src/darkmode.js":                    `document.addEventListener("alpine:init", () => window.darkFixture = true);`,
		"assets/js/src/dependency-loader.js":           `(() => { window.loaderFixture = Promise.resolve(true) })();`,
	}
	for name, content := range sources {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(results) != 6 {
		t.Fatalf("artifact count = %d, want 6", len(results))
	}
	bundle, err := os.ReadFile(filepath.Join(root, "assets", "js", "goshtoso.min.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(bundle), generatedHeader) {
		t.Fatalf("bundle missing generated header: %q", bundle)
	}
	siteBundle, err := os.ReadFile(filepath.Join(root, "site", "assets", "js", "goshtoso-demo.min.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(siteBundle), generatedHeader) {
		t.Fatalf("site bundle missing generated header: %q", siteBundle)
	}
	assertSplitBundleContents(t, bundle, siteBundle)
	if strings.Contains(string(bundle), "\n  ") || strings.Contains(string(siteBundle), "\n  ") {
		t.Fatalf("bundles were not minified: component=%s site=%s", bundle, siteBundle)
	}
	if _, err := Build(root, true); err != nil {
		t.Fatalf("clean check: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "js", "goshtoso.min.js"), []byte("drift"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site", "assets", "js", "goshtoso-demo.min.js"), []byte("drift"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root, true); err == nil ||
		!strings.Contains(err.Error(), "assets/js/goshtoso.min.js") ||
		!strings.Contains(err.Error(), "site/assets/js/goshtoso-demo.min.js") {
		t.Fatalf("split-bundle drift check error = %v", err)
	}
}

func assertSplitBundleContents(t *testing.T, componentBundle, siteBundle []byte) {
	t.Helper()

	componentFixtures := []string{
		"comboboxFixture", "actionGroupFixture", "structuredInputFixture", "tooltipFixture",
		"dataFixture", "carouselFixture", "dropdownFixture", "paletteFixture", "selectFixture", "tabsFixture",
	}
	demoFixtures := []string{
		"actionGroupDemoFixture", "avatarShowcaseFixture", "logFeedFixture",
		"chatFixture", "profileImagesFixture", "tickerPaneFixture", "themePageFixture",
	}
	assertBundleContainsOnly(t, "component bundle", componentBundle, componentFixtures, demoFixtures)
	assertBundleContainsOnly(t, "site bundle", siteBundle, demoFixtures, componentFixtures)
	assertBundleOrder(t, "component bundle", componentBundle, componentFixtures)
	assertBundleOrder(t, "site bundle", siteBundle, demoFixtures)
}

func assertBundleContainsOnly(t *testing.T, label string, bundle []byte, required, forbidden []string) {
	t.Helper()

	for _, fixture := range required {
		if !strings.Contains(string(bundle), fixture) {
			t.Fatalf("%s missing %s: %s", label, fixture, bundle)
		}
	}
	for _, fixture := range forbidden {
		if strings.Contains(string(bundle), fixture) {
			t.Fatalf("%s contains forbidden %s: %s", label, fixture, bundle)
		}
	}
}

func assertBundleOrder(t *testing.T, label string, bundle []byte, fixtures []string) {
	t.Helper()

	for index := 1; index < len(fixtures); index++ {
		if strings.Index(string(bundle), fixtures[index-1]) > strings.Index(string(bundle), fixtures[index]) {
			t.Fatalf("%s sources out of order at %s then %s: %s", label, fixtures[index-1], fixtures[index], bundle)
		}
	}
}
