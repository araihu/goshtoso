//go:build e2e && scrollregion && axe

package e2e

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

const (
	scrollRegionAxeCoreVersion        = "4.10.3"
	scrollRegionAxeArchiveSHA256      = "0f2b4d7dcdf7d1219df8d1959ad68e565f51d14c3f0d88bb71cd59abeb956292"
	scrollRegionAxeScriptSHA256       = "880970c081707360e64f34cea25ff91892f5bc95675b0776925b9709dd8a68bb"
	scrollRegionAxeArchiveEnvironment = "GOSHTOSO_AXE_CORE_TGZ"
)

type scrollRegionAxeCore struct {
	Source string
}

type scrollRegionAxeRule struct {
	ID    string `json:"id"`
	Nodes []struct {
		Target         []string `json:"target"`
		HTML           string   `json:"html"`
		FailureSummary string   `json:"failureSummary"`
	} `json:"nodes"`
}

type scrollRegionAxeResult struct {
	Raw        string `json:"-"`
	TestEngine struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"testEngine"`
	Violations []scrollRegionAxeRule `json:"violations"`
	Incomplete []scrollRegionAxeRule `json:"incomplete"`
	Passes     []scrollRegionAxeRule `json:"passes"`
}

type scrollRegionContrastSample struct {
	Text            string  `json:"text"`
	Foreground      string  `json:"foreground"`
	Background      string  `json:"background"`
	ForegroundRGB   []int   `json:"foregroundRGB"`
	BackgroundRGB   []int   `json:"backgroundRGB"`
	BackgroundAlpha int     `json:"backgroundAlpha"`
	Contrast        float64 `json:"contrast"`
}

func TestScrollRegionAxeNamedPublicViewport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	axe := loadScrollRegionAxeCore(t)
	for _, theme := range []string{"araihu", "goshtoso", "minimal", "modern"} {
		for _, dark := range []bool{false, true} {
			t.Run(theme+"/dark="+strconv.FormatBool(dark), func(t *testing.T) {
				page, _, failures := newScrollRegionTestPage(t)
				_, err := page.Evaluate(`([theme, dark]) => {
					document.documentElement.dataset.theme = theme;
					document.documentElement.classList.toggle("dark", dark);
				}`, []any{theme, dark})
				require.NoError(t, err)
				require.NoError(t, injectScrollRegionAxe(page, axe.Source))
				result := runScrollRegionAxe(t, page)
				require.Equal(t, "axe-core", result.TestEngine.Name)
				require.Equal(t, scrollRegionAxeCoreVersion, result.TestEngine.Version)
				require.Emptyf(t, result.Violations, "axe-core violations on the public Scroll Region examples: %s", describeScrollRegionAxeRules(result.Violations))
				requireScrollRegionAxeCompleteness(t, result)
				require.NotEmpty(t, result.Passes, "axe-core must report evaluated criteria")
				requireScrollRegionAxeRulePasses(t, result.Passes,
					"aria-prohibited-attr",
					"aria-required-attr",
					"aria-roles",
					"aria-valid-attr",
					"aria-valid-attr-value",
					"landmark-unique",
					"scrollable-region-focusable",
				)
				samples := scrollRegionContrastSamples(t, page)
				for _, sample := range samples {
					require.Equalf(t, 255, sample.BackgroundAlpha, "text %q must have an opaque rendered background: %#v", sample.Text, sample)
					require.GreaterOrEqualf(t, sample.Contrast, 4.5, "text %q must meet normal-text contrast against its rendered background: %#v", sample.Text, sample)
				}
				requireScrollRegionPageHealthy(t, page, failures)
			})
		}
	}
}

func requireScrollRegionAxeCompleteness(t *testing.T, result scrollRegionAxeResult) {
	t.Helper()
	if len(result.Incomplete) == 0 {
		return
	}

	// axe-core 4.10.3 cannot evaluate the browser's CSS Color 4 serialization
	// used by Minimal and Modern. Keep this exception exact and prove rendered
	// contrast below; any other incomplete criterion remains a failure.
	require.Len(t, result.Incomplete, 1, "unexpected axe-core incomplete findings: %s", describeScrollRegionAxeRules(result.Incomplete))
	require.Equal(t, "color-contrast", result.Incomplete[0].ID)
	require.Empty(t, result.Incomplete[0].Nodes)
	require.Contains(t, result.Raw, "color-contrast-evaluate", "color-contrast must be incomplete only because axe-core reported its evaluator error")
}

func requireScrollRegionAxeRulePasses(t *testing.T, rules []scrollRegionAxeRule, want ...string) {
	t.Helper()
	passed := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		passed[rule.ID] = struct{}{}
	}
	for _, rule := range want {
		_, ok := passed[rule]
		require.Truef(t, ok, "axe-core did not pass required Scroll Region criterion %q; got %#v", rule, rules)
	}
}

func describeScrollRegionAxeRules(rules []scrollRegionAxeRule) string {
	var findings []string
	for _, rule := range rules {
		for _, node := range rule.Nodes {
			findings = append(findings, rule.ID+" target="+strings.Join(node.Target, " ")+" summary="+node.FailureSummary)
		}
	}
	return strings.Join(findings, "; ")
}

func scrollRegionContrastSamples(t *testing.T, page playwright.Page) []scrollRegionContrastSample {
	t.Helper()
	raw, err := page.Evaluate(`() => {
		const selectors = [
			'#scroll-region-default [data-goshtoso-scroll-viewport] li span',
			'#scroll-region-no-overflow [data-goshtoso-scroll-viewport] p',
			'#scroll-region-indicators-disabled [data-goshtoso-scroll-viewport] li span',
		];
		const canvas = document.createElement('canvas');
		canvas.width = canvas.height = 1;
		const context = canvas.getContext('2d', {willReadFrequently: true});
		const toRGB = color => {
			context.clearRect(0, 0, 1, 1);
			context.fillStyle = color;
			context.fillRect(0, 0, 1, 1);
			return [...context.getImageData(0, 0, 1, 1).data];
		};
		const relativeLuminance = rgb => rgb.slice(0, 3).map(channel => {
			const linear = channel / 255;
			return linear <= 0.04045 ? linear / 12.92 : ((linear + 0.055) / 1.055) ** 2.4;
		}).reduce((sum, channel, index) => sum + channel * [0.2126, 0.7152, 0.0722][index], 0);
		const opaqueBackground = element => {
			for (let node = element; node; node = node.parentElement) {
				const background = getComputedStyle(node).backgroundColor;
				const rgb = toRGB(background);
				if (rgb[3] === 255) return {background, rgb};
			}
			return {background: 'transparent', rgb: [0, 0, 0, 0]};
		};
		return selectors.flatMap(selector => [...document.querySelectorAll(selector)]).map(element => {
			const foreground = getComputedStyle(element).color;
			const foregroundRGB = toRGB(foreground);
			const resolvedBackground = opaqueBackground(element);
			const foregroundLuminance = relativeLuminance(foregroundRGB);
			const backgroundLuminance = relativeLuminance(resolvedBackground.rgb);
			return {
				text: element.textContent.trim(),
				foreground,
				background: resolvedBackground.background,
				foregroundRGB,
				backgroundRGB: resolvedBackground.rgb.slice(0, 3),
				backgroundAlpha: resolvedBackground.rgb[3],
				contrast: (Math.max(foregroundLuminance, backgroundLuminance) + 0.05) / (Math.min(foregroundLuminance, backgroundLuminance) + 0.05),
			};
		});
	}`, nil)
	require.NoError(t, err)
	encoded, err := json.Marshal(raw)
	require.NoError(t, err)
	var samples []scrollRegionContrastSample
	require.NoError(t, json.Unmarshal(encoded, &samples))
	require.NotEmpty(t, samples, "Scroll Region public examples must expose rendered text to contrast-check")
	return samples
}

func loadScrollRegionAxeCore(t *testing.T) scrollRegionAxeCore {
	t.Helper()
	path := os.Getenv(scrollRegionAxeArchiveEnvironment)
	require.NotEmptyf(t, path, "set %s to the authenticated axe-core %s archive", scrollRegionAxeArchiveEnvironment, scrollRegionAxeCoreVersion)

	archive, err := os.ReadFile(path)
	require.NoError(t, err)
	archiveDigest := sha256.Sum256(archive)
	require.Equal(t, scrollRegionAxeArchiveSHA256, hex.EncodeToString(archiveDigest[:]))

	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	require.NoError(t, err)
	t.Cleanup(func() { _ = gzipReader.Close() })
	tarReader := tar.NewReader(gzipReader)
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
	sourceDigest := sha256.Sum256(source)
	require.Equal(t, scrollRegionAxeScriptSHA256, hex.EncodeToString(sourceDigest[:]))
	return scrollRegionAxeCore{Source: string(source)}
}

func injectScrollRegionAxe(page playwright.Page, source string) error {
	ready, err := page.Evaluate(`() => Boolean(window.axe && window.axe.run)`, nil)
	if err != nil {
		return err
	}
	if ready == true {
		return nil
	}
	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{Content: &source})
	return err
}

func runScrollRegionAxe(t *testing.T, page playwright.Page) scrollRegionAxeResult {
	t.Helper()
	raw, err := page.Evaluate(`async () => {
		if (!window.axe || !window.axe.run) throw new Error("authenticated axe-core is unavailable");
		const selector = "#scroll-region-fragment [data-goshtoso-scroll-region]";
		if (!document.querySelector(selector)) throw new Error("public Scroll Region examples are unavailable");
		return await window.axe.run({include: [[selector]]}, {
			resultTypes: ["violations", "incomplete", "passes"],
		});
	}`, nil)
	require.NoError(t, err)
	encoded, err := json.Marshal(raw)
	require.NoError(t, err)
	var result scrollRegionAxeResult
	require.NoError(t, json.Unmarshal(encoded, &result))
	result.Raw = string(encoded)
	return result
}
