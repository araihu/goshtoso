package e2e

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

var (
	baseURL             = "" // set dynamically in TestMain
	goshtosoDocsVersion = ""
	screenshotDir       = "test-results/screenshots"
)

const avatarBrokenImagePath = "/assets/images/does-not-exist-404.png"

const pageSettleGraceMilliseconds = 250

// Shared singleton state — initialized once in TestMain, shared across all tests.
var (
	sharedPW      *playwright.Playwright
	sharedBrowser playwright.Browser
	serverCmd     *exec.Cmd
	shutdownToken string
)

// freePort finds an available TCP port
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port, nil
}

func TestMain(m *testing.M) {
	// Build server
	projectRoot, _ := filepath.Abs("../..")
	buildArgs := []string{"build", "-o", "bin/server"}
	docsVersion, err := pinnedGoshtosoVersion(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve pinned Goshtoso version: %v\n", err)
		os.Exit(1)
	}
	goshtosoDocsVersion = docsVersion
	buildArgs = append(buildArgs,
		"-ldflags=-X github.com/araihu/goshtoso/site/internal/buildinfo.goDocsVersion="+docsVersion,
	)
	if coverDir := os.Getenv("GOSHTOSO_E2E_COVERDIR"); coverDir != "" {
		buildArgs = append(buildArgs, "-cover")
		if coverPkg := os.Getenv("GOSHTOSO_E2E_COVERPKG"); coverPkg != "" {
			buildArgs = append(buildArgs, "-coverpkg="+e2eCoverPkg(coverPkg))
		}
	}
	buildArgs = append(buildArgs, "./cmd/server")
	buildCmd := exec.Command("go", buildArgs...)
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build server: %s\n%s\n", err, output)
		os.Exit(1)
	}

	// Pick a random free port so tests don't conflict with manual dev server on 8090
	port, err := freePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find free port: %v\n", err)
		os.Exit(1)
	}
	baseURL = fmt.Sprintf("http://localhost:%d", port)

	// Start server
	serverBin := filepath.Join(projectRoot, "bin", "server")
	serverCmd = exec.Command(serverBin, "-port", fmt.Sprintf("%d", port))
	serverCmd.Dir = projectRoot
	if coverDir := os.Getenv("GOSHTOSO_E2E_COVERDIR"); coverDir != "" {
		if err := os.MkdirAll(coverDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create coverage dir: %v\n", err)
			os.Exit(1)
		}
		shutdownToken = randomToken()
		serverCmd.Env = append(os.Environ(),
			"GOCOVERDIR="+coverDir,
			"GOSHTOSO_E2E_SHUTDOWN_TOKEN="+shutdownToken,
		)
	}
	serverCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := serverCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
	// Wait for ready
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(baseURL); err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Launch shared Playwright + browser
	pw, err := playwright.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start playwright: %v\n", err)
		os.Exit(1)
	}
	sharedPW = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: new(true),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to launch browser: %v\n", err)
		_ = pw.Stop()
		os.Exit(1)
	}
	sharedBrowser = browser

	// Run all tests
	code := m.Run()

	// Cleanup
	_ = browser.Close()
	_ = pw.Stop()
	if err := stopServer(serverCmd, baseURL, shutdownToken); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop server: %v\n", err)
	}

	os.Exit(code)
}

func pinnedGoshtosoVersion(projectRoot string) (string, error) {
	cmd := exec.Command(
		"go",
		"list",
		"-m",
		"-f",
		"{{.Version}}",
		"github.com/araihu/goshtoso",
	)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("empty module version")
	}
	return version, nil
}

func stopServer(cmd *exec.Cmd, url, token string) error {
	if cmd == nil || cmd.Process == nil {
		if token == "" {
			return nil
		}
		return postShutdown(url, token)
	}
	if token != "" {
		if err := postShutdown(url, token); err == nil {
			return cmd.Wait()
		}
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
		return err
	}
	return cmd.Wait()
}

func postShutdown(url, token string) error {
	req, err := http.NewRequest(http.MethodPost, url+"/__e2e/shutdown?token="+token, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("shutdown returned %s", resp.Status)
	}
	return nil
}

func randomToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func e2eCoverPkg(coverPkg string) string {
	return "github.com/araihu/goshtoso/site/cmd/server," + coverPkg
}

// setupServer is now a no-op since TestMain handles it.
// Kept for backward compatibility with existing tests.
func setupServer(t *testing.T) func() {
	return func() {}
}

// setupPlaywright returns the shared browser. The cleanup func is a no-op
// since the browser lives for the entire test run.
// Kept for backward compatibility with existing tests.
func setupPlaywright(t *testing.T) (*playwright.Playwright, playwright.Browser, func()) {
	return sharedPW, sharedBrowser, func() {}
}

// newPage creates a new page (tab) in the shared browser with short timeouts.
// The caller should defer page.Close() to clean up the tab.
func newPage(t *testing.T, browser playwright.Browser, opts ...playwright.BrowserNewPageOptions) playwright.Page {
	var page playwright.Page
	var err error
	if len(opts) > 0 {
		page, err = browser.NewPage(opts[0])
	} else {
		page, err = browser.NewPage()
	}
	require.NoError(t, err)
	page.SetDefaultTimeout(2000)
	page.SetDefaultNavigationTimeout(3000)
	t.Cleanup(func() { _ = page.Close() })
	return page
}

func waitForPageSettled(t *testing.T, page playwright.Page) {
	t.Helper()
	require.NoError(t, page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(3000),
	}))
	waitForPageFailureWindow(t, page)
}

// waitForPageFailureWindow gives asynchronous console and page-error handlers
// one bounded turn without adding Playwright's network-idle quiet period.
func waitForPageFailureWindow(t *testing.T, page playwright.Page) {
	t.Helper()
	// Network idle may already have been reached before the caller starts
	// waiting. Give timers and Playwright's asynchronous event callbacks one
	// bounded turn before pageFailures takes its final snapshot.
	_, err := page.Evaluate(
		"milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds))",
		pageSettleGraceMilliseconds,
	)
	require.NoError(t, err)
}

type pageFailures struct {
	mu       sync.Mutex
	messages []string
}

func watchPageFailures(page playwright.Page) *pageFailures {
	failures := &pageFailures{}
	add := func(message string) {
		failures.mu.Lock()
		defer failures.mu.Unlock()
		failures.messages = append(failures.messages, message)
	}

	page.OnPageError(func(err error) {
		add(fmt.Sprintf("page error: %v", err))
	})
	page.OnConsole(func(message playwright.ConsoleMessage) {
		if message.Type() == "error" {
			add(consoleFailureMessage(message))
		}
	})
	page.OnRequestFailed(func(request playwright.Request) {
		add(fmt.Sprintf(
			"request failed: %s %s: %v",
			request.Method(),
			request.URL(),
			request.Failure(),
		))
	})
	page.OnResponse(func(response playwright.Response) {
		if response.Status() >= http.StatusBadRequest {
			add(fmt.Sprintf(
				"HTTP response: %d %s: %s",
				response.Status(),
				response.StatusText(),
				response.URL(),
			))
		}
	})

	return failures
}

func consoleFailureMessage(message playwright.ConsoleMessage) string {
	location := message.Location()
	if location == nil {
		return fmt.Sprintf("console error: %s", message.Text())
	}
	return fmt.Sprintf(
		"console error: %s [url=%s line=%d column=%d]",
		message.Text(),
		location.URL,
		location.LineNumber,
		location.ColumnNumber,
	)
}

// filterIgnorable drops only the intentional broken-image fixture used by the
// Avatar fallback demo. Category, status, and exact local URL must all match.
func filterIgnorable(messages []string) []string {
	kept := make([]string, 0, len(messages))
	fixtureURL := baseURL + avatarBrokenImagePath
	for _, message := range messages {
		knownHTTP404 := message == "HTTP response: 404 Not Found: "+fixtureURL
		knownConsole404 := strings.HasPrefix(message,
			"console error: Failed to load resource: the server responded with a status of 404 (Not Found) "+
				"[url="+fixtureURL+" ")
		if knownHTTP404 || knownConsole404 {
			continue
		}
		kept = append(kept, message)
	}
	return kept
}

func (failures *pageFailures) RequireEmpty(t *testing.T) {
	t.Helper()

	failures.mu.Lock()
	messages := append([]string(nil), failures.messages...)
	failures.mu.Unlock()

	require.Empty(
		t,
		filterIgnorable(messages),
		"unexpected page failures: %s",
		strings.Join(messages, "; "),
	)
}

// clickUntil clicks loc and waits for jsCondition to hold, retrying the click
// if it was dropped by an HTMX swap rebind race. When a control is replaced by a
// swap (outerHTML or OOB), htmx re-binds the new element a beat after it appears
// in the DOM; a click landing in that window fires no request and the UI state
// never advances. Because a dropped click leaves state unchanged (no request)
// while a real click lands in tens of ms, retrying only ever re-fires genuinely
// lost clicks — it never double-advances stateful controls. Fast path: 1 click.
func clickUntil(t *testing.T, page playwright.Page, loc playwright.Locator, jsCondition string) {
	t.Helper()
	for attempt := 1; attempt <= 5; attempt++ {
		require.NoError(t, loc.Click())
		if _, err := page.WaitForFunction(jsCondition, nil, playwright.PageWaitForFunctionOptions{
			Timeout: playwright.Float(2000),
		}); err == nil {
			return
		}
	}
	t.Fatalf("clickUntil: condition never satisfied after 5 click attempts: %s", jsCondition)
}

// takeScreenshot captures a screenshot for debugging
func takeScreenshot(t *testing.T, page playwright.Page, name string) {
	_ = os.MkdirAll(screenshotDir, 0755)
	path := filepath.Join(screenshotDir, fmt.Sprintf("%s-%d.png", name, time.Now().Unix()))
	_, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     new(path),
		FullPage: new(true),
	})
	if err != nil {
		t.Logf("failed to take screenshot: %v", err)
	} else {
		t.Logf("Screenshot saved: %s", path)
	}
}
