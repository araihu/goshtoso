package skillgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesExternalSkillReference(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "types.go"), `package alert

import "github.com/a-h/templ"

// Config controls the alert.
type Config struct {
	// Message is the alert body.
	Message string
}

func Alert(config Config) templ.Component { return nil }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	legacy := mustRead(t, ".claude/skills/using-goshtoso/components-reference.md")
	external := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")

	if legacy != external {
		t.Fatal("external skill reference differs from legacy compatibility reference")
	}
	if !strings.Contains(external, `import "github.com/araihu/goshtoso/components/alert"`) {
		t.Fatalf("external reference missing component import:\n%s", external)
	}
}

func TestRunDocumentsFunctionalOptions(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "button", "types.go"), `package button

type Tone string

const TonePrimary Tone = "primary"

type config struct {
	tone Tone
}

type Option interface {
	apply(*config)
}

func WithTone(tone Tone) Option { return nil }
`)
	mustWrite(t, filepath.Join("components", "button", "button_templ.go"), `package button

import "github.com/a-h/templ"

func Button(options ...Option) templ.Component { return nil }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	if !strings.Contains(reference, "**Options:** `WithTone(tone Tone)`") {
		t.Fatalf("generated reference missing functional option signature:\n%s", reference)
	}
}

func TestRunDocumentsConcreteComponentConstructors(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "types.go"), `package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type Config struct{}

type Instance struct{}

func (Instance) Kind() components.Kind { return components.KindAlert }

func (Instance) Render(context.Context, io.Writer) error { return nil }

func Alert(config Config) Instance { return Instance{} }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	if !strings.Contains(reference, "`Alert(config Config)`") {
		t.Fatalf("generated reference missing concrete component constructor:\n%s", reference)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(contents)
}
