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
	sourceDir := filepath.Join(root, "assets", "js", "src")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{
		"combobox.js":          `(() => { window.comboboxFixture = true })();`,
		"action-group.js":      `(() => { window.actionGroupFixture = true })();`,
		"darkmode.js":          `document.addEventListener("alpine:init", () => window.darkFixture = true);`,
		"dependency-loader.js": `(() => { window.loaderFixture = Promise.resolve(true) })();`,
	}
	for name, content := range sources {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("artifact count = %d, want 5", len(results))
	}
	bundle, err := os.ReadFile(filepath.Join(root, "assets", "js", "goshtoso.min.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(bundle), generatedHeader) {
		t.Fatalf("bundle missing generated header: %q", bundle)
	}
	if !strings.Contains(string(bundle), "comboboxFixture") || !strings.Contains(string(bundle), "actionGroupFixture") {
		t.Fatalf("bundle missing ordered sources: %s", bundle)
	}
	if strings.Contains(string(bundle), "\n  ") {
		t.Fatalf("bundle was not minified: %s", bundle)
	}
	if _, err := Build(root, true); err != nil {
		t.Fatalf("clean check: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "js", "goshtoso.min.js"), []byte("drift"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(root, true); err == nil || !strings.Contains(err.Error(), "goshtoso.min.js") {
		t.Fatalf("drift check error = %v", err)
	}
}
