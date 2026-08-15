//go:build e2e && full

package e2e

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/araihu/goshtoso/site/internal/pages/catalog"
	demothemes "github.com/araihu/goshtoso/site/internal/themes"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

var conformanceViewportWidths = []int{
	390,
	639, 640, 641,
	704,
	767, 768, 769,
	896,
	1023, 1024, 1025,
	1152,
	1279, 1280, 1281,
	1408,
	1440,
	1535, 1536, 1537,
}

func TestConformanceLedgerCurrentSourceRouteThemeModeResponsiveSmoke(t *testing.T) {
	pages := catalog.ComponentPages()
	require.Len(t, pages, 51, "current source component route inventory")
	themes := demothemes.All()
	require.Len(t, themes, 16, "current source theme inventory")

	seenRoutes := make(map[string]struct{}, len(pages))
	for _, entry := range pages {
		entry := entry
		t.Run(entry.Key, func(t *testing.T) {
			if _, duplicate := seenRoutes[entry.Path]; duplicate {
				t.Fatalf("duplicate route %s", entry.Path)
			}
			seenRoutes[entry.Path] = struct{}{}

			page := newIsolatedPage(t)
			dismissCookieBanner(t, page)
			var errorMu sync.Mutex
			var browserErrors []string
			expectedAvatar404 := false
			page.On("pageerror", func(exception error) {
				errorMu.Lock()
				browserErrors = append(browserErrors, "pageerror: "+exception.Error())
				errorMu.Unlock()
			})
			page.On("console", func(message playwright.ConsoleMessage) {
				if message.Type() == "error" {
					errorMu.Lock()
					browserErrors = append(browserErrors, "console: "+message.Text())
					errorMu.Unlock()
				}
			})
			page.On("response", func(response playwright.Response) {
				if entry.Path == "/components/avatar" && response.Status() == 404 && strings.HasSuffix(response.URL(), avatarBrokenImagePath) {
					errorMu.Lock()
					expectedAvatar404 = true
					errorMu.Unlock()
				}
			})

			response, err := page.Goto(baseURL+entry.Path, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, 200, response.Status())
			require.NoError(t, waitForAlpine(page))
			require.Equal(t, 1, mustLocatorCount(t, page.Locator("#main-content")))
			require.NotEmpty(t, mustText(t, page.Locator("#main-content")), "initial HTML for %s", entry.Path)

			for _, theme := range themes {
				for _, dark := range []bool{false, true} {
					result, err := page.Evaluate(`([theme, dark]) => {
						const root = document.documentElement;
						root.setAttribute('data-theme', theme);
						root.classList.toggle('dark', dark);
						const body = getComputedStyle(document.body);
						return {
							theme: root.getAttribute('data-theme'),
							dark: root.classList.contains('dark'),
							surface: getComputedStyle(root).getPropertyValue('--color-surface').trim(),
							color: body.color,
							background: body.backgroundColor,
						};
					}`, []any{theme.Key, dark})
					require.NoError(t, err)
					state := result.(map[string]any)
					require.Equal(t, theme.Key, state["theme"], "%s theme", entry.Path)
					require.Equal(t, dark, state["dark"], "%s dark mode", entry.Path)
					require.NotEmpty(t, state["surface"], "%s %s dark=%t semantic surface", entry.Path, theme.Key, dark)
					require.NotEmpty(t, state["color"], "%s %s dark=%t computed color", entry.Path, theme.Key, dark)
					require.NotEmpty(t, state["background"], "%s %s dark=%t computed background", entry.Path, theme.Key, dark)
				}
			}

			for _, width := range conformanceViewportWidths {
				require.NoError(t, page.SetViewportSize(width, 900))
				got, err := page.Evaluate(`() => ({width: innerWidth, main: !!document.querySelector('#main-content')})`, nil)
				require.NoError(t, err)
				viewport := got.(map[string]any)
				require.Equal(t, width, viewport["width"], "%s viewport edge", entry.Path)
				require.Equal(t, true, viewport["main"], "%s main at viewport %d", entry.Path, width)
			}

			session, err := page.Context().NewCDPSession(page)
			require.NoError(t, err)
			for _, zoom := range []float64{1, 2} {
				_, err := session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": zoom})
				require.NoError(t, err)
				_, err = page.WaitForFunction(`zoom => Math.abs((window.visualViewport?.scale || 1) - zoom) < 0.05`, zoom)
				require.NoError(t, err, "%s zoom %.0f", entry.Path, zoom)
			}
			_, err = session.Send("Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
			require.NoError(t, err)

			for _, motion := range []struct {
				value *playwright.ReducedMotion
				want  bool
			}{
				{value: playwright.ReducedMotionNoPreference, want: false},
				{value: playwright.ReducedMotionReduce, want: true},
			} {
				require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: motion.value}))
				got, err := page.Evaluate(`() => matchMedia('(prefers-reduced-motion: reduce)').matches`, nil)
				require.NoError(t, err)
				require.Equal(t, motion.want, got, "%s runtime reduced-motion", entry.Path)
			}

			errorMu.Lock()
			if expectedAvatar404 {
				browserErrors = removeOne(browserErrors, "console: Failed to load resource: the server responded with a status of 404 (Not Found)")
			}
			sort.Strings(browserErrors)
			require.Empty(t, browserErrors, "%s browser errors: %v", entry.Path, browserErrors)
			errorMu.Unlock()
		})
	}
}

func removeOne(values []string, target string) []string {
	for index, value := range values {
		if value == target {
			return append(values[:index], values[index+1:]...)
		}
	}
	return values
}

func TestConformanceLedgerViewportAuthority(t *testing.T) {
	want := []int{390, 639, 640, 641, 704, 767, 768, 769, 896, 1023, 1024, 1025, 1152, 1279, 1280, 1281, 1408, 1440, 1535, 1536, 1537}
	require.Equal(t, want, conformanceViewportWidths, fmt.Sprintf("compiled breakpoint edge matrix: %v", conformanceViewportWidths))
}
