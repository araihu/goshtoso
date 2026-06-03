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

	first := cfg.GetID()
	second := cfg.GetID()

	if first == "" {
		t.Fatal("GetID returned empty ID")
	}
	if first != second {
		t.Fatalf("GetID = %q then %q; want deterministic result", first, second)
	}
	if strings.Contains(first, " ") {
		t.Fatalf("GetID contains whitespace: %q", first)
	}
}

func TestConfigGetIDHonorsExplicitID(t *testing.T) {
	cfg := Config{ID: "install-snippet"}

	if got := cfg.GetID(); got != "install-snippet" {
		t.Fatalf("GetID = %q; want explicit ID", got)
	}
}

func TestConfigGetIDConcurrentCallsAreStable(t *testing.T) {
	cfg := Config{Language: "bash", Label: "Install", Code: "go test ./components/..."}
	want := cfg.GetID()

	var wg sync.WaitGroup
	errs := make(chan string, 32)
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := cfg.GetID(); got != want {
				errs <- got
			}
		}()
	}
	wg.Wait()
	close(errs)

	for got := range errs {
		t.Fatalf("concurrent GetID = %q; want %q", got, want)
	}
}

func TestCodeBlockCopyButtonHasDistinctAccessibleName(t *testing.T) {
	html := renderCodeBlock(t, Config{
		ID:       "component-example",
		Language: "go",
		Label:    "Button usage",
		Code:     `@button.Button(button.Config{})`,
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
