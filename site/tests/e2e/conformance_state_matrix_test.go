//go:build e2e && full

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/internal/conformanceledger"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestConformanceLedgerStateFixturesRenderEverySourceState(t *testing.T) {
	fixtures := conformanceStateFixtures()
	require.Len(t, fixtures, 347)
	seen := make(map[string]struct{}, len(fixtures))
	for _, fixture := range fixtures {
		fixture := fixture
		if _, duplicate := seen[fixture.State]; duplicate {
			t.Fatalf("duplicate state fixture %s", fixture.State)
		}
		seen[fixture.State] = struct{}{}
		t.Run(fixture.State, func(t *testing.T) {
			var rendered strings.Builder
			err := fixture.Component.Render(context.Background(), &rendered)
			require.NoError(t, err)
		})
	}
}

func TestConformanceLedgerPublicStateBFull(t *testing.T) {
	fixtures := conformanceStateFixtures()
	require.Len(t, fixtures, 347)

	repoRoot, err := filepath.Abs("../../..")
	require.NoError(t, err)
	commit := conformanceGitIdentity(t, repoRoot, "HEAD^{commit}")
	tree := conformanceGitIdentity(t, repoRoot, "HEAD^{tree}")
	browserIdentity := "Chromium " + sharedBrowser.Version()

	evidenceDir := os.Getenv("CONFORMANCE_BFULL_EVIDENCE_DIR")
	if evidenceDir == "" {
		evidenceDir = t.TempDir()
	}
	require.NoError(t, os.MkdirAll(evidenceDir, 0o700))

	diagnostic := os.Getenv("CONFORMANCE_BFULL_DIAGNOSTIC") == "1"
	diagnosticStateCap := conformanceDiagnosticStateCap(t, diagnostic, len(fixtures))
	selectedFixtures := fixtures
	if diagnostic {
		selectedFixtures = fixtures[:diagnosticStateCap]
		if requested := os.Getenv("CONFORMANCE_BFULL_DIAGNOSTIC_STATES"); requested != "" {
			selectedFixtures = conformanceDiagnosticStateSelection(t, fixtures, requested)
			diagnosticStateCap = len(selectedFixtures)
		}
	}
	identityMode := conformanceledger.BFullIdentityCommitted
	if diagnostic {
		identityMode = conformanceledger.BFullIdentityDirtyBound
	}
	if requested := os.Getenv("CONFORMANCE_BFULL_IDENTITY_MODE"); requested != "" {
		identityMode = requested
	}
	if !diagnostic {
		require.Equal(t, conformanceledger.BFullIdentityCommitted, identityMode, "closure runs must use a clean committed identity")
	}
	identity, err := conformanceledger.BuildBFullIdentity(repoRoot, commit, tree, identityMode, filepath.Join(evidenceDir, "dirty-worktree-manifest.json"))
	require.NoError(t, err)
	consumerFixture := conformanceBuildConsumerFixture(t, evidenceDir)
	axeCore := conformanceLoadAxeCore(t)

	documents := map[string]string{}
	for _, theme := range []string{"araihu", "goshtoso", "minimal", "modern"} {
		for _, mode := range []string{"light", "dark"} {
			documents[theme+"/"+mode] = conformanceStateDocument(t, selectedFixtures, theme, mode, consumerFixture)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /sprites/ui.svg", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "image/svg+xml")
		_, _ = io.WriteString(writer, `<svg xmlns="http://www.w3.org/2000/svg"><symbol id="check" viewBox="0 0 24 24"><path d="M3 12l6 6L21 4"/></symbol></svg>`)
	})
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		theme := request.URL.Query().Get("theme")
		mode := request.URL.Query().Get("mode")
		if documents[theme+"/"+mode] == "" {
			http.Error(writer, "unknown conformance axes", http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(writer, documents[theme+"/"+mode])
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	page := newIsolatedPage(t)
	if diagnostic {
		page.SetDefaultTimeout(750)
	} else {
		page.SetDefaultTimeout(3_000)
	}
	var errorMu sync.Mutex
	var browserErrors []conformanceledger.BFullBrowserError
	page.On("pageerror", func(exception error) {
		errorMu.Lock()
		browserErrors = append(browserErrors, conformanceledger.BFullBrowserError{Kind: "page", Message: conformancePageErrorMessage(exception)})
		errorMu.Unlock()
	})
	page.On("console", func(message playwright.ConsoleMessage) {
		if message.Type() != "error" {
			return
		}
		kind := "console"
		if strings.Contains(strings.ToLower(message.Text()), "content security policy") {
			kind = "csp"
		}
		errorMu.Lock()
		browserErrors = append(browserErrors, conformanceledger.BFullBrowserError{Kind: kind, Message: message.Text()})
		errorMu.Unlock()
	})
	page.On("requestfailed", func(request playwright.Request) {
		errorMu.Lock()
		browserErrors = append(browserErrors, conformanceledger.BFullBrowserError{Kind: "network", Message: request.URL()})
		errorMu.Unlock()
	})
	page.On("response", func(response playwright.Response) {
		if response.Status() >= 400 {
			errorMu.Lock()
			browserErrors = append(browserErrors, conformanceledger.BFullBrowserError{Kind: "network", Message: fmt.Sprintf("%d %s", response.Status(), response.URL())})
			errorMu.Unlock()
		}
	})
	inventory, err := conformanceledger.DeriveInventory(repoRoot)
	require.NoError(t, err)
	axes, err := conformanceledger.ExpectedBFullAxesFromInventory(inventory)
	require.NoError(t, err)
	require.Equal(t, 699552, len(axes.States)*len(axes.Themes)*len(axes.Modes)*len(axes.Viewports)*len(axes.Zooms)*len(axes.Motions)*len(axes.Inputs))

	manifest := conformanceledger.BFullManifest{SchemaVersion: conformanceledger.SchemaVersion, SourceCommit: commit, SourceTree: tree, Identity: identity, Browser: browserIdentity}
	if diagnostic {
		manifest.Diagnostic = &conformanceledger.BFullDiagnostic{NonClosure: true, StateCap: diagnosticStateCap, Reason: "bounded dirty-bound feasibility slice; never closure evidence"}
	}
	manifest.Cells = make([]conformanceledger.BFullCell, 0, 699552)
	session, err := page.Context().NewCDPSession(page)
	require.NoError(t, err)

	batchNumber := 0
	for _, theme := range axes.Themes {
		for _, mode := range axes.Modes {
			for _, viewport := range axes.Viewports {
				for _, zoom := range axes.Zooms {
					for _, motion := range axes.Motions {
						batchNumber++
						if diagnostic && batchNumber > 1 {
							continue
						}
						require.NoError(t, page.SetViewportSize(viewport, 900))
						reducedMotion := playwright.ReducedMotionNoPreference
						if motion == "reduced" {
							reducedMotion = playwright.ReducedMotionReduce
						}
						require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: reducedMotion}))

						errorMu.Lock()
						errorStart := len(browserErrors)
						errorMu.Unlock()
						response, err := page.Goto(server.URL+fmt.Sprintf("?theme=%s&mode=%s", theme, mode), playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
						require.NoError(t, err)
						require.Equal(t, 200, response.Status())
						conformanceSetUserAgentZoom(t, page, session, zoom)
						initialHTML, err := response.Body()
						require.NoError(t, err)
						initialHTMLPath := filepath.Join(evidenceDir, fmt.Sprintf("initial-%04d-%s-%s-%d-%d-%s.html", batchNumber, theme, mode, viewport, zoom, motion))
						require.NoError(t, os.WriteFile(initialHTMLPath, initialHTML, 0o600))
						initialHTMLSHA256, err := conformanceledger.SHA256File(initialHTMLPath)
						require.NoError(t, err)
						firstPaint := conformanceInspectStates(t, page, theme, mode, motion, zoom)
						zoomErr := conformanceVerifyUserAgentZoom(page, zoom)
						require.NoError(t, waitForAlpine(page))
						require.Equal(t, len(selectedFixtures), mustLocatorCount(t, page.Locator("[data-conformance-state]")))
						require.NoError(t, conformanceInjectAxe(page, axeCore.Source))
						inputs := conformanceExecuteInputs(t, page, session, selectedFixtures, inventory)
						persistenceResponse, err := page.Reload(playwright.PageReloadOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
						require.NoError(t, err)
						require.Equal(t, 200, persistenceResponse.Status())
						require.NoError(t, waitForAlpine(page))
						require.NoError(t, conformanceInjectAxe(page, axeCore.Source))
						persistence := conformanceInspectStates(t, page, theme, mode, motion, zoom)
						accessibility := conformanceRunAxe(t, page, axeCore, len(selectedFixtures))
						if !diagnostic {
							require.Empty(t, accessibility.Violations)
							require.Empty(t, accessibility.Incomplete)
						}
						errorMu.Lock()
						batchErrors := append([]conformanceledger.BFullBrowserError(nil), browserErrors[errorStart:]...)
						errorMu.Unlock()
						if zoomErr != nil {
							batchErrors = append(batchErrors, conformanceledger.BFullBrowserError{Kind: "zoom-reflow", Message: zoomErr.Error()})
						}
						if !diagnostic {
							require.Empty(t, batchErrors)
						}

						evidence := conformanceledger.BFullBatchEvidence{
							SchemaVersion: conformanceledger.SchemaVersion, SourceCommit: commit, SourceTree: tree, Browser: browserIdentity,
							Identity: identity,
							Theme:    theme, Mode: mode, Viewport: viewport, Zoom: zoom, Motion: motion,
							InitialHTMLPath: initialHTMLPath, InitialHTMLSHA256: initialHTMLSHA256,
							FirstPaint: firstPaint, Persistence: persistence,
							Accessibility: accessibility, Errors: batchErrors, Inputs: inputs,
						}
						evidencePath := filepath.Join(evidenceDir, fmt.Sprintf("batch-%04d-%s-%s-%d-%d-%s.json", batchNumber, theme, mode, viewport, zoom, motion))
						conformanceWriteJSON(t, evidencePath, evidence)
						evidenceSHA256, err := conformanceledger.SHA256File(evidencePath)
						require.NoError(t, err)
						manifest.Batches = append(manifest.Batches, conformanceledger.BFullBatch{Theme: theme, Mode: mode, Viewport: viewport, Zoom: zoom, Motion: motion, EvidencePath: evidencePath, EvidenceSHA256: evidenceSHA256})
						for _, input := range inputs {
							manifest.Cells = append(manifest.Cells, conformanceledger.BFullCell{State: input.State, Theme: theme, Mode: mode, Viewport: viewport, Zoom: zoom, Motion: motion, Input: input.Input, Applicability: input.Applicability, ReceiptStatus: input.ReceiptStatus, Rationale: input.Rationale, EvidenceSHA256: evidenceSHA256})
						}
					}
				}
			}
		}
	}
	_, err = session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
	require.NoError(t, err)
	if !diagnostic {
		manifest.Coverage = conformanceCollectBFullCoverage(t, inventory, consumerFixture, evidenceDir)
		require.NoError(t, conformanceledger.ValidateBFullCoverage(manifest.Coverage, conformanceledger.RequiredBFullCoverageFromInventory(inventory)))
	}

	manifestPath := filepath.Join(evidenceDir, "b-full-manifest.json")
	conformanceWriteJSON(t, manifestPath, manifest)
	if diagnostic {
		t.Logf("NON-CLOSURE diagnostic B-FULL evidence: %s identity=%s state_cap=%d", manifestPath, identity.Mode, diagnosticStateCap)
		return
	}
	require.Len(t, manifest.Batches, 504)
	require.Len(t, manifest.Cells, 699552)
	require.NoError(t, conformanceledger.ValidateBFullManifest(manifest, commit, tree, axes))
	manifestSHA256, err := conformanceledger.SHA256File(manifestPath)
	require.NoError(t, err)
	t.Logf("B-FULL manifest=%s sha256=%s cells=%d batches=%d", manifestPath, manifestSHA256, len(manifest.Cells), len(manifest.Batches))
}
func TestConformanceNamespaceFragmentPreservesAlpineExpressions(t *testing.T) {
	fragment := `<div id="panel" aria-labelledby="panel label"><label for="input">State</label><input id="input"><template x-for="(slide, index) in slides"><h3 x-bind:id="'slide-' + index"></h3></template><a href="#panel">State</a></div>`
	got := conformanceNamespaceFragment(fragment, "carousel/default")
	for _, want := range []string{
		`id="state-carousel-default-panel"`,
		`aria-labelledby="state-carousel-default-panel state-carousel-default-label"`,
		`for="state-carousel-default-input"`,
		`id="state-carousel-default-input"`,
		`href="#state-carousel-default-panel"`,
		`x-for="(slide, index) in slides"`,
		`x-bind:id="'slide-' + index"`,
	} {
		require.Contains(t, got, want)
	}
}
