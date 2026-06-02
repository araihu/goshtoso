package e2e

import (
	"testing"
)

// TestThemeContrastResolvesColors guards the contrast-checker color resolver.
// Regression: Tailwind v4 palette tokens are oklch; the old normalizeColor read
// ctx.fillStyle, which echoes oklch() back verbatim under CSS Color 4, so every
// token fell back to #000000 and every ratio collapsed to 1.00:1. The pixel-
// readback resolver must produce real, varied colors.
func TestThemeContrastResolvesColors(t *testing.T) {
	_, browser, cleanup := setupPlaywright(t)
	defer cleanup()
	page := newPage(t, browser)
	if _, err := page.Goto(baseURL + "/docs/theme"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	if _, err := page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil); err != nil {
		t.Fatalf("alpine load: %v", err)
	}

	// Drive the Alpine component directly: resolve colors and read the matrix.
	res, err := page.Evaluate(`() => {
		const root = document.querySelector('[x-data="themePage"]');
		const c = Alpine.$data(root);
		c.refreshResolved();
		const m = c.contrastMatrix();
		const colors = m.map(x => x.color);
		const ratios = m.map(x => x.ratio);
		return {
			count: m.length,
			nonBlack: colors.filter(h => h && h.toLowerCase() !== '#000000').length,
			distinctColors: new Set(colors).size,
			distinctRatios: new Set(ratios.map(r => r.toFixed(2))).size,
			maxRatio: Math.max(...ratios),
		};
	}`, nil)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	m, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type %T: %v", res, res)
	}
	// playwright-go decodes integer JS numbers as int, floats as float64.
	num := func(k string) float64 {
		switch v := m[k].(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		default:
			return 0
		}
	}

	if num("count") < 10 {
		t.Fatalf("expected a full token matrix, got count=%v", m["count"])
	}
	// Before the fix every token resolved to #000000. Demand most are real.
	if num("nonBlack") < num("count")*0.7 {
		t.Fatalf("too many tokens fell back to black: nonBlack=%v of count=%v", m["nonBlack"], m["count"])
	}
	if num("distinctColors") < 4 {
		t.Fatalf("colors not resolving to distinct values: distinctColors=%v", m["distinctColors"])
	}
	// Before the fix every ratio was 1.00:1.
	if num("distinctRatios") < 3 {
		t.Fatalf("ratios collapsed (oklch not resolved): distinctRatios=%v", m["distinctRatios"])
	}
	if num("maxRatio") < 3 {
		t.Fatalf("no meaningful contrast found: maxRatio=%v", m["maxRatio"])
	}
}
