package codeblock

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
)

func renderCodeBlock(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := CodeBlock(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render CodeBlock: %v", err)
	}
	return buf.String()
}

func TestConfigGetIDIsDeterministicWithoutMutableCounter(t *testing.T) {
	cfg := Config{Language: "go", Label: "Example", Code: "fmt.Println(\"hi\")"}

	first := cfg.getID()
	second := cfg.getID()

	if first == "" {
		t.Fatal("getID returned empty ID")
	}
	if first != second {
		t.Fatalf("getID = %q then %q; want deterministic result", first, second)
	}
	if strings.Contains(first, " ") {
		t.Fatalf("getID contains whitespace: %q", first)
	}
}

func TestConfigGetIDHonorsExplicitID(t *testing.T) {
	cfg := Config{ID: "install-snippet"}

	if got := cfg.getID(); got != "install-snippet" {
		t.Fatalf("getID = %q; want explicit ID", got)
	}
}

func TestConfigGetIDConcurrentCallsAreStable(t *testing.T) {
	cfg := Config{Language: "bash", Label: "Install", Code: "go test ./components/..."}
	want := cfg.getID()

	var wg sync.WaitGroup
	errs := make(chan string, 32)
	for range 32 {
		wg.Go(func() {
			if got := cfg.getID(); got != want {
				errs <- got
			}
		})
	}
	wg.Wait()
	close(errs)

	for got := range errs {
		t.Fatalf("concurrent getID = %q; want %q", got, want)
	}
}

func TestCodeBlockCopyButtonHasDistinctAccessibleName(t *testing.T) {
	html := renderCodeBlock(t, Config{
		ID:       "component-example",
		Language: "go",
		Label:    "Button usage",
		Code:     `@button.Button()`,
	})

	if !strings.Contains(html, `aria-label="Copy Button usage code"`) {
		t.Fatalf("copy button missing distinct aria-label:\n%s", html)
	}
}

func TestCodeBlockEscapesIDInAlpineExpression(t *testing.T) {
	html := renderCodeBlock(t, Config{
		ID:   "code'block",
		Code: "x",
	})

	if strings.Contains(html, `document.getElementById('code'block')`) {
		t.Fatalf("raw ID leaked into JS string:\n%s", html)
	}
	if !strings.Contains(html, `document.getElementById(&#39;code\&#39;block&#39;)`) &&
		!strings.Contains(html, `document.getElementById('code\'block')`) {
		t.Fatalf("escaped ID not found in Alpine expression:\n%s", html)
	}
}

func TestCodeBlockCompactDensityReducesHeaderAndBodySpacing(t *testing.T) {
	html := renderCodeBlock(t, Config{
		Language: "bash",
		Label:    "Install",
		Code:     "go get github.com/araihu/goshtoso@latest",
		Density:  DensityCompact,
	})

	if !strings.Contains(html, `data-density="compact"`) {
		t.Fatalf("compact code block missing density identity:\n%s", html)
	}
	if !strings.Contains(html, `data-code-block-header`) || !strings.Contains(html, `px-3 py-1.5`) {
		t.Fatalf("compact code block missing compact header spacing:\n%s", html)
	}
	if !strings.Contains(html, `class="codeblock codeblock-compact overflow-x-auto"`) {
		t.Fatalf("compact code block missing compact body class:\n%s", html)
	}
}

func TestCodeBlockCopiedFeedbackRejectsRawPaletteClass(t *testing.T) {
	html := renderCodeBlock(t, Config{
		ID:       "semantic-copy-feedback",
		Language: "go",
		Label:    "main.go",
		Code:     `fmt.Println("copied")`,
	})

	if strings.Contains(html, "text-green-500") {
		t.Fatalf("copied feedback uses forbidden raw palette class text-green-500:\n%s", html)
	}
}

func TestCodeBlockCopiedFeedbackUsesSemanticSuccessTextRoles(t *testing.T) {
	html := renderCodeBlock(t, Config{
		ID:       "semantic-copy-feedback",
		Language: "go",
		Label:    "main.go",
		Code:     `fmt.Println("copied")`,
	})

	if !strings.Contains(html, "text-success-text dark:text-success-text-dark") {
		t.Fatalf("copied feedback missing semantic success text roles:\n%s", html)
	}
}
