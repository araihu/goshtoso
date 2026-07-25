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
	if !strings.Contains(external, "`Alert(config Config)`") {
		t.Fatalf("external reference missing legacy templ.Component constructor:\n%s", external)
	}
}

func TestRunIntroducesTheComponentModel(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "types.go"), `package alert

import "github.com/a-h/templ"

func Alert() templ.Component { return nil }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	for _, phrase := range []string{
		"Theme",
		"Primitive",
		"Kind",
		"configuration dimension",
		"docs/COMPONENT_MODEL.md",
	} {
		if !strings.Contains(reference, phrase) {
			t.Errorf("generated introduction missing %q", phrase)
		}
	}
	if t.Failed() {
		t.Logf("generated reference:\n%s", reference)
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

func TestRunRejectsConcreteTypesWithWrongMethodSignatures(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "report", "component.go"), `package report

type Report struct{}

func (Report) Kind() string { return "report" }

func (Report) Render() string { return "" }

func BuildReport() Report { return Report{} }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	if strings.Contains(reference, "`BuildReport()`") {
		t.Fatalf("generated reference includes constructor with wrong component method signatures:\n%s", reference)
	}
}

func TestRunRespectsConcreteTypeMethodSets(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "component.go"), `package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type Instance struct{}

func (Instance) Kind() components.Kind { return components.KindAlert }

func (*Instance) Render(context.Context, io.Writer) error { return nil }

func PointerAlert() *Instance { return &Instance{} }

func ValueAlert() Instance { return Instance{} }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	hasPointer := strings.Contains(reference, "`PointerAlert()`")
	hasValue := strings.Contains(reference, "`ValueAlert()`")
	if !hasPointer || hasValue {
		t.Fatalf(
			"generated reference method-set entries: PointerAlert=%t ValueAlert=%t\n%s",
			hasPointer,
			hasValue,
			reference,
		)
	}
}

func TestRunResolvesComponentContractImportsByPath(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "types.go"), `package alert

type Instance struct{}
`)
	mustWrite(t, filepath.Join("components", "alert", "kind.go"), `package alert

import ui "github.com/araihu/goshtoso/components"

func (Instance) Kind() ui.Kind { return ui.KindAlert }
`)
	mustWrite(t, filepath.Join("components", "alert", "render.go"), `package alert

import (
	ctx "context"
	sink "io"
)

func (Instance) Render(ctx.Context, sink.Writer) error { return nil }
`)
	mustWrite(t, filepath.Join("components", "alert", "constructors.go"), `package alert

import view "github.com/a-h/templ"

func ConcreteAlert() Instance { return Instance{} }

func LegacyAlert() view.Component { return nil }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	for _, entry := range []string{"`ConcreteAlert()`", "`LegacyAlert()`"} {
		if !strings.Contains(reference, entry) {
			t.Fatalf("generated reference missing aliased-import entry %s:\n%s", entry, reference)
		}
	}
}

func TestRunRejectsUnrelatedContractImportAliases(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "report", "component.go"), `package report

import (
	context "example.com/context"
	io "example.com/io"
	components "example.com/components"
	templ "example.com/templ"
)

type Instance struct{}

func (Instance) Kind() components.Kind { return "" }

func (Instance) Render(context.Context, io.Writer) error { return nil }

func ConcreteReport() Instance { return Instance{} }

func LegacyReport() templ.Component { return nil }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	for _, entry := range []string{"`ConcreteReport()`", "`LegacyReport()`"} {
		if strings.Contains(reference, entry) {
			t.Fatalf("generated reference includes entry using unrelated import aliases %s:\n%s", entry, reference)
		}
	}
}

func TestRunRejectsShadowedPredeclaredError(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "shadow.go"), `package alert

type error string
`)
	mustWrite(t, filepath.Join("components", "alert", "component.go"), `package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type Instance struct{}

func (Instance) Kind() components.Kind { return components.KindAlert }

func (Instance) Render(context.Context, io.Writer) error { return "" }

func Alert() Instance { return Instance{} }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	if strings.Contains(reference, "`Alert()`") {
		t.Fatalf("generated reference includes constructor with shadowed error result:\n%s", reference)
	}
}

func TestRunRequiresSingleLogicalConstructorResult(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "component.go"), `package alert

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
)

type Instance struct{}

func (Instance) Kind() components.Kind { return components.KindAlert }

func (Instance) Render(context.Context, io.Writer) error { return nil }

func NamedLegacy() (component templ.Component) { return nil }

func NamedConcrete() (component Instance) { return Instance{} }

func LegacyWithError() (templ.Component, error) { return nil, nil }

func LegacyGrouped() (first, second templ.Component) { return nil, nil }

func ConcreteWithError() (Instance, error) { return Instance{}, nil }

func ConcreteGrouped() (first, second Instance) { return Instance{}, Instance{} }

func DoublePointer() **Instance { return nil }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	for _, entry := range []string{"`NamedLegacy()`", "`NamedConcrete()`"} {
		if !strings.Contains(reference, entry) {
			t.Errorf("generated reference missing valid named single result %s", entry)
		}
	}
	for _, entry := range []string{
		"`LegacyWithError()`",
		"`LegacyGrouped()`",
		"`ConcreteWithError()`",
		"`ConcreteGrouped()`",
		"`DoublePointer()`",
	} {
		if strings.Contains(reference, entry) {
			t.Errorf("generated reference includes invalid constructor result %s", entry)
		}
	}
	if t.Failed() {
		t.Logf("generated reference:\n%s", reference)
	}
}

func TestRunIgnoresTestOnlyComponentDeclarations(t *testing.T) {
	t.Chdir(t.TempDir())

	mustWrite(t, filepath.Join("components", "alert", "component.go"), `package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type Instance struct{}

func (Instance) Kind() components.Kind { return components.KindAlert }

func (Instance) Render(context.Context, io.Writer) error { return nil }

func Alert() Instance { return Instance{} }
`)
	mustWrite(t, filepath.Join("components", "alert", "external_test.go"), `package alert_test

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type error string

type ExternalInstance struct{}

func (ExternalInstance) Kind() components.Kind { return components.KindAlert }

func (ExternalInstance) Render(context.Context, io.Writer) error { return "" }

func ExternalTestOnly() ExternalInstance { return ExternalInstance{} }
`)
	mustWrite(t, filepath.Join("components", "alert", "internal_test.go"), `package alert

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

type error string

type InternalInstance struct{}

func (InternalInstance) Kind() components.Kind { return components.KindAlert }

func (InternalInstance) Render(context.Context, io.Writer) error { return "" }

func InternalTestOnly() InternalInstance { return InternalInstance{} }
`)

	if err := Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	reference := mustRead(t, ".agents/skills/using-goshtoso/references/components-reference.md")
	if !strings.Contains(reference, "`Alert()`") {
		t.Errorf("generated reference missing production constructor")
	}
	for _, entry := range []string{"`ExternalTestOnly()`", "`InternalTestOnly()`"} {
		if strings.Contains(reference, entry) {
			t.Errorf("generated reference includes test-only constructor %s", entry)
		}
	}
	if t.Failed() {
		t.Logf("generated reference:\n%s", reference)
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
