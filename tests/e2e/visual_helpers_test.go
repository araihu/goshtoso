package e2e

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// ScreenshotConfig holds configuration for screenshot comparison
type ScreenshotConfig struct {
	OriginalURL    string
	GoshtosoURL    string
	ComponentName  string
	ViewportWidth  int
	ViewportHeight int
	Threshold      float64 // Pixel match threshold (0.0 - 1.0)
}

// DefaultScreenshotConfig returns default configuration
func DefaultScreenshotConfig(componentName string) ScreenshotConfig {
	return ScreenshotConfig{
		ViewportWidth:  1280,
		ViewportHeight: 800,
		Threshold:      0.95, // 95% match required
		ComponentName:  componentName,
	}
}

// CompareScreenshots takes screenshots of both original and Goshtoso implementations
// and returns comparison results
func CompareScreenshots(t *testing.T, cfg ScreenshotConfig) *ComparisonResult {
	t.Helper()

	result := &ComparisonResult{
		ComponentName: cfg.ComponentName,
		Timestamp:     time.Now(),
	}

	// Create screenshots directory
	screenshotDir := filepath.Join("test-results", "screenshots", cfg.ComponentName)
	require.NoError(t, os.MkdirAll(screenshotDir, 0755))

	// Setup browser
	pw, browser, cleanup := setupPlaywrightWithViewport(t, cfg.ViewportWidth, cfg.ViewportHeight)
	defer cleanup()

	// Capture original screenshot
	originalPath := filepath.Join(screenshotDir, "original.png")
	originalBytes := captureScreenshot(t, browser, cfg.OriginalURL, originalPath, "original")
	result.OriginalScreenshotPath = originalPath

	// Capture Goshtoso screenshot
	goshtosoPath := filepath.Join(screenshotDir, "goshtoso.png")
	goshtosoBytes := captureScreenshot(t, browser, cfg.GoshtosoURL, goshtosoPath, "goshtoso")
	result.GoshtosoScreenshotPath = goshtosoPath

	// Compare screenshots
	if len(originalBytes) > 0 && len(goshtosoBytes) > 0 {
		matchPercent, diffPath := compareImages(t, originalBytes, goshtosoBytes, screenshotDir)
		result.MatchPercentage = matchPercent
		result.DiffScreenshotPath = diffPath
		result.Passed = matchPercent >= cfg.Threshold

		t.Logf("Visual parity for %s: %.2f%% match (threshold: %.2f%%)",
			cfg.ComponentName, matchPercent*100, cfg.Threshold*100)
	}

	_ = pw // Use pw to avoid unused variable error
	return result
}

// ComparisonResult holds the results of a visual comparison
type ComparisonResult struct {
	ComponentName          string
	Timestamp              time.Time
	OriginalScreenshotPath string
	GoshtosoScreenshotPath string
	DiffScreenshotPath     string
	MatchPercentage        float64
	Passed                 bool
	Errors                 []string
}

// captureScreenshot navigates to a URL and captures a screenshot
func captureScreenshot(t *testing.T, browser playwright.Browser, url, savePath, label string) []byte {
	t.Helper()

	page := newPage(t, browser)
	defer page.Close()

	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(3000),
	})
	require.NoError(t, err, "failed to navigate to %s (%s)", url, label)

	// Wait for any animations to complete
	page.WaitForTimeout(50)

	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type:     playwright.ScreenshotTypePng,
		FullPage: playwright.Bool(false),
	})
	require.NoError(t, err, "failed to capture screenshot for %s", label)

	// Save screenshot
	err = os.WriteFile(savePath, screenshot, 0644)
	require.NoError(t, err, "failed to save screenshot to %s", savePath)

	t.Logf("✓ Captured %s screenshot: %s", label, savePath)
	return screenshot
}

// compareImages compares two PNG images and returns match percentage
func compareImages(t *testing.T, img1Bytes, img2Bytes []byte, outputDir string) (float64, string) {
	t.Helper()

	// Decode images
	img1, err := png.Decode(bytes.NewReader(img1Bytes))
	require.NoError(t, err, "failed to decode image 1")

	img2, err := png.Decode(bytes.NewReader(img2Bytes))
	require.NoError(t, err, "failed to decode image 2")

	// Get bounds
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Check if dimensions match
	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		t.Logf("Image dimensions differ: %dx%d vs %dx%d",
			bounds1.Dx(), bounds1.Dy(), bounds2.Dx(), bounds2.Dy())
		return 0.0, ""
	}

	width := bounds1.Dx()
	height := bounds1.Dy()
	totalPixels := width * height

	// Create diff image
	diffImg := image.NewRGBA(bounds1)
	draw.Draw(diffImg, bounds1, img1, bounds1.Min, draw.Src)

	// Compare pixels
	differentPixels := 0
	tolerance := uint32(10) // Color difference tolerance (0-255)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c1 := color.RGBAModel.Convert(img1.At(x, y)).(color.RGBA)
			c2 := color.RGBAModel.Convert(img2.At(x, y)).(color.RGBA)

			// Check if colors are different (with tolerance)
			rDiff := abs(int(c1.R) - int(c2.R))
			gDiff := abs(int(c1.G) - int(c2.G))
			bDiff := abs(int(c1.B) - int(c2.B))
			aDiff := abs(int(c1.A) - int(c2.A))

			if uint32(rDiff) > tolerance || uint32(gDiff) > tolerance ||
				uint32(bDiff) > tolerance || uint32(aDiff) > tolerance {
				differentPixels++
				// Mark difference in red
				diffImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			}
		}
	}

	// Calculate match percentage
	matchPercent := 1.0 - float64(differentPixels)/float64(totalPixels)

	// Save diff image
	diffPath := filepath.Join(outputDir, "diff.png")
	diffFile, err := os.Create(diffPath)
	require.NoError(t, err)
	defer diffFile.Close()

	err = png.Encode(diffFile, diffImg)
	require.NoError(t, err)

	return matchPercent, diffPath
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// setupPlaywrightWithViewport initializes Playwright with specific viewport
func setupPlaywrightWithViewport(t *testing.T, width, height int) (*playwright.Playwright, playwright.Browser, func()) {
	pw, err := playwright.Run()
	require.NoError(t, err, "failed to start playwright")

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err, "failed to launch browser")

	cleanup := func() {
		browser.Close()
		pw.Stop()
	}

	return pw, browser, cleanup
}

// TakeComponentScreenshot captures a screenshot of a specific component
func TakeComponentScreenshot(t *testing.T, page playwright.Page, selector, componentName string) string {
	t.Helper()

	screenshotDir := filepath.Join("test-results", "screenshots", componentName)
	require.NoError(t, os.MkdirAll(screenshotDir, 0755))

	screenshotPath := filepath.Join(screenshotDir, fmt.Sprintf("%s-%d.png", componentName, time.Now().Unix()))

	locator := page.Locator(selector)
	_, err := locator.Screenshot(playwright.LocatorScreenshotOptions{
		Path: playwright.String(screenshotPath),
		Type: playwright.ScreenshotTypePng,
	})
	require.NoError(t, err, "failed to capture component screenshot")

	t.Logf("✓ Component screenshot saved: %s", screenshotPath)
	return screenshotPath
}

// CreateSideBySideComparison creates a side-by-side comparison image
func CreateSideBySideComparison(t *testing.T, img1Path, img2Path, outputPath string) {
	t.Helper()

	// Read images
	img1Bytes, err := os.ReadFile(img1Path)
	require.NoError(t, err)

	img2Bytes, err := os.ReadFile(img2Path)
	require.NoError(t, err)

	img1, err := png.Decode(bytes.NewReader(img1Bytes))
	require.NoError(t, err)

	img2, err := png.Decode(bytes.NewReader(img2Bytes))
	require.NoError(t, err)

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Create canvas
	width := bounds1.Dx() + bounds2.Dx() + 20 // Gap between images
	height := max(bounds1.Dy(), bounds2.Dy())

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw first image
	draw.Draw(canvas, bounds1, img1, bounds1.Min, draw.Src)

	// Draw second image (offset by first width + gap)
	offsetBounds := image.Rect(bounds1.Dx()+20, 0, width, bounds2.Dy())
	draw.Draw(canvas, offsetBounds, img2, bounds2.Min, draw.Src)

	// Add labels
	// Note: Adding text would require a font package, skipping for simplicity

	// Save comparison
	outFile, err := os.Create(outputPath)
	require.NoError(t, err)
	defer outFile.Close()

	err = png.Encode(outFile, canvas)
	require.NoError(t, err)

	t.Logf("✓ Side-by-side comparison saved: %s", outputPath)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
