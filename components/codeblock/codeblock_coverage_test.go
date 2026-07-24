package codeblock

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func renderToString(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := CodeBlock(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render CodeBlock: %v", err)
	}
	return buf.String()
}

// TestRenderMaxHeightAttribute covers the conditional style attribute branch in
// the generated templ: present when MaxHeight is set, absent otherwise.
func TestRenderMaxHeightAttribute(t *testing.T) {
	withMax := renderToString(t, Config{Language: "go", Code: "x", MaxHeight: "180px"})
	if !strings.Contains(withMax, "max-height: 180px; overflow-y: auto;") {
		t.Fatalf("expected max-height style attr in render:\n%s", withMax)
	}

	noMax := renderToString(t, Config{Language: "go", Code: "x"})
	if strings.Contains(noMax, "max-height:") {
		t.Fatalf("did not expect max-height style when MaxHeight unset:\n%s", noMax)
	}
}

// TestRenderHeaderAndCopyButton asserts the label, copy aria-label, and the
// highlighted code id all land in the rendered markup.
func TestRenderHeaderAndCopyButton(t *testing.T) {
	cfg := Config{ID: "demo-block", Language: "go", Label: "main.go", Code: `fmt.Println("hi")`}
	html := renderToString(t, cfg)

	if !strings.Contains(html, ">main.go<") {
		t.Fatalf("label not rendered in header:\n%s", html)
	}
	if !strings.Contains(html, `aria-label="Copy main.go code"`) {
		t.Fatalf("copy button aria-label missing:\n%s", html)
	}
	if !strings.Contains(html, `id="demo-block"`) {
		t.Fatalf("code container id missing:\n%s", html)
	}
	if !strings.Contains(html, "ch-chroma") {
		t.Fatalf("highlighted code missing:\n%s", html)
	}
}

// TestGetIDBaseSelection exercises the slug-base fallback chain in getID:
// explicit ID wins, else label slug, else language slug, else "snippet".
func TestGetIDBaseSelection(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantSub string // substring that must appear in the generated ID
	}{
		{
			name:    "explicit id wins verbatim",
			cfg:     Config{ID: "install-snippet", Language: "go", Label: "Example"},
			wantSub: "install-snippet",
		},
		{
			name:    "label drives the slug base",
			cfg:     Config{Label: "Button Usage", Language: "go", Code: "x"},
			wantSub: "codeblock-button-usage-",
		},
		{
			name:    "language drives base when label empty",
			cfg:     Config{Language: "Bash", Code: "go test ./..."},
			wantSub: "codeblock-bash-",
		},
		{
			name:    "snippet fallback when label and language slug empty",
			cfg:     Config{Label: "!!!", Language: "***", Code: "x"},
			wantSub: "codeblock-snippet-",
		},
		{
			name:    "snippet fallback when everything empty",
			cfg:     Config{},
			wantSub: "codeblock-snippet-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.getID()
			if !strings.HasPrefix(got, tt.wantSub) && !strings.Contains(got, tt.wantSub) {
				t.Fatalf("getID() = %q; want to contain %q", got, tt.wantSub)
			}
		})
	}
}

// TestGetIDDiffersByContent confirms the hash suffix changes the ID when the
// code differs even though the base slug is identical.
func TestGetIDDiffersByContent(t *testing.T) {
	a := Config{Language: "go", Label: "Same", Code: "one"}.getID()
	b := Config{Language: "go", Label: "Same", Code: "two"}.getID()
	if a == b {
		t.Fatalf("expected distinct IDs for differing code, both = %q", a)
	}
	if !strings.HasPrefix(a, "codeblock-same-") || !strings.HasPrefix(b, "codeblock-same-") {
		t.Fatalf("unexpected base slug: a=%q b=%q", a, b)
	}
}

// TestGetLabelDefaultsToLanguage covers both getLabel branches.
func TestGetLabelDefaultsToLanguage(t *testing.T) {
	if got := (Config{Label: "Custom", Language: "go"}).getLabel(); got != "Custom" {
		t.Fatalf("getLabel() = %q; want explicit label", got)
	}
	if got := (Config{Language: "bash"}).getLabel(); got != "bash" {
		t.Fatalf("getLabel() = %q; want language fallback", got)
	}
	if got := (Config{}).getLabel(); got != "" {
		t.Fatalf("getLabel() = %q; want empty when both unset", got)
	}
}

// TestCopyLabelFallsBackToID covers the copyLabel branch where getLabel is
// empty and the accessible name is built from the generated ID instead.
func TestCopyLabelFallsBackToID(t *testing.T) {
	withLabel := Config{Label: "Install", Language: "bash"}.copyLabel()
	if withLabel != "Copy Install code" {
		t.Fatalf("copyLabel() = %q; want %q", withLabel, "Copy Install code")
	}

	cfg := Config{Code: "x"} // no label, no language
	got := cfg.copyLabel()
	id := cfg.getID()
	want := "Copy " + id + " code"
	if got != want {
		t.Fatalf("copyLabel() = %q; want %q (ID fallback)", got, want)
	}
}

// TestMaxHeightStyle covers both branches of maxHeightStyle.
func TestMaxHeightStyle(t *testing.T) {
	if got := (Config{MaxHeight: "200px"}).maxHeightStyle(); got != "max-height: 200px; overflow-y: auto;" {
		t.Fatalf("maxHeightStyle() = %q", got)
	}
	if got := (Config{}).maxHeightStyle(); got != "" {
		t.Fatalf("maxHeightStyle() = %q; want empty", got)
	}
}

// TestSlugPart exercises trimming, lowercasing, separator collapsing, and the
// empty result for symbol-only input.
func TestSlugPart(t *testing.T) {
	tests := map[string]string{
		"Hello World":  "hello-world",
		"Go/Templ":     "go-templ",
		"--Trim--":     "trim",
		"***":          "",
		"":             "",
		"MixedCASE123": "mixedcase123",
	}
	for in, want := range tests {
		if got := slugPart(in); got != want {
			t.Fatalf("slugPart(%q) = %q; want %q", in, got, want)
		}
	}
}

// TestJSStringSingle covers every escape branch of jsStringSingle: backslash,
// single quote, newline, carriage return, U+2028, U+2029, and the default
// passthrough for ordinary runes.
func TestJSStringSingle(t *testing.T) {
	ls := string(rune(0x2028)) // line separator
	ps := string(rune(0x2029)) // paragraph separator

	in := "a'b\\c\nd\re" + ls + "f" + ps + "g中"
	want := "a\\'b\\\\c\\nd\\re\\u2028f\\u2029g中"
	if got := jsStringSingle(in); got != want {
		t.Fatalf("jsStringSingle(%q) = %q; want %q", in, got, want)
	}

	if got := jsStringSingle("plain"); got != "plain" {
		t.Fatalf("jsStringSingle(plain) = %q; want unchanged", got)
	}
}

// TestPickLexerSelection covers the templ alias, the empty fallback, a known
// language, and an unknown language (which falls back, not nil).
func TestPickLexerSelection(t *testing.T) {
	goLexer := pickLexer("go")
	if goLexer == nil {
		t.Fatal("pickLexer(go) returned nil")
	}

	templLexer := pickLexer("templ")
	if templLexer == nil {
		t.Fatal("pickLexer(templ) returned nil")
	}
	if templLexer.Config().Name != goLexer.Config().Name {
		t.Fatalf("pickLexer(templ) = %q; want same lexer as go (%q)",
			templLexer.Config().Name, goLexer.Config().Name)
	}

	if pickLexer("") == nil {
		t.Fatal("pickLexer(empty) returned nil; want fallback")
	}
	if pickLexer("totally-unknown-lang") == nil {
		t.Fatal("pickLexer(unknown) returned nil; want fallback")
	}
}

// TestRenderHighlightsKnownLanguage confirms Render emits Chroma class-mode
// markup with the ch- prefix for a recognized language.
func TestRenderHighlightsKnownLanguage(t *testing.T) {
	out := Render("package main\n\nfunc main() {}", "go")
	if !strings.Contains(out, "ch-chroma") {
		t.Fatalf("Render output missing ch-chroma wrapper:\n%s", out)
	}
	if !strings.Contains(out, `<span class="ch-`) {
		t.Fatalf("Render output missing ch- token spans:\n%s", out)
	}
}

// TestRenderFallbackLanguageStillHighlights confirms unknown/empty languages
// route through the fallback lexer (chroma), not plainPre.
func TestRenderFallbackLanguageStillHighlights(t *testing.T) {
	out := Render("just some text", "")
	if !strings.Contains(out, "ch-chroma") {
		t.Fatalf("fallback Render missing ch-chroma wrapper:\n%s", out)
	}

	unknown := Render("more text", "no-such-lang")
	if !strings.Contains(unknown, "ch-chroma") {
		t.Fatalf("unknown-lang Render missing ch-chroma wrapper:\n%s", unknown)
	}
}

// TestRenderTemplAliasUsesGo confirms the templ alias highlights via the Go
// lexer, producing Chroma markup.
func TestRenderTemplAliasUsesGo(t *testing.T) {
	out := Render("func main() {}", "templ")
	if !strings.Contains(out, "ch-chroma") {
		t.Fatalf("templ Render missing ch-chroma wrapper:\n%s", out)
	}
}

// TestPlainPreEscapesHTML directly exercises plainPre, the unhighlighted
// fallback path (0% covered): it must HTML-escape the source.
func TestPlainPreEscapesHTML(t *testing.T) {
	out := plainPre(`<script>alert("x")</script> & 'quote'`)
	if strings.Contains(out, "<script>") {
		t.Fatalf("plainPre did not escape <script>:\n%s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Fatalf("plainPre missing escaped tag:\n%s", out)
	}
	if !strings.HasPrefix(out, `<pre class="chroma"><code>`) {
		t.Fatalf("plainPre missing pre/code wrapper:\n%s", out)
	}
	if !strings.HasSuffix(out, "</code></pre>") {
		t.Fatalf("plainPre missing closing tags:\n%s", out)
	}
}
