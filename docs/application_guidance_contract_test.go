package docs_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var expectedAgentSkillArtifacts = []string{
	"SKILL.md",
	"agents/openai.yaml",
	"references/adversarial-acceptance.md",
	"references/application-patterns.md",
	"references/components-reference.md",
	"references/design-intelligence.md",
	"references/ecosystem-discovery.md",
	"references/runtime-integration.md",
	"references/visual-acceptance.md",
}

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

func TestAgentSkillDiscoveryPassIsStreamingSafe(t *testing.T) {
	skill := readDoc(t, "../.agents/skills/using-goshtoso/SKILL.md")
	for _, want := range []string{
		"## Required Discovery Pass",
		"go list -m all",
		"Goshtoso Charts",
		"Margo",
		"Goshtoso App Shells",
		"cmd/iconpack",
		"reuse`, `compose`, or `gap",
		"Custom HTML, CSS, or JavaScript requires a concrete gap",
		"streaming `skills use` client provides only this file",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("streamed agent skill missing discovery guidance %q", want)
		}
	}
}

func TestAgentSkillChoosesCurrentNavigationAndActionPrimitives(t *testing.T) {
	skill := readDoc(t, "../.agents/skills/using-goshtoso/SKILL.md")
	for _, want := range []string{
		"Use `dropdown` for a menu",
		"Use `popover` for arbitrary consumer-owned content",
		"Use `splitbutton` for one always-visible dominant action",
		"Use `actiongroup` for one required primary action",
		"`navbar.Config.Secondary`",
		"`navbar.SecondaryRow`",
		"keep responsive substitution consumer-owned",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("agent skill missing component-selection guidance %q", want)
		}
	}
}

func TestEcosystemDiscoveryRecordsCurrentPublicBaseline(t *testing.T) {
	discovery := readDoc(t, "../.agents/skills/using-goshtoso/references/ecosystem-discovery.md")
	for _, want := range []string{
		"Goshtoso `v0.2.5`",
		"Goshtoso App Shells `v0.1.6`",
		"Margo `v0.0.6`",
		"Goshtoso Charts `v0.0.2`",
		"`components/interactive/<type>`",
		"`margo/ssg`",
	} {
		if !strings.Contains(discovery, want) {
			t.Errorf("ecosystem discovery missing current public baseline %q", want)
		}
	}

	for _, stale := range []string{
		"Goshtoso `v0.1.13`",
		"Goshtoso App Shells `v0.1.4`",
		"Margo `v0.0.5`",
		"Goshtoso Charts `v0.0.1`",
	} {
		if strings.Contains(discovery, stale) {
			t.Errorf("ecosystem discovery retains stale baseline %q", stale)
		}
	}
}

func TestAgentSkillInstallExampleTargetsCurrentRelease(t *testing.T) {
	skill := readDoc(t, "../.agents/skills/using-goshtoso/SKILL.md")
	if !strings.Contains(skill, "go get github.com/araihu/goshtoso@v0.2.6") {
		t.Error("agent skill install example must target v0.2.6")
	}
}

func TestAgentSkillDistributionCLI(t *testing.T) {
	npx, err := exec.LookPath("npx")
	if err != nil {
		t.Fatalf("find npx: %v", err)
	}
	productionRoot := filepath.Join("..", ".agents", "skills", "using-goshtoso")
	production, err := agentSkillInventory(productionRoot)
	if err != nil {
		t.Fatalf("inventory production skill: %v", err)
	}
	if strings.Join(production, "\n") != strings.Join(expectedAgentSkillArtifacts, "\n") {
		t.Fatalf("production skill inventory mismatch\nwant:\n%s\ngot:\n%s",
			strings.Join(expectedAgentSkillArtifacts, "\n"), strings.Join(production, "\n"))
	}

	source := filepath.Join(t.TempDir(), "source")
	sourceRoot := filepath.Join(source, ".agents", "skills", "using-goshtoso")
	for _, relative := range expectedAgentSkillArtifacts {
		contents, readErr := os.ReadFile(filepath.Join(productionRoot, filepath.FromSlash(relative)))
		if readErr != nil {
			t.Fatalf("read source artifact %s: %v", relative, readErr)
		}
		destination := filepath.Join(sourceRoot, filepath.FromSlash(relative))
		if mkdirErr := os.MkdirAll(filepath.Dir(destination), 0o755); mkdirErr != nil {
			t.Fatalf("create source artifact directory: %v", mkdirErr)
		}
		if writeErr := os.WriteFile(destination, contents, 0o644); writeErr != nil {
			t.Fatalf("copy source artifact %s: %v", relative, writeErr)
		}
	}
	gitSource := exec.Command("git", "init", "-q")
	gitSource.Dir = source
	if output, initErr := gitSource.CombinedOutput(); initErr != nil {
		t.Fatalf("initialize isolated skill source: %v\n%s", initErr, output)
	}

	use := exec.Command(npx, "--yes", "skills", "use", source, "--skill", "using-goshtoso")
	streamed, err := use.CombinedOutput()
	if err != nil {
		t.Fatalf("stream using-goshtoso: %v\n%s", err, streamed)
	}
	for _, want := range []string{
		"## Required Discovery Pass",
		"go list -m -versions",
		"Goshtoso Charts",
		"Margo",
		"Goshtoso App Shells",
		"cmd/iconpack",
		"reuse`, `compose`, or `gap",
		"Custom HTML, CSS, or JavaScript requires a concrete gap",
	} {
		if !strings.Contains(string(streamed), want) {
			t.Errorf("streamed skill missing %q", want)
		}
	}

	target := t.TempDir()
	gitInit := exec.Command("git", "init", "-q")
	gitInit.Dir = target
	if output, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("initialize install probe: %v\n%s", err, output)
	}
	install := exec.Command(npx, "--yes", "skills", "add", source,
		"--skill", "using-goshtoso", "--agent", "codex", "--copy", "-y")
	install.Dir = target
	if output, err := install.CombinedOutput(); err != nil {
		t.Fatalf("install using-goshtoso: %v\n%s", err, output)
	}

	root := filepath.Join(target, ".agents", "skills", "using-goshtoso")
	installed, err := agentSkillInventory(root)
	if err != nil {
		t.Fatalf("inventory installed skill: %v", err)
	}
	if strings.Join(installed, "\n") != strings.Join(expectedAgentSkillArtifacts, "\n") {
		t.Fatalf("installed skill inventory mismatch\nwant:\n%s\ngot:\n%s",
			strings.Join(expectedAgentSkillArtifacts, "\n"), strings.Join(installed, "\n"))
	}
}

func agentSkillInventory(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func TestEcosystemDiscoveryUsesSiblingReferencePaths(t *testing.T) {
	discovery := readDoc(t, "../.agents/skills/using-goshtoso/references/ecosystem-discovery.md")
	for _, broken := range []string{
		"`references/components-reference.md`",
		"`references/application-patterns.md`",
	} {
		if strings.Contains(discovery, broken) {
			t.Errorf("ecosystem discovery contains nested reference path %q", broken)
		}
	}
}

func TestPublicLandingGuidanceTriesLandingShellBeforeBrandSite(t *testing.T) {
	skill := readDoc(t, "../.agents/skills/using-goshtoso/SKILL.md")
	landing := strings.Index(skill, "try App Shells\n`landingshell` first")
	brand := strings.Index(skill, "examples/brand-site")
	if landing < 0 || brand < 0 || landing > brand {
		t.Error("agent skill must try landingshell before selecting brand-site")
	}
	for _, want := range []string{"Record that concrete gap first", `@"$GOSHTOSO_VERSION"`} {
		if !strings.Contains(skill, want) {
			t.Errorf("public landing guidance missing %q", want)
		}
	}
}

func TestReusableDocumentationShellGuidanceDistinguishesFrameFromPage(t *testing.T) {
	for path, wants := range map[string][]string{
		"../.agents/skills/using-goshtoso/SKILL.md": {
			"github.com/araihu/goshtoso-app-shells/componentdocshell",
			"github.com/araihu/goshtoso-app-shells/componentpage",
			"the catalog shell is a",
			"separate pattern",
		},
		"USAGE.md": {
			"Reusable documentation shells",
			"componentdocshell",
			"componentpage",
		},
		"../site/internal/pages/demo/contentpages/docs/agents.templ": {
			"Reusable documentation shells",
			"goshtoso-app-shells",
		},
	} {
		content := readDoc(t, path)
		for _, want := range wants {
			if !strings.Contains(content, want) {
				t.Errorf("%s missing reusable shell guidance %q", path, want)
			}
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
	for _, artifact := range expectedAgentSkillArtifacts {
		path := ".agents/skills/using-goshtoso/" + artifact
		if !strings.Contains(checklist, path) {
			t.Errorf("release checklist missing packaged skill artifact %q", path)
		}
	}
	for _, want := range []string{
		"references/application-patterns.md",
		"references/visual-acceptance.md",
		"references/design-intelligence.md",
		"references/ecosystem-discovery.md",
		"references/runtime-integration.md",
		"complete required",
		"every `references/*.md` file",
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
