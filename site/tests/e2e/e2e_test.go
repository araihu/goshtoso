package e2e

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

var (
	baseURL       = "" // set dynamically in TestMain
	screenshotDir = "test-results/screenshots"
)

// Shared singleton state — initialized once in TestMain, shared across all tests.
var (
	sharedPW      *playwright.Playwright
	sharedBrowser playwright.Browser
	serverCmd     *exec.Cmd
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
	buildCmd := exec.Command("go", "build", "-o", "bin/server", "./cmd/server")
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
	if serverCmd != nil && serverCmd.Process != nil {
		_ = syscall.Kill(-serverCmd.Process.Pid, syscall.SIGTERM)
		_ = serverCmd.Wait()
	}

	os.Exit(code)
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
