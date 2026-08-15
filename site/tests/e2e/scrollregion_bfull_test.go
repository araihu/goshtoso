//go:build e2e && scrollregion && bfull && axe

package e2e

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

const (
	scrollRegionBFullRoute                  = "/components/scroll-region"
	scrollRegionBFullConsumerTheme          = "consumer-scrollregion"
	scrollRegionBFullConsumerAttribute      = "data-goshtoso-scrollregion-consumer-theme"
	scrollRegionBFullConsumerAttributeValue = "t-gs-011"
	scrollRegionBFullUAZoomFactor           = 2
	scrollRegionBFullChromiumZoomLevel      = 3.8017840169239308 // log(2) / log(1.2)
	scrollRegionBFullDiagnosticEnvironment  = "GOSHTOSO_SCROLLREGION_BFULL_DIAGNOSTIC"
	scrollRegionBFullDiagnosticMaxCellsEnv  = "GOSHTOSO_SCROLLREGION_BFULL_DIAGNOSTIC_MAX_CELLS_PER_ZOOM"
	scrollRegionBFullThemeInitialSource     = "server-routed-html"
	scrollRegionBFullThemePersistenceNA     = "locked-site-selector-disabled: ComponentDocsLayout Appearance.DisableThemeSelector=true; t-gs-011-theme is server-routed initial HTML only"
	scrollRegionBFullDarkPersistence        = "product-ui"
	scrollRegionBFullReceiptSchema          = "goshtoso.t-gs-011.bfull.receipt.v6"
	scrollRegionBFullProvenanceStructural   = "structural-unattested"
	scrollRegionBFullClosureUnattested      = "unattested-non-closure"
	scrollRegionBFullStorageEvidenceSchema  = "goshtoso.t-gs-011.bfull.storage.v1"
	scrollRegionBFullActionEvidenceSchema   = "goshtoso.t-gs-011.bfull.action.v1"
	scrollRegionBFullPaintEvidenceSchema    = "goshtoso.t-gs-011.bfull.first-paint.v1"
)

type scrollRegionBFullTheme struct {
	ID          string
	ServerTheme string
	ConsumerCSS bool
}

type scrollRegionBFullZoom struct {
	ID             string
	Factor         int
	RealUA         bool
	Context        playwright.BrowserContext
	Close          func()
	NewContext     func(*testing.T) playwright.BrowserContext
	BrowserVersion string
}

type scrollRegionBFullArtifact struct {
	Path           string                         `json:"path"`
	SHA256         string                         `json:"sha256"`
	Width          int                            `json:"width,omitempty"`
	Height         int                            `json:"height,omitempty"`
	CapturedRegion string                         `json:"captured_region,omitempty"`
	Capture        *scrollRegionBFullCaptureProof `json:"capture,omitempty"`
}

// scrollRegionBFullPixelAnchor maps a known flat card surface from DOM CSS
// coordinates into the emitted PNG. It is deliberately a content anchor, not
// merely a region bounding box: a valid page crop must include actual Activity
// rows rather than browser/window chrome at the same dimensions.
type scrollRegionBFullPixelAnchor struct {
	Name      string   `json:"name"`
	DOMText   string   `json:"dom_text"`
	CSSX      float64  `json:"css_x"`
	CSSY      float64  `json:"css_y"`
	PixelX    int      `json:"pixel_x"`
	PixelY    int      `json:"pixel_y"`
	Expected  [4]uint8 `json:"expected_rgba"`
	Tolerance uint8    `json:"tolerance"`
}

// scrollRegionBFullBoundaryCue proves the crop contains the actual rendered
// start/end affordance for the visible scroll boundary, not only card pixels.
// It intentionally records geometry rather than a gradient color sample.
type scrollRegionBFullBoundaryCue struct {
	State       string  `json:"state"`
	Visible     bool    `json:"visible"`
	CSSX        float64 `json:"css_x"`
	CSSY        float64 `json:"css_y"`
	CSSWidth    float64 `json:"css_width"`
	CSSHeight   float64 `json:"css_height"`
	PixelX      int     `json:"pixel_x"`
	PixelY      int     `json:"pixel_y"`
	PixelWidth  int     `json:"pixel_width"`
	PixelHeight int     `json:"pixel_height"`
}

// scrollRegionBFullCaptureProof records the exact coordinate relationship
// between the browser's rendered visual viewport and the CDP page-surface PNG.
// It is carried in the receipt so later inspection can reproduce the crop.
type scrollRegionBFullCaptureProof struct {
	Method               string                         `json:"method"`
	VisualViewportLeft   float64                        `json:"visual_viewport_left"`
	VisualViewportTop    float64                        `json:"visual_viewport_top"`
	VisualViewportWidth  float64                        `json:"visual_viewport_width"`
	VisualViewportHeight float64                        `json:"visual_viewport_height"`
	DevicePixelRatio     float64                        `json:"device_pixel_ratio"`
	SourceWidth          int                            `json:"source_width"`
	SourceHeight         int                            `json:"source_height"`
	ScaleX               float64                        `json:"scale_x"`
	ScaleY               float64                        `json:"scale_y"`
	CropCSSX             float64                        `json:"crop_css_x"`
	CropCSSY             float64                        `json:"crop_css_y"`
	CropCSSWidth         float64                        `json:"crop_css_width"`
	CropCSSHeight        float64                        `json:"crop_css_height"`
	CropPixelX           int                            `json:"crop_pixel_x"`
	CropPixelY           int                            `json:"crop_pixel_y"`
	CropPixelWidth       int                            `json:"crop_pixel_width"`
	CropPixelHeight      int                            `json:"crop_pixel_height"`
	Anchors              []scrollRegionBFullPixelAnchor `json:"anchors"`
	BoundaryCue          scrollRegionBFullBoundaryCue   `json:"boundary_cue"`
}

type scrollRegionBFullCapture struct {
	ExpectedWidth  float64
	ExpectedHeight float64
	Proof          scrollRegionBFullCaptureProof
}

type scrollRegionBFullCellReceipt struct {
	CellID          string                            `json:"cell_id"`
	Route           string                            `json:"route"`
	Theme           string                            `json:"theme"`
	Mode            string                            `json:"mode"`
	ViewportWidth   int                               `json:"viewportWidth"`
	Zoom            string                            `json:"zoom"`
	States          []string                          `json:"states"`
	Inputs          []string                          `json:"inputs"`
	FirstHTMLSHA256 string                            `json:"firstHTMLSHA256"`
	SetupActions    []scrollRegionBFullAction         `json:"setup_actions"`
	Persistence     scrollRegionBFullPersistenceProof `json:"persistence"`
	Screenshot      scrollRegionBFullArtifact         `json:"screenshot"`
	Trace           *scrollRegionBFullArtifact        `json:"trace,omitempty"`
	NAs             map[string]string                 `json:"notApplicable"`
}

type scrollRegionBFullPaint struct {
	Phase       string `json:"phase"`
	ReadyState  string `json:"ready_state"`
	Theme       string `json:"theme"`
	ThemeSource string `json:"theme_source"`
	Dark        bool   `json:"dark"`
	Visible     bool   `json:"visible"`
	Role        string `json:"role"`
	Name        string `json:"name"`
}

type scrollRegionBFullPersistenceProof struct {
	ThemeInitialSource            string                      `json:"theme_initial_source"`
	ThemePersistenceNotApplicable string                      `json:"theme_persistence_not_applicable"`
	DarkPersistence               string                      `json:"dark_persistence"`
	StorageBefore                 map[string]string           `json:"storage_before"`
	Actions                       []scrollRegionBFullAction   `json:"actions"`
	FreshLoadInitialHTML          scrollRegionBFullPaint      `json:"fresh_load_initial_html"`
	FreshLoadFirstPaint           scrollRegionBFullPaint      `json:"fresh_load_first_paint"`
	FreshLoadSettled              scrollRegionBFullPaint      `json:"fresh_load_settled"`
	PageAInitialHTML              scrollRegionBFullArtifact   `json:"page_a_initial_html"`
	PageAStorageBefore            scrollRegionBFullArtifact   `json:"page_a_storage_before"`
	PageAActions                  []scrollRegionBFullArtifact `json:"page_a_actions"`
	FreshLoadStorage              scrollRegionBFullArtifact   `json:"fresh_load_storage"`
	FreshLoadInitialHTMLArtifact  scrollRegionBFullArtifact   `json:"fresh_load_initial_html_artifact"`
	FreshLoadPaintArtifact        scrollRegionBFullArtifact   `json:"fresh_load_paint_artifact"`
}

type scrollRegionBFullAction struct {
	Before string `json:"before"`
	Action string `json:"action"`
	Return string `json:"return"`
}

// scrollRegionBFullPageState is captured from the real page around each public
// consent/dark action. It carries no test-owned state mutation.
type scrollRegionBFullPageState struct {
	Cookie        string `json:"cookie"`
	DarkMode      string `json:"dark_mode"`
	Dark          bool   `json:"dark"`
	DialogVisible bool   `json:"dialog_visible"`
}

type scrollRegionBFullStorageEvidence struct {
	Schema string                     `json:"schema"`
	CellID string                     `json:"cell_id"`
	Phase  string                     `json:"phase"`
	State  scrollRegionBFullPageState `json:"state"`
}

type scrollRegionBFullActionEvidence struct {
	Schema string                     `json:"schema"`
	CellID string                     `json:"cell_id"`
	Phase  string                     `json:"phase"`
	Before scrollRegionBFullPageState `json:"before"`
	Action scrollRegionBFullAction    `json:"action"`
	After  scrollRegionBFullPageState `json:"after"`
}

type scrollRegionBFullPaintEvidence struct {
	Schema  string                   `json:"schema"`
	CellID  string                   `json:"cell_id"`
	Events  []scrollRegionBFullPaint `json:"events"`
	Settled scrollRegionBFullPaint   `json:"settled"`
}

// scrollRegionBFullPaintEvidenceFromObserver keeps the complete, ordered
// browser observer transcript. In particular, dark restoration is proved by
// its candidate-owned root mutation, which occurs between the initial server
// HTML snapshot and the first animation frame.
func scrollRegionBFullPaintEvidenceFromObserver(cellID string, events []scrollRegionBFullPaint, settled scrollRegionBFullPaint) scrollRegionBFullPaintEvidence {
	return scrollRegionBFullPaintEvidence{
		Schema:  scrollRegionBFullPaintEvidenceSchema,
		CellID:  cellID,
		Events:  append([]scrollRegionBFullPaint(nil), events...),
		Settled: settled,
	}
}

type scrollRegionBFullReceipt struct {
	Schema          string                               `json:"schema"`
	Closure         string                               `json:"closure"`
	ProvenanceClass string                               `json:"provenance_class"`
	ExpectedCells   int                                  `json:"expected_cells"`
	Binding         scrollRegionBFullIdentityBinding     `json:"binding"`
	ToolVersions    scrollRegionBFullToolVersions        `json:"tool_versions"`
	Widths          []int                                `json:"widths"`
	Cells           []scrollRegionBFullCellReceipt       `json:"cells"`
	TraceByCell     map[string]scrollRegionBFullArtifact `json:"trace_by_cell"`
	WrapperSHA256   string                               `json:"wrapper_sha256"`
}

type scrollRegionBFullToolVersions struct {
	GoRuntime        string            `json:"go_runtime"`
	PlaywrightGo     string            `json:"playwright_go"`
	BrowserByZoom    map[string]string `json:"browser_by_zoom"`
	AxeCore          string            `json:"axe_core"`
	AxeArchiveSHA256 string            `json:"axe_archive_sha256"`
	AxeScriptSHA256  string            `json:"axe_script_sha256"`
}

type scrollRegionBFullRecorder struct {
	directory string
	widths    []int
	plan      scrollRegionBFullRunPlan
	cells     []scrollRegionBFullCellReceipt
	traces    map[string]scrollRegionBFullArtifact
	binding   scrollRegionBFullIdentityBinding
	tools     scrollRegionBFullToolVersions
}

// First-paint evidence may observe the page but cannot preseed its theme,
// dark mode, scroll state, storage, or root classes. Otherwise a reload merely
// proves the test-owned init script rather than product persistence.
func TestScrollRegionBFullFirstPaintObserverCannotMutateProductState(t *testing.T) {
	repository := scrollRegionRepositoryRoot(t)
	source, err := os.ReadFile(filepath.Join(repository, "site", "tests", "e2e", "scrollregion_bfull_test.go"))
	require.NoError(t, err)
	require.NotContains(t, string(source), "scrollRegionBFullInit"+"Script", "first-paint harness must not retain a state-seeding init script")
	require.NotContains(t, string(source), "#site"+"-theme-trigger", "locked component docs must not fabricate a removed public theme selector")
	require.NotContains(t, string(source), "requireScrollRegion"+"SelectTheme", "B-FULL must not claim client theme persistence through a nonexistent selector")
	require.Contains(t, string(source), `Name:        playwright.String("scrollregion-bfull-" + scrollRegionBFullArtifactName(cellID))`, "trace names must not contain raw route separators that discard Playwright trace entries")
	observer := scrollRegionBFullFirstPaintObserverScript()
	for _, required := range []string{"func requireScrollRegionWideConsumerHorizontalAccess", "validateScrollRegionBFullHorizontalContract", "allowInternalHorizontalRange", `Press("ArrowRight")`, `Input.dispatchTouchEvent`} {
		require.Contains(t, string(source), required, "wide consumer horizontal contract must retain real keyboard and CDP touch actions")
	}
	consumerCSS, err := os.ReadFile(filepath.Join(repository, "tests", "external", "scrollregion-a11y", "consumer-scrollregion.css"))
	require.NoError(t, err)
	require.Contains(t, string(consumerCSS), "t-gs-011-wide-consumer-content", "maintained consumer fixture must deliberately expose wide content for horizontal access proof")
	for _, prohibited := range []string{"localStorage." + "setItem(\"theme\"", "localStorage." + "setItem(\"darkMode\"", "root." + "setAttribute(\"data-theme\"", "root." + "classList.toggle(\"dark\""} {
		require.NotContains(t, observer, prohibited, "first-paint observer must not directly mutate product state")
	}
	// A dark fresh load is causal only when the raw observer sequence retains
	// the product root mutation between the pre-bootstrap initial HTML and the
	// first browser frame. Reducing the artifact to its first two entries would
	// make that validator requirement impossible to satisfy.
	initial := scrollRegionBFullPaint{Phase: "init", Dark: false}
	mutation := scrollRegionBFullPaint{Phase: "root-mutation", Dark: true}
	frame := scrollRegionBFullPaint{Phase: "first-animation-frame", Dark: true}
	settled := scrollRegionBFullPaint{Phase: "settled", Dark: true}
	evidence := scrollRegionBFullPaintEvidenceFromObserver("fixture-cell", []scrollRegionBFullPaint{initial, mutation, frame}, settled)
	require.Equal(t, []scrollRegionBFullPaint{initial, mutation, frame}, evidence.Events)
	_, err = validateScrollRegionBFullPaintTranscript(evidence.Events, initial, frame)
	require.NoError(t, err)
	_, err = validateScrollRegionBFullPaintTranscript([]scrollRegionBFullPaint{initial, mutation, settled}, initial, frame)
	require.Error(t, err, "the serialized wrapper first paint must be an event in the raw transcript")
	require.NoError(t, validateScrollRegionBFullDarkPaintTranscript(evidence.Events, initial, frame, "", ""))
	require.Error(t, validateScrollRegionBFullDarkPaintTranscript([]scrollRegionBFullPaint{initial, frame, mutation}, initial, frame, "", ""), "a mutation after the first frame cannot prove dark first-paint restoration")
	consumer, ok := scrollRegionBFullThemeByID(scrollRegionBFullConsumerTheme)
	require.True(t, ok)
	cellURL, err := url.Parse(scrollRegionBFullCellRoutedURL(consumer, true, 390, scrollRegionBFullZoom{ID: "ua-200"}))
	require.NoError(t, err)
	require.Equal(t, scrollRegionBFullCellID(consumer, true, 390, scrollRegionBFullZoom{ID: "ua-200"}), cellURL.Query().Get("t-gs-011-cell"))
	require.Equal(t, consumer.ID, cellURL.Query().Get("t-gs-011-consumer"))
	require.Equal(t, "dark", cellURL.Query().Get("t-gs-011-mode"))
	require.Equal(t, "390", cellURL.Query().Get("t-gs-011-width"))
	require.Equal(t, "ua-200", cellURL.Query().Get("t-gs-011-zoom"))
	// Consumer fixture deliberately owns a horizontal content range. The page
	// must still have no horizontal range; ordinary cells must not acquire one.
	require.NoError(t, validateScrollRegionBFullHorizontalContract(false, 390, 390, 0))
	require.Error(t, validateScrollRegionBFullHorizontalContract(false, 390, 704, 314))
	require.NoError(t, validateScrollRegionBFullHorizontalContract(true, 390, 704, 314))
	require.Error(t, validateScrollRegionBFullHorizontalContract(true, 390, 390, 0))
}

// scrollRegionBFullRunPlan makes diagnostic sampling opt-in and receipt-visible.
// A full run has no cap and must produce every literal B-FULL cell.
type scrollRegionBFullRunPlan struct {
	Diagnostic        bool
	MaxCellsPerZoom   int
	ExpectedCells     int
	FullExpectedCells int
}

func TestScrollRegionBFull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping T-GS-011 B-FULL browser contract in short mode")
	}
	repositoryRoot := scrollRegionRepositoryRoot(t)
	widths := scrollRegionBFullWidths(t)
	plan, err := scrollRegionBFullPlan(widths)
	require.NoError(t, err)
	binding, err := resolveScrollRegionBFullIdentity(repositoryRoot, !plan.Diagnostic)
	require.NoError(t, err)
	if binding.Identity == nil {
		t.Logf("T-GS-011 B-FULL binding=%s: %s", binding.Binding, binding.Note)
	} else {
		t.Logf("T-GS-011 B-FULL binding=%s candidate_tree=%s", binding.Binding, binding.Identity.CandidateTree)
	}

	axe := loadScrollRegionAxeCore(t)
	recorder := newScrollRegionBFullRecorder(t, widths, binding, plan)
	consumerCSS := scrollRegionBFullConsumerStylesheet(t)
	t.Logf("T-GS-011 B-FULL provenance=%s closure=%s widths=%v themes=%v zooms=[default ua-200]", scrollRegionBFullProvenanceStructural, scrollRegionBFullPlanClosure(plan), widths, scrollRegionBFullThemeIDs())

	baseline := scrollRegionBFullZoom{
		ID:             "default",
		Factor:         1,
		BrowserVersion: sharedBrowser.Version(),
		NewContext:     newScrollRegionBFullDefaultContext,
	}
	runScrollRegionBFullZoom(t, baseline, widths, consumerCSS, axe, recorder)

	uaZoom := newScrollRegionBFullUAZoom(t, recorder.directory)
	// This bootstrap context exists only to read the browser version. Every
	// matrix cell creates and closes its own persistent profile below; retaining
	// this one would leak a profile and defeat the isolation contract.
	uaZoom.Close()
	uaZoom.Context = nil
	runScrollRegionBFullZoom(t, uaZoom, widths, consumerCSS, axe, recorder)

	pins, err := scrollRegionSourceDependencyPins(repositoryRoot)
	require.NoError(t, err)
	recorder.tools = scrollRegionBFullToolVersions{
		GoRuntime:        runtime.Version(),
		PlaywrightGo:     pins.PlaywrightGo,
		BrowserByZoom:    map[string]string{baseline.ID: baseline.BrowserVersion, uaZoom.ID: uaZoom.BrowserVersion},
		AxeCore:          scrollRegionAxeCoreVersion,
		AxeArchiveSHA256: scrollRegionAxeArchiveSHA256,
		AxeScriptSHA256:  scrollRegionAxeScriptSHA256,
	}
	recorder.write(t, repositoryRoot)
}

func scrollRegionBFullPlan(widths []int) (scrollRegionBFullRunPlan, error) {
	fullExpected := len(scrollRegionBFullThemes()) * 2 * len(widths) * 2
	diagnostic := strings.TrimSpace(os.Getenv(scrollRegionBFullDiagnosticEnvironment))
	rawCap := strings.TrimSpace(os.Getenv(scrollRegionBFullDiagnosticMaxCellsEnv))
	if rawCap == "" {
		if diagnostic != "" {
			return scrollRegionBFullRunPlan{}, fmt.Errorf("%s requires %s", scrollRegionBFullDiagnosticEnvironment, scrollRegionBFullDiagnosticMaxCellsEnv)
		}
		return scrollRegionBFullRunPlan{ExpectedCells: fullExpected, FullExpectedCells: fullExpected}, nil
	}
	if diagnostic != "1" {
		return scrollRegionBFullRunPlan{}, fmt.Errorf("%s is diagnostic-only; set %s=1 to bypass closure", scrollRegionBFullDiagnosticMaxCellsEnv, scrollRegionBFullDiagnosticEnvironment)
	}
	cap, err := strconv.Atoi(rawCap)
	if err != nil || cap < 1 {
		return scrollRegionBFullRunPlan{}, fmt.Errorf("%s must be a positive integer", scrollRegionBFullDiagnosticMaxCellsEnv)
	}
	if cap >= len(scrollRegionBFullThemes())*2*len(widths) {
		return scrollRegionBFullRunPlan{}, fmt.Errorf("%s must be below the full per-zoom matrix", scrollRegionBFullDiagnosticMaxCellsEnv)
	}
	return scrollRegionBFullRunPlan{
		Diagnostic:        true,
		MaxCellsPerZoom:   cap,
		ExpectedCells:     cap * 2,
		FullExpectedCells: fullExpected,
	}, nil
}

func scrollRegionBFullPlanClosure(plan scrollRegionBFullRunPlan) string {
	if plan.Diagnostic {
		return "diagnostic-non-closure"
	}
	// Raw Playwright archives and PNGs are structurally validated, but their
	// bytes have no independent browser-attestation authority. Literal coverage
	// therefore remains non-closure until such an authority exists.
	return scrollRegionBFullClosureUnattested
}

func TestScrollRegionBFullPlanRejectsUnmarkedCap(t *testing.T) {
	t.Setenv(scrollRegionBFullDiagnosticEnvironment, "")
	t.Setenv(scrollRegionBFullDiagnosticMaxCellsEnv, "1")
	_, err := scrollRegionBFullPlan([]int{390})
	require.ErrorContains(t, err, "diagnostic-only")
}

func TestScrollRegionBFullPlanMarksExplicitDiagnosticNonClosure(t *testing.T) {
	t.Setenv(scrollRegionBFullDiagnosticEnvironment, "1")
	t.Setenv(scrollRegionBFullDiagnosticMaxCellsEnv, "1")
	plan, err := scrollRegionBFullPlan([]int{390})
	require.NoError(t, err)
	require.True(t, plan.Diagnostic)
	require.Equal(t, 2, plan.ExpectedCells)
	require.Equal(t, 16, plan.FullExpectedCells)
	require.Equal(t, "diagnostic-non-closure", scrollRegionBFullPlanClosure(plan))

	t.Setenv(scrollRegionBFullDiagnosticEnvironment, "")
	t.Setenv(scrollRegionBFullDiagnosticMaxCellsEnv, "")
	fullPlan, err := scrollRegionBFullPlan([]int{390})
	require.NoError(t, err)
	require.Equal(t, "unattested-non-closure", scrollRegionBFullPlanClosure(fullPlan), "structural trace/PNG evidence cannot claim independent full closure")
}

// Regression fixture copied from the V4 UA-200 receipt. It visibly contains
// browser chrome rather than the named Activity-history region; an evidence
// validator must reject it even though it is a decodable, non-uniform PNG.
func TestScrollRegionBFullWrongChromeCropFailsVisualBinding(t *testing.T) {
	repositoryRoot := scrollRegionRepositoryRoot(t)
	encoded, err := os.ReadFile(filepath.Join(repositoryRoot, "site", "tests", "e2e", "testdata", "scrollregion-ua200-wrong-chrome-crop.b64"))
	require.NoError(t, err)
	content, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(encoded), "\n", ""))
	require.NoError(t, err)
	decoded, format, err := image.Decode(bytes.NewReader(content))
	require.NoError(t, err)
	require.Equal(t, "png", format)

	err = validateScrollRegionBFullCapturedRegion(decoded, scrollRegionBFullCapture{
		ExpectedWidth:  109,
		ExpectedHeight: 236,
		Proof: scrollRegionBFullCaptureProof{
			BoundaryCue: scrollRegionBFullBoundaryCue{State: "end", Visible: true, PixelX: 1, PixelY: 1, PixelWidth: 1, PixelHeight: 1},
			Anchors: []scrollRegionBFullPixelAnchor{{
				Name:     "first-activity-card",
				PixelX:   54,
				PixelY:   118,
				Expected: [4]uint8{255, 255, 255, 255},
			}},
		},
	})
	require.ErrorContains(t, err, "named-region anchor")
}

// TestScrollRegionFooterUA200ResponsiveCompatibility proves the generic
// component-doc footer remains readable at the same real-UA-200/390 boundary
// that found the ScrollRegion visual blocker. It covers a second component
// route so responsive footer behavior cannot be inferred from one page.
func TestScrollRegionFooterUA200ResponsiveCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-UA-200 footer browser regression in short mode")
	}
	for _, route := range []string{scrollRegionBFullRoute, "/components/button"} {
		t.Run(route, func(t *testing.T) {
			zoom := newScrollRegionBFullUAZoom(t, t.TempDir())
			t.Cleanup(zoom.Close)
			page, err := zoom.Context.NewPage()
			require.NoError(t, err)
			t.Cleanup(func() { _ = page.Close() })
			page.SetDefaultTimeout(2500)
			page.SetDefaultNavigationTimeout(5000)
			require.NoError(t, page.SetViewportSize(390, 900))
			require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(scrollRegionBFullFirstPaintObserverScript())}))

			response, err := page.Goto(baseURL+route, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, 200, response.Status())
			requireScrollRegionStorageConsent(t, page, "footer-"+route)
			footer := page.Locator("#main-content footer")
			require.NoError(t, footer.ScrollIntoViewIfNeeded())

			state, err := footer.Evaluate(`(footer) => {
				const nav = footer.querySelector('nav');
				if (!nav) throw new Error('component-doc footer nav is absent');
				const visual = window.visualViewport;
				const left = visual ? visual.offsetLeft : 0;
				const right = left + (visual ? visual.width : window.innerWidth);
				const scroll = document.scrollingElement || document.documentElement;
				const originalScrollLeft = scroll.scrollLeft;
				scroll.scrollLeft = 99999;
				const attemptedScrollLeft = scroll.scrollLeft;
				scroll.scrollLeft = originalScrollLeft;
				const links = Array.from(nav.querySelectorAll('a')).map(link => {
					const rect = link.getBoundingClientRect();
					return {text: link.textContent.trim(), href: link.getAttribute('href'), left: rect.left, right: rect.right, width: rect.width};
				});
				return {
					flexWrap: getComputedStyle(nav).flexWrap,
					navClientWidth: nav.clientWidth,
					navScrollWidth: nav.scrollWidth,
					footerClientWidth: footer.clientWidth,
					footerScrollWidth: footer.scrollWidth,
					links,
					visualLeft: left,
					visualRight: right,
					scrollingClientWidth: scroll.clientWidth,
					scrollingScrollWidth: scroll.scrollWidth,
					attemptedScrollLeft,
				};
			}`, nil)
			require.NoError(t, err)
			values := state.(map[string]any)
			require.Equal(t, "wrap", values["flexWrap"], "footer legal navigation must wrap rather than overflow at real UA-200")
			require.LessOrEqual(t, scrollRegionBFullNumber(t, values["navScrollWidth"]), scrollRegionBFullNumber(t, values["navClientWidth"])+1)
			require.LessOrEqual(t, scrollRegionBFullNumber(t, values["footerScrollWidth"]), scrollRegionBFullNumber(t, values["footerClientWidth"])+1)
			require.LessOrEqual(t, scrollRegionBFullNumber(t, values["scrollingScrollWidth"]), scrollRegionBFullNumber(t, values["scrollingClientWidth"])+1, "page must not be horizontally scrollable")
			require.EqualValues(t, 0, values["attemptedScrollLeft"], "page scrolling element must reject horizontal scroll")
			links := values["links"].([]any)
			require.Len(t, links, 3)
			require.Equal(t, []string{"Privacy", "Attributions", "License"}, []string{links[0].(map[string]any)["text"].(string), links[1].(map[string]any)["text"].(string), links[2].(map[string]any)["text"].(string)}, "wrapping must preserve the DOM reading order")
			for _, raw := range links {
				link := raw.(map[string]any)
				require.Greater(t, scrollRegionBFullNumber(t, link["width"]), 0.0)
				require.GreaterOrEqual(t, scrollRegionBFullNumber(t, link["left"]), scrollRegionBFullNumber(t, values["visualLeft"])-1)
				require.LessOrEqual(t, scrollRegionBFullNumber(t, link["right"]), scrollRegionBFullNumber(t, values["visualRight"])+1)
			}
		})
	}
}

// TestScrollRegionUA200HorizontalAccessContract keeps ordinary maintained
// prose visible at real browser zoom while leaving generic wide consumer
// content reachable through the public horizontal scroll axis.
func TestScrollRegionUA200HorizontalAccessContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-UA-200 ScrollRegion axis regression in short mode")
	}
	zooms := []scrollRegionBFullZoom{
		{ID: "default", Factor: 1, NewContext: newScrollRegionBFullDefaultContext},
		{ID: "ua-200", Factor: scrollRegionBFullUAZoomFactor, RealUA: true, NewContext: newScrollRegionBFullUAContext},
	}
	theme, ok := scrollRegionBFullThemeByID("araihu")
	require.True(t, ok)
	for _, zoom := range zooms {
		for _, dark := range []bool{false, true} {
			t.Run(zoom.ID+"/"+scrollRegionBFullMode(dark), func(t *testing.T) {
				context := zoom.NewContext(t)
				t.Cleanup(func() { _ = context.Close() })
				page, err := context.NewPage()
				require.NoError(t, err)
				t.Cleanup(func() { _ = page.Close() })
				require.NoError(t, page.SetViewportSize(390, 900))
				require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce}))
				require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(scrollRegionBFullFirstPaintObserverScript())}))
				failures := watchPageFailures(page)
				response, err := page.Goto(scrollRegionBFullRoutedURL(theme), playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
				require.NoError(t, err)
				require.NotNil(t, response)
				require.Equal(t, 200, response.Status())
				requireScrollRegionStorageConsent(t, page, "axis-"+zoom.ID+"-"+scrollRegionBFullMode(dark))
				requireScrollRegionDarkModeThroughPublicUI(t, page, "axis-"+zoom.ID+"-"+scrollRegionBFullMode(dark), dark)
				requireScrollRegionPageHealthy(t, page, failures)
				freshPage, _ := requireScrollRegionFreshPersistedPage(t, context, theme, dark, 390, zoom, scrollRegionBFullCellRoutedURL(theme, dark, 390, zoom), "")
				t.Cleanup(func() { _ = freshPage.Close() })
				page = freshPage
				freshFailures := watchPageFailures(page)
				viewport := page.Locator("#scroll-region-default [data-goshtoso-scroll-viewport]")
				require.NoError(t, viewport.WaitFor())
				_, _, _, _ = requireScrollRegionFirstPaint(t, page, viewport, theme, dark, 390, zoom)
				requireScrollRegionHorizontalAccess(t, page, viewport)
				requireScrollRegionPageHealthy(t, page, freshFailures)

				// The maintained external consumer stylesheet deliberately makes
				// its Activity cards wider than the narrow viewport. This proves
				// the public horizontal axis by genuine keyboard and CDP touch
				// actions without relabeling wrapped documentation prose.
				consumer, ok := scrollRegionBFullThemeByID(scrollRegionBFullConsumerTheme)
				require.True(t, ok)
				consumerCSS := scrollRegionBFullConsumerStylesheet(t)
				wideRoute := scrollRegionBFullCellRoutedURL(consumer, dark, 390, zoom)
				widePage, err := context.NewPage()
				require.NoError(t, err)
				t.Cleanup(func() { _ = widePage.Close() })
				require.NoError(t, widePage.SetViewportSize(390, 900))
				require.NoError(t, widePage.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce}))
				require.NoError(t, widePage.AddInitScript(playwright.Script{Content: playwright.String(scrollRegionBFullFirstPaintObserverScript())}))
				installScrollRegionBFullConsumerRoute(t, widePage, consumer, wideRoute, consumerCSS)
				wideResponse, err := widePage.Goto(wideRoute, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
				require.NoError(t, err)
				require.Equal(t, 200, wideResponse.Status())
				requireScrollRegionStorageConsent(t, widePage, "wide-"+zoom.ID+"-"+scrollRegionBFullMode(dark))
				requireScrollRegionDarkModeThroughPublicUI(t, widePage, "wide-"+zoom.ID+"-"+scrollRegionBFullMode(dark), dark)
				wideFresh, _ := requireScrollRegionFreshPersistedPage(t, context, consumer, dark, 390, zoom, wideRoute, consumerCSS)
				t.Cleanup(func() { _ = wideFresh.Close() })
				requireScrollRegionWideConsumerHorizontalAccess(t, wideFresh, wideFresh.Locator("#scroll-region-default [data-goshtoso-scroll-viewport]"))
			})
		}
	}
}

// requireScrollRegionWideConsumerHorizontalAccess exercises the public
// horizontal ScrollRegion axis against the maintained external fixture. All
// movement is issued through Playwright keyboard/CDP touch input; it never
// writes scrollLeft as an assertion shortcut.
func requireScrollRegionWideConsumerHorizontalAccess(t *testing.T, page playwright.Page, viewport playwright.Locator) {
	t.Helper()
	require.NoError(t, viewport.WaitFor())
	before, err := viewport.Evaluate(`el => ({left: el.scrollLeft, clientWidth: el.clientWidth, scrollWidth: el.scrollWidth, overflowX: getComputedStyle(el).overflowX})`, nil)
	require.NoError(t, err)
	beforeValues := before.(map[string]any)
	require.EqualValues(t, 0, beforeValues["left"])
	require.Greater(t, scrollRegionBFullNumber(t, beforeValues["scrollWidth"]), scrollRegionBFullNumber(t, beforeValues["clientWidth"])+1, "wide consumer fixture must expose a real horizontal range")
	require.Equal(t, "auto", beforeValues["overflowX"])

	require.NoError(t, viewport.Focus())
	require.NoError(t, viewport.Press("ArrowRight"))
	_, err = page.WaitForFunction(`selector => document.querySelector(selector).scrollLeft > 0`, "#scroll-region-default [data-goshtoso-scroll-viewport]")
	require.NoError(t, err, "real ArrowRight must move the focused wide consumer viewport")
	keyboard, err := viewport.Evaluate(`el => el.scrollLeft`, nil)
	require.NoError(t, err)
	require.Greater(t, scrollRegionBFullNumber(t, keyboard), 0.0)

	box, err := viewport.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, box)
	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Detach() })
	_, err = session.Send("Input.dispatchTouchEvent", map[string]any{"type": "touchStart", "touchPoints": []map[string]float64{{"x": box.X + box.Width - 12, "y": box.Y + box.Height/2}}})
	require.NoError(t, err)
	_, err = session.Send("Input.dispatchTouchEvent", map[string]any{"type": "touchMove", "touchPoints": []map[string]float64{{"x": box.X + 12, "y": box.Y + box.Height/2}}})
	require.NoError(t, err)
	_, err = session.Send("Input.dispatchTouchEvent", map[string]any{"type": "touchEnd", "touchPoints": []map[string]float64{}})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`selector => document.querySelector(selector).scrollLeft > 0`, "#scroll-region-default [data-goshtoso-scroll-viewport]")
	require.NoError(t, err, "real CDP touch swipe must retain horizontal movement")
	touch, err := viewport.Evaluate(`el => ({left: el.scrollLeft, max: el.scrollWidth - el.clientWidth})`, nil)
	require.NoError(t, err)
	touchValues := touch.(map[string]any)
	require.Greater(t, scrollRegionBFullNumber(t, touchValues["left"]), 0.0)
	require.LessOrEqual(t, scrollRegionBFullNumber(t, touchValues["left"]), scrollRegionBFullNumber(t, touchValues["max"])+1)
}

func requireScrollRegionHorizontalAccess(t *testing.T, page playwright.Page, defaultViewport playwright.Locator) {
	t.Helper()
	for selector, requireVisibleDeployment := range map[string]bool{
		"#scroll-region-default [data-goshtoso-scroll-viewport]":     true,
		"#scroll-region-no-overflow [data-goshtoso-scroll-viewport]": false,
	} {
		selector, requireVisibleDeployment := selector, requireVisibleDeployment
		t.Run(selector, func(t *testing.T) {
			viewport := page.Locator(selector)
			if selector == "#scroll-region-default [data-goshtoso-scroll-viewport]" {
				viewport = defaultViewport
			}
			require.NoError(t, viewport.WaitFor())
			state, err := viewport.Evaluate(`(el, requireVisibleDeployment) => {
				const rect = el.getBoundingClientRect();
				const original = el.scrollLeft;
				el.scrollLeft = original + 99999;
				const attempted = el.scrollLeft;
				el.scrollLeft = original;
				const deployment = Array.from(el.querySelectorAll('li')).flatMap(item => Array.from(item.querySelectorAll('span'))).find(node => node.textContent.includes('Deployment'));
				const textRects = deployment ? Array.from(deployment.getClientRects()).map(part => ({left: part.left, right: part.right, width: part.width, height: part.height})) : [];
				return {
					clientWidth: el.clientWidth,
					scrollWidth: el.scrollWidth,
					attemptedScrollLeft: attempted,
					overflowX: getComputedStyle(el).overflowX,
					text: deployment ? deployment.textContent.trim() : '',
					textRects,
					viewportLeft: rect.left,
					viewportRight: rect.right,
					requireVisibleDeployment,
				};
			}`, requireVisibleDeployment)
			require.NoError(t, err)
			values := state.(map[string]any)
			require.LessOrEqual(t, scrollRegionBFullNumber(t, values["scrollWidth"]), scrollRegionBFullNumber(t, values["clientWidth"])+1, "maintained demo copy must wrap without a horizontal range: %#v", values)
			require.EqualValues(t, 0, values["attemptedScrollLeft"], "wrapped demo copy must not acquire a horizontal range: %#v", values)
			require.Equal(t, "auto", values["overflowX"])
			if requireVisibleDeployment {
				require.Contains(t, values["text"], "Deployment state recorded.")
				rects := values["textRects"].([]any)
				require.NotEmpty(t, rects)
				for _, raw := range rects {
					part := raw.(map[string]any)
					require.Greater(t, scrollRegionBFullNumber(t, part["width"]), 0.0)
					require.Greater(t, scrollRegionBFullNumber(t, part["height"]), 0.0)
					require.GreaterOrEqual(t, scrollRegionBFullNumber(t, part["left"]), scrollRegionBFullNumber(t, values["viewportLeft"])-1)
					require.LessOrEqual(t, scrollRegionBFullNumber(t, part["right"]), scrollRegionBFullNumber(t, values["viewportRight"])+1)
				}
			}
		})
	}
}

func TestScrollRegionBFullATReceiptHarness(t *testing.T) {
	repositoryRoot := scrollRegionRepositoryRoot(t)
	templatePath := filepath.Join(repositoryRoot, "tests", "external", "scrollregion-a11y", "at-receipt.template.json")
	template, _, err := readScrollRegionATEvidenceReceipt(templatePath)
	require.NoError(t, err)
	require.Equal(t, scrollRegionATReceiptSchema, template.Schema)
	require.Equal(t, "template-not-evidence", template.Status)
	require.Len(t, template.Captures, 2)
	require.True(t, strings.HasPrefix(template.Identity.Head, "REPLACE_WITH_"))
	require.True(t, strings.HasPrefix(template.Challenge, "REPLACE_WITH_"))

	receiptPath := os.Getenv("GOSHTOSO_SCROLLREGION_AT_RECEIPT")
	identityPath := os.Getenv("GOSHTOSO_SCROLLREGION_AT_IDENTITY")
	challengePath := os.Getenv(scrollRegionATChallengeEnvironment)
	replayRegistry := os.Getenv(scrollRegionATReplayRegistryEnvironment)
	if receiptPath == "" || identityPath == "" || challengePath == "" || replayRegistry == "" {
		t.Skip("PENDING_EXTERNAL_AT_CAPTURE: generate an independent challenge, capture macOS Safari+VoiceOver and macOS Chromium+VoiceOver through cmd/scrollregionatcapture, assemble the v3 receipt, then run scripts/validate-scrollregion-at-receipt.sh with an external owner-only challenge registry")
	}
	require.NoError(t, validateScrollRegionATReceipt(repositoryRoot, identityPath, receiptPath))
}

func scrollRegionBFullThemes() []scrollRegionBFullTheme {
	return []scrollRegionBFullTheme{
		{ID: "araihu", ServerTheme: "araihu"},
		{ID: "goshtoso", ServerTheme: "goshtoso"},
		{ID: "minimal", ServerTheme: "minimal"},
		{ID: scrollRegionBFullConsumerTheme, ServerTheme: "araihu", ConsumerCSS: true},
	}
}

// scrollRegionBFullRoutedURL asks the maintained server to bind the literal
// visual-theme axis into raw initial HTML. It never relies on the removed
// component-doc selector or on localStorage theme persistence.
func scrollRegionBFullRoutedURL(theme scrollRegionBFullTheme) string {
	query := url.Values{}
	query.Set("t-gs-011-theme", theme.ServerTheme)
	if theme.ConsumerCSS {
		query.Set("t-gs-011-consumer", "scrollregion")
	}
	return baseURL + scrollRegionBFullRoute + "?" + query.Encode()
}

// scrollRegionBFullCellRoutedURL embeds every literal cell axis in both Page
// A and fresh Page B navigations. The server ignores these evidence-only
// parameters, but Playwright's raw network trace retains them and the receipt
// validator can therefore reject a trace copied from another cell.
func scrollRegionBFullCellRoutedURL(theme scrollRegionBFullTheme, dark bool, width int, zoom scrollRegionBFullZoom) string {
	query := url.Values{}
	query.Set("t-gs-011-theme", theme.ServerTheme)
	if theme.ConsumerCSS {
		query.Set("t-gs-011-consumer", theme.ID)
	}
	query.Set("t-gs-011-mode", scrollRegionBFullMode(dark))
	query.Set("t-gs-011-width", strconv.Itoa(width))
	query.Set("t-gs-011-zoom", zoom.ID)
	query.Set("t-gs-011-cell", scrollRegionBFullCellID(theme, dark, width, zoom))
	return baseURL + scrollRegionBFullRoute + "?" + query.Encode()
}

func scrollRegionBFullThemeIDs() []string {
	themes := scrollRegionBFullThemes()
	ids := make([]string, 0, len(themes))
	for _, theme := range themes {
		ids = append(ids, theme.ID)
	}
	return ids
}

func scrollRegionBFullWidths(t *testing.T) []int {
	t.Helper()
	repositoryRoot := scrollRegionRepositoryRoot(t)
	stylesheet, err := os.ReadFile(filepath.Join(repositoryRoot, "assets", "styles.css"))
	require.NoError(t, err)

	breakpointPattern := regexp.MustCompile(`@media\s*\(width >= ([0-9]+(?:\.[0-9]+)?)rem\)`)
	matches := breakpointPattern.FindAllStringSubmatch(string(stylesheet), -1)
	require.NotEmpty(t, matches, "compiled Goshtoso stylesheet must declare responsive breakpoints")

	breakpoints := make(map[int]struct{}, len(matches))
	for _, match := range matches {
		rem, parseErr := strconv.ParseFloat(match[1], 64)
		require.NoError(t, parseErr)
		breakpoints[int(rem*16+0.5)] = struct{}{}
	}
	require.Contains(t, breakpoints, 640, "known compiled Goshtoso breakpoint must remain 640px")

	widthSet := map[int]struct{}{390: {}, 768: {}, 1440: {}}
	orderedBreakpoints := make([]int, 0, len(breakpoints))
	for breakpoint := range breakpoints {
		orderedBreakpoints = append(orderedBreakpoints, breakpoint)
		for _, edge := range []int{breakpoint - 1, breakpoint, breakpoint + 1} {
			widthSet[edge] = struct{}{}
		}
	}
	sort.Ints(orderedBreakpoints)
	for index := 0; index+1 < len(orderedBreakpoints); index++ {
		left, right := orderedBreakpoints[index], orderedBreakpoints[index+1]
		if right-left > 2 {
			widthSet[left+(right-left)/2] = struct{}{}
		}
	}

	widths := make([]int, 0, len(widthSet))
	for width := range widthSet {
		widths = append(widths, width)
	}
	sort.Ints(widths)
	return widths
}

func scrollRegionRepositoryRoot(t *testing.T) string {
	t.Helper()
	siteRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	return filepath.Dir(siteRoot)
}

func scrollRegionBFullConsumerStylesheet(t *testing.T) string {
	t.Helper()
	path := filepath.Join(scrollRegionRepositoryRoot(t), "tests", "external", "scrollregion-a11y", "consumer-scrollregion.css")
	stylesheet, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(stylesheet), scrollRegionBFullConsumerAttribute)
	require.NotContains(t, strings.ToLower(string(stylesheet)), "modern")
	return string(stylesheet)
}

func newScrollRegionBFullDefaultContext(t *testing.T) playwright.BrowserContext {
	t.Helper()
	context, err := sharedBrowser.NewContext(playwright.BrowserNewContextOptions{
		HasTouch: new(true),
		Viewport: &playwright.Size{Width: 1440, Height: 900},
	})
	require.NoError(t, err)
	return context
}

func newScrollRegionBFullUAZoom(t *testing.T, _ string) scrollRegionBFullZoom {
	t.Helper()
	context := newScrollRegionBFullUAContext(t)
	return scrollRegionBFullZoom{
		ID:             "ua-200",
		Factor:         scrollRegionBFullUAZoomFactor,
		RealUA:         true,
		Context:        context,
		Close:          func() { _ = context.Close() },
		NewContext:     newScrollRegionBFullUAContext,
		BrowserVersion: context.Browser().Version(),
	}
}

// newScrollRegionBFullUAContext creates one fresh persistent Chromium profile
// for one literal B-FULL cell. Browser zoom preferences live in the profile,
// so reusing it would also reuse consent cookies and violate cell isolation.
func newScrollRegionBFullUAContext(t *testing.T) playwright.BrowserContext {
	t.Helper()
	profileDirectory := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(profileDirectory, "Default"), 0o755))
	preferences, err := json.Marshal(map[string]any{
		"partition": map[string]any{
			"per_host_zoom_levels": map[string]any{
				"x": map[string]any{
					"localhost": map[string]any{
						"last_modified": "13430990000000000",
						"zoom_level":    scrollRegionBFullChromiumZoomLevel,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(profileDirectory, "Default", "Preferences"), preferences, 0o600))

	context, err := sharedPW.Chromium.LaunchPersistentContext(profileDirectory, playwright.BrowserTypeLaunchPersistentContextOptions{
		// Chromium's persistent-profile zoom preference is ignored in headless
		// mode. Keep this headed so the focused UA contract can verify actual
		// browser metrics rather than mislabeling a CDP page-scale emulation.
		Headless: new(false),
		HasTouch: new(true),
		Viewport: &playwright.Size{Width: 1440, Height: 900},
		Timeout:  playwright.Float(10000),
	})
	require.NoError(t, err, "real 200%% UA zoom requires headed bundled Chromium; run under xvfb-run on Linux")
	return context
}

func runScrollRegionBFullZoom(t *testing.T, zoom scrollRegionBFullZoom, widths []int, consumerCSS string, axe scrollRegionAxeCore, recorder *scrollRegionBFullRecorder) {
	t.Helper()
	require.NotNil(t, zoom.NewContext, "B-FULL zoom %q must create an isolated context per cell", zoom.ID)
	runCells := 0
	for _, theme := range scrollRegionBFullThemes() {
		if recorder.plan.Diagnostic && runCells >= recorder.plan.MaxCellsPerZoom {
			break
		}
		for _, dark := range []bool{false, true} {
			if recorder.plan.Diagnostic && runCells >= recorder.plan.MaxCellsPerZoom {
				break
			}
			for _, width := range widths {
				if recorder.plan.Diagnostic && runCells >= recorder.plan.MaxCellsPerZoom {
					break
				}
				label := fmt.Sprintf("theme=%s/mode=%s/width=%d/zoom=%s", theme.ID, scrollRegionBFullMode(dark), width, zoom.ID)
				if !t.Run(label, func(t *testing.T) {
					context := zoom.NewContext(t)
					t.Cleanup(func() { _ = context.Close() })
					cellID := scrollRegionBFullCellID(theme, dark, width, zoom)
					tracePath := filepath.Join(recorder.directory, "scrollregion-bfull-"+scrollRegionBFullArtifactName(cellID)+".trace.zip")
					require.NoError(t, context.Tracing().Start(playwright.TracingStartOptions{
						Name:        playwright.String("scrollregion-bfull-" + scrollRegionBFullArtifactName(cellID)),
						Screenshots: playwright.Bool(true),
						Snapshots:   playwright.Bool(true),
						Sources:     playwright.Bool(true),
					}))
					runScrollRegionBFullCell(t, zoom, context, theme, dark, width, consumerCSS, axe, recorder)
					require.NoError(t, context.Tracing().Stop(tracePath))
					trace := scrollRegionBFullArtifactForFile(t, tracePath, []byte("PK\x03\x04"))
					require.NotEmpty(t, recorder.cells)
					require.Equal(t, cellID, recorder.cells[len(recorder.cells)-1].CellID)
					recorder.cells[len(recorder.cells)-1].Trace = &trace
					recorder.traces[cellID] = trace
				}) {
					t.FailNow()
				}
				runCells++
			}
		}
	}
	if recorder.plan.Diagnostic {
		require.Equal(t, recorder.plan.MaxCellsPerZoom, runCells, "each diagnostic zoom must execute exactly its declared capped cell count")
	} else {
		require.Equal(t, len(scrollRegionBFullThemes())*2*len(widths), runCells, "full B-FULL zoom must execute every literal theme/mode/width cell")
	}
}

func scrollRegionBFullArtifactName(cellID string) string {
	replacer := strings.NewReplacer("/", "-", "|", "-", "?", "-", "=", "-")
	return replacer.Replace(cellID)
}

func runScrollRegionBFullCell(t *testing.T, zoom scrollRegionBFullZoom, context playwright.BrowserContext, theme scrollRegionBFullTheme, dark bool, width int, consumerCSS string, axe scrollRegionAxeCore, recorder *scrollRegionBFullRecorder) {
	t.Helper()
	cellID := scrollRegionBFullCellID(theme, dark, width, zoom)
	routedURL := scrollRegionBFullCellRoutedURL(theme, dark, width, zoom)
	page, err := context.NewPage()
	require.NoError(t, err)
	t.Cleanup(func() { _ = page.Close() })
	page.SetDefaultTimeout(2500)
	page.SetDefaultNavigationTimeout(5000)
	require.NoError(t, page.SetViewportSize(width, 900))
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce}))
	require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(scrollRegionBFullFirstPaintObserverScript())}))
	if theme.ConsumerCSS {
		installScrollRegionBFullConsumerRoute(t, page, theme, routedURL, consumerCSS)
	}
	firstLoadFailures := watchPageFailures(page)

	response, err := page.Goto(routedURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, 200, response.Status())
	initialHTML, err := response.Body()
	require.NoError(t, err)
	requireScrollRegionInitialHTML(t, initialHTML, theme)
	storageBeforeState := scrollRegionBFullPageStateFor(t, page)
	require.True(t, storageBeforeState.DialogVisible, "fresh cell context must expose first-run consent")
	require.NotContains(t, storageBeforeState.Cookie, "gt_storage=allowed", "fresh cell context must not inherit another cell consent cookie")
	consentAction := requireScrollRegionStorageConsent(t, page, cellID)
	stateActions := requireScrollRegionDarkModeThroughPublicUI(t, page, cellID, dark)
	requireScrollRegionPageHealthy(t, page, firstLoadFailures)

	// The locked site theme arrives in server-routed initial HTML. Only dark
	// mode has a maintained public control, so Page A persists that state via a
	// real click and Page B proves the product-owned storage restoration.
	freshPage, freshResponse := requireScrollRegionFreshPersistedPage(t, context, theme, dark, width, zoom, routedURL, consumerCSS)
	t.Cleanup(func() { _ = freshPage.Close() })
	freshHTML, err := freshResponse.Body()
	require.NoError(t, err)
	requireScrollRegionInitialHTML(t, freshHTML, theme)
	freshStorageState := scrollRegionBFullPageStateFor(t, freshPage)
	require.False(t, freshStorageState.DialogVisible, "fresh Page B must inherit Page A real consent through its cell-owned context")
	require.Contains(t, freshStorageState.Cookie, "gt_storage=allowed")
	require.Equal(t, strconv.FormatBool(dark), freshStorageState.DarkMode)
	page = freshPage
	failures := watchPageFailures(page)
	defaultRoot := page.Locator("#scroll-region-default [data-goshtoso-scroll-region]")
	viewport := defaultRoot.Locator("[data-goshtoso-scroll-viewport]")
	require.NoError(t, defaultRoot.WaitFor())
	initialPaint, firstPaint, settledPaint, observerEvents := requireScrollRegionFirstPaint(t, page, viewport, theme, dark, width, zoom)
	capture := requireScrollRegionVisualCapture(t, page, defaultRoot, viewport, theme.ConsumerCSS)
	initialScreenshot := recorder.screenshot(t, page, capture, theme, dark, width, zoom, "post-consent-region")

	requireScrollRegionAccessibleNames(t, page)
	// Scan every theme/mode/width/zoom cell at its fresh start boundary. The
	// subsequent middle/end checks intentionally make gradient boundary cues
	// visible; those cues can visually occlude underlying copy, which is not a
	// stable color-contrast target for axe. Roles, names, state truth, and the
	// rendered focus boundary are asserted independently after each real action.
	requireScrollRegionAxeContract(t, page, axe)
	requireScrollRegionNoOverflow(t, page)
	requireScrollRegionNativeInputs(t, page, viewport)
	requireScrollRegionReducedMotion(t, page, viewport)
	requireScrollRegionPageHealthy(t, page, failures)
	actionEvidence := append([]scrollRegionBFullActionEvidence{consentAction}, stateActions...)
	actions := make([]scrollRegionBFullAction, 0, len(actionEvidence))
	actionArtifacts := make([]scrollRegionBFullArtifact, 0, len(actionEvidence))
	for index, evidence := range actionEvidence {
		actions = append(actions, evidence.Action)
		actionArtifacts = append(actionArtifacts, recorder.jsonArtifact(t, cellID, fmt.Sprintf("page-a-action-%02d", index), evidence))
	}
	pageAInitialArtifact := recorder.bytesArtifact(t, cellID, "page-a-initial", initialHTML)
	pageAStorageArtifact := recorder.jsonArtifact(t, cellID, "page-a-storage-before", scrollRegionBFullStorageEvidence{Schema: scrollRegionBFullStorageEvidenceSchema, CellID: cellID, Phase: "page-a-storage-before", State: storageBeforeState})
	pageBStorageArtifact := recorder.jsonArtifact(t, cellID, "page-b-storage", scrollRegionBFullStorageEvidence{Schema: scrollRegionBFullStorageEvidenceSchema, CellID: cellID, Phase: "page-b-storage", State: freshStorageState})
	pageBInitialArtifact := recorder.bytesArtifact(t, cellID, "page-b-initial", freshHTML)
	paintArtifact := recorder.jsonArtifact(t, cellID, "page-b-first-paint", scrollRegionBFullPaintEvidenceFromObserver(cellID, observerEvents, settledPaint))
	nas := map[string]string{
		"dismiss-and-Escape": "Source-grounded N/A: ScrollRegion owns no dismissible surface; Escape has no owned dismissal outcome.",
		"focus-return":       "Source-grounded N/A: ScrollRegion does not open a transient surface; focus remains on its stable owned viewport across boundary actions.",
		"theme-persistence":  scrollRegionBFullThemePersistenceNA,
	}

	recorder.cells = append(recorder.cells, scrollRegionBFullCellReceipt{
		CellID:          cellID,
		Route:           scrollRegionBFullRoute,
		Theme:           theme.ID,
		Mode:            scrollRegionBFullMode(dark),
		ViewportWidth:   width,
		Zoom:            zoom.ID,
		States:          []string{"default", "no-overflow", "start", "middle", "end", "focused"},
		Inputs:          []string{"mouse", "keyboard", "cdp-touch"},
		FirstHTMLSHA256: scrollRegionBFullSHA256(freshHTML),
		SetupActions:    actions,
		Persistence: scrollRegionBFullPersistenceProof{
			ThemeInitialSource:            scrollRegionBFullThemeInitialSource,
			ThemePersistenceNotApplicable: scrollRegionBFullThemePersistenceNA,
			DarkPersistence:               scrollRegionBFullDarkPersistence,
			StorageBefore:                 map[string]string{"darkMode": storageBeforeState.DarkMode, "cookie": storageBeforeState.Cookie},
			Actions:                       actions,
			FreshLoadInitialHTML:          initialPaint,
			FreshLoadFirstPaint:           firstPaint,
			FreshLoadSettled:              settledPaint,
			PageAInitialHTML:              pageAInitialArtifact,
			PageAStorageBefore:            pageAStorageArtifact,
			PageAActions:                  actionArtifacts,
			FreshLoadStorage:              pageBStorageArtifact,
			FreshLoadInitialHTMLArtifact:  pageBInitialArtifact,
			FreshLoadPaintArtifact:        paintArtifact,
		},
		Screenshot: initialScreenshot,
		NAs:        nas,
	})
}

func scrollRegionBFullCellID(theme scrollRegionBFullTheme, dark bool, width int, zoom scrollRegionBFullZoom) string {
	return fmt.Sprintf("%s|%s|%s|%d|%s", scrollRegionBFullRoute, theme.ID, scrollRegionBFullMode(dark), width, zoom.ID)
}

func scrollRegionBFullMode(dark bool) string {
	if dark {
		return "dark"
	}
	return "light"
}

func scrollRegionBFullFirstPaintObserverScript() string {
	return `(() => {
		const snapshot = phase => {
			const root = document.documentElement;
			const viewport = document.querySelector('#scroll-region-default [data-goshtoso-scroll-viewport]');
			const box = viewport ? viewport.getBoundingClientRect() : null;
			return {
				phase,
				at: performance.now(),
				readyState: document.readyState,
				theme: root ? root.getAttribute("data-theme") || "" : "",
				themeSource: root ? root.getAttribute("data-goshtoso-theme-initial-source") || "" : "",
				dark: !!root && root.classList.contains("dark"),
				visible: !!box && box.width > 0 && box.height > 0,
				role: viewport ? viewport.getAttribute("role") || "" : "",
				name: viewport ? viewport.getAttribute("aria-label") || "" : "",
			};
		};
		const proof = { events: [] };
		window.__goshtosoTGS011FirstPaint = proof;
		const installRootObserver = () => {
			const root = document.documentElement;
			if (!root) {
				const documentObserver = new MutationObserver(() => {
					if (!document.documentElement) return;
					documentObserver.disconnect();
					installRootObserver();
				});
				documentObserver.observe(document, { childList: true });
				return;
			}
			proof.events.push(snapshot("init"));
			new MutationObserver(() => proof.events.push(snapshot("root-mutation"))).observe(root, {
				attributes: true,
				attributeFilter: ["data-theme", "class"],
			});
		};
		installRootObserver();
		document.addEventListener("DOMContentLoaded", () => proof.events.push(snapshot("dom-content-loaded")), { once: true });
		requestAnimationFrame(() => proof.events.push(snapshot("first-animation-frame")));
	})();`
}

// requireScrollRegionStorageConsent uses the maintained first-run control as a
// real browser action. It intentionally does not inject, hide, or remove the
// dialog, so visual evidence cannot ignore a fixed interactive surface.
func requireScrollRegionStorageConsent(t *testing.T, page playwright.Page, cellID string) scrollRegionBFullActionEvidence {
	t.Helper()
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Browser storage"})
	allow := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Allow browser storage"})
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.NoError(t, allow.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	before := scrollRegionBFullPageStateFor(t, page)
	require.True(t, before.DialogVisible)
	require.NotContains(t, before.Cookie, "gt_storage=allowed")

	require.NoError(t, allow.Click())
	require.NoError(t, dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}))
	after := scrollRegionBFullPageStateFor(t, page)
	require.False(t, after.DialogVisible)
	require.Contains(t, after.Cookie, "gt_storage=allowed")
	return scrollRegionBFullActionEvidence{
		Schema: scrollRegionBFullActionEvidenceSchema,
		CellID: cellID,
		Phase:  "page-a-consent",
		Before: before,
		Action: scrollRegionBFullAction{
			Before: "browser-storage-dialog-visible; gt_storage is absent",
			Action: "Playwright mouse click Allow browser storage",
			Return: "browser-storage-dialog-hidden; gt_storage=allowed",
		},
		After: after,
	}
}

// requireScrollRegionDarkModeThroughPublicUI establishes the one persisted
// appearance state the locked component docs actually expose. Theme selection
// is intentionally absent: that axis is supplied by the server-routed raw HTML
// response and must never be relabeled as client persistence.
func requireScrollRegionDarkModeThroughPublicUI(t *testing.T, page playwright.Page, cellID string, dark bool) []scrollRegionBFullActionEvidence {
	t.Helper()
	actions := make([]scrollRegionBFullActionEvidence, 0, 2)
	currentDark := scrollRegionBFullDarkMode(t, page)
	if currentDark == dark {
		actions = append(actions, requireScrollRegionToggleDarkMode(t, page, cellID, "page-a-dark-reset", !dark))
	}
	actions = append(actions, requireScrollRegionToggleDarkMode(t, page, cellID, "page-a-dark-set", dark))
	state := scrollRegionBFullStorageState(t, page)
	require.Equal(t, strconv.FormatBool(dark), state["darkMode"])
	return actions
}

func requireScrollRegionToggleDarkMode(t *testing.T, page playwright.Page, cellID, phase string, want bool) scrollRegionBFullActionEvidence {
	t.Helper()
	before := scrollRegionBFullPageStateFor(t, page)
	toggle := page.Locator("#darkModeToggleBtn")
	require.NoError(t, toggle.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}))
	require.NoError(t, toggle.Click())
	_, err := page.WaitForFunction(`want => document.documentElement.classList.contains("dark") === want && localStorage.getItem("darkMode") === String(want)`, want)
	require.NoError(t, err)
	after := scrollRegionBFullPageStateFor(t, page)
	require.Equal(t, want, after.Dark)
	require.Equal(t, strconv.FormatBool(want), after.DarkMode)
	return scrollRegionBFullActionEvidence{
		Schema: scrollRegionBFullActionEvidenceSchema,
		CellID: cellID,
		Phase:  phase,
		Before: before,
		Action: scrollRegionBFullAction{Before: "dark=" + strconv.FormatBool(before.Dark), Action: "Playwright mouse click dark mode toggle", Return: "dark=" + strconv.FormatBool(want) + "; localStorage darkMode=" + strconv.FormatBool(want)},
		After:  after,
	}
}

func scrollRegionBFullDarkMode(t *testing.T, page playwright.Page) bool {
	t.Helper()
	value, err := page.Evaluate(`() => document.documentElement.classList.contains("dark")`, nil)
	require.NoError(t, err)
	dark, ok := value.(bool)
	require.True(t, ok)
	return dark
}

func scrollRegionBFullStorageState(t *testing.T, page playwright.Page) map[string]string {
	t.Helper()
	state := scrollRegionBFullPageStateFor(t, page)
	return map[string]string{"darkMode": state.DarkMode, "cookie": state.Cookie}
}

func scrollRegionBFullPageStateFor(t *testing.T, page playwright.Page) scrollRegionBFullPageState {
	t.Helper()
	value, err := page.Evaluate(`() => {
		const dialog = document.querySelector('[role="dialog"][aria-labelledby="cookie-banner-title"]');
		const style = dialog ? getComputedStyle(dialog) : null;
		const rect = dialog ? dialog.getBoundingClientRect() : null;
		return {
			darkMode: localStorage.getItem("darkMode") || "",
			cookie: document.cookie || "",
			dark: document.documentElement.classList.contains("dark"),
			dialogVisible: !!dialog && style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0,
		};
	}`, nil)
	require.NoError(t, err)
	values := value.(map[string]any)
	return scrollRegionBFullPageState{
		DarkMode:      values["darkMode"].(string),
		Cookie:        values["cookie"].(string),
		Dark:          values["dark"].(bool),
		DialogVisible: values["dialogVisible"].(bool),
	}
}

func requireScrollRegionFreshPersistedPage(t *testing.T, context playwright.BrowserContext, theme scrollRegionBFullTheme, dark bool, width int, zoom scrollRegionBFullZoom, routedURL, consumerCSS string) (playwright.Page, playwright.Response) {
	t.Helper()
	page, err := context.NewPage()
	require.NoError(t, err)
	page.SetDefaultTimeout(2500)
	page.SetDefaultNavigationTimeout(5000)
	require.NoError(t, page.SetViewportSize(width, 900))
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce}))
	require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(scrollRegionBFullFirstPaintObserverScript())}))
	if theme.ConsumerCSS {
		installScrollRegionBFullConsumerRoute(t, page, theme, routedURL, consumerCSS)
	}
	require.Equal(t, routedURL, scrollRegionBFullCellRoutedURL(theme, dark, width, zoom), "fresh Page B must retain the literal trace-bound cell URL")
	response, err := page.Goto(routedURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, 200, response.Status())
	return page, response
}

func installScrollRegionBFullConsumerRoute(t *testing.T, page playwright.Page, theme scrollRegionBFullTheme, routedURL, stylesheet string) {
	t.Helper()
	require.True(t, theme.ConsumerCSS)
	require.NotContains(t, strings.ToLower(stylesheet), "</style", "consumer fixture stylesheet must be safe to embed in its own initial document style element")
	require.NoError(t, page.Route(routedURL, func(route playwright.Route) {
		response, err := route.Fetch()
		if err != nil {
			t.Errorf("fetch consumer fixture route: %v", err)
			_ = route.Abort()
			return
		}
		body, err := response.Body()
		if err != nil {
			t.Errorf("read consumer fixture route body: %v", err)
			_ = route.Abort()
			return
		}
		const marker = "</head>"
		injection := `<style id="t-gs-011-consumer-scrollregion-css">` + stylesheet + `</style>`
		updated := strings.Replace(string(body), marker, injection+marker, 1)
		if updated == string(body) {
			t.Errorf("consumer fixture route did not contain %q", marker)
			_ = route.Abort()
			return
		}
		if !strings.Contains(updated, scrollRegionBFullConsumerAttribute+`="`+scrollRegionBFullConsumerAttributeValue+`"`) {
			t.Errorf("server-routed consumer response did not bind its explicit consumer identity into initial HTML")
			_ = route.Abort()
			return
		}
		if err := route.Fulfill(playwright.RouteFulfillOptions{Response: response, Body: []byte(updated)}); err != nil {
			t.Errorf("fulfill consumer fixture route: %v", err)
		}
	}))
}

func requireScrollRegionInitialHTML(t *testing.T, body []byte, theme scrollRegionBFullTheme) {
	t.Helper()
	for _, want := range []string{
		`id="scroll-region-fragment"`,
		`data-goshtoso-scroll-region`,
		`data-goshtoso-scroll-viewport`,
		`tabindex="0"`,
		`role="region"`,
		`aria-label="Activity history"`,
		`data-theme="` + theme.ServerTheme + `"`,
		`data-goshtoso-theme-initial-source="server-routed-html"`,
	} {
		require.Contains(t, string(body), want, "initial server HTML must already contain the public Scroll Region contract")
	}
	if theme.ConsumerCSS {
		require.Contains(t, string(body), `id="t-gs-011-consumer-scrollregion-css"`, "the intentional consumer stylesheet must be in initial HTML before first paint")
		require.Contains(t, string(body), scrollRegionBFullConsumerAttribute+`="`+scrollRegionBFullConsumerAttributeValue+`"`, "the server-routed consumer identity must be in initial HTML before first paint")
	}
}

func requireScrollRegionFirstPaint(t *testing.T, page playwright.Page, viewport playwright.Locator, theme scrollRegionBFullTheme, dark bool, width int, zoom scrollRegionBFullZoom) (scrollRegionBFullPaint, scrollRegionBFullPaint, scrollRegionBFullPaint, []scrollRegionBFullPaint) {
	t.Helper()
	state, err := viewport.Evaluate(`(el, expected) => {
		const root = document.documentElement;
		const box = el.getBoundingClientRect();
		const observer = window.__goshtosoTGS011FirstPaint;
		if (!observer || !Array.isArray(observer.events) || observer.events.length === 0) throw new Error("first-paint observer missing");
		return {
			events: observer.events,
			settled: {
				phase: "settled",
				readyState: document.readyState,
				theme: root.getAttribute("data-theme"),
				themeSource: root.getAttribute("data-goshtoso-theme-initial-source") || "",
				dark: root.classList.contains("dark"),
				visible: box.width > 0 && box.height > 0,
				role: el.getAttribute("role"),
				name: el.getAttribute("aria-label"),
			},
			consumer: root.getAttribute("`+scrollRegionBFullConsumerAttribute+`"),
			consumerSurface: getComputedStyle(root).getPropertyValue("--color-surface").trim(),
			viewportWidth: innerWidth,
			visualScale: visualViewport ? visualViewport.scale : null,
			dpr: devicePixelRatio,
		};
	}`, nil)
	require.NoError(t, err)
	values := state.(map[string]any)
	rawEvents := values["events"].([]any)
	require.NotEmpty(t, rawEvents)
	events := make([]scrollRegionBFullPaint, 0, len(rawEvents))
	for _, raw := range rawEvents {
		events = append(events, scrollRegionBFullPaintFromBrowser(t, raw))
	}
	initial := events[0]
	require.Equal(t, "init", initial.Phase, "first-paint observer must run before product bootstrap mutates root state")
	require.Equal(t, theme.ServerTheme, initial.Theme, "earliest observable document state must retain server-routed theme HTML")
	require.Equal(t, scrollRegionBFullThemeInitialSource, initial.ThemeSource, "earliest observable document state must expose source provenance")
	require.False(t, initial.Dark, "raw initial server HTML must precede product-owned dark-mode restoration")
	if dark {
		rootMutation := false
		for _, event := range events {
			if event.Phase == "root-mutation" && event.Dark {
				rootMutation = true
			}
		}
		require.True(t, rootMutation, "dark mode must be restored by a product-owned root mutation after server HTML")
	}
	var firstPaint scrollRegionBFullPaint
	for _, event := range events {
		if event.Phase == "first-animation-frame" {
			firstPaint = event
			break
		}
	}
	require.Equal(t, "first-animation-frame", firstPaint.Phase, "observer must retain the earliest animation-frame paint record")
	require.Equal(t, theme.ServerTheme, firstPaint.Theme)
	require.Equal(t, scrollRegionBFullThemeInitialSource, firstPaint.ThemeSource)
	require.Equal(t, dark, firstPaint.Dark, "fresh page must restore dark mode through candidate-owned bootstrap before first animation-frame paint")
	require.Equal(t, true, firstPaint.Visible)
	require.Equal(t, "region", firstPaint.Role)
	require.Equal(t, "Activity history", firstPaint.Name)
	settled := scrollRegionBFullPaintFromBrowser(t, values["settled"])
	require.NotEqual(t, "loading", settled.ReadyState)
	require.Equal(t, true, settled.Visible)
	require.Equal(t, "region", settled.Role)
	require.Equal(t, "Activity history", settled.Name)
	require.Equal(t, theme.ServerTheme, settled.Theme)
	require.Equal(t, dark, settled.Dark)
	require.Equal(t, scrollRegionBFullThemeInitialSource, settled.ThemeSource)
	if theme.ConsumerCSS {
		require.Equal(t, scrollRegionBFullConsumerAttributeValue, values["consumer"])
		require.NotEmpty(t, values["consumerSurface"], "consumer stylesheet must apply before the first route paint")
	} else {
		require.Empty(t, values["consumer"])
	}
	if zoom.RealUA {
		viewportWidth := scrollRegionBFullNumber(t, values["viewportWidth"])
		require.LessOrEqual(t, viewportWidth, float64(width)/float64(zoom.Factor)+1, "real UA zoom must reduce the available CSS viewport and force content expansion")
		require.EqualValues(t, 1, values["visualScale"], "UA zoom must not be mislabeled CDP page scale")
		dpr := scrollRegionBFullNumber(t, values["dpr"])
		require.GreaterOrEqual(t, dpr, float64(zoom.Factor)-0.01)
	}
	return initial, firstPaint, settled, events
}

func scrollRegionBFullPaintFromBrowser(t *testing.T, raw any) scrollRegionBFullPaint {
	t.Helper()
	values := raw.(map[string]any)
	paint := scrollRegionBFullPaint{
		Phase:       values["phase"].(string),
		ReadyState:  values["readyState"].(string),
		Theme:       values["theme"].(string),
		ThemeSource: values["themeSource"].(string),
		Dark:        values["dark"].(bool),
	}
	if visible, exists := values["visible"]; exists {
		paint.Visible = visible.(bool)
	}
	if role, exists := values["role"]; exists {
		paint.Role = role.(string)
	}
	if name, exists := values["name"]; exists {
		paint.Name = name.(string)
	}
	return paint
}

// requireScrollRegionVisualCapture establishes a DOM-to-page-surface capture
// plan. The later CDP screenshot must map each CSS anchor into exact PNG pixels
// before it can be recorded as named Scroll Region visual evidence.
func requireScrollRegionVisualCapture(t *testing.T, page playwright.Page, root, viewport playwright.Locator, allowInternalHorizontalRange bool) scrollRegionBFullCapture {
	t.Helper()
	require.NoError(t, root.ScrollIntoViewIfNeeded())
	state, err := root.Evaluate(`(el, allowInternalHorizontalRange) => {
		const viewport = el.querySelector('[data-goshtoso-scroll-viewport]');
		if (!viewport) throw new Error('named Scroll Region viewport is absent');
		const rect = el.getBoundingClientRect();
		const cards = Array.from(viewport.querySelectorAll('li'));
		if (cards.length < 2) throw new Error('named Scroll Region needs two rendered activity cards for pixel anchoring');
		const viewportRect = viewport.getBoundingClientRect();
		const visual = window.visualViewport;
		const left = visual ? visual.offsetLeft : 0;
		const top = visual ? visual.offsetTop : 0;
		const right = left + (visual ? visual.width : window.innerWidth);
		const bottom = top + (visual ? visual.height : window.innerHeight);
		const visualWidth = visual ? visual.width : window.innerWidth;
		const visualHeight = visual ? visual.height : window.innerHeight;
		const rgba = (element) => {
			const canvas = document.createElement('canvas');
			canvas.width = 1;
			canvas.height = 1;
			const context = canvas.getContext('2d', {willReadFrequently: true});
			context.fillStyle = getComputedStyle(element).backgroundColor;
			context.fillRect(0, 0, 1, 1);
			return Array.from(context.getImageData(0, 0, 1, 1).data);
		};
		const cardAnchor = (name, card) => {
			const cardRect = card.getBoundingClientRect();
			const x = Math.max(viewportRect.left + 6, Math.min(viewportRect.right - 6, cardRect.left + Math.min(Math.max(6, cardRect.width / 8), 10)));
			const y = cardRect.top + cardRect.height / 2;
			return {
				name,
				domText: card.textContent.trim(),
				cssX: x - left,
				cssY: y - top,
				rgba: rgba(card),
			};
		};
		const fullyVisibleCards = cards.filter((card) => {
			const cardRect = card.getBoundingClientRect();
			return cardRect.left >= viewportRect.left + 2 &&
				cardRect.right <= viewportRect.right - 2 &&
				cardRect.top >= viewportRect.top + 2 &&
				cardRect.bottom <= viewportRect.bottom - 2;
		});
		const anchorCards = allowInternalHorizontalRange ? cards.filter((card) => {
			const cardRect = card.getBoundingClientRect();
			return cardRect.right > viewportRect.left + 6 && cardRect.left < viewportRect.right - 6 && cardRect.top >= viewportRect.top + 2 && cardRect.bottom <= viewportRect.bottom - 2;
		}) : fullyVisibleCards;
		if (anchorCards.length === 0) throw new Error('named Scroll Region needs at least one visible Activity card for pixel anchoring');
		const endIndicator = el.querySelector('[data-goshtoso-scroll-end-indicator]');
		if (!endIndicator || endIndicator.hidden) throw new Error('named Scroll Region end cue must be visibly rendered at the initial scroll boundary');
		const endRect = endIndicator.getBoundingClientRect();
		if (endRect.left < rect.left || endRect.right > rect.right || endRect.top < rect.top || endRect.bottom > rect.bottom) throw new Error('named Scroll Region end cue must be inside the captured root');
		const inset = Math.max(2, Math.min(6, Math.floor(Math.min(rect.width, rect.height) / 12)));
		const points = [
			[rect.left + inset, rect.top + inset],
			[rect.right - inset, rect.top + inset],
			[rect.left + rect.width / 2, rect.top + rect.height / 2],
			[rect.left + inset, rect.bottom - inset],
			[rect.right - inset, rect.bottom - inset],
		];
		const occluders = points.flatMap(([x, y]) => {
			const hit = document.elementFromPoint(x, y);
			return hit && (hit === el || el.contains(hit)) ? [] : [hit ? hit.outerHTML.slice(0, 160) : 'no-hit'];
		});
			const scrolling = document.scrollingElement || document.documentElement;
			const originalViewportScrollLeft = viewport.scrollLeft;
			viewport.scrollLeft = originalViewportScrollLeft + 99999;
			const attemptedViewportScrollLeft = viewport.scrollLeft;
			viewport.scrollLeft = originalViewportScrollLeft;
		const originalScrollLeft = scrolling.scrollLeft;
		const originalVisualLeft = visual ? visual.pageLeft : window.scrollX;
		scrolling.scrollLeft = originalScrollLeft + 99999;
		const attemptedScrollLeft = scrolling.scrollLeft;
		const attemptedVisualLeft = visual ? visual.pageLeft : window.scrollX;
		scrolling.scrollLeft = originalScrollLeft;
		const padding = 6;
		const cropX = rect.left - left - padding;
		const cropY = rect.top - top - padding;
		return {
			visible: rect.width > 0 && rect.height > 0,
			insideVisualViewport: rect.left >= left + padding && rect.top >= top + padding && rect.right <= right - padding && rect.bottom <= bottom - padding,
			role: viewport.getAttribute('role'),
			name: viewport.getAttribute('aria-label') || viewport.getAttribute('aria-labelledby') || '',
			documentScrollingElement: scrolling.tagName.toLowerCase(),
			pageCanScrollHorizontally: Math.abs(attemptedScrollLeft - originalScrollLeft) > 1 || Math.abs(attemptedVisualLeft - originalVisualLeft) > 1,
				scrollingClientWidth: scrolling.clientWidth,
				scrollingScrollWidth: scrolling.scrollWidth,
				viewportClientWidth: viewport.clientWidth,
				viewportScrollWidth: viewport.scrollWidth,
				viewportOverflowX: getComputedStyle(viewport).overflowX,
				attemptedViewportScrollLeft,
			attemptedScrollLeft,
			occluders,
			visual: {left, top, width: visualWidth, height: visualHeight, devicePixelRatio: window.devicePixelRatio},
			crop: {x: cropX, y: cropY, width: rect.width + padding * 2, height: rect.height + padding * 2},
			anchors: anchorCards.slice(0, 2).map((card, index) => cardAnchor('visible-activity-card-' + (index + 1), card)),
			cue: {
				state: 'end',
				visible: true,
				cssX: endRect.left - left,
				cssY: endRect.top - top,
				cssWidth: endRect.width,
				cssHeight: endRect.height,
			},
		};
	}`, allowInternalHorizontalRange)
	require.NoError(t, err)
	values, ok := state.(map[string]any)
	require.True(t, ok, "unexpected visual-capture state %T", state)
	require.Equal(t, true, values["visible"])
	require.Equal(t, true, values["insideVisualViewport"], "Scroll Region must be fully inside the effective visual viewport before capture: %#v", values)
	require.Equal(t, "region", values["role"])
	require.NotEmpty(t, values["name"])
	require.Equal(t, false, values["pageCanScrollHorizontally"], "the actual document scrolling element and visual viewport must reject horizontal scroll at the visual evidence cell: %#v", values)
	require.NoError(t, validateScrollRegionBFullHorizontalContract(
		allowInternalHorizontalRange,
		scrollRegionBFullNumber(t, values["viewportClientWidth"]),
		scrollRegionBFullNumber(t, values["viewportScrollWidth"]),
		scrollRegionBFullNumber(t, values["attemptedViewportScrollLeft"]),
	), "Scroll Region horizontal contract: %#v", values)
	require.Equal(t, "auto", values["viewportOverflowX"])
	occluders, ok := values["occluders"].([]any)
	require.True(t, ok)
	require.Empty(t, occluders, "visual capture must not be covered by unrelated page content: %#v", values)
	visual, ok := values["visual"].(map[string]any)
	require.True(t, ok)
	clip, ok := values["crop"].(map[string]any)
	require.True(t, ok)
	anchors, ok := values["anchors"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, anchors, "capture must bind at least one visible Activity card")
	cue, ok := values["cue"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "end", cue["state"])
	require.Equal(t, true, cue["visible"])
	capture := scrollRegionBFullCapture{
		ExpectedWidth:  scrollRegionBFullNumber(t, clip["width"]),
		ExpectedHeight: scrollRegionBFullNumber(t, clip["height"]),
		Proof: scrollRegionBFullCaptureProof{
			Method:               "cdp-page-capture-screenshot",
			VisualViewportLeft:   scrollRegionBFullNumber(t, visual["left"]),
			VisualViewportTop:    scrollRegionBFullNumber(t, visual["top"]),
			VisualViewportWidth:  scrollRegionBFullNumber(t, visual["width"]),
			VisualViewportHeight: scrollRegionBFullNumber(t, visual["height"]),
			DevicePixelRatio:     scrollRegionBFullNumber(t, visual["devicePixelRatio"]),
			CropCSSX:             scrollRegionBFullNumber(t, clip["x"]),
			CropCSSY:             scrollRegionBFullNumber(t, clip["y"]),
			CropCSSWidth:         scrollRegionBFullNumber(t, clip["width"]),
			CropCSSHeight:        scrollRegionBFullNumber(t, clip["height"]),
		},
	}
	for _, raw := range anchors {
		anchorValues, ok := raw.(map[string]any)
		require.True(t, ok)
		rawRGBA, ok := anchorValues["rgba"].([]any)
		require.True(t, ok)
		require.Len(t, rawRGBA, 4)
		anchor := scrollRegionBFullPixelAnchor{
			Name:      anchorValues["name"].(string),
			DOMText:   anchorValues["domText"].(string),
			CSSX:      scrollRegionBFullNumber(t, anchorValues["cssX"]),
			CSSY:      scrollRegionBFullNumber(t, anchorValues["cssY"]),
			Tolerance: 16,
		}
		for index := range anchor.Expected {
			anchor.Expected[index] = uint8(scrollRegionBFullNumber(t, rawRGBA[index]))
		}
		require.Contains(t, anchor.DOMText, "Activity", "pixel anchor must bind visible named-region content")
		capture.Proof.Anchors = append(capture.Proof.Anchors, anchor)
	}
	capture.Proof.BoundaryCue = scrollRegionBFullBoundaryCue{
		State:     cue["state"].(string),
		Visible:   cue["visible"].(bool),
		CSSX:      scrollRegionBFullNumber(t, cue["cssX"]),
		CSSY:      scrollRegionBFullNumber(t, cue["cssY"]),
		CSSWidth:  scrollRegionBFullNumber(t, cue["cssWidth"]),
		CSSHeight: scrollRegionBFullNumber(t, cue["cssHeight"]),
	}
	require.Greater(t, capture.Proof.BoundaryCue.CSSWidth, 0.0)
	require.Greater(t, capture.Proof.BoundaryCue.CSSHeight, 0.0)
	require.GreaterOrEqual(t, capture.Proof.CropCSSX, 0.0, "capture padding must remain within the visual viewport")
	require.GreaterOrEqual(t, capture.Proof.CropCSSY, 0.0, "capture padding must remain within the visual viewport")
	require.LessOrEqual(t, capture.Proof.CropCSSX+capture.Proof.CropCSSWidth, capture.Proof.VisualViewportWidth)
	require.LessOrEqual(t, capture.Proof.CropCSSY+capture.Proof.CropCSSHeight, capture.Proof.VisualViewportHeight)
	return capture
}

// validateScrollRegionBFullHorizontalContract distinguishes a page-level
// regression from the maintained consumer's intentional internal range.
// Non-consumer documentation states must remain vertically scroll-only.
func validateScrollRegionBFullHorizontalContract(allowInternalRange bool, clientWidth, scrollWidth, attemptedScrollLeft float64) error {
	hasRange := scrollWidth > clientWidth+1 && attemptedScrollLeft > 0
	if allowInternalRange {
		if !hasRange {
			return fmt.Errorf("wide consumer Scroll Region must retain an intentional internal horizontal range")
		}
		return nil
	}
	if hasRange || scrollWidth > clientWidth+1 || attemptedScrollLeft > 0 {
		return fmt.Errorf("non-wide Scroll Region must not acquire an internal horizontal range")
	}
	return nil
}

func scrollRegionBFullNumber(t *testing.T, value any) float64 {
	t.Helper()
	switch number := value.(type) {
	case int:
		return float64(number)
	case int64:
		return float64(number)
	case float64:
		return number
	case float32:
		return float64(number)
	default:
		t.Fatalf("expected browser number, got %T (%#v)", value, value)
		return 0
	}
}

func requireScrollRegionAccessibleNames(t *testing.T, page playwright.Page) {
	t.Helper()
	actual, err := page.Locator("#scroll-region-fragment [data-goshtoso-scroll-viewport][role='region']").EvaluateAll(`els => els.map(el => ({
		name: el.getAttribute("aria-label") || el.getAttribute("aria-labelledby") || "",
		tabIndex: el.getAttribute("tabindex"),
	}))`, nil)
	require.NoError(t, err)
	values, ok := actual.([]any)
	require.True(t, ok)
	require.Len(t, values, 3)
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := raw.(map[string]any)
		name := value["name"].(string)
		require.NotEmpty(t, name)
		require.Equal(t, "0", value["tabIndex"])
		_, duplicate := seen[name]
		require.Falsef(t, duplicate, "every maintained demo region must have a unique landmark name: %q", name)
		seen[name] = struct{}{}
	}
}

func requireScrollRegionNoOverflow(t *testing.T, page playwright.Page) {
	t.Helper()
	viewport := page.Locator("#scroll-region-no-overflow [data-goshtoso-scroll-viewport]")
	require.NoError(t, viewport.Focus())
	box, err := viewport.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, box)
	require.NoError(t, page.Mouse().Move(box.X+box.Width/2, box.Y+box.Height/2))
	require.NoError(t, page.Mouse().Wheel(0, 120))
	require.NoError(t, viewport.Press("End"))
	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Detach() })
	_, err = session.Send("Input.synthesizeScrollGesture", map[string]any{
		"x":                 int(box.X + box.Width/2),
		"y":                 int(box.Y + box.Height/2),
		"yDistance":         -120,
		"speed":             800,
		"gestureSourceType": "touch",
	})
	require.NoError(t, err)
	state, err := viewport.Evaluate(`el => {
		const root = el.closest("[data-goshtoso-scroll-region]");
		return {
			scrollTop: el.scrollTop,
			clientHeight: el.clientHeight,
			scrollHeight: el.scrollHeight,
			clientWidth: el.clientWidth,
			scrollWidth: el.scrollWidth,
			overflowsHorizontally: el.scrollWidth > el.clientWidth + 1,
			overflowX: getComputedStyle(el).overflowX,
			overflows: el.scrollHeight > el.clientHeight + 1,
			startHidden: root.querySelector("[data-goshtoso-scroll-start-indicator]").hidden,
			endHidden: root.querySelector("[data-goshtoso-scroll-end-indicator]").hidden,
		};
	}`, nil)
	require.NoError(t, err)
	values := state.(map[string]any)
	require.Equalf(t, false, values["overflows"], "the no-overflow public state must remain non-scrollable at this B-FULL cell: %#v", values)
	require.Equalf(t, false, values["overflowsHorizontally"], "the no-overflow public state must not gain a horizontal scroll axis: %#v", values)
	require.Equal(t, "auto", values["overflowX"])
	require.EqualValues(t, 0, values["scrollTop"])
	require.Equal(t, true, values["startHidden"])
	require.Equal(t, true, values["endHidden"])
}

func requireScrollRegionNativeInputs(t *testing.T, page playwright.Page, viewport playwright.Locator) {
	t.Helper()
	require.NoError(t, viewport.ScrollIntoViewIfNeeded())
	require.NoError(t, viewport.Focus())
	requireScrollRegionFocusIndicator(t, viewport)
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "start")

	// Keyboard owns the complete start -> middle -> end boundary path. The
	// subsequent real navigation deliberately returns to the source-owned start
	// state; no synthetic scrollTop assignment stands in for a user outcome.
	require.NoError(t, viewport.Press("PageDown"))
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "middle")
	require.NoError(t, viewport.Press("End"))
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "end")
	scrollRegionBFullReturnToStart(t, page)

	// Mouse runs from a fresh, observable start boundary and returns through a
	// fresh navigation before touch begins. This prevents an edge-layout page
	// scroll from being misreported as a scroll-region mouse outcome.
	require.NoError(t, viewport.ScrollIntoViewIfNeeded())
	require.NoError(t, viewport.Hover())
	box, err := viewport.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, box)
	require.NoError(t, page.Mouse().Wheel(0, 180))
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "middle")
	scrollRegionBFullReturnToStart(t, page)

	// CDP synthesizes a native touch scroll gesture rather than dispatching a
	// DOM event. Its before/action/outcome begins from another fresh start.
	require.NoError(t, viewport.ScrollIntoViewIfNeeded())
	box, err = viewport.BoundingBox()
	require.NoError(t, err)
	require.NotNil(t, box)
	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Detach() })
	_, err = session.Send("Input.synthesizeScrollGesture", map[string]any{
		"x":                 int(box.X + box.Width/2),
		"y":                 int(box.Y + box.Height/2),
		"yDistance":         -120,
		"speed":             800,
		"gestureSourceType": "touch",
	})
	require.NoError(t, err)
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "middle")

	require.NoError(t, viewport.Press("Escape"))
	state, err := viewport.Evaluate(`el => ({
		modal: Boolean(document.querySelector("#scroll-region-fragment [role=dialog][aria-modal=true]")),
	})`, nil)
	require.NoError(t, err)
	values := state.(map[string]any)
	require.Equal(t, false, values["modal"], "Escape is source-grounded N/A: ScrollRegion owns no dismissible surface")
}

func scrollRegionBFullReturnToStart(t *testing.T, page playwright.Page) {
	t.Helper()
	response, err := page.Reload(playwright.PageReloadOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, 200, response.Status())
	root := page.Locator("#scroll-region-default [data-goshtoso-scroll-region]")
	require.NoError(t, root.WaitFor())
	waitForScrollRegionPosition(t, page, "#scroll-region-default", "start")
}

func requireScrollRegionFocusIndicator(t *testing.T, viewport playwright.Locator) {
	t.Helper()
	state, err := viewport.Evaluate(`el => {
		const canvas = document.createElement("canvas");
		canvas.width = canvas.height = 1;
		const context = canvas.getContext("2d", {willReadFrequently: true});
		const rgba = value => {
			context.clearRect(0, 0, 1, 1);
			context.fillStyle = value;
			context.fillRect(0, 0, 1, 1);
			return [...context.getImageData(0, 0, 1, 1).data];
		};
		const luminance = rgb => rgb.slice(0, 3).map(channel => {
			const linear = channel / 255;
			return linear <= 0.04045 ? linear / 12.92 : ((linear + 0.055) / 1.055) ** 2.4;
		}).reduce((sum, channel, index) => sum + channel * [0.2126, 0.7152, 0.0722][index], 0);
		const composite = (front, back) => {
			const alpha = front[3] / 255 + (back[3] / 255) * (1 - front[3] / 255);
			if (alpha === 0) return [0, 0, 0, 0];
			return [
				(front[0] * (front[3] / 255) + back[0] * (back[3] / 255) * (1 - front[3] / 255)) / alpha,
				(front[1] * (front[3] / 255) + back[1] * (back[3] / 255) * (1 - front[3] / 255)) / alpha,
				(front[2] * (front[3] / 255) + back[2] * (back[3] / 255) * (1 - front[3] / 255)) / alpha,
				alpha * 255,
			];
		};
		const exterior = node => {
			const layers = [];
			for (let current = node; current; current = current.parentElement) layers.push(rgba(getComputedStyle(current).backgroundColor));
			let background = [255, 255, 255, 255];
			for (let index = layers.length - 1; index >= 0; index--) background = composite(layers[index], background);
			return background;
		};
		const style = getComputedStyle(el);
		const offset = Number.parseFloat(style.outlineOffset) || 0;
		const width = Number.parseFloat(style.outlineWidth) || 0;
		const outline = rgba(style.outlineColor);
		const adjacent = exterior(el.parentElement);
		const renderedOutline = composite(outline, adjacent);
		const contrast = (Math.max(luminance(renderedOutline), luminance(adjacent)) + .05) / (Math.min(luminance(renderedOutline), luminance(adjacent)) + .05);
		return {
			active: document.activeElement === el,
			style: style.outlineStyle,
			width: style.outlineWidth,
			offset: style.outlineOffset,
			outside: offset > 0 && width > 0,
			contrast,
			adjacentAlpha: adjacent[3],
			adjacent: adjacent.map(value => Math.round(value)),
			renderedOutline: renderedOutline.map(value => Math.round(value)),
		};
	}`, nil)
	require.NoError(t, err)
	values := state.(map[string]any)
	require.Equal(t, true, values["active"])
	require.NotEqual(t, "none", values["style"])
	require.NotEqual(t, "0px", values["width"])
	require.Equal(t, true, values["outside"], "Scroll Region uses an outside focus outline; contrast must be measured against its parent surface")
	require.EqualValues(t, 255, values["adjacentAlpha"])
	require.GreaterOrEqual(t, values["contrast"].(float64), 3.0, "focus indicator must meet non-text contrast against its actual outside adjacent surface: %#v", values)
}

func requireScrollRegionReducedMotion(t *testing.T, page playwright.Page, viewport playwright.Locator) {
	t.Helper()
	state, err := viewport.Evaluate(`el => {
		const root = el.closest("[data-goshtoso-scroll-region]");
		const values = [root, el, ...root.querySelectorAll("[data-goshtoso-scroll-start-indicator], [data-goshtoso-scroll-end-indicator]")].map(node => {
			const style = getComputedStyle(node);
			return {animationName: style.animationName, animationDuration: style.animationDuration, transitionDuration: style.transitionDuration};
		});
		return {reduced: matchMedia("(prefers-reduced-motion: reduce)").matches, values};
	}`, nil)
	require.NoError(t, err)
	values := state.(map[string]any)
	require.Equal(t, true, values["reduced"])
	for _, raw := range values["values"].([]any) {
		style := raw.(map[string]any)
		require.Equal(t, "none", style["animationName"], "ScrollRegion state updates must not animate under reduced motion")
		require.Equal(t, "0s", style["animationDuration"])
		require.Equal(t, "0s", style["transitionDuration"])
	}
}

func requireScrollRegionAxeContract(t *testing.T, page playwright.Page, axe scrollRegionAxeCore) {
	t.Helper()
	require.NoError(t, injectScrollRegionAxe(page, axe.Source))
	result := runScrollRegionAxe(t, page)
	require.Equal(t, "axe-core", result.TestEngine.Name)
	require.Equal(t, scrollRegionAxeCoreVersion, result.TestEngine.Version)
	require.Emptyf(t, result.Violations, "axe-core violations on public Scroll Region examples: %s", describeScrollRegionAxeRules(result.Violations))
	requireScrollRegionAxeCompleteness(t, result)
	requireScrollRegionAxeRulePasses(t, result.Passes,
		"aria-prohibited-attr",
		"aria-required-attr",
		"aria-roles",
		"aria-valid-attr",
		"aria-valid-attr-value",
		"landmark-unique",
		"scrollable-region-focusable",
	)
	for _, sample := range scrollRegionContrastSamples(t, page) {
		require.Equalf(t, 255, sample.BackgroundAlpha, "text %q must have an opaque rendered background: %#v", sample.Text, sample)
		require.GreaterOrEqualf(t, sample.Contrast, 4.5, "text %q must meet normal-text contrast: %#v", sample.Text, sample)
	}
}

func newScrollRegionBFullRecorder(t *testing.T, widths []int, binding scrollRegionBFullIdentityBinding, plan scrollRegionBFullRunPlan) *scrollRegionBFullRecorder {
	t.Helper()
	directory := os.Getenv("GOSHTOSO_SCROLLREGION_BFULL_ARTIFACT_DIR")
	if directory == "" {
		directory = t.TempDir()
	}
	require.NoError(t, os.MkdirAll(directory, 0o755))
	return &scrollRegionBFullRecorder{
		directory: directory,
		widths:    widths,
		plan:      plan,
		traces:    make(map[string]scrollRegionBFullArtifact),
		binding:   binding,
	}
}

// jsonArtifact persists one direct browser observation before it is folded into
// the receipt. The receipt wrapper binds the exact bytes by SHA-256; semantic
// validation later decodes these bytes instead of trusting the wrapper fields.
func (recorder *scrollRegionBFullRecorder) jsonArtifact(t *testing.T, cellID, phase string, value any) scrollRegionBFullArtifact {
	t.Helper()
	content, err := json.Marshal(value)
	require.NoError(t, err)
	path := filepath.Join(recorder.directory, "scrollregion-bfull-"+scrollRegionBFullArtifactName(cellID)+"-"+phase+".json")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	require.NoError(t, err, "raw evidence artifact names must not overwrite prior bytes")
	_, err = file.Write(content)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return scrollRegionBFullArtifact{Path: path, SHA256: scrollRegionBFullSHA256(content)}
}

func (recorder *scrollRegionBFullRecorder) bytesArtifact(t *testing.T, cellID, phase string, content []byte) scrollRegionBFullArtifact {
	t.Helper()
	path := filepath.Join(recorder.directory, "scrollregion-bfull-"+scrollRegionBFullArtifactName(cellID)+"-"+phase+".html")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	require.NoError(t, err, "raw evidence artifact names must not overwrite prior bytes")
	_, err = file.Write(content)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return scrollRegionBFullArtifact{Path: path, SHA256: scrollRegionBFullSHA256(content)}
}

func (recorder *scrollRegionBFullRecorder) screenshot(t *testing.T, page playwright.Page, capture scrollRegionBFullCapture, theme scrollRegionBFullTheme, dark bool, width int, zoom scrollRegionBFullZoom, phase string) scrollRegionBFullArtifact {
	t.Helper()
	name := fmt.Sprintf("%s-%s-%d-%s-%s.png", theme.ID, scrollRegionBFullMode(dark), width, zoom.ID, phase)
	path := filepath.Join(recorder.directory, name)
	pageSurface := scrollRegionBFullCDPPageSurface(t, page)
	decoded, format, err := image.Decode(bytes.NewReader(pageSurface))
	require.NoError(t, err)
	require.Equal(t, "png", format)
	sourceBounds := decoded.Bounds()
	require.Greater(t, sourceBounds.Dx(), 0)
	require.Greater(t, sourceBounds.Dy(), 0)
	require.Greater(t, capture.Proof.VisualViewportWidth, 0.0)
	require.Greater(t, capture.Proof.VisualViewportHeight, 0.0)

	capture.Proof.SourceWidth = sourceBounds.Dx()
	capture.Proof.SourceHeight = sourceBounds.Dy()
	capture.Proof.ScaleX = float64(sourceBounds.Dx()) / capture.Proof.VisualViewportWidth
	capture.Proof.ScaleY = float64(sourceBounds.Dy()) / capture.Proof.VisualViewportHeight
	require.InDelta(t, capture.Proof.ScaleX, capture.Proof.ScaleY, 0.02, "CDP page-surface PNG must map consistently to the browser visual viewport")
	require.Greater(t, capture.Proof.ScaleX, 0.0)
	require.Greater(t, capture.Proof.ScaleY, 0.0)

	capture.Proof.CropPixelX = scrollRegionBFullRoundedPixel(capture.Proof.CropCSSX * capture.Proof.ScaleX)
	capture.Proof.CropPixelY = scrollRegionBFullRoundedPixel(capture.Proof.CropCSSY * capture.Proof.ScaleY)
	capture.Proof.CropPixelWidth = scrollRegionBFullRoundedPixel(capture.Proof.CropCSSWidth * capture.Proof.ScaleX)
	capture.Proof.CropPixelHeight = scrollRegionBFullRoundedPixel(capture.Proof.CropCSSHeight * capture.Proof.ScaleY)
	require.Greater(t, capture.Proof.CropPixelWidth, 100)
	require.Greater(t, capture.Proof.CropPixelHeight, 80)
	cropBounds := image.Rect(
		capture.Proof.CropPixelX,
		capture.Proof.CropPixelY,
		capture.Proof.CropPixelX+capture.Proof.CropPixelWidth,
		capture.Proof.CropPixelY+capture.Proof.CropPixelHeight,
	)
	require.Truef(t, cropBounds.In(sourceBounds), "named Scroll Region crop must remain within the CDP page surface: crop=%v source=%v proof=%#v", cropBounds, sourceBounds, capture.Proof)
	for index := range capture.Proof.Anchors {
		anchor := &capture.Proof.Anchors[index]
		anchor.PixelX = scrollRegionBFullRoundedPixel((anchor.CSSX - capture.Proof.CropCSSX) * capture.Proof.ScaleX)
		anchor.PixelY = scrollRegionBFullRoundedPixel((anchor.CSSY - capture.Proof.CropCSSY) * capture.Proof.ScaleY)
		require.GreaterOrEqualf(t, anchor.PixelX, 0, "named-region anchor %q must be inside crop", anchor.Name)
		require.GreaterOrEqualf(t, anchor.PixelY, 0, "named-region anchor %q must be inside crop", anchor.Name)
		require.Lessf(t, anchor.PixelX, capture.Proof.CropPixelWidth, "named-region anchor %q must be inside crop", anchor.Name)
		require.Lessf(t, anchor.PixelY, capture.Proof.CropPixelHeight, "named-region anchor %q must be inside crop", anchor.Name)
	}
	cue := &capture.Proof.BoundaryCue
	require.Equal(t, "end", cue.State, "initial visual capture must bind the rendered end-boundary cue")
	require.True(t, cue.Visible, "initial visual capture must bind a visible end-boundary cue")
	cue.PixelX = scrollRegionBFullRoundedPixel((cue.CSSX - capture.Proof.CropCSSX) * capture.Proof.ScaleX)
	cue.PixelY = scrollRegionBFullRoundedPixel((cue.CSSY - capture.Proof.CropCSSY) * capture.Proof.ScaleY)
	cue.PixelWidth = scrollRegionBFullRoundedPixel(cue.CSSWidth * capture.Proof.ScaleX)
	cue.PixelHeight = scrollRegionBFullRoundedPixel(cue.CSSHeight * capture.Proof.ScaleY)
	require.Greater(t, cue.PixelWidth, 0, "named-region boundary cue must retain pixel width")
	require.Greater(t, cue.PixelHeight, 0, "named-region boundary cue must retain pixel height")
	require.GreaterOrEqual(t, cue.PixelX, 0, "named-region boundary cue must be inside crop")
	require.GreaterOrEqual(t, cue.PixelY, 0, "named-region boundary cue must be inside crop")
	require.LessOrEqual(t, cue.PixelX+cue.PixelWidth, capture.Proof.CropPixelWidth, "named-region boundary cue must be inside crop")
	require.LessOrEqual(t, cue.PixelY+cue.PixelHeight, capture.Proof.CropPixelHeight, "named-region boundary cue must be inside crop")

	cropped := image.NewRGBA(image.Rect(0, 0, cropBounds.Dx(), cropBounds.Dy()))
	draw.Draw(cropped, cropped.Bounds(), decoded, cropBounds.Min, draw.Src)
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, cropped))
	require.NoError(t, os.WriteFile(path, encoded.Bytes(), 0o644))
	return scrollRegionBFullScreenshotArtifact(t, path, capture)
}

func scrollRegionBFullCDPPageSurface(t *testing.T, page playwright.Page) []byte {
	t.Helper()
	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err, "named-region visual evidence requires Chromium CDP page-surface capture")
	t.Cleanup(func() { _ = session.Detach() })
	response, err := session.Send("Page.captureScreenshot", map[string]any{
		"format":      "png",
		"fromSurface": true,
	})
	require.NoError(t, err)
	values, ok := response.(map[string]any)
	require.True(t, ok, "unexpected CDP Page.captureScreenshot response %T", response)
	data, ok := values["data"].(string)
	require.True(t, ok)
	content, err := base64.StdEncoding.DecodeString(data)
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(content, []byte("\x89PNG\r\n\x1a\n")), "CDP Page.captureScreenshot must return PNG page pixels")
	return content
}

func scrollRegionBFullRoundedPixel(value float64) int {
	return int(math.Round(value))
}

func scrollRegionBFullScreenshotArtifact(t *testing.T, path string, capture scrollRegionBFullCapture) scrollRegionBFullArtifact {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Truef(t, bytes.HasPrefix(content, []byte("\x89PNG\r\n\x1a\n")), "screenshot %s has wrong PNG magic", path)
	decoded, format, err := image.Decode(bytes.NewReader(content))
	require.NoError(t, err)
	require.Equal(t, "png", format)
	bounds := decoded.Bounds()
	require.GreaterOrEqual(t, bounds.Dx(), int(math.Floor(capture.ExpectedWidth))-1, "screenshot must include the complete bounded Scroll Region width")
	require.GreaterOrEqual(t, bounds.Dy(), int(math.Floor(capture.ExpectedHeight))-1, "screenshot must include the complete bounded Scroll Region height")
	require.GreaterOrEqual(t, bounds.Dx(), 100)
	require.GreaterOrEqual(t, bounds.Dy(), 100)
	require.Equal(t, "cdp-page-capture-screenshot", capture.Proof.Method)
	require.Equal(t, capture.Proof.CropPixelWidth, bounds.Dx(), "PNG width must equal mapped CDP crop width")
	require.Equal(t, capture.Proof.CropPixelHeight, bounds.Dy(), "PNG height must equal mapped CDP crop height")
	require.NoError(t, validateScrollRegionBFullCapturedRegion(decoded, capture))
	return scrollRegionBFullArtifact{
		Path:           path,
		SHA256:         scrollRegionBFullSHA256(content),
		Width:          bounds.Dx(),
		Height:         bounds.Dy(),
		CapturedRegion: "named-scroll-region",
		Capture:        &capture.Proof,
	}
}

func validateScrollRegionBFullCapturedRegion(decoded image.Image, capture scrollRegionBFullCapture) error {
	bounds := decoded.Bounds()
	if capture.ExpectedWidth > 0 && math.Abs(float64(bounds.Dx())-capture.ExpectedWidth) > 1 && capture.Proof.CropPixelWidth == 0 {
		return fmt.Errorf("named-region PNG width %d does not match expected %.2f", bounds.Dx(), capture.ExpectedWidth)
	}
	if capture.ExpectedHeight > 0 && math.Abs(float64(bounds.Dy())-capture.ExpectedHeight) > 1 && capture.Proof.CropPixelHeight == 0 {
		return fmt.Errorf("named-region PNG height %d does not match expected %.2f", bounds.Dy(), capture.ExpectedHeight)
	}
	if len(capture.Proof.Anchors) == 0 {
		return fmt.Errorf("named-region anchor proof is absent")
	}
	cue := capture.Proof.BoundaryCue
	if cue.State != "end" || !cue.Visible {
		return fmt.Errorf("named-region visible end-boundary cue proof is absent")
	}
	if cue.PixelWidth <= 0 || cue.PixelHeight <= 0 || cue.PixelX < bounds.Min.X || cue.PixelY < bounds.Min.Y || cue.PixelX+cue.PixelWidth > bounds.Max.X || cue.PixelY+cue.PixelHeight > bounds.Max.Y {
		return fmt.Errorf("named-region end-boundary cue lies outside PNG bounds")
	}
	for _, anchor := range capture.Proof.Anchors {
		if anchor.PixelX < bounds.Min.X || anchor.PixelX >= bounds.Max.X || anchor.PixelY < bounds.Min.Y || anchor.PixelY >= bounds.Max.Y {
			return fmt.Errorf("named-region anchor %q lies outside PNG bounds", anchor.Name)
		}
		actual := scrollRegionBFullRGBA8(decoded.At(anchor.PixelX, anchor.PixelY))
		for index := range actual {
			difference := int(actual[index]) - int(anchor.Expected[index])
			if difference < 0 {
				difference = -difference
			}
			if difference > int(anchor.Tolerance) {
				return fmt.Errorf("named-region anchor %q pixel mismatch: got %#v, want %#v within %d", anchor.Name, actual, anchor.Expected, anchor.Tolerance)
			}
		}
	}
	return nil
}

func scrollRegionBFullRGBA8(color color.Color) [4]uint8 {
	red, green, blue, alpha := color.RGBA()
	return [4]uint8{uint8(red >> 8), uint8(green >> 8), uint8(blue >> 8), uint8(alpha >> 8)}
}

func scrollRegionBFullArtifactForFile(t *testing.T, path string, magic []byte) scrollRegionBFullArtifact {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Truef(t, bytes.HasPrefix(content, magic), "artifact %s has wrong format magic", path)
	return scrollRegionBFullArtifact{Path: path, SHA256: scrollRegionBFullSHA256(content)}
}

func scrollRegionBFullSHA256(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func (recorder *scrollRegionBFullRecorder) write(t *testing.T, repositoryRoot string) {
	t.Helper()
	require.Len(t, recorder.cells, recorder.plan.ExpectedCells, "receipt must record exact declared cell count")
	cellIDs := make(map[string]struct{}, len(recorder.cells))
	for _, cell := range recorder.cells {
		require.NotEmpty(t, cell.CellID)
		_, duplicate := cellIDs[cell.CellID]
		require.Falsef(t, duplicate, "duplicate B-FULL cell receipt %q", cell.CellID)
		cellIDs[cell.CellID] = struct{}{}
	}
	require.Len(t, recorder.traces, recorder.plan.ExpectedCells, "every isolated B-FULL cell must emit its own authenticated trace artifact")
	if recorder.plan.Diagnostic {
		require.Equal(t, "diagnostic-unbound-dirty-worktree", recorder.binding.Binding)
		require.Nil(t, recorder.binding.Identity, "diagnostic receipt must make no final source identity claim")
		require.Less(t, recorder.plan.ExpectedCells, recorder.plan.FullExpectedCells)
	} else {
		require.NotNil(t, recorder.binding.Identity, "literal full-coverage receipt requires a sealed identity")
		require.Equal(t, recorder.plan.FullExpectedCells, recorder.plan.ExpectedCells)
		require.NoError(t, recorder.binding.revalidate(repositoryRoot), "source identity must remain exact until receipt bytes are written")
	}
	require.NotEmpty(t, recorder.tools.PlaywrightGo, "receipt must record source-derived Playwright Go pin")
	receipt := scrollRegionBFullReceipt{
		Schema:          scrollRegionBFullReceiptSchema,
		Closure:         scrollRegionBFullPlanClosure(recorder.plan),
		ProvenanceClass: scrollRegionBFullProvenanceStructural,
		ExpectedCells:   recorder.plan.ExpectedCells,
		Binding:         recorder.binding,
		ToolVersions:    recorder.tools,
		Widths:          recorder.widths,
		Cells:           recorder.cells,
		TraceByCell:     recorder.traces,
	}
	digest, err := scrollRegionBFullReceiptDigest(receipt)
	require.NoError(t, err)
	receipt.WrapperSHA256 = digest
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	require.NoError(t, err)
	validated, err := validateScrollRegionBFullReceiptWrapper(encoded)
	require.NoError(t, err)
	require.Equal(t, receipt, validated, "receipt wrapper verification must bind every serialized field before it is written")
	path := filepath.Join(recorder.directory, "scrollregion-bfull-receipt.json")
	require.NoError(t, os.WriteFile(path, append(encoded, '\n'), 0o644))
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	_, err = validateScrollRegionBFullReceiptWrapper(written)
	require.NoError(t, err)
}

func scrollRegionBFullReceiptDigest(receipt scrollRegionBFullReceipt) (string, error) {
	receipt.WrapperSHA256 = ""
	canonical, err := json.Marshal(receipt)
	if err != nil {
		return "", err
	}
	return scrollRegionBFullSHA256(canonical), nil
}

// validateScrollRegionBFullReceiptWrapper checks wrapper integrity and
// structural consistency, not independent browser provenance. The current
// trace and PNG parser can reject malformed or cross-cell artifacts, but a
// claimant can reproduce structurally valid bytes; only non-closure receipts
// are accepted until independent attestation exists.
func validateScrollRegionBFullReceiptWrapper(content []byte) (scrollRegionBFullReceipt, error) {
	var receipt scrollRegionBFullReceipt
	if err := scrollRegionDecodeStrictJSON(content, &receipt); err != nil {
		return scrollRegionBFullReceipt{}, fmt.Errorf("decode B-FULL receipt wrapper: %w", err)
	}
	if !scrollRegionSHA256Pattern.MatchString(receipt.WrapperSHA256) {
		return scrollRegionBFullReceipt{}, fmt.Errorf("B-FULL receipt wrapper SHA-256 is invalid")
	}
	want, err := scrollRegionBFullReceiptDigest(receipt)
	if err != nil {
		return scrollRegionBFullReceipt{}, fmt.Errorf("canonicalize B-FULL receipt wrapper: %w", err)
	}
	if receipt.WrapperSHA256 != want {
		return scrollRegionBFullReceipt{}, fmt.Errorf("B-FULL receipt wrapper SHA-256 mismatch")
	}
	if receipt.Schema != scrollRegionBFullReceiptSchema {
		return scrollRegionBFullReceipt{}, fmt.Errorf("B-FULL receipt schema %q is unsupported", receipt.Schema)
	}
	if receipt.ProvenanceClass != scrollRegionBFullProvenanceStructural {
		return scrollRegionBFullReceipt{}, fmt.Errorf("B-FULL receipt provenance %q is unsupported; current Playwright trace/PNG evidence is structural-unattested", receipt.ProvenanceClass)
	}
	if err := validateScrollRegionBFullThemeEvidence(receipt); err != nil {
		return scrollRegionBFullReceipt{}, err
	}
	if err := validateScrollRegionBFullReceiptAxes(receipt); err != nil {
		return scrollRegionBFullReceipt{}, err
	}
	if err := validateScrollRegionBFullRawPersistenceEvidence(receipt); err != nil {
		return scrollRegionBFullReceipt{}, err
	}
	return receipt, nil
}

// validateScrollRegionBFullThemeEvidence keeps the locked site boundary in
// the authenticated wrapper. A receipt cannot turn a server-routed theme axis
// into a fictional client persistence claim merely by changing labels.
func validateScrollRegionBFullThemeEvidence(receipt scrollRegionBFullReceipt) error {
	for _, cell := range receipt.Cells {
		proof := cell.Persistence
		if proof.ThemeInitialSource != scrollRegionBFullThemeInitialSource {
			return fmt.Errorf("B-FULL cell %q theme initial source must be %q", cell.CellID, scrollRegionBFullThemeInitialSource)
		}
		if proof.ThemePersistenceNotApplicable != scrollRegionBFullThemePersistenceNA || cell.NAs["theme-persistence"] != scrollRegionBFullThemePersistenceNA {
			return fmt.Errorf("B-FULL cell %q theme persistence must be source-grounded not-applicable for locked component docs", cell.CellID)
		}
		if proof.DarkPersistence != scrollRegionBFullDarkPersistence {
			return fmt.Errorf("B-FULL cell %q dark persistence must be %q", cell.CellID, scrollRegionBFullDarkPersistence)
		}
		if _, exists := proof.StorageBefore["theme"]; exists {
			return fmt.Errorf("B-FULL cell %q must not record site theme storage as persistence evidence", cell.CellID)
		}
		for _, action := range proof.Actions {
			if strings.Contains(strings.ToLower(action.Action+" "+action.Return), "theme") {
				return fmt.Errorf("B-FULL cell %q must not claim client theme persistence action", cell.CellID)
			}
		}
		for _, paint := range []scrollRegionBFullPaint{proof.FreshLoadInitialHTML, proof.FreshLoadFirstPaint, proof.FreshLoadSettled} {
			if paint.ThemeSource != scrollRegionBFullThemeInitialSource {
				return fmt.Errorf("B-FULL cell %q paint %q lacks server-routed theme source", cell.CellID, paint.Phase)
			}
		}
	}
	return nil
}

// validateScrollRegionBFullReceiptAxes derives the literal matrix from the
// serialized width axis and maintained theme catalog. A wrapper label alone
// cannot claim an exact cell count or axes.
func validateScrollRegionBFullReceiptAxes(receipt scrollRegionBFullReceipt) error {
	if receipt.ExpectedCells != len(receipt.Cells) {
		return fmt.Errorf("B-FULL receipt expected_cells=%d does not equal serialized cells=%d", receipt.ExpectedCells, len(receipt.Cells))
	}
	if len(receipt.Widths) == 0 {
		return fmt.Errorf("B-FULL receipt width axis is empty")
	}
	widths := make(map[int]struct{}, len(receipt.Widths))
	for index, width := range receipt.Widths {
		if width <= 0 || (index > 0 && receipt.Widths[index-1] >= width) {
			return fmt.Errorf("B-FULL receipt width axis must be positive, unique, and sorted")
		}
		widths[width] = struct{}{}
	}
	seen := make(map[string]struct{}, len(receipt.Cells))
	for _, cell := range receipt.Cells {
		theme, ok := scrollRegionBFullThemeByID(cell.Theme)
		if !ok || cell.Route != scrollRegionBFullRoute || (cell.Mode != "light" && cell.Mode != "dark") || (cell.Zoom != "default" && cell.Zoom != "ua-200") {
			return fmt.Errorf("B-FULL cell %q has unsupported route/theme/mode/zoom axes", cell.CellID)
		}
		if _, ok := widths[cell.ViewportWidth]; !ok {
			return fmt.Errorf("B-FULL cell %q width %d is absent from receipt width axis", cell.CellID, cell.ViewportWidth)
		}
		dark := cell.Mode == "dark"
		wantID := scrollRegionBFullCellID(theme, dark, cell.ViewportWidth, scrollRegionBFullZoom{ID: cell.Zoom})
		if cell.CellID != wantID {
			return fmt.Errorf("B-FULL cell ID %q does not derive from its literal axes %q", cell.CellID, wantID)
		}
		if _, duplicate := seen[cell.CellID]; duplicate {
			return fmt.Errorf("B-FULL receipt contains duplicate cell %q", cell.CellID)
		}
		seen[cell.CellID] = struct{}{}
		if strings.Join(cell.States, ",") != "default,no-overflow,start,middle,end,focused" || strings.Join(cell.Inputs, ",") != "mouse,keyboard,cdp-touch" {
			return fmt.Errorf("B-FULL cell %q does not record the required state/input contract", cell.CellID)
		}
	}
	fullExpected := len(scrollRegionBFullThemes()) * 2 * len(receipt.Widths) * 2
	switch receipt.Closure {
	case "full-closure":
		// Wrapper hashes, structural ZIP parsing, and DOM-to-PNG coordinates are
		// all claimant-recomputable. No independent artifact attestation class is
		// implemented yet, so current receipts must never assert full closure.
		return fmt.Errorf("full B-FULL closure requires independent artifact attestation; structural trace/PNG evidence is unattested")
	case scrollRegionBFullClosureUnattested:
		if receipt.ExpectedCells != fullExpected {
			return fmt.Errorf("unattested literal B-FULL receipt expected_cells=%d, want literal matrix=%d", receipt.ExpectedCells, fullExpected)
		}
		for _, theme := range scrollRegionBFullThemes() {
			for _, dark := range []bool{false, true} {
				for _, width := range receipt.Widths {
					for _, zoom := range []string{"default", "ua-200"} {
						id := scrollRegionBFullCellID(theme, dark, width, scrollRegionBFullZoom{ID: zoom})
						if _, found := seen[id]; !found {
							return fmt.Errorf("unattested literal B-FULL receipt is missing cell %q", id)
						}
					}
				}
			}
		}
	case "diagnostic-non-closure":
		if receipt.ExpectedCells <= 0 || receipt.ExpectedCells >= fullExpected {
			return fmt.Errorf("diagnostic B-FULL receipt cell count must be non-zero and below literal full matrix")
		}
	default:
		return fmt.Errorf("B-FULL receipt closure %q is unsupported", receipt.Closure)
	}
	return nil
}

func scrollRegionBFullThemeByID(id string) (scrollRegionBFullTheme, bool) {
	for _, theme := range scrollRegionBFullThemes() {
		if theme.ID == id {
			return theme, true
		}
	}
	return scrollRegionBFullTheme{}, false
}

// validateScrollRegionBFullRawPersistenceEvidence decodes every direct Page A
// and Page B artifact and re-derives the consent, dark-mode, fresh-load, and
// first-paint claims. WrapperSHA256 is checked above for transport integrity;
// it is deliberately not treated as semantic authentication.
func validateScrollRegionBFullRawPersistenceEvidence(receipt scrollRegionBFullReceipt) error {
	seenPaths := make(map[string]string)
	seenHashes := make(map[string]string)
	if len(receipt.TraceByCell) != len(receipt.Cells) {
		return fmt.Errorf("B-FULL receipt must bind one trace artifact per isolated cell")
	}
	for _, cell := range receipt.Cells {
		if cell.Trace == nil {
			return fmt.Errorf("B-FULL cell %q lacks isolated trace artifact", cell.CellID)
		}
		trace, ok := receipt.TraceByCell[cell.CellID]
		if !ok || trace != *cell.Trace {
			return fmt.Errorf("B-FULL cell %q trace is not wrapper-bound by exact cell ID", cell.CellID)
		}
		traceBytes, err := readScrollRegionBFullArtifact(*cell.Trace, "cell trace", seenPaths, seenHashes)
		if err != nil {
			return err
		}
		if err := validateScrollRegionBFullPlaywrightTrace(cell, traceBytes); err != nil {
			return err
		}
		screenshot, err := readScrollRegionBFullArtifact(cell.Screenshot, "cell screenshot", seenPaths, seenHashes)
		if err != nil {
			return err
		}
		if err := validateScrollRegionBFullReceiptScreenshot(cell, screenshot); err != nil {
			return err
		}
		if err := validateScrollRegionBFullCellPersistence(cell, seenPaths, seenHashes); err != nil {
			return err
		}
	}
	return nil
}

// validateScrollRegionBFullReceiptScreenshot replays the semantic named-region
// crop checks against receipt bytes. Hashing a producer-validated PNG is not
// sufficient: a recomputed wrapper could otherwise pair a generic image with
// a different literal cell.
func validateScrollRegionBFullReceiptScreenshot(cell scrollRegionBFullCellReceipt, content []byte) error {
	artifact := cell.Screenshot
	if artifact.CapturedRegion != "named-scroll-region" || artifact.Capture == nil || artifact.Width <= 0 || artifact.Height <= 0 {
		return fmt.Errorf("B-FULL cell %q screenshot lacks named-region capture provenance", cell.CellID)
	}
	decoded, format, err := image.Decode(bytes.NewReader(content))
	if err != nil || format != "png" {
		return fmt.Errorf("B-FULL cell %q screenshot is not a decodable PNG", cell.CellID)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != artifact.Width || bounds.Dy() != artifact.Height {
		return fmt.Errorf("B-FULL cell %q screenshot dimensions do not bind capture metadata", cell.CellID)
	}
	capture := scrollRegionBFullCapture{ExpectedWidth: float64(artifact.Width), ExpectedHeight: float64(artifact.Height), Proof: *artifact.Capture}
	if err := validateScrollRegionBFullCapturedRegion(decoded, capture); err != nil {
		return fmt.Errorf("B-FULL cell %q screenshot does not prove its named region: %w", cell.CellID, err)
	}
	return nil
}

// validateScrollRegionBFullPlaywrightTrace accepts only an actual Playwright
// trace archive. A ZIP magic prefix, a screenshot hash, or a claimant-authored
// JSON wrapper does not establish Page A/Page B actions or first-paint context.
func validateScrollRegionBFullPlaywrightTrace(cell scrollRegionBFullCellReceipt, raw []byte) error {
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return fmt.Errorf("B-FULL cell %q trace is not a valid Playwright ZIP: %w", cell.CellID, err)
	}
	entries := make(map[string]*zip.File, len(archive.File))
	for _, file := range archive.File {
		entries[file.Name] = file
	}
	traceFile, traceOK := entries["trace.trace"]
	networkFile, networkOK := entries["trace.network"]
	if !traceOK || !networkOK {
		return fmt.Errorf("B-FULL cell %q trace lacks required Playwright trace.trace/trace.network entries", cell.CellID)
	}
	hasScreencastResource := false
	for name := range entries {
		if strings.HasPrefix(name, "resources/page@") && (strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".png")) {
			hasScreencastResource = true
			break
		}
	}
	if !hasScreencastResource {
		return fmt.Errorf("B-FULL cell %q trace lacks Playwright screencast screenshot resource", cell.CellID)
	}
	readEntry := func(file *zip.File) ([]byte, error) {
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	trace, err := readEntry(traceFile)
	if err != nil {
		return fmt.Errorf("read B-FULL cell %q trace.trace: %w", cell.CellID, err)
	}
	network, err := readEntry(networkFile)
	if err != nil {
		return fmt.Errorf("read B-FULL cell %q trace.network: %w", cell.CellID, err)
	}
	if !bytes.Contains(network, []byte(`"type":"resource-snapshot"`)) || !bytes.Contains(network, []byte(`"_resourceType":"document"`)) {
		return fmt.Errorf("B-FULL cell %q trace.network lacks Playwright document resource snapshot", cell.CellID)
	}

	theme, ok := scrollRegionBFullThemeByID(cell.Theme)
	if !ok {
		return fmt.Errorf("B-FULL cell %q has unknown trace theme", cell.CellID)
	}
	expectedURL, err := url.Parse(scrollRegionBFullCellRoutedURL(theme, cell.Mode == "dark", cell.ViewportWidth, scrollRegionBFullZoom{ID: cell.Zoom}))
	if err != nil {
		return fmt.Errorf("B-FULL cell %q construct literal trace URL: %w", cell.CellID, err)
	}
	urlsMatch := func(rawURL string) bool {
		parsed, parseErr := url.Parse(rawURL)
		return parseErr == nil && parsed.Path == expectedURL.Path && parsed.Query().Encode() == expectedURL.Query().Encode()
	}
	type traceCall struct {
		Method string
		PageID string
	}
	calls := make(map[string]traceCall)
	after := make(map[string]struct{})
	pageWidths := make(map[string]bool)
	pageA, pageB := "", ""
	contextOptions := false
	consentCalls, darkCalls := make([]string, 0, 1), make([]string, 0, 2)
	for _, line := range bytes.Split(trace, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var event struct {
			Type              string         `json:"type"`
			CallID            string         `json:"callId"`
			Class             string         `json:"class"`
			Method            string         `json:"method"`
			BrowserName       string         `json:"browserName"`
			PlaywrightVersion string         `json:"playwrightVersion"`
			SDKLanguage       string         `json:"sdkLanguage"`
			Version           int            `json:"version"`
			PageID            string         `json:"pageId"`
			Params            map[string]any `json:"params"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			return fmt.Errorf("B-FULL cell %q trace.trace contains non-Playwright JSON event: %w", cell.CellID, err)
		}
		if event.Type == "context-options" && event.Version >= 1 && strings.TrimSpace(event.BrowserName) != "" && strings.TrimSpace(event.PlaywrightVersion) != "" && strings.TrimSpace(event.SDKLanguage) != "" {
			contextOptions = true
		}
		if event.Type == "after" && event.CallID != "" {
			after[event.CallID] = struct{}{}
			continue
		}
		if event.Type != "before" {
			continue
		}
		if event.CallID == "" || event.PageID == "" {
			continue
		}
		calls[event.CallID] = traceCall{Method: event.Method, PageID: event.PageID}
		selector, _ := event.Params["selector"].(string)
		if event.Method == "setViewportSize" {
			if viewport, ok := event.Params["viewportSize"].(map[string]any); ok {
				if width, ok := viewport["width"].(float64); ok && int(width) == cell.ViewportWidth {
					pageWidths[event.PageID] = true
				}
			}
		}
		if event.Method == "goto" {
			urlValue, _ := event.Params["url"].(string)
			if urlsMatch(urlValue) {
				if pageA == "" {
					pageA = event.PageID
				} else if pageB == "" && event.PageID != pageA {
					pageB = event.PageID
				} else {
					return fmt.Errorf("B-FULL cell %q trace has extra or reused literal Page A/Page B navigation", cell.CellID)
				}
			}
		}
		if event.Method == "click" && strings.Contains(selector, "Allow browser storage") {
			consentCalls = append(consentCalls, event.CallID)
		}
		if event.Method == "click" && strings.Contains(selector, "darkModeToggleBtn") {
			darkCalls = append(darkCalls, event.CallID)
		}
	}
	if !contextOptions || pageA == "" || pageB == "" || pageA == pageB || !pageWidths[pageA] || !pageWidths[pageB] {
		return fmt.Errorf("B-FULL cell %q trace lacks literal trace cell context/Page A/Page B/viewport provenance", cell.CellID)
	}
	if len(consentCalls) != 1 || calls[consentCalls[0]].PageID != pageA {
		return fmt.Errorf("B-FULL cell %q trace lacks one Page A consent action", cell.CellID)
	}
	wantDarkCalls := 2
	if cell.Mode == "dark" {
		wantDarkCalls = 1
	}
	if len(darkCalls) != wantDarkCalls {
		return fmt.Errorf("B-FULL cell %q trace dark action count does not derive from mode", cell.CellID)
	}
	for _, callID := range append(consentCalls, darkCalls...) {
		call := calls[callID]
		if call.PageID != pageA {
			return fmt.Errorf("B-FULL cell %q trace action belongs outside Page A", cell.CellID)
		}
		if _, complete := after[callID]; !complete {
			return fmt.Errorf("B-FULL cell %q trace action %q lacks a successful after event", cell.CellID, call.Method)
		}
	}
	if !bytes.Contains(network, []byte(expectedURL.String())) || bytes.Count(network, []byte(`"_resourceType":"document"`)) < 2 {
		return fmt.Errorf("B-FULL cell %q trace network does not bind both literal Page A/Page B document responses", cell.CellID)
	}
	return nil
}

func readScrollRegionBFullArtifact(artifact scrollRegionBFullArtifact, label string, seenPaths, seenHashes map[string]string) ([]byte, error) {
	if strings.TrimSpace(artifact.Path) == "" || !scrollRegionSHA256Pattern.MatchString(artifact.SHA256) {
		return nil, fmt.Errorf("B-FULL %s artifact has invalid path or SHA-256", label)
	}
	if prior, duplicate := seenPaths[artifact.Path]; duplicate {
		return nil, fmt.Errorf("B-FULL %s artifact path %q is reused by %s", label, artifact.Path, prior)
	}
	content, err := os.ReadFile(artifact.Path)
	if err != nil {
		return nil, fmt.Errorf("read B-FULL %s artifact %q: %w", label, artifact.Path, err)
	}
	if actual := scrollRegionBFullSHA256(content); actual != artifact.SHA256 {
		return nil, fmt.Errorf("B-FULL %s artifact SHA-256 mismatch", label)
	}
	seenPaths[artifact.Path] = label
	// Page A and fresh Page B can legitimately have identical server HTML
	// bytes. Path reuse is forbidden; content-hash reuse is recorded but not
	// treated as a forgery for this deterministic response evidence.
	seenHashes[artifact.SHA256] = label
	return content, nil
}

// validateScrollRegionBFullPaintTranscript establishes that the wrapper's
// initial and first-frame values are retained in the ordered raw observer
// transcript, rather than being claimant-authored replacement values.
func validateScrollRegionBFullPaintTranscript(events []scrollRegionBFullPaint, initial, first scrollRegionBFullPaint) (int, error) {
	if len(events) < 2 || !reflect.DeepEqual(events[0], initial) {
		return 0, fmt.Errorf("first-paint transcript lacks its raw initial server HTML event")
	}
	firstIndex := -1
	for index, observed := range events {
		if observed.Phase == "first-animation-frame" {
			firstIndex = index
			break
		}
	}
	if firstIndex < 1 || !reflect.DeepEqual(events[firstIndex], first) {
		return 0, fmt.Errorf("first-paint transcript does not bind the first animation frame")
	}
	return firstIndex, nil
}

// validateScrollRegionBFullDarkPaintTranscript establishes that dark mode was
// restored by the candidate after the raw server HTML, and before the first
// rendered frame. The ordered browser observer transcript is the authority;
// a receipt cannot assert a later root mutation as first-paint evidence.
func validateScrollRegionBFullDarkPaintTranscript(events []scrollRegionBFullPaint, initial, first scrollRegionBFullPaint, theme, themeSource string) error {
	firstIndex, err := validateScrollRegionBFullPaintTranscript(events, initial, first)
	if err != nil {
		return err
	}
	for _, observed := range events[1:firstIndex] {
		if observed.Phase != "root-mutation" || !observed.Dark {
			continue
		}
		if (theme == "" || observed.Theme == theme) && (themeSource == "" || observed.ThemeSource == themeSource) {
			return nil
		}
	}
	return fmt.Errorf("first-paint transcript lacks a dark root mutation before its first animation frame")
}

func validateScrollRegionBFullCellPersistence(cell scrollRegionBFullCellReceipt, seenPaths, seenHashes map[string]string) error {
	proof := cell.Persistence
	theme, _ := scrollRegionBFullThemeByID(cell.Theme)
	pageAHTML, err := readScrollRegionBFullArtifact(proof.PageAInitialHTML, "Page A initial HTML", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	pageBHTML, err := readScrollRegionBFullArtifact(proof.FreshLoadInitialHTMLArtifact, "Page B initial HTML", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	if cell.FirstHTMLSHA256 != scrollRegionBFullSHA256(pageBHTML) {
		return fmt.Errorf("B-FULL cell %q Page B firstHTMLSHA256 does not bind raw initial HTML", cell.CellID)
	}
	for _, html := range [][]byte{pageAHTML, pageBHTML} {
		for _, want := range []string{`data-theme="` + theme.ServerTheme + `"`, `data-goshtoso-theme-initial-source="` + scrollRegionBFullThemeInitialSource + `"`, `role="region"`, `aria-label="Activity history"`} {
			if !bytes.Contains(html, []byte(want)) {
				return fmt.Errorf("B-FULL cell %q raw initial HTML lacks %q", cell.CellID, want)
			}
		}
	}

	pageAStorageBytes, err := readScrollRegionBFullArtifact(proof.PageAStorageBefore, "Page A storage-before", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	var pageAStorage scrollRegionBFullStorageEvidence
	if err := scrollRegionDecodeStrictJSON(pageAStorageBytes, &pageAStorage); err != nil {
		return fmt.Errorf("B-FULL cell %q Page A storage-before schema: %w", cell.CellID, err)
	}
	if pageAStorage.Schema != scrollRegionBFullStorageEvidenceSchema || pageAStorage.CellID != cell.CellID || pageAStorage.Phase != "page-a-storage-before" || !pageAStorage.State.DialogVisible || strings.Contains(pageAStorage.State.Cookie, "gt_storage=allowed") || pageAStorage.State.DarkMode != "" || !reflect.DeepEqual(proof.StorageBefore, map[string]string{"darkMode": pageAStorage.State.DarkMode, "cookie": pageAStorage.State.Cookie}) {
		return fmt.Errorf("B-FULL cell %q Page A storage-before does not prove isolated first-run consent state", cell.CellID)
	}
	if len(proof.PageAActions) != len(proof.Actions) || len(proof.PageAActions) < 2 || len(cell.SetupActions) != len(proof.Actions) || !reflect.DeepEqual(cell.SetupActions, proof.Actions) {
		return fmt.Errorf("B-FULL cell %q does not bind raw Page A action count and wrapper actions", cell.CellID)
	}
	previous := pageAStorage.State
	for index, artifact := range proof.PageAActions {
		raw, err := readScrollRegionBFullArtifact(artifact, fmt.Sprintf("Page A action %d", index), seenPaths, seenHashes)
		if err != nil {
			return err
		}
		var action scrollRegionBFullActionEvidence
		if err := scrollRegionDecodeStrictJSON(raw, &action); err != nil {
			return fmt.Errorf("B-FULL cell %q Page A action %d schema: %w", cell.CellID, index, err)
		}
		if action.Schema != scrollRegionBFullActionEvidenceSchema || action.CellID != cell.CellID || !reflect.DeepEqual(action.Action, proof.Actions[index]) || !reflect.DeepEqual(action.Before, previous) {
			return fmt.Errorf("B-FULL cell %q Page A action %d does not bind before/action chain", cell.CellID, index)
		}
		if index == 0 {
			if action.Phase != "page-a-consent" || action.Action.Action != "Playwright mouse click Allow browser storage" || action.After.DialogVisible || !strings.Contains(action.After.Cookie, "gt_storage=allowed") {
				return fmt.Errorf("B-FULL cell %q Page A consent action does not prove real dialog dismissal", cell.CellID)
			}
		} else if !strings.HasPrefix(action.Phase, "page-a-dark-") || action.Action.Action != "Playwright mouse click dark mode toggle" {
			return fmt.Errorf("B-FULL cell %q Page A dark action is not a maintained public control", cell.CellID)
		}
		previous = action.After
	}
	wantDark := cell.Mode == "dark"
	if previous.Dark != wantDark || previous.DarkMode != strconv.FormatBool(wantDark) || !strings.Contains(previous.Cookie, "gt_storage=allowed") {
		return fmt.Errorf("B-FULL cell %q Page A actions do not establish requested dark persistence", cell.CellID)
	}

	pageBStorageBytes, err := readScrollRegionBFullArtifact(proof.FreshLoadStorage, "Page B storage", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	var pageBStorage scrollRegionBFullStorageEvidence
	if err := scrollRegionDecodeStrictJSON(pageBStorageBytes, &pageBStorage); err != nil {
		return fmt.Errorf("B-FULL cell %q Page B storage schema: %w", cell.CellID, err)
	}
	if pageBStorage.Schema != scrollRegionBFullStorageEvidenceSchema || pageBStorage.CellID != cell.CellID || pageBStorage.Phase != "page-b-storage" || pageBStorage.State.DialogVisible || pageBStorage.State.Dark != wantDark || pageBStorage.State.DarkMode != strconv.FormatBool(wantDark) || !strings.Contains(pageBStorage.State.Cookie, "gt_storage=allowed") {
		return fmt.Errorf("B-FULL cell %q Page B storage does not prove Page A public-action persistence", cell.CellID)
	}

	paintBytes, err := readScrollRegionBFullArtifact(proof.FreshLoadPaintArtifact, "Page B first-paint", seenPaths, seenHashes)
	if err != nil {
		return err
	}
	var paint scrollRegionBFullPaintEvidence
	if err := scrollRegionDecodeStrictJSON(paintBytes, &paint); err != nil {
		return fmt.Errorf("B-FULL cell %q Page B first-paint schema: %w", cell.CellID, err)
	}
	if paint.Schema != scrollRegionBFullPaintEvidenceSchema || paint.CellID != cell.CellID || len(paint.Events) < 2 || !reflect.DeepEqual(paint.Settled, proof.FreshLoadSettled) {
		return fmt.Errorf("B-FULL cell %q raw Page B first-paint evidence does not bind wrapper paints", cell.CellID)
	}
	initial := paint.Events[0]
	if initial.Phase != "init" || initial.Theme != theme.ServerTheme || initial.ThemeSource != scrollRegionBFullThemeInitialSource || initial.Dark || initial.Visible || initial.Role != "" || initial.Name != "" {
		return fmt.Errorf("B-FULL cell %q raw Page B initial observer state is not pre-rendered server-theme evidence", cell.CellID)
	}
	if wantDark {
		if err := validateScrollRegionBFullDarkPaintTranscript(paint.Events, proof.FreshLoadInitialHTML, proof.FreshLoadFirstPaint, theme.ServerTheme, scrollRegionBFullThemeInitialSource); err != nil {
			return fmt.Errorf("B-FULL cell %q %w", cell.CellID, err)
		}
	} else if _, err := validateScrollRegionBFullPaintTranscript(paint.Events, proof.FreshLoadInitialHTML, proof.FreshLoadFirstPaint); err != nil {
		return fmt.Errorf("B-FULL cell %q %w", cell.CellID, err)
	}
	for _, observed := range []scrollRegionBFullPaint{proof.FreshLoadFirstPaint, paint.Settled} {
		if observed.Theme != theme.ServerTheme || observed.ThemeSource != scrollRegionBFullThemeInitialSource || observed.Dark != wantDark || !observed.Visible || observed.Role != "region" || observed.Name != "Activity history" {
			return fmt.Errorf("B-FULL cell %q raw Page B paint does not prove theme/mode/region correlation", cell.CellID)
		}
	}
	return nil
}
