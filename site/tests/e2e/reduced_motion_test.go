//go:build e2e && (full || reducedmotion)

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/carousel"
	"github.com/araihu/goshtoso/components/head"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/components/tooltip"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type motionMode struct {
	name    string
	theme   string
	dark    bool
	reduced bool
}

type transitionProbe struct {
	Index              int
	DurationMS         float64
	Property           string
	TransitionRunCount int
	Opacity            float64
	SettleMS           float64
	Frames             int
}

func motionNumber(t *testing.T, value any) float64 {
	t.Helper()
	switch number := value.(type) {
	case int:
		return float64(number)
	case int64:
		return float64(number)
	case float64:
		return number
	default:
		require.FailNowf(t, "unexpected Playwright number type", "got %T (%v)", value, value)
		return 0
	}
}

func TestReducedMotionSuppressesCarouselAutoplay(t *testing.T) {
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}))

	_, err := page.Goto(baseURL+"/components/carousel", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	require.NoError(t, waitForCarouselIndex(page, "#carousel-autoplay-c", 1))

	time.Sleep(4250 * time.Millisecond)
	state, err := page.Locator("#carousel-autoplay-c").Evaluate(`element => {
		const data = Alpine.$data(element);
		return {
			index: data.currentSlideIndex,
			timerActive: data.autoplayInterval !== null,
		};
	}`, nil)
	require.NoError(t, err)
	values := state.(map[string]any)
	assert.EqualValues(t, 1, values["index"], "reduced motion must not advance autoplay after four seconds")
	assert.Equal(t, false, values["timerActive"], "reduced motion must not retain an autoplay timer")

	_, err = page.Locator("#carousel-autoplay-c").Evaluate(`element => {
		Alpine.$data(element).currentSlideIndex = 1;
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, page.Locator("#carousel-autoplay-c button[aria-label='next slide']").Click())
	require.NoError(t, waitForCarouselIndex(page, "#carousel-autoplay-c", 2))
}

func TestMotionBehaviorAcrossFoundationThemes(t *testing.T) {
	server := httptest.NewServer(motionHandler(t))
	t.Cleanup(server.Close)

	modes := make([]motionMode, 0, 12)
	for _, theme := range []string{"araihu", "modern", "goshtoso"} {
		for _, dark := range []bool{false, true} {
			for _, reduced := range []bool{false, true} {
				mode := motionMode{theme: theme, dark: dark, reduced: reduced}
				mode.name = fmt.Sprintf("%s-dark=%t-reduced=%t", theme, dark, reduced)
				modes = append(modes, mode)
			}
		}
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			page := newPage(t, sharedBrowser)
			motion := playwright.ReducedMotionNoPreference
			if mode.reduced {
				motion = playwright.ReducedMotionReduce
			}
			require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: motion}))
			_, err := page.Goto(fmt.Sprintf("%s/?theme=%s&dark=%t", server.URL, mode.theme, mode.dark), playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
			require.NoError(t, err)
			_, err = page.WaitForFunction(`() => typeof Alpine !== 'undefined' && Alpine.$data(document.querySelector('#motion-carousel')).slides.length === 3`, nil)
			require.NoError(t, err)

			assertCarouselMotionMode(t, page, mode)
			assertCardCarouselMotionMode(t, page, mode)
			assertToastMotionMode(t, page, mode)
			assertTooltipMotionMode(t, page, mode)
		})
	}
}

func TestMotionRuntimeTracksPreferenceChanges(t *testing.T) {
	server := httptest.NewServer(motionHandler(t))
	t.Cleanup(server.Close)
	page := newPage(t, sharedBrowser)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionNoPreference}))
	_, err := page.Goto(server.URL+"/?theme=goshtoso&dark=false", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => {
		const carousel = Alpine.$data(document.querySelector('#motion-carousel'));
		const toasts = Alpine.$data(document.querySelector('#motion-toasts'));
		return carousel.autoplayInterval !== null && !carousel.reducedMotion && !toasts.reducedMotion;
	}`, nil)
	require.NoError(t, err)

	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionReduce}))
	_, err = page.WaitForFunction(`() => {
		const carousel = Alpine.$data(document.querySelector('#motion-carousel'));
		const toasts = Alpine.$data(document.querySelector('#motion-toasts'));
		return carousel.reducedMotion && carousel.autoplayInterval === null && toasts.reducedMotion;
	}`, nil)
	require.NoError(t, err)
	_, err = page.Locator("#motion-carousel").Evaluate(`element => { Alpine.$data(element).currentSlideIndex = 1 }`, nil)
	require.NoError(t, err)
	time.Sleep(550 * time.Millisecond)
	index, err := page.Locator("#motion-carousel").Evaluate(`element => Alpine.$data(element).currentSlideIndex`, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 1, index, "live reduce preference must cancel autoplay")
	require.NoError(t, page.Locator("#motion-carousel button[aria-label='next slide']").Click())
	require.NoError(t, waitForCarouselIndex(page, "#motion-carousel", 2))

	_, err = page.Locator("#motion-carousel").Evaluate(`element => {
		const data = Alpine.$data(element);
		data.currentSlideIndex = 1;
		data.motionTestAdvances = 0;
		const next = data.next;
		data.next = function () { this.motionTestAdvances += 1; return next.call(this) };
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: playwright.ReducedMotionNoPreference}))
	_, err = page.WaitForFunction(`() => {
		const carousel = Alpine.$data(document.querySelector('#motion-carousel'));
		const toasts = Alpine.$data(document.querySelector('#motion-toasts'));
		return !carousel.reducedMotion && carousel.autoplayInterval !== null && carousel.motionTestAdvances > 0 && !toasts.reducedMotion;
	}`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(2000)})
	require.NoError(t, err)
}

func assertCarouselMotionMode(t *testing.T, page playwright.Page, mode motionMode) {
	t.Helper()
	root := page.Locator("#motion-carousel")
	state, err := root.Evaluate(`element => {
		const data = Alpine.$data(element);
		data.motionTestAdvances = 0;
		const next = data.next;
		data.next = function () { this.motionTestAdvances += 1; return next.call(this) };
		return { reduced: data.reducedMotion, timer: data.autoplayInterval !== null };
	}`, nil)
	require.NoError(t, err)
	initial := state.(map[string]any)
	assert.Equal(t, mode.reduced, initial["reduced"])

	if mode.reduced {
		time.Sleep(550 * time.Millisecond)
		state, err = root.Evaluate(`element => {
			const data = Alpine.$data(element);
			return { advances: data.motionTestAdvances, timer: data.autoplayInterval !== null };
		}`, nil)
		require.NoError(t, err)
		result := state.(map[string]any)
		assert.EqualValues(t, 0, result["advances"], "reduced autoplay advances")
		assert.Equal(t, false, result["timer"], "reduced autoplay timer")
	} else {
		_, err = page.WaitForFunction(`() => Alpine.$data(document.querySelector('#motion-carousel')).motionTestAdvances > 0`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1500)})
		require.NoError(t, err, "normal mode must preserve autoplay")
	}

	_, err = root.Evaluate(`element => {
		const data = Alpine.$data(element);
		clearInterval(data.autoplayInterval);
		data.autoplayInterval = null;
		data.isPaused = true;
		data.currentSlideIndex = 1;
	}`, nil)
	require.NoError(t, err)
	probe := probeCarouselTransition(t, root)
	assert.Equal(t, 2, probe.Index, "manual Carousel next")
	assertTransitionProbe(t, mode, "Carousel", probe, 600)
	t.Logf("motion family=carousel mode=%s transition_property=%s duration_ms=%.1f transition_runs=%d opacity=%.3f settle_ms=%.1f frames=%d", mode.name, probe.Property, probe.DurationMS, probe.TransitionRunCount, probe.Opacity, probe.SettleMS, probe.Frames)

	_, err = root.Evaluate(`element => { Alpine.$data(element).currentSlideIndex = 1 }`, nil)
	require.NoError(t, err)
	next := root.Locator("button[aria-label='next slide']")
	require.NoError(t, next.Focus())
	require.NoError(t, page.Keyboard().Press("Enter"))
	require.NoError(t, waitForCarouselIndex(page, "#motion-carousel", 2))

	touchIndex, err := root.Evaluate(`element => {
		const data = Alpine.$data(element);
		data.currentSlideIndex = 1;
		data.handleTouchStart({touches: [{clientX: 240}]});
		data.handleTouchMove({touches: [{clientX: 120}]});
		data.handleTouchEnd();
		return data.currentSlideIndex;
	}`, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 2, touchIndex, "Carousel touch movement")
}

func assertCardCarouselMotionMode(t *testing.T, page playwright.Page, mode motionMode) {
	t.Helper()
	root := page.Locator("#motion-card")
	timer, err := root.Evaluate(`element => Alpine.$data(element).autoplayInterval !== null`, nil)
	require.NoError(t, err)
	assert.Equal(t, false, timer, "Card Carousel must remain manual-only")

	probe := probeCarouselTransition(t, root)
	assert.Equal(t, 2, probe.Index, "manual Card Carousel next")
	assertTransitionProbe(t, mode, "Card Carousel", probe, 250)
	t.Logf("motion family=card-carousel mode=%s transition_property=%s duration_ms=%.1f transition_runs=%d opacity=%.3f settle_ms=%.1f frames=%d", mode.name, probe.Property, probe.DurationMS, probe.TransitionRunCount, probe.Opacity, probe.SettleMS, probe.Frames)

	_, err = root.Evaluate(`element => { Alpine.$data(element).currentSlideIndex = 1 }`, nil)
	require.NoError(t, err)
	next := root.Locator("button[aria-label='next slide']")
	require.NoError(t, next.Focus())
	require.NoError(t, page.Keyboard().Press("Enter"))
	require.NoError(t, waitForCarouselIndex(page, "#motion-card", 2))

	touchIndex, err := root.Evaluate(`element => {
		const data = Alpine.$data(element);
		data.currentSlideIndex = 1;
		data.handleTouchStart({touches: [{clientX: 240}]});
		data.handleTouchMove({touches: [{clientX: 120}]});
		data.handleTouchEnd();
		return data.currentSlideIndex;
	}`, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 2, touchIndex, "Card Carousel touch movement")
}

func probeCarouselTransition(t *testing.T, root playwright.Locator) transitionProbe {
	t.Helper()
	value, err := root.Evaluate(`async element => {
		const data = Alpine.$data(element);
		data.currentSlideIndex = 1;
		let transitionRuns = 0;
		const slides = Array.from(element.querySelectorAll("div[x-show='currentSlideIndex == index + 1']"));
		for (const slide of slides) {
			slide.addEventListener('transitionrun', () => transitionRuns++);
		}
		const startedAt = performance.now();
		element.querySelector("button[aria-label='next slide']").click();
		let activeStyle;
		let frames = 0;
		while (frames < 120) {
			frames++;
			await new Promise(requestAnimationFrame);
			activeStyle = getComputedStyle(slides[data.currentSlideIndex - 1]);
			if (Number.parseFloat(activeStyle.opacity) > .99) break;
		}
		const duration = value => Math.max(...value.split(',').map(part => {
			part = part.trim();
			return part.endsWith('ms') ? Number.parseFloat(part) : Number.parseFloat(part) * 1000;
		}));
		return {
			index: data.currentSlideIndex,
			duration: duration(activeStyle.transitionDuration),
			property: activeStyle.transitionProperty,
			transitionRuns,
			opacity: Number.parseFloat(activeStyle.opacity),
			settle: performance.now() - startedAt,
			frames,
		};
	}`, nil)
	require.NoError(t, err)
	result := value.(map[string]any)
	return transitionProbe{
		Index:              int(motionNumber(t, result["index"])),
		DurationMS:         motionNumber(t, result["duration"]),
		Property:           result["property"].(string),
		TransitionRunCount: int(motionNumber(t, result["transitionRuns"])),
		Opacity:            motionNumber(t, result["opacity"]),
		SettleMS:           motionNumber(t, result["settle"]),
		Frames:             int(motionNumber(t, result["frames"])),
	}
}

func assertTransitionProbe(t *testing.T, mode motionMode, family string, probe transitionProbe, normalMinimum float64) {
	t.Helper()
	if mode.reduced {
		assert.Equalf(t, "none", probe.Property, "%s reduced transition property in %s", family, mode.name)
		assert.Equalf(t, 0, probe.TransitionRunCount, "%s reduced transition events in %s", family, mode.name)
		assert.InDeltaf(t, 1.0, probe.Opacity, 0.01, "%s reduced final opacity in %s", family, mode.name)
		assert.LessOrEqualf(t, probe.SettleMS, 75.0, "%s reduced time to final state in %s", family, mode.name)
		return
	}
	assert.NotEqualf(t, "none", probe.Property, "%s normal transition property in %s", family, mode.name)
	assert.Greaterf(t, probe.TransitionRunCount, 0, "%s normal transition events in %s", family, mode.name)
	assert.GreaterOrEqualf(t, probe.SettleMS, normalMinimum, "%s normal time to final state in %s", family, mode.name)
}

func assertToastMotionMode(t *testing.T, page playwright.Page, mode motionMode) {
	t.Helper()
	container := page.Locator("#motion-toasts")
	reduced, err := container.Evaluate(`element => Alpine.$data(element).reducedMotion`, nil)
	require.NoError(t, err)
	assert.Equal(t, mode.reduced, reduced)

	value, err := page.Evaluate(`async () => {
		let transitionRuns = 0;
		const countTransition = event => {
			if (event.target.closest('#motion-toasts')) transitionRuns++;
		};
		document.addEventListener('transitionrun', countTransition, true);
		window.dispatchEvent(new CustomEvent('notify', { detail: { kind: 'toast', tone: 'success', title: 'Manual motion', message: 'Dismiss me' } }));
		let alert = null;
		for (let attempt = 0; attempt < 30 && !alert; attempt++) {
			await new Promise(requestAnimationFrame);
			alert = document.querySelector('#motion-toasts [role="alert"]');
		}
		const startedAt = performance.now();
		let style;
		let frames = 0;
		while (frames < 120) {
			frames++;
			await new Promise(requestAnimationFrame);
			style = getComputedStyle(alert);
			if (style.translate === 'none' || style.translate === '0px' || style.translate === '0px 0px') break;
		}
		const duration = value => Math.max(...value.split(',').map(part => part.trim().endsWith('ms') ? Number.parseFloat(part) : Number.parseFloat(part) * 1000));
		document.removeEventListener('transitionrun', countTransition, true);
		return {
			duration: duration(style.transitionDuration),
			property: style.transitionProperty,
			transitionRuns,
			opacity: Number.parseFloat(style.opacity),
			settle: performance.now() - startedAt,
			frames,
		};
	}`, nil)
	require.NoError(t, err)
	transitionResult := value.(map[string]any)
	toastProbe := transitionProbe{
		DurationMS:         motionNumber(t, transitionResult["duration"]),
		Property:           transitionResult["property"].(string),
		TransitionRunCount: int(motionNumber(t, transitionResult["transitionRuns"])),
		Opacity:            motionNumber(t, transitionResult["opacity"]),
		SettleMS:           motionNumber(t, transitionResult["settle"]),
		Frames:             int(motionNumber(t, transitionResult["frames"])),
	}
	assertTransitionProbe(t, mode, "Toast", toastProbe, 250)

	alert := container.Locator("[role='alert']").First()
	require.NoError(t, alert.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateAttached}))
	_, err = alert.Evaluate(`element => { window.motionDismissStart = performance.now(); element.querySelector("button[aria-label='dismiss notification']").click() }`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => document.querySelectorAll('#motion-toasts [role="alert"]').length === 0`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1500)})
	require.NoError(t, err)
	elapsed, err := page.Evaluate(`() => performance.now() - window.motionDismissStart`, nil)
	require.NoError(t, err)
	dismissMS := motionNumber(t, elapsed)
	if mode.reduced {
		assert.LessOrEqual(t, dismissMS, 150.0, "reduced Toast removal latency")
	} else {
		assert.GreaterOrEqual(t, dismissMS, 300.0, "normal Toast removal latency")
	}

	_, err = page.Evaluate(`() => {
		window.motionLifecycleStart = performance.now();
		window.dispatchEvent(new CustomEvent('notify', { detail: { kind: 'toast', tone: 'info', title: 'Lifecycle motion', message: 'Auto dismiss' } }));
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, container.Locator("[role='alert']").First().WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateAttached}))
	time.Sleep(300 * time.Millisecond)
	count, err := container.Locator("[role='alert']").Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Toast display lifecycle must not collapse under reduced motion")
	_, err = page.WaitForFunction(`() => document.querySelectorAll('#motion-toasts [role="alert"]').length === 0`, nil, playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(1800)})
	require.NoError(t, err)
	lifecycleElapsed, err := page.Evaluate(`() => performance.now() - window.motionLifecycleStart`, nil)
	require.NoError(t, err)
	lifecycleMS := motionNumber(t, lifecycleElapsed)
	assert.GreaterOrEqual(t, lifecycleMS, 600.0, "Toast display duration must remain intact")
	t.Logf("motion family=toast mode=%s transition_property=%s duration_ms=%.1f transition_runs=%d opacity=%.3f settle_ms=%.1f frames=%d dismiss_ms=%.1f lifecycle_ms=%.1f", mode.name, toastProbe.Property, toastProbe.DurationMS, toastProbe.TransitionRunCount, toastProbe.Opacity, toastProbe.SettleMS, toastProbe.Frames, dismissMS, lifecycleMS)
}

func assertTooltipMotionMode(t *testing.T, page playwright.Page, mode motionMode) {
	t.Helper()
	hoverTrigger := page.Locator("button[aria-describedby='motion-hover']")
	hoverTip := page.Locator("#motion-hover")
	_, err := hoverTip.Evaluate(`element => {
		element.motionTransitionRuns = 0;
		element.motionStartedAt = performance.now();
		element.addEventListener('transitionrun', () => element.motionTransitionRuns++);
	}`, nil)
	require.NoError(t, err)
	require.NoError(t, hoverTrigger.Hover())
	hoverValue, err := hoverTip.Evaluate(`async element => {
		let style;
		let frames = 0;
		while (frames < 120) {
			frames++;
			await new Promise(requestAnimationFrame);
			style = getComputedStyle(element);
			if (Number.parseFloat(style.opacity) > .99) break;
		}
		const value = style.transitionDuration.split(',')[0].trim();
		return {
			duration: value.endsWith('ms') ? Number.parseFloat(value) : Number.parseFloat(value) * 1000,
			property: style.transitionProperty,
			transitionRuns: element.motionTransitionRuns,
			opacity: Number.parseFloat(style.opacity),
			settle: performance.now() - element.motionStartedAt,
			frames,
		};
	}`, nil)
	require.NoError(t, err)
	hoverResult := hoverValue.(map[string]any)
	hoverProbe := transitionProbe{
		DurationMS:         motionNumber(t, hoverResult["duration"]),
		Property:           hoverResult["property"].(string),
		TransitionRunCount: int(motionNumber(t, hoverResult["transitionRuns"])),
		Opacity:            motionNumber(t, hoverResult["opacity"]),
		SettleMS:           motionNumber(t, hoverResult["settle"]),
		Frames:             int(motionNumber(t, hoverResult["frames"])),
	}
	assertTransitionProbe(t, mode, "hover Tooltip", hoverProbe, 100)
	require.NoError(t, page.Mouse().Move(1, 1))
	_, err = page.WaitForFunction(`() => Number.parseFloat(getComputedStyle(document.querySelector('#motion-hover')).opacity) < .01`, nil)
	require.NoError(t, err)

	clickValue, err := page.Evaluate(`async () => {
		const trigger = document.querySelector("button[aria-describedby='motion-click']");
		const tooltip = document.querySelector('#motion-click');
		let transitionRuns = 0;
		tooltip.addEventListener('transitionrun', () => transitionRuns++);
		const startedAt = performance.now();
		trigger.click();
		let style;
		let frames = 0;
		while (frames < 120) {
			frames++;
			await new Promise(requestAnimationFrame);
			style = getComputedStyle(tooltip);
			if (Number.parseFloat(style.opacity) > .99) break;
		}
		const value = style.transitionDuration.split(',')[0].trim();
		return {
			duration: value.endsWith('ms') ? Number.parseFloat(value) : Number.parseFloat(value) * 1000,
			property: style.transitionProperty,
			transitionRuns,
			opacity: Number.parseFloat(style.opacity),
			settle: performance.now() - startedAt,
			frames,
		};
	}`, nil)
	require.NoError(t, err)
	clickResult := clickValue.(map[string]any)
	clickProbe := transitionProbe{
		DurationMS:         motionNumber(t, clickResult["duration"]),
		Property:           clickResult["property"].(string),
		TransitionRunCount: int(motionNumber(t, clickResult["transitionRuns"])),
		Opacity:            motionNumber(t, clickResult["opacity"]),
		SettleMS:           motionNumber(t, clickResult["settle"]),
		Frames:             int(motionNumber(t, clickResult["frames"])),
	}
	assertTransitionProbe(t, mode, "click Tooltip", clickProbe, 100)
	_, err = page.Evaluate(`async () => {
		document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }));
		await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
	}`, nil)
	require.NoError(t, err)
	_, err = page.WaitForFunction(`() => getComputedStyle(document.querySelector('#motion-click')).display === 'none'`, nil)
	require.NoError(t, err)
	t.Logf("motion family=tooltip mode=%s hover_property=%s hover_duration_ms=%.1f hover_transition_runs=%d hover_opacity=%.3f hover_settle_ms=%.1f hover_frames=%d click_property=%s click_duration_ms=%.1f click_transition_runs=%d click_opacity=%.3f click_settle_ms=%.1f click_frames=%d", mode.name, hoverProbe.Property, hoverProbe.DurationMS, hoverProbe.TransitionRunCount, hoverProbe.Opacity, hoverProbe.SettleMS, hoverProbe.Frames, clickProbe.Property, clickProbe.DurationMS, clickProbe.TransitionRunCount, clickProbe.Opacity, clickProbe.SettleMS, clickProbe.Frames)
}

func motionHandler(t *testing.T) http.Handler {
	t.Helper()
	body := motionFixture(t)
	dependencies := renderMotionComponent(t, head.Dependencies(head.WithLocalRuntime()))
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", assets.Handler())
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		darkClass := ""
		if request.URL.Query().Get("dark") == "true" {
			darkClass = " dark"
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(writer, `<!doctype html><html data-theme=%q class=%q><head><meta charset="utf-8">%s</head><body class="bg-surface text-on-surface dark:bg-surface-dark dark:text-on-surface-dark"><main class="space-y-8 p-8">%s</main></body></html>`, request.URL.Query().Get("theme"), darkClass, dependencies, body)
	})
	return mux
}

func motionFixture(t *testing.T) string {
	t.Helper()
	slides := []carousel.Slide{
		{ImgSrc: "/assets/motion-one.webp", ImgAlt: "One", Title: "One"},
		{ImgSrc: "/assets/motion-two.webp", ImgAlt: "Two", Title: "Two"},
		{ImgSrc: "/assets/motion-three.webp", ImgAlt: "Three", Title: "Three"},
	}
	components := []templ.Component{
		carousel.Carousel(carousel.Config{ID: "motion-carousel", Slides: slides, Autoplay: &carousel.AutoplayConfig{Interval: 180}, Touch: true, Height: "h-48"}),
		carousel.CardCarousel(carousel.CardConfig{ID: "motion-card", Slides: slides, Touch: true, Height: "h-48"}),
		toast.ToastContainer(toast.ContainerConfig{ID: "motion-toasts", DisplayDuration: 650}),
		tooltip.Tooltip("motion-hover", "Hover motion", tooltip.WithTriggerLabel("Hover motion")),
		tooltip.Tooltip("motion-click", "Click motion", tooltip.WithActivation(tooltip.ActivationClick), tooltip.WithTriggerLabel("Click motion")),
	}
	var body strings.Builder
	for _, component := range components {
		body.WriteString(renderMotionComponent(t, component))
	}
	return body.String()
}

func renderMotionComponent(t *testing.T, component templ.Component) string {
	t.Helper()
	var output strings.Builder
	require.NoError(t, component.Render(context.Background(), &output))
	return output.String()
}
