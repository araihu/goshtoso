package docs_test

import (
	"os"
	"strings"
	"testing"
)

func TestAgentSkillRoutesFromIntegrationToApplicationPatterns(t *testing.T) {
	skill := readDoc(t, "../.agents/skills/using-goshtoso/SKILL.md")

	for _, want := range []string{
		"## From First Component to Application",
		"references/application-patterns.md",
		"references/visual-acceptance.md",
		`mux.Handle("GET /assets/", assets.Handler())`,
		"silently",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("agent skill missing application guidance %q", want)
		}
	}
}

func TestReadmeMakesApplicationGuidanceDiscoverable(t *testing.T) {
	readme := readDoc(t, "../README.md")
	for _, want := range []string{
		"application-patterns.md",
		"App Shell",
		`mux.Handle("GET /assets/", assets.Handler())`,
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README missing application guidance %q", want)
		}
	}
	if strings.Contains(readme, "distribution checks") {
		t.Error("README sends consumers to maintainer-only distribution checks")
	}
}

func TestReleaseChecklistProtectsApplicationReferences(t *testing.T) {
	checklist := readDoc(t, "RELEASE_CHECKLIST.md")
	for _, want := range []string{
		"references/application-patterns.md",
		"references/visual-acceptance.md",
		"application recipes",
	} {
		if !strings.Contains(checklist, want) {
			t.Errorf("release checklist missing application artifact %q", want)
		}
	}
}

func TestApplicationPatternReferencesCoverApprovedArchetypesAndStates(t *testing.T) {
	patterns := readDoc(t, "../.agents/skills/using-goshtoso/references/application-patterns.md")
	for _, want := range []string{
		"## App Shell",
		"## Operations List",
		"## Detail Workspace",
		"## Multi-step Workflow",
		"390 px",
		"1440 px",
		"loading",
		"empty",
		"error",
		"success",
		"What stays application-specific",
	} {
		if !strings.Contains(patterns, want) {
			t.Errorf("application patterns reference missing %q", want)
		}
	}

	acceptance := readDoc(t, "../.agents/skills/using-goshtoso/references/visual-acceptance.md")
	for _, want := range []string{
		"390 px",
		"1440 px",
		"Goshtoso",
		"Minimal",
		"light",
		"dark",
		"console",
		"keyboard",
		"axe",
	} {
		if !strings.Contains(acceptance, want) {
			t.Errorf("visual acceptance reference missing %q", want)
		}
	}
}

func TestEmbeddedCSSContractSafelistsAuditedApplicationUtilities(t *testing.T) {
	content, err := os.ReadFile("../css/main.css")
	if err != nil {
		t.Fatalf("read css/main.css: %v", err)
	}
	cssSource := string(content)

	for _, utility := range []string{
		"max-w-7xl",
		"xl:grid-cols-4",
		"lg:col-span-2",
		"min-h-64",
		"sm:text-4xl",
		"first:pt-0",
		"last:pb-0",
		"min-w-[220px]",
		"sm:col-span-2",
	} {
		if !strings.Contains(cssSource, utility) {
			t.Errorf("css/main.css missing guaranteed application utility %q", utility)
		}
	}
}

func TestGeneratedEmbeddedCSSContainsAuditedApplicationSelectors(t *testing.T) {
	content, err := os.ReadFile("../assets/styles.css")
	if err != nil {
		t.Fatalf("read assets/styles.css: %v", err)
	}
	generated := string(content)

	for _, selector := range []string{
		`.max-w-7xl`,
		`.xl\:grid-cols-4`,
		`.lg\:col-span-2`,
		`.min-h-64`,
		`.sm\:text-4xl`,
		`.first\:pt-0`,
		`.last\:pb-0`,
		`.min-w-\[220px\]`,
		`.sm\:col-span-2`,
	} {
		if !strings.Contains(generated, selector) {
			t.Errorf("assets/styles.css missing guaranteed selector %q", selector)
		}
	}
}

func TestRoundTwoConsumerGuidancePublishesRecoveredContracts(t *testing.T) {
	skill := readDoc(t, "../.agents/skills/using-goshtoso/SKILL.md")
	for _, want := range []string{
		"Go 1.26.5 or newer",
		"go mod tidy",
		"link.AppearanceButton",
		"htmx:beforeSwap",
		"adjacent text",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("agent skill missing recovered contract %q", want)
		}
	}
	if strings.Index(skill, "templ generate") > strings.Index(skill, "go mod tidy") {
		t.Error("agent skill must generate consumer templ before dependency tidying")
	}

	patterns := readDoc(t, "../.agents/skills/using-goshtoso/references/application-patterns.md")
	for _, want := range []string{
		"AppShell renders the `<header>` landmark",
		"--color-surface",
		"--color-surface-dark-alt",
		"Base status tokens are shared across modes",
		"text-danger-text dark:text-danger-text-dark",
		"Decision Queue",
		"Interruption-safe Workflow",
		"Content-first Review",
		`badge.Config{Label: "Healthy"}`,
	} {
		if !strings.Contains(patterns, want) {
			t.Errorf("application patterns missing recovered contract %q", want)
		}
	}
	if strings.Contains(patterns, "badge.Config{Text:") {
		t.Error("application patterns uses obsolete badge.Config.Text field")
	}

	acceptance := readDoc(t, "../.agents/skills/using-goshtoso/references/visual-acceptance.md")
	for _, want := range []string{
		"Inter",
		"Geist",
		"Roboto",
		"Georgia/Times",
		"ghost card",
		"broad diffuse",
		"internal overflow",
	} {
		if !strings.Contains(acceptance, want) {
			t.Errorf("visual acceptance missing convergence check %q", want)
		}
	}
}
