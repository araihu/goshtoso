//go:build e2e && full

package e2e

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/table"
	"github.com/araihu/goshtoso/site/internal/conformanceledger"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

type conformanceStateFixture struct {
	State     string
	Component templ.Component
}

func conformanceMetadata(openGraphType head.OpenGraphType, twitterCard head.TwitterCard) head.MetadataConfig {
	return head.MetadataConfig{
		Title:         "Conformance state",
		Description:   "Source-derived conformance state fixture.",
		CanonicalURL:  "https://example.test/conformance-state",
		OpenGraphType: openGraphType,
		TwitterCard:   twitterCard,
		Image: head.SocialImage{
			URL:      "https://example.test/conformance-state.png",
			MIMEType: "image/png",
			Width:    1200,
			Height:   630,
			Alt:      "Conformance state",
		},
	}
}

func conformanceDependencyFixture(dependency head.Dependency) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		var rendered strings.Builder
		if err := head.Dependencies(head.WithLocalRuntime(), head.WithoutDependency(dependency)).Render(ctx, &rendered); err != nil {
			return err
		}
		_, err := fmt.Fprintf(writer, `<pre data-conformance-dependency=%q>%s</pre>`, dependency, html.EscapeString(rendered.String()))
		return err
	})
}

func conformanceDependenciesFixture(minimal bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		var rendered strings.Builder
		var component templ.Component = head.Dependencies(head.WithLocalRuntime())
		if minimal {
			component = head.DependenciesMinimal(head.WithLocalRuntime())
		}
		if err := component.Render(ctx, &rendered); err != nil {
			return err
		}
		_, err := fmt.Fprintf(writer, `<pre data-conformance-dependencies=%q>%s</pre>`, map[bool]string{false: "full", true: "minimal"}[minimal], html.EscapeString(rendered.String()))
		return err
	})
}

func conformanceWithChildren(component templ.Component, label string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
		return component.Render(templ.WithChildren(ctx, templ.Raw(html.EscapeString(label))), writer)
	})
}

func conformanceColumns() []table.Column {
	return []table.Column{{Key: "name", Label: "Name", Sortable: true}}
}

func conformanceRows() []table.Row {
	return []table.Row{{ID: "row", Cells: map[string]table.Cell{"name": {Text: "State"}}}}
}

type conformanceAxeCore struct {
	Source        string
	ArchivePath   string
	ArchiveSHA256 string
	ScriptSHA256  string
}

type conformanceConsumerFixture struct {
	Evidence conformanceledger.BFullConsumerFixture
	CSS      string
}

func conformanceBuildConsumerFixture(t *testing.T, evidenceDir string) conformanceConsumerFixture {
	t.Helper()
	const css = `[data-conformance-consumer-fixture="t-gs-010-modern-consumer"] .conformance-state {
	border-color: rgb(15 118 110);
	outline: 1px solid color-mix(in srgb, rgb(15 118 110) 45%, transparent);
}`
	path := filepath.Join(evidenceDir, "modern-consumer-override.css")
	require.NoError(t, os.WriteFile(path, []byte(css), 0o600))
	digest, err := conformanceledger.SHA256File(path)
	require.NoError(t, err)
	return conformanceConsumerFixture{
		Evidence: conformanceledger.BFullConsumerFixture{
			ID:               "t-gs-010-modern-consumer",
			BaseTheme:        "modern",
			StylesheetPath:   path,
			StylesheetSHA256: digest,
		},
		CSS: css,
	}
}

func conformanceLoadAxeCore(t *testing.T) conformanceAxeCore {
	t.Helper()
	path := os.Getenv("CONFORMANCE_AXE_CORE_TGZ")
	if path == "" {
		path = "/private/tmp/araihu-design-implementation/t-gs-010-axe/axe-core-4.10.3.tgz"
	}
	digest, err := conformanceledger.SHA256File(path)
	require.NoError(t, err)
	require.Equal(t, conformanceledger.AxeCoreArchiveSHA256, digest)

	archive, err := os.Open(path)
	require.NoError(t, err)
	defer archive.Close()
	reader, err := gzip.NewReader(archive)
	require.NoError(t, err)
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	var source []byte
	for {
		header, readErr := tarReader.Next()
		if readErr == io.EOF {
			break
		}
		require.NoError(t, readErr)
		if header.Name != "package/axe.min.js" {
			continue
		}
		source, err = io.ReadAll(tarReader)
		require.NoError(t, err)
		break
	}
	require.NotEmpty(t, source, "authenticated axe-core archive must contain package/axe.min.js")
	scriptDigest := sha256.Sum256(source)
	return conformanceAxeCore{Source: string(source), ArchivePath: path, ArchiveSHA256: digest, ScriptSHA256: hex.EncodeToString(scriptDigest[:])}
}

func conformanceStateDocument(t *testing.T, fixtures []conformanceStateFixture, theme, mode string, consumer conformanceConsumerFixture) string {
	t.Helper()
	var body strings.Builder
	for _, fixture := range fixtures {
		var rendered strings.Builder
		require.NoError(t, fixture.Component.Render(context.Background(), &rendered), fixture.State)
		// These markers are fixture-owned measurement anchors, not a guessed
		// button/input fallback. The inner anchor contains the component bytes
		// whose text and adjacent surface B-FULL measures.
		fmt.Fprintf(&body, `<section class="conformance-state min-w-0 border border-outline p-3" data-conformance-state=%q><h2 class="sr-only">%s</h2><div data-conformance-target><div data-conformance-paint-target>%s</div></div></section>`, fixture.State, html.EscapeString(fixture.State), conformanceNamespaceFragment(rendered.String(), fixture.State))
	}
	dependencies := renderComponentFragment(t, head.Dependencies(head.WithLocalRuntime()))
	darkClass := ""
	if mode == "dark" {
		darkClass = ` class="dark"`
	}
	consumerCSS, consumerAttr := "", ""
	if theme == "modern" {
		consumerAttr = fmt.Sprintf(` data-conformance-consumer-fixture=%q`, consumer.Evidence.ID)
		consumerCSS = `<style data-conformance-consumer-stylesheet="` + html.EscapeString(consumer.Evidence.StylesheetSHA256) + `">` + consumer.CSS + `</style>`
	}
	return `<!doctype html><html lang="en" data-theme="` + html.EscapeString(theme) + `"` + darkClass + `><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Goshtoso conformance state fixture</title>` + dependencies + consumerCSS + `</head><body class="bg-surface text-text"><main id="conformance-states" class="grid gap-3"` + consumerAttr + `><h1 class="sr-only">Goshtoso conformance state fixtures</h1>` + body.String() + `</main></body></html>`
}

// These patterns deliberately require an HTML attribute boundary.  A loose
// `for=` matcher also matches Alpine's `x-for=`, corrupting the expression
// before the browser gets to execute it.  Dynamic bindings are source code,
// not static identifiers, and must never be rewritten by fixture isolation.
var conformanceIDAttribute = regexp.MustCompile(`(?i)(^|[\s<])(id|for|aria-labelledby|aria-describedby|aria-controls|aria-owns|list)="([^"]+)"`)
var conformanceFragmentReference = regexp.MustCompile(`(?i)(^|[\s<])(href|x-bind:aria-controls)="#([^"]+)"`)

func conformanceNamespaceFragment(fragment, state string) string {
	prefix := "state-" + strings.NewReplacer("/", "-", ".", "-", "_", "-").Replace(strings.ToLower(state)) + "-"
	fragment = conformanceIDAttribute.ReplaceAllStringFunc(fragment, func(attribute string) string {
		parts := conformanceIDAttribute.FindStringSubmatch(attribute)
		values := parts[3]
		for _, value := range strings.Fields(values) {
			values = strings.ReplaceAll(values, value, prefix+value)
		}
		return parts[1] + parts[2] + `="` + values + `"`
	})
	return conformanceFragmentReference.ReplaceAllString(fragment, "$"+"{1}$"+"{2}=\"#"+prefix+"$"+"{3}\"")
}
func conformanceGitIdentity(t *testing.T, repoRoot, revision string) string {
	t.Helper()
	command := exec.Command("git", "rev-parse", revision)
	command.Dir = repoRoot
	output, err := command.Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(output))
}

// conformancePageErrorMessage retains the browser-side stack when Playwright
// provides one. A bare text error is not enough evidence to assign a finding
// to a component or to distinguish a fixture precondition from a runtime
// defect.
func conformancePageErrorMessage(exception error) string {
	var playwrightError *playwright.Error
	if errors.As(exception, &playwrightError) && strings.TrimSpace(playwrightError.Stack) != "" {
		return playwrightError.Message + "\n" + playwrightError.Stack
	}
	return exception.Error()
}

func conformanceWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	content, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(content, '\n'), 0o600))
}

func conformanceDiagnosticStateCap(t *testing.T, diagnostic bool, total int) int {
	t.Helper()
	raw := os.Getenv("CONFORMANCE_BFULL_DIAGNOSTIC_STATE_CAP")
	states := os.Getenv("CONFORMANCE_BFULL_DIAGNOSTIC_STATES")
	if !diagnostic {
		require.Empty(t, raw, "state cap is only permitted for explicitly non-closure diagnostics")
		require.Empty(t, states, "state selection is only permitted for explicitly non-closure diagnostics")
		return total
	}
	if raw == "" {
		return min(6, total)
	}
	var cap int
	_, err := fmt.Sscanf(raw, "%d", &cap)
	require.NoError(t, err)
	require.Greater(t, cap, 0)
	require.LessOrEqual(t, cap, total)
	return cap
}

func conformanceDiagnosticStateSelection(t *testing.T, fixtures []conformanceStateFixture, requested string) []conformanceStateFixture {
	t.Helper()
	byState := make(map[string]conformanceStateFixture, len(fixtures))
	for _, fixture := range fixtures {
		byState[fixture.State] = fixture
	}
	var selected []conformanceStateFixture
	seen := map[string]struct{}{}
	for _, state := range strings.Split(requested, ",") {
		state = strings.TrimSpace(state)
		require.NotEmpty(t, state, "diagnostic state selection cannot contain an empty state")
		_, duplicate := seen[state]
		require.False(t, duplicate, "diagnostic state selection cannot contain duplicates")
		fixture, ok := byState[state]
		require.True(t, ok, "diagnostic state selection must name a source-derived state: %s", state)
		seen[state] = struct{}{}
		selected = append(selected, fixture)
	}
	require.NotEmpty(t, selected)
	return selected
}

type conformanceRouteLayout struct {
	MainVisible bool   `json:"main_visible"`
	OverflowX   bool   `json:"overflow_x"`
	Background  string `json:"background"`
}

// conformanceCollectBFullCoverage is intentionally outside the 347-state
// fixture page: package/renderable/Kind/route/lifecycle closure must be tied
// to maintained source inventory and real site routes, not inferred from a
// Cartesian state batch that happens to contain similar markup.
func conformanceCollectBFullCoverage(t *testing.T, inventory conformanceledger.Inventory, consumer conformanceConsumerFixture, evidenceDir string) conformanceledger.BFullCoverage {
	t.Helper()
	values := func(items []conformanceledger.SourceItem) []string {
		out := make([]string, 0, len(items))
		for _, item := range items {
			out = append(out, item.Value)
		}
		return out
	}
	coverage := conformanceledger.BFullCoverage{
		Packages:        values(inventory.Packages),
		Renderables:     values(inventory.Renderables),
		Kinds:           values(inventory.Kinds),
		Routes:          values(inventory.Routes),
		LifecycleStates: values(inventory.LifecycleStates),
		ConsumerFixture: consumer.Evidence,
	}
	page := newIsolatedPage(t)
	dismissCookieBanner(t, page)
	page.SetDefaultTimeout(3_000)
	require.NoError(t, page.SetViewportSize(1440, 900))
	for index, route := range inventory.Routes {
		response, err := page.Goto(baseURL+route.Value, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
		require.NoError(t, err, route.Value)
		require.NotNil(t, response, route.Value)
		require.Equal(t, 200, response.Status(), route.Value)
		body, err := response.Body()
		require.NoError(t, err, route.Value)
		name := strings.NewReplacer("/", "-", ".", "-", "_", "-").Replace(strings.TrimPrefix(route.Value, "/"))
		path := filepath.Join(evidenceDir, fmt.Sprintf("route-%02d-%s.initial.html", index, name))
		require.NoError(t, os.WriteFile(path, body, 0o600), route.Value)
		digest, err := conformanceledger.SHA256File(path)
		require.NoError(t, err, route.Value)
		require.NoError(t, waitForAlpine(page), route.Value)
		layout, err := conformanceEvaluateJSON[conformanceRouteLayout](page, `() => {
			const main = document.querySelector('#main-content');
			if (!main) return {main_visible: false, overflow_x: true, background: ''};
			const style = getComputedStyle(main);
			const body = getComputedStyle(document.body);
			const rect = main.getBoundingClientRect();
			return {
				main_visible: style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0,
				overflow_x: main.scrollWidth > main.clientWidth + 1 || document.documentElement.scrollWidth > document.documentElement.clientWidth + 1,
				background: style.backgroundColor !== 'rgba(0, 0, 0, 0)' ? style.backgroundColor : body.backgroundColor,
			};
		}`, nil)
		require.NoError(t, err, route.Value)
		coverage.RouteEvidence = append(coverage.RouteEvidence, conformanceledger.BFullRouteObservation{
			Route: route.Value, URL: response.URL(), SourceResponsePath: path, SourceResponseSHA256: digest,
			MainVisible: layout.MainVisible, MainOverflowX: layout.OverflowX, Background: layout.Background,
		})
	}
	for _, at := range []string{"safari-voiceover", "chromium-screen-reader"} {
		for _, exemplar := range conformanceledger.RequiredATExemplars {
			coverage.ATExemplars = append(coverage.ATExemplars, conformanceledger.BFullATExemplarObservation{
				AT: at, Name: exemplar.Name, Route: exemplar.Route, State: exemplar.State,
				Browser: "external authenticated AT receipt required",
			})
		}
	}
	return coverage
}

func conformanceEvaluateJSON[T any](page playwright.Page, expression string, arg any) (T, error) {
	var result T
	value, err := page.Evaluate(expression, arg)
	if err != nil {
		return result, err
	}
	content, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return result, err
	}
	return result, nil
}

const conformanceStateInspectionScript = `async config => {
	await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
	const visible = element => {
		if (!element) return false;
		const style = getComputedStyle(element), rect = element.getBoundingClientRect();
		return style.display !== 'none' && style.visibility !== 'hidden' && Number(style.opacity || 1) !== 0 && rect.width > 0 && rect.height > 0;
	};
	const parseColor = value => {
		const parts = String(value || '').match(/[\\d.]+/g);
		return parts && parts.length >= 3 ? [Number(parts[0]), Number(parts[1]), Number(parts[2]), parts.length > 3 ? Number(parts[3]) : 1] : null;
	};
	const encodeColor = color => color ? 'rgba(' + color.slice(0, 3).map(value => Math.round(value)).join(', ') + ', ' + color[3] + ')' : '';
	const composite = chain => {
		let result = [0, 0, 0, 0];
		for (let index = chain.length - 1; index >= 0; index--) {
			const layer = chain[index], alpha = layer[3] + result[3] * (1 - layer[3]);
			if (!alpha) continue;
			result = [0, 1, 2].map(channel => (layer[channel] * layer[3] + result[channel] * result[3] * (1 - layer[3])) / alpha).concat(alpha);
		}
		return result[3] === 1 ? result : null;
	};
	const backgrounds = element => {
		const chain = [];
		for (let current = element; current; current = current.parentElement) {
			const value = parseColor(getComputedStyle(current).backgroundColor);
			if (value) chain.push(value);
		}
		return {chain, composited: composite(chain)};
	};
	const contrast = (left, right) => {
		if (!left || !right) return null;
		const linear = value => { value /= 255; return value <= .04045 ? value / 12.92 : Math.pow((value + .055) / 1.055, 2.4); };
		const luminance = value => .2126 * linear(value[0]) + .7152 * linear(value[1]) + .0722 * linear(value[2]);
		const a = luminance(left), b = luminance(right);
		return (Math.max(a, b) + .05) / (Math.min(a, b) + .05);
	};
	const duration = value => Math.max(0, ...String(value || '0s').split(',').map(item => {
		item = item.trim();
		return item.endsWith('ms') ? Number.parseFloat(item) / 1000 : Number.parseFloat(item) || 0;
	}));
	const animationTime = element => Math.max(-1, ...element.getAnimations({subtree: true}).filter(animation => animation.playState === 'running').map(animation => Number(animation.currentTime) || 0));
	const pass = passed => ({applicability: 'applicable', passed});
	const na = rationale => ({applicability: 'not-applicable', rationale});
	return Promise.all([...document.querySelectorAll('[data-conformance-state]')].map(async section => {
		const style = getComputedStyle(section), rect = section.getBoundingClientRect();
		const base = '[data-conformance-state="' + CSS.escape(section.dataset.conformanceState) + '"]';
		const target = section.querySelector('[data-conformance-target]');
		if (!target || !visible(target)) throw new Error('state ' + section.dataset.conformanceState + ' lacks source-owned data-conformance-target');
		const paintTarget = target.querySelector('[data-conformance-paint-target]');
		if (!paintTarget || !visible(paintTarget)) throw new Error('state ' + section.dataset.conformanceState + ' lacks source-owned data-conformance-paint-target');
		const targetStyle = getComputedStyle(target), targetRect = target.getBoundingClientRect();
		const paintStyle = getComputedStyle(paintTarget);
		const paintBackground = backgrounds(paintTarget), targetBackground = backgrounds(target);
		const textRatio = contrast(parseColor(paintStyle.color), paintBackground.composited);
		const borderWidth = Math.max(Number.parseFloat(targetStyle.borderTopWidth) || 0, Number.parseFloat(targetStyle.borderRightWidth) || 0, Number.parseFloat(targetStyle.borderBottomWidth) || 0, Number.parseFloat(targetStyle.borderLeftWidth) || 0);
		const boundaryRatio = contrast(parseColor(targetStyle.borderTopColor), targetBackground.composited);
		const hasBoundary = borderWidth > 0 || (targetStyle.boxShadow && targetStyle.boxShadow !== 'none');
		const motionDuration = Math.max(duration(targetStyle.animationDuration), duration(targetStyle.transitionDuration));
		const hasMotion = (targetStyle.animationName && targetStyle.animationName !== 'none') || (targetStyle.transitionProperty && targetStyle.transitionProperty !== 'none' && motionDuration > 0);
		const beforeMotionMS = animationTime(target);
		await new Promise(resolve => requestAnimationFrame(resolve));
		const actionMotionMS = animationTime(target);
		const observedMotionDeltaMS = beforeMotionMS >= 0 && actionMotionMS >= 0 ? actionMotionMS - beforeMotionMS : 0;
		const overlay = [...section.querySelectorAll('[role="dialog"],[role="alertdialog"],[role="menu"],[role="tooltip"]')].find(visible);
		const targetSelector = base + ' [data-conformance-target]';
		const paintSelector = targetSelector + ' [data-conformance-paint-target]';
		return {
			state: section.dataset.conformanceState,
			exists: section.isConnected,
			dom_nodes: section.querySelectorAll('*').length + 1,
			visible: visible(section),
			width: Math.round(rect.width),
			height: Math.round(rect.height),
			color: style.color,
			background: style.backgroundColor,
			overflow_x: section.scrollWidth > section.clientWidth + 1 || document.documentElement.scrollWidth > document.documentElement.clientWidth + 1,
			theme: document.documentElement.dataset.theme || '',
			mode: document.documentElement.classList.contains('dark') ? 'dark' : 'light',
			motion: config.motion,
			zoom: config.zoom,
			text_contrast: textRatio === null ? na('source-rendered text has no computable opaque contrast pair') : pass(textRatio >= 4.5),
			boundary_contrast: !hasBoundary ? na('source-rendered state has no visual boundary contract') : (boundaryRatio === null ? na('source-rendered boundary has no computable opaque contrast pair') : pass(boundaryRatio >= 3)),
			motion_outcome: !hasMotion ? na('source-rendered state has no animation or transition contract') : pass(config.motion === 'reduced' ? motionDuration <= .1 : observedMotionDeltaMS > 0),
			overlay_provenance: !overlay ? na('source-rendered state has no visible overlay in this phase') : pass(overlay.closest('[data-conformance-state]') === section && Boolean(overlay.getAttribute('data-goshtoso-overlay') || overlay.getAttribute('aria-labelledby'))),
			raw: {
				target_selector: targetSelector,
				text: [{selector: paintSelector, adjacent_selector: targetSelector, foreground: paintStyle.color, background: encodeColor(paintBackground.chain[0]), background_chain: paintBackground.chain.map(encodeColor), composited_background: encodeColor(paintBackground.composited)}],
				boundaries: [{selector: paintSelector, adjacent_selector: targetSelector, foreground: targetStyle.borderTopColor, background: encodeColor(targetBackground.chain[0]), background_chain: targetBackground.chain.map(encodeColor), composited_background: encodeColor(targetBackground.composited)}],
				motion: hasMotion ? {selector: targetSelector, descendant_selector: paintSelector, before_ms: beforeMotionMS, action_ms: actionMotionMS, observed_delta_ms: observedMotionDeltaMS, action_token: 'requestAnimationFrame', reduced_ms: config.motion === 'reduced' ? actionMotionMS : 0} : {},
				overlay: !overlay ? {} : {selector: base + ' [role="' + overlay.getAttribute('role') + '"]', role: overlay.getAttribute('role'), source_selector: targetSelector, source_token: overlay.getAttribute('data-goshtoso-overlay') || overlay.getAttribute('aria-labelledby') || ''},
			},
		};
	}));
}`

func conformanceInspectStates(t *testing.T, page playwright.Page, theme, mode, motion string, zoom int) []conformanceledger.BFullStateObservation {
	t.Helper()
	observations, err := conformanceEvaluateJSON[[]conformanceledger.BFullStateObservation](page, conformanceStateInspectionScript, map[string]any{"theme": theme, "mode": mode, "motion": motion, "zoom": zoom})
	require.NoError(t, err)
	return observations
}

func conformanceInjectAxe(page playwright.Page, source string) error {
	ready, err := conformanceEvaluateJSON[bool](page, `() => Boolean(window.axe && window.axe.run)`, nil)
	if err != nil {
		return err
	}
	if ready {
		return nil
	}
	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{Content: &source})
	return err
}

type conformanceAxeRule struct {
	ID    string `json:"id"`
	Nodes []struct {
		Target []string `json:"target"`
	} `json:"nodes"`
}

type conformanceAxeResult struct {
	Violations   []conformanceAxeRule `json:"violations"`
	Incomplete   []conformanceAxeRule `json:"incomplete"`
	Passes       []conformanceAxeRule `json:"passes"`
	Inapplicable []conformanceAxeRule `json:"inapplicable"`
}

func conformanceRunAxe(t *testing.T, page playwright.Page, axe conformanceAxeCore, stateCount int) conformanceledger.BFullAccessibilityScan {
	t.Helper()
	result, err := conformanceEvaluateJSON[conformanceAxeResult](page, `async () => {
		if (!window.axe || !window.axe.run) throw new Error('authenticated axe-core unavailable');
		return await window.axe.run(document, {resultTypes: ['violations', 'incomplete', 'passes', 'inapplicable']});
	}`, nil)
	require.NoError(t, err)
	scan := conformanceledger.BFullAccessibilityScan{
		Engine: "axe-core", Version: conformanceledger.AxeCoreVersion, ArchivePath: axe.ArchivePath,
		ArchiveSHA256: axe.ArchiveSHA256, ScriptSHA256: axe.ScriptSHA256, ScannedStates: stateCount,
	}
	seen := map[string]struct{}{}
	add := func(outcome string, rules []conformanceAxeRule) {
		for _, rule := range rules {
			if rule.ID == "" {
				continue
			}
			if _, exists := seen[rule.ID]; !exists {
				seen[rule.ID] = struct{}{}
				scan.Rules = append(scan.Rules, rule.ID)
			}
			scan.ChecklistResults = append(scan.ChecklistResults, conformanceledger.BFullChecklistResult{
				Criterion: "automated axe-core rule result", URL: conformanceledger.ChecklistA11Y, RuleID: rule.ID, Outcome: outcome, Targets: len(rule.Nodes),
			})
			for _, node := range rule.Nodes {
				finding := rule.ID
				if len(node.Target) > 0 {
					finding += ":" + strings.Join(node.Target, " ")
				}
				if outcome == "violation" {
					scan.Violations = append(scan.Violations, finding)
				}
				if outcome == "incomplete" {
					scan.Incomplete = append(scan.Incomplete, finding)
				}
			}
		}
	}
	add("violation", result.Violations)
	add("incomplete", result.Incomplete)
	add("pass", result.Passes)
	add("inapplicable", result.Inapplicable)
	sort.Strings(scan.Rules)
	sort.Strings(scan.Violations)
	sort.Strings(scan.Incomplete)
	sort.Slice(scan.ChecklistResults, func(i, j int) bool {
		if scan.ChecklistResults[i].RuleID == scan.ChecklistResults[j].RuleID {
			return scan.ChecklistResults[i].Outcome < scan.ChecklistResults[j].Outcome
		}
		return scan.ChecklistResults[i].RuleID < scan.ChecklistResults[j].RuleID
	})
	return scan
}

func conformanceSetUserAgentZoom(t *testing.T, page playwright.Page, session playwright.CDPSession, zoom int) {
	t.Helper()
	_, err := session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": float64(zoom) / 100})
	require.NoError(t, err)
	page.WaitForTimeout(25)
}

type conformanceZoomObservation struct {
	Scale  float64 `json:"scale"`
	Reflow bool    `json:"reflow"`
}

func conformanceVerifyUserAgentZoom(page playwright.Page, zoom int) error {
	observation, err := conformanceEvaluateJSON[conformanceZoomObservation](page, `() => {
		const viewport = window.visualViewport;
		const scale = viewport ? viewport.scale : 1;
		const width = viewport ? viewport.width : document.documentElement.clientWidth;
		const reflow = document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1 &&
			[...document.querySelectorAll('[data-conformance-state]')].every(section => section.getBoundingClientRect().width <= width + 1);
		return {scale, reflow};
	}`, nil)
	if err != nil {
		return err
	}
	expected := float64(zoom) / 100
	delta := observation.Scale - expected
	if delta < 0 {
		delta = -delta
	}
	if delta > .1 {
		return fmt.Errorf("visualViewport.scale = %.2f, want %.2f", observation.Scale, expected)
	}
	if !observation.Reflow {
		return fmt.Errorf("user-agent zoom %d%% did not reflow without horizontal overflow", zoom)
	}
	return nil
}

type conformancePreparedTarget struct {
	Found     bool              `json:"found"`
	Selector  string            `json:"selector"`
	Role      string            `json:"role"`
	Name      string            `json:"name"`
	ARIAState map[string]string `json:"aria_state"`
	Rationale string            `json:"rationale"`
}

const conformanceActionTimeoutMS = 500

const conformancePrepareTargetScript = `index => {
	const section = document.querySelectorAll('[data-conformance-state]')[index];
	if (!section) return {found: false, rationale: 'source-rendered state section is absent'};
	const visible = element => {
		const style = getComputedStyle(element), rect = element.getBoundingClientRect();
		return style.display !== 'none' && style.visibility !== 'hidden' && Number(style.opacity || 1) !== 0 && rect.width > 0 && rect.height > 0;
	};
	const candidates = [...section.querySelectorAll('button:not([disabled]):not([aria-disabled="true"]),input:not([type="hidden"]):not([disabled]):not([aria-disabled="true"]),select:not([disabled]):not([aria-disabled="true"]),textarea:not([disabled]):not([aria-disabled="true"]),[role="button"]:not([aria-disabled="true"]),[data-goshtoso-scroll-viewport][tabindex]:not([tabindex="-1"])')];
	const target = candidates.find(visible);
	if (!target) return {found: false, rationale: 'source-rendered state exposes no visible enabled interactive control'};
	target.setAttribute('data-conformance-target', String(index));
	if (!target.__conformanceInputEvents) {
		target.__conformanceInputEvents = [];
		for (const type of ['pointerover', 'pointerout', 'mouseover', 'mouseout', 'mousemove', 'pointerdown', 'pointerup', 'touchstart', 'touchend', 'keydown', 'keyup', 'focusin', 'focusout']) {
			target.addEventListener(type, event => target.__conformanceInputEvents.push(event.type), true);
		}
	}
	target.__conformanceInputEvents.length = 0;
	const role = target.getAttribute('role') || ({BUTTON: 'button', INPUT: target.type === 'checkbox' ? 'checkbox' : target.type === 'radio' ? 'radio' : 'textbox', SELECT: 'combobox', TEXTAREA: 'textbox'}[target.tagName] || target.tagName.toLowerCase());
	const labelled = target.getAttribute('aria-labelledby');
	let name = labelled ? labelled.split(/\\s+/).map(id => document.getElementById(id)?.textContent?.trim() || '').join(' ').trim() : '';
	if (!name) name = target.getAttribute('aria-label')?.trim() || '';
	if (!name && target.id) name = document.querySelector('label[for="' + CSS.escape(target.id) + '"]')?.textContent?.trim() || '';
	if (!name) name = (target.textContent || target.getAttribute('alt') || target.getAttribute('placeholder') || target.getAttribute('title') || '').trim();
	const ariaState = {};
	for (const attribute of target.attributes) if (attribute.name.startsWith('aria-')) ariaState[attribute.name] = attribute.value;
	if ('disabled' in target) ariaState.disabled = String(target.disabled);
	if ('checked' in target) ariaState.checked = String(target.checked);
	if ('value' in target) ariaState.value = String(target.value);
	return {found: true, selector: '[data-conformance-target="' + index + '"]', role, name, aria_state: ariaState};
}`

const conformanceSnapshotScript = `selector => {
	const target = document.querySelector(selector);
	if (!target) return {target_connected: false, target_visible: false, aria_state: {}, hovered: false, event_types: []};
	const style = getComputedStyle(target), rect = target.getBoundingClientRect(), active = document.activeElement;
	const ariaState = {};
	for (const attribute of target.attributes) if (attribute.name.startsWith('aria-')) ariaState[attribute.name] = attribute.value;
	if ('disabled' in target) ariaState.disabled = String(target.disabled);
	if ('checked' in target) ariaState.checked = String(target.checked);
	if ('value' in target) ariaState.value = String(target.value);
	return {
		target_connected: target.isConnected,
		target_visible: style.display !== 'none' && style.visibility !== 'hidden' && Number(style.opacity || 1) !== 0 && rect.width > 0 && rect.height > 0,
		target_focused: active === target,
		active_selector: active === target ? selector : (active && active.id ? '#' + active.id : active ? active.tagName.toLowerCase() : ''),
		aria_state: ariaState,
		hovered: target.matches(':hover'),
		event_types: Array.isArray(target.__conformanceInputEvents) ? [...target.__conformanceInputEvents] : [],
	};
}`

func conformancePrepareTarget(page playwright.Page, index int) (conformancePreparedTarget, error) {
	return conformanceEvaluateJSON[conformancePreparedTarget](page, conformancePrepareTargetScript, index)
}

func conformanceSnapshot(page playwright.Page, selector string) (conformanceledger.BFullInteractionOutcome, error) {
	return conformanceEvaluateJSON[conformanceledger.BFullInteractionOutcome](page, conformanceSnapshotScript, selector)
}

type conformanceFocusPaintGeometry struct {
	Focused bool    `json:"focused"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Outline string  `json:"outline"`
	Surface string  `json:"surface"`
}

func conformanceCaptureFocusVisiblePaint(page playwright.Page, selector string) (conformanceledger.BFullFocusVisiblePaint, error) {
	geometry, err := conformanceEvaluateJSON[conformanceFocusPaintGeometry](page, `selector => {
		const target = document.querySelector(selector);
		if (!target) return {focused:false};
		const style = getComputedStyle(target), rect = target.getBoundingClientRect();
		const surface = current => { for (; current; current = current.parentElement) { const value = getComputedStyle(current).backgroundColor; if (value && value !== 'rgba(0, 0, 0, 0)') return value; } return 'rgb(255, 255, 255)'; };
		const shadowColor = (style.boxShadow || '').match(/rgba?\([^)]*\)/)?.[0] || '';
		return {focused:target.matches(':focus-visible'), x:rect.x, y:rect.y, width:rect.width, height:rect.height, outline:style.outlineWidth !== '0px' ? style.outlineColor : shadowColor, surface:surface(target.parentElement)};
	}`, selector)
	if err != nil {
		return conformanceledger.BFullFocusVisiblePaint{}, err
	}
	if !geometry.Focused || geometry.Outline == "" || geometry.Width <= 0 || geometry.Height <= 0 {
		return conformanceledger.BFullFocusVisiblePaint{}, fmt.Errorf("target has no visible keyboard focus paint")
	}
	const margin = 8.0
	shot, err := page.Screenshot(playwright.PageScreenshotOptions{Clip: &playwright.Rect{X: geometry.X - margin, Y: geometry.Y - margin, Width: geometry.Width + margin*2, Height: geometry.Height + margin*2}, Scale: playwright.ScreenshotScaleCss})
	if err != nil {
		return conformanceledger.BFullFocusVisiblePaint{}, err
	}
	decoded, err := png.Decode(bytes.NewReader(shot))
	if err != nil {
		return conformanceledger.BFullFocusVisiblePaint{}, err
	}
	r, g, b, ok := conformanceRGB(geometry.Outline)
	if !ok {
		return conformanceledger.BFullFocusVisiblePaint{}, fmt.Errorf("parse focus outline %q", geometry.Outline)
	}
	pixels := 0
	for y := decoded.Bounds().Min.Y; y < decoded.Bounds().Max.Y; y++ {
		for x := decoded.Bounds().Min.X; x < decoded.Bounds().Max.X; x++ {
			if float64(x) >= margin && float64(x) < margin+geometry.Width && float64(y) >= margin && float64(y) < margin+geometry.Height {
				continue
			}
			red, green, blue, _ := decoded.At(x, y).RGBA()
			if absInt(int(red>>8)-r) <= 4 && absInt(int(green>>8)-g) <= 4 && absInt(int(blue>>8)-b) <= 4 {
				pixels++
			}
		}
	}
	return conformanceledger.BFullFocusVisiblePaint{TargetSelector: selector, OutlineRGBA: geometry.Outline, SurfaceRGBA: geometry.Surface, OutlinePixels: pixels}, nil
}

func conformanceRGB(value string) (int, int, int, bool) {
	values := regexp.MustCompile(`\d+(?:\.\d+)?`).FindAllString(value, -1)
	if len(values) < 3 {
		return 0, 0, 0, false
	}
	rgb := [3]int{}
	for index := range rgb {
		parsed, err := strconv.ParseFloat(values[index], 64)
		if err != nil || parsed < 0 || parsed > 255 {
			return 0, 0, 0, false
		}
		rgb[index] = int(parsed + .5)
	}
	return rgb[0], rgb[1], rgb[2], true
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func conformanceNoTarget(state, input, source, rationale string) conformanceledger.BFullInputObservation {
	rationale = source + "; " + rationale
	na := conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.NotApplicable, Rationale: rationale}
	return conformanceledger.BFullInputObservation{
		State: state, Input: input, Applicability: conformanceledger.Applicable, ReceiptStatus: conformanceledger.StatusFailed,
		Rationale: rationale, ARIAState: map[string]string{}, SourceGrounding: source, FocusVisible: na, MovementReturn: na,
		Escape: conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.NotApplicable, Rationale: rationale},
	}
}

func conformanceFailure(state, input, source, driver string, target conformancePreparedTarget, before conformanceledger.BFullInteractionOutcome, err error) conformanceledger.BFullInputObservation {
	rationale := source + "; genuine " + input + " driver failed: " + err.Error()
	na := conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.NotApplicable, Rationale: rationale}
	escape := conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.NotApplicable, Rationale: rationale}
	if input == "keyboard" {
		escape = conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: rationale}
	}
	return conformanceledger.BFullInputObservation{
		State: state, Input: input, Applicability: conformanceledger.Applicable, ReceiptStatus: conformanceledger.StatusFailed,
		Rationale: rationale, TargetSelector: target.Selector, TargetRole: target.Role, AccessibleName: target.Name, ARIAState: target.ARIAState,
		Driver: driver, SourceGrounding: source, Before: before, Action: before, Return: before,
		FocusVisible: na, MovementReturn: na, Escape: escape,
	}
}

func conformanceMouseInput(page playwright.Page, state, source string, target conformancePreparedTarget) conformanceledger.BFullInputObservation {
	const driver = "Playwright Mouse.Move"
	locator := page.Locator(target.Selector)
	if err := locator.ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)}); err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, conformanceledger.BFullInteractionOutcome{}, err)
	}
	box, err := locator.BoundingBox(playwright.LocatorBoundingBoxOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)})
	if err != nil || box == nil {
		if err == nil {
			err = fmt.Errorf("target has no bounding box")
		}
		return conformanceFailure(state, "mouse", source, driver, target, conformanceledger.BFullInteractionOutcome{}, err)
	}
	_ = page.Mouse().Move(1, 1)
	before, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, before, err)
	}
	x, y := box.X+box.Width/2, box.Y+box.Height/2
	if err := page.Mouse().Move(x, y, playwright.MouseMoveOptions{Steps: playwright.Int(3)}); err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, before, err)
	}
	action, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, before, err)
	}
	if err := page.Mouse().Move(1, 1); err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, before, err)
	}
	if err := page.Mouse().Move(x, y, playwright.MouseMoveOptions{Steps: playwright.Int(3)}); err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, before, err)
	}
	returned, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "mouse", source, driver, target, before, err)
	}
	return conformanceledger.BFullInputObservation{
		State: state, Input: "mouse", Applicability: conformanceledger.Applicable, ReceiptStatus: conformanceledger.StatusExecuted,
		TargetSelector: target.Selector, TargetRole: target.Role, AccessibleName: target.Name, ARIAState: target.ARIAState, EventCount: len(returned.EventTypes),
		Driver: driver, SourceGrounding: source, Before: before, Action: action, Return: returned,
		FocusVisible:   conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.NotApplicable, Rationale: "mouse driver has no focus-visible contract"},
		MovementReturn: conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.Applicable, Passed: returned.Hovered && len(action.EventTypes) > 0},
		Escape:         conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.NotApplicable, Rationale: "mouse driver does not exercise Escape"},
	}
}

func conformanceKeyboardInput(page playwright.Page, state, source string, index int, target conformancePreparedTarget) conformanceledger.BFullInputObservation {
	const driver = "Playwright Keyboard.Press"
	locator := page.Locator(target.Selector)
	if err := locator.ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)}); err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, conformanceledger.BFullInteractionOutcome{}, err)
	}
	if err := locator.Focus(playwright.LocatorFocusOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)}); err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, conformanceledger.BFullInteractionOutcome{}, err)
	}
	before, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, before, err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, before, err)
	}
	action, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, before, err)
	}
	if err := page.Keyboard().Press("Shift+Tab"); err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, before, err)
	}
	returned, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, before, err)
	}
	paint, err := conformanceCaptureFocusVisiblePaint(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "keyboard", source, driver, target, before, err)
	}
	returned.FocusVisiblePaint = paint
	return conformanceledger.BFullInputObservation{
		State: state, Input: "keyboard", Applicability: conformanceledger.Applicable, ReceiptStatus: conformanceledger.StatusExecuted,
		TargetSelector: target.Selector, TargetRole: target.Role, AccessibleName: target.Name, ARIAState: target.ARIAState, EventCount: len(returned.EventTypes),
		Driver: driver, SourceGrounding: source, Before: before, Action: action, Return: returned,
		FocusVisible:   conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.Applicable, Passed: returned.TargetFocused && len(action.EventTypes) > 0},
		MovementReturn: conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.NotApplicable, Rationale: "keyboard driver has no hover movement-return contract"},
		Escape:         conformanceExecuteEscape(page, source, index),
	}
}

func conformanceTouchInput(page playwright.Page, session playwright.CDPSession, state, source string, target conformancePreparedTarget) conformanceledger.BFullInputObservation {
	const driver = "CDP Input.dispatchTouchEvent"
	locator := page.Locator(target.Selector)
	if err := locator.ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)}); err != nil {
		return conformanceFailure(state, "touch", source, driver, target, conformanceledger.BFullInteractionOutcome{}, err)
	}
	box, err := locator.BoundingBox(playwright.LocatorBoundingBoxOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)})
	if err != nil || box == nil {
		if err == nil {
			err = fmt.Errorf("target has no bounding box")
		}
		return conformanceFailure(state, "touch", source, driver, target, conformanceledger.BFullInteractionOutcome{}, err)
	}
	before, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "touch", source, driver, target, before, err)
	}
	if _, err := session.Send("Emulation.setTouchEmulationEnabled", map[string]any{"enabled": true, "maxTouchPoints": 1}); err != nil {
		return conformanceFailure(state, "touch", source, driver, target, before, err)
	}
	defer func() { _, _ = session.Send("Emulation.setTouchEmulationEnabled", map[string]any{"enabled": false}) }()
	point := map[string]any{"x": box.X + box.Width/2, "y": box.Y + box.Height/2, "id": 1}
	if _, err := session.Send("Input.dispatchTouchEvent", map[string]any{"type": "touchStart", "touchPoints": []map[string]any{point}}); err != nil {
		return conformanceFailure(state, "touch", source, driver, target, before, err)
	}
	action, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "touch", source, driver, target, before, err)
	}
	if _, err := session.Send("Input.dispatchTouchEvent", map[string]any{"type": "touchEnd", "touchPoints": []map[string]any{}}); err != nil {
		return conformanceFailure(state, "touch", source, driver, target, before, err)
	}
	page.WaitForTimeout(20)
	returned, err := conformanceSnapshot(page, target.Selector)
	if err != nil {
		return conformanceFailure(state, "touch", source, driver, target, before, err)
	}
	return conformanceledger.BFullInputObservation{
		State: state, Input: "touch", Applicability: conformanceledger.Applicable, ReceiptStatus: conformanceledger.StatusExecuted,
		TargetSelector: target.Selector, TargetRole: target.Role, AccessibleName: target.Name, ARIAState: target.ARIAState, EventCount: len(returned.EventTypes),
		Driver: driver, SourceGrounding: source, Before: before, Action: action, Return: returned,
		FocusVisible:   conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.NotApplicable, Rationale: "touch input has no focus-visible contract"},
		MovementReturn: conformanceledger.BFullSemanticAssertion{Applicability: conformanceledger.NotApplicable, Rationale: "touch input has no hover movement-return contract"},
		Escape:         conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.NotApplicable, Rationale: "touch input has no Escape key contract"},
	}
}

type conformanceEscapeSurface struct {
	Found       bool    `json:"found"`
	Selector    string  `json:"selector"`
	Role        string  `json:"role"`
	LiveText    string  `json:"live_text"`
	RadiusPX    float64 `json:"radius_px"`
	SourceToken string  `json:"source_token"`
}

const conformanceEscapeTriggerScript = `index => {
	const section = document.querySelectorAll('[data-conformance-state]')[index];
	if (!section) return {found: false};
	const visible = element => {
		const style = getComputedStyle(element), rect = element.getBoundingClientRect();
		return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
	};
	const trigger = [...section.querySelectorAll('button:not([disabled]):not([aria-disabled="true"])')].find(element => visible(element) && (element.hasAttribute('aria-haspopup') || element.hasAttribute('aria-controls') || /open|show|menu|dialog/i.test(element.textContent || '')));
	if (!trigger) return {found: false};
	trigger.setAttribute('data-conformance-escape-trigger', String(index));
	return {found: true, selector: '[data-conformance-escape-trigger="' + index + '"]'};
}`

const conformanceEscapeSurfaceScript = `index => {
	const trigger = document.querySelector('[data-conformance-escape-trigger="' + index + '"]');
	if (!trigger) return {found: false};
	const visible = element => {
		const style = getComputedStyle(element), rect = element.getBoundingClientRect();
		return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
	};
	const controlled = trigger.getAttribute('aria-controls');
	const surface = [...document.querySelectorAll('[role="dialog"],[role="alertdialog"],[role="menu"],[role="tooltip"]')].find(element => visible(element) && (!controlled || element.id === controlled));
	if (!surface) return {found: false};
	surface.setAttribute('data-conformance-escape-surface', String(index));
	const style = getComputedStyle(surface);
	return {
		found: true,
		selector: '[data-conformance-escape-surface="' + index + '"]',
		role: surface.getAttribute('role') || '',
		live_text: (surface.textContent || surface.getAttribute('aria-label') || '').trim(),
		radius_px: Number.parseFloat(style.borderTopLeftRadius) || 0,
		source_token: surface.id || surface.getAttribute('aria-labelledby') || surface.getAttribute('data-goshtoso-overlay') || '',
	};
}`

const conformanceEscapeClosedScript = `index => {
	const trigger = document.querySelector('[data-conformance-escape-trigger="' + index + '"]');
	const surface = document.querySelector('[data-conformance-escape-surface="' + index + '"]');
	const visible = element => {
		if (!element) return false;
		const style = getComputedStyle(element), rect = element.getBoundingClientRect();
		return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
	};
	return {closed: !visible(surface), focus_returned: document.activeElement === trigger};
}`

func conformanceExecuteEscape(page playwright.Page, source string, index int) conformanceledger.BFullEscapeOutcome {
	trigger, err := conformanceEvaluateJSON[conformanceEscapeSurface](page, conformanceEscapeTriggerScript, index)
	if err != nil {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: source + "; inspect Escape trigger: " + err.Error()}
	}
	if !trigger.Found {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.NotApplicable, Rationale: source + "; source-rendered state has no visible control that can open an Escape-dismissible surface"}
	}
	if err := page.Locator(trigger.Selector).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(conformanceActionTimeoutMS)}); err != nil {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: source + "; genuine trigger click failed: " + err.Error()}
	}
	page.WaitForTimeout(25)
	surface, err := conformanceEvaluateJSON[conformanceEscapeSurface](page, conformanceEscapeSurfaceScript, index)
	if err != nil {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: source + "; inspect opened surface: " + err.Error()}
	}
	if !surface.Found {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.NotApplicable, Rationale: source + "; genuine trigger did not open a visible role-backed Escape surface"}
	}
	if err := page.Keyboard().Press("Escape"); err != nil {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: source + "; genuine Escape key failed: " + err.Error(), Opened: true, SurfaceSelector: surface.Selector}
	}
	if err := page.Locator(surface.Selector).WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden, Timeout: playwright.Float(conformanceActionTimeoutMS)}); err != nil {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: source + "; actual Escape did not close its surface: " + err.Error(), Opened: true, SurfaceSelector: surface.Selector}
	}
	closed, err := conformanceEvaluateJSON[struct {
		Closed        bool `json:"closed"`
		FocusReturned bool `json:"focus_returned"`
	}](page, conformanceEscapeClosedScript, index)
	if err != nil {
		return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Rationale: source + "; inspect Escape return: " + err.Error(), Opened: true, SurfaceSelector: surface.Selector}
	}
	tooltip := conformanceledger.BFullTooltipEscapeEvidence{}
	if surface.Role == "tooltip" {
		tooltip = conformanceledger.BFullTooltipEscapeEvidence{Selector: surface.Selector, Role: surface.Role, LiveText: surface.LiveText, RadiusPX: surface.RadiusPX, SourceToken: surface.SourceToken}
	}
	return conformanceledger.BFullEscapeOutcome{Applicability: conformanceledger.Applicable, Passed: closed.Closed && closed.FocusReturned, Opened: true, Closed: closed.Closed, FocusReturned: closed.FocusReturned, SurfaceSelector: surface.Selector, Tooltip: tooltip}
}

func conformanceExecuteInputs(t *testing.T, page playwright.Page, session playwright.CDPSession, fixtures []conformanceStateFixture, inventory conformanceledger.Inventory) []conformanceledger.BFullInputObservation {
	t.Helper()
	authorities := make(map[string]string, len(inventory.States))
	for _, item := range inventory.States {
		authorities[item.Value] = item.Source.Path + "#" + item.Source.Symbol
	}
	observations := make([]conformanceledger.BFullInputObservation, 0, len(fixtures)*3)
	for index, fixture := range fixtures {
		source := authorities[fixture.State]
		require.NotEmpty(t, source, "state fixture %s must have a maintained source authority", fixture.State)
		target, err := conformancePrepareTarget(page, index)
		require.NoError(t, err)
		if !target.Found {
			for _, input := range []string{"mouse", "keyboard", "touch"} {
				observations = append(observations, conformanceNoTarget(fixture.State, input, source, target.Rationale))
			}
			continue
		}
		observations = append(observations,
			conformanceMouseInput(page, fixture.State, source, target),
			conformanceKeyboardInput(page, fixture.State, source, index, target),
			conformanceTouchInput(page, session, fixture.State, source, target),
		)
	}
	return observations
}
