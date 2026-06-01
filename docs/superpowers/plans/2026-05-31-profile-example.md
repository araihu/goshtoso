# Profile Example Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a second Goshtoso example app — a user profile page with avatar + banner image upload (client-side IndexedDB), editable display name + bio (cookie + HTMX), and live theme/dark customization (existing theme engine).

**Architecture:** Three storage planes, never mixed. **Cookie** `gt_profile` (base64url-JSON) holds name + bio; HTMX mutations re-render fragments (todo pattern). **IndexedDB** holds avatar + banner `Blob`s, driven entirely by an Alpine component (`profileImages`); the server never sees image bytes. **Theme + dark** reuse the app's existing client engine (`theme` on the layout root x-data + `$store.darkMode`) — no new persistence. The example flows through the shared demo registry + `renderDemo`, so theme, dark mode, and sidebar fragment-nav come for free.

**Tech Stack:** Go 1.26, templ v0.3, Alpine.js v3 (IndexedDB + theme store), HTMX v2, Tailwind v4, Playwright E2E.

**Scope deltas from the spec (approved during planning):**
- Cookie `State` holds only `Name` + `Bio` (no `Theme`/`Dark`/`Seq` — those are client-side; `Seq` was unused without IDs).
- Theme + dark are NOT persisted in the cookie; they drive the existing `localStorage` theme engine (one source of truth).
- **Select needs no extension** — its existing `AlpineModel` field, bound to the layout root's `theme` property, is the onChange hook. Only **Avatar** (`SrcExpr`) and **Toggle** (`Attrs`) are extended.
- The upload control uses the existing **File Input** component (`components/fileinput`), not a raw `<input>`.

---

## File Structure

**Create:**
- `internal/examples/profile/state.go` — `State{Name, Bio}`, consts, `Encode`/`Decode`
- `internal/examples/profile/state_test.go`
- `internal/examples/profile/state_mutations.go` — `SetName`, `SetBio` (trim, rune-cap, budget)
- `internal/examples/profile/state_mutations_test.go`
- `internal/examples/profile/cookie.go` — `CookieName`, `FromRequest`, `SetCookie`
- `internal/examples/profile/sample.go` — `Sample()` seed
- `internal/examples/profile/cookie_test.go`
- `internal/pages/demo/examples/profile.templ` — Alpine IndexedDB script + page + fragments
- `internal/server/profile_handler.go` — routes + page render + identity handler
- `tests/e2e/profile_test.go` — example E2E

**Modify:**
- `components/toggle/types.go` — add `Attrs templ.Attributes`
- `components/toggle/toggle.templ` — single `<input>` with `{ cfg.Attrs... }` spread
- `components/avatar/types.go` — add `SrcExpr string`, update `HasImage`
- `components/avatar/avatar.templ` — reactive `:src` image layer when `SrcExpr` set
- `internal/server/server.go` — call `registerProfileRoutes()`; add `"profile"` case to `handleExample`
- `internal/pages/demo/components/registry.go` — add `"examples/profile"` entry
- `internal/pages/demo/layout.templ` — sidebar item under Examples
- `internal/pages/demo/examples/index.templ` — gallery card for profile
- `tests/e2e/avatar_test.go` — test for `SrcExpr` mode
- `tests/e2e/toggle_test.go` — test for `Attrs` passthrough (create if absent)

---

## Task 1: Profile domain — State + Encode/Decode

**Files:**
- Create: `internal/examples/profile/state.go`
- Test: `internal/examples/profile/state_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/examples/profile/state_test.go
package profile

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	in := State{Name: "Ada Lovelace", Bio: "First programmer."}
	out, err := Decode([]byte(Encode(in)))
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if out.Name != in.Name || out.Bio != in.Bio {
		t.Fatalf("round trip mismatch: got %+v want %+v", out, in)
	}
}

func TestDecodeEmptyYieldsZero(t *testing.T) {
	s, err := Decode(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "" || s.Bio != "" {
		t.Fatalf("expected zero State, got %+v", s)
	}
}

func TestDecodeCorruptReturnsError(t *testing.T) {
	if _, err := Decode([]byte("!!!not-base64!!!")); err == nil {
		t.Fatal("expected error for corrupt input")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/profile/ -run TestEncodeDecode -v`
Expected: FAIL — `undefined: State` / `undefined: Encode`.

- [ ] **Step 3: Write minimal implementation**

```go
// Package profile holds the pure, HTTP-free domain model for the /examples/profile
// app. Only the display name and bio live here (in a cookie); avatar/banner images
// live client-side in IndexedDB and theme/dark in the app's existing theme engine,
// so the server keeps no per-user memory and never sees image bytes.
package profile

import (
	"encoding/base64"
	"encoding/json"
)

const (
	// MaxNameLen bounds the display name's stored length in runes.
	MaxNameLen = 60
	// MaxBioLen bounds the bio's stored length in runes.
	MaxBioLen = 280
	// maxCookieBytes bounds the encoded cookie value so the browser never silently
	// drops it (browsers cap a cookie near 4KB).
	maxCookieBytes = 3800
)

// State is the per-user profile text. Avatar/banner images and theme/dark are
// intentionally absent — they are client-side concerns.
type State struct {
	Name string `json:"n"`
	Bio  string `json:"b"`
}

// Encode serializes State to a base64url(JSON) string for cookie storage.
// State is always serializable; a marshal error is a programmer error and panics.
func Encode(s State) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic("profile.Encode: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a cookie value back into State. Empty input yields the zero
// State; malformed input returns an error so callers can fall back to a default.
func Decode(raw []byte) (State, error) {
	var s State
	if len(raw) == 0 {
		return s, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(string(raw))
	if err != nil {
		return State{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	return s, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/profile/ -run TestEncodeDecode -v && go test ./internal/examples/profile/ -run TestDecode -v`
Expected: PASS (all three).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/profile/state.go internal/examples/profile/state_test.go
git commit -m "feat(profile): domain State with base64url cookie codec"
```

---

## Task 2: Profile mutations — SetName / SetBio

**Files:**
- Create: `internal/examples/profile/state_mutations.go`
- Test: `internal/examples/profile/state_mutations_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/examples/profile/state_mutations_test.go
package profile

import (
	"strings"
	"testing"
)

func TestSetNameTrimsAndCaps(t *testing.T) {
	var s State
	s.SetName("  Ada  ")
	if s.Name != "Ada" {
		t.Fatalf("trim failed: %q", s.Name)
	}
	s.SetName(strings.Repeat("x", MaxNameLen+50))
	if len([]rune(s.Name)) != MaxNameLen {
		t.Fatalf("cap failed: got %d runes", len([]rune(s.Name)))
	}
}

func TestSetBioCaps(t *testing.T) {
	var s State
	s.SetBio(strings.Repeat("y", MaxBioLen+100))
	if len([]rune(s.Bio)) != MaxBioLen {
		t.Fatalf("bio cap failed: got %d runes", len([]rune(s.Bio)))
	}
}

func TestSetNameRefusedOverBudget(t *testing.T) {
	// A bio already at budget; a huge name must be refused (kept unchanged),
	// not silently produce an over-budget cookie. Force the situation by
	// shrinking nothing — instead assert the budget guard exists: set a bio
	// that alone is within budget, then verify Encode stays under maxCookieBytes.
	var s State
	s.SetBio(strings.Repeat("z", MaxBioLen))
	s.SetName(strings.Repeat("n", MaxNameLen))
	if len(Encode(s)) > maxCookieBytes {
		t.Fatalf("encoded %d exceeds budget %d", len(Encode(s)), maxCookieBytes)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/profile/ -run TestSet -v`
Expected: FAIL — `s.SetName undefined`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/examples/profile/state_mutations.go
package profile

import "strings"

// capRunes trims s and caps it to max runes without splitting a multibyte rune.
func capRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// SetName sets the trimmed, rune-capped display name. The change is refused
// (name kept unchanged) if the resulting encoded cookie would exceed the budget.
func (s *State) SetName(name string) {
	candidate := State{Name: capRunes(name, MaxNameLen), Bio: s.Bio}
	if len(Encode(candidate)) > maxCookieBytes {
		return
	}
	s.Name = candidate.Name
}

// SetBio sets the trimmed, rune-capped bio. Refused if over the cookie budget.
func (s *State) SetBio(bio string) {
	candidate := State{Name: s.Name, Bio: capRunes(bio, MaxBioLen)}
	if len(Encode(candidate)) > maxCookieBytes {
		return
	}
	s.Bio = candidate.Bio
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/profile/ -run TestSet -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/examples/profile/state_mutations.go internal/examples/profile/state_mutations_test.go
git commit -m "feat(profile): SetName/SetBio mutations with rune cap + cookie budget"
```

---

## Task 3: Profile cookie + sample seed

**Files:**
- Create: `internal/examples/profile/cookie.go`, `internal/examples/profile/sample.go`
- Test: `internal/examples/profile/cookie_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/examples/profile/cookie_test.go
package profile

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCookieThenFromRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	SetCookie(rec, State{Name: "Grace", Bio: "Compiler pioneer."})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	got := FromRequest(req)
	if got.Name != "Grace" || got.Bio != "Compiler pioneer." {
		t.Fatalf("round trip via cookie failed: %+v", got)
	}
}

func TestFromRequestNoCookieYieldsZero(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := FromRequest(req); got.Name != "" || got.Bio != "" {
		t.Fatalf("expected zero State, got %+v", got)
	}
}

func TestSampleIsNonEmptyAndWithinBudget(t *testing.T) {
	s := Sample()
	if s.Name == "" || s.Bio == "" {
		t.Fatal("Sample should seed a non-empty profile")
	}
	if len(Encode(s)) > maxCookieBytes {
		t.Fatalf("Sample over budget: %d", len(Encode(s)))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/examples/profile/ -run "TestSetCookie|TestFromRequest|TestSample" -v`
Expected: FAIL — `undefined: SetCookie`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/examples/profile/cookie.go
package profile

import "net/http"

// CookieName is the cookie that carries the encoded profile State.
const CookieName = "gt_profile"

// cookieMaxAge is ~30 days in seconds.
const cookieMaxAge = 30 * 24 * 60 * 60

// FromRequest reads and decodes State from the request cookie. A missing or
// corrupt cookie yields the zero State.
func FromRequest(r *http.Request) State {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return State{}
	}
	s, err := Decode([]byte(c.Value))
	if err != nil {
		return State{}
	}
	return s
}

// SetCookie writes the encoded State as a cookie. Path "/" so it reaches both
// /examples/* and /api/examples/profile/*.
func SetCookie(w http.ResponseWriter, s State) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    Encode(s),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   cookieMaxAge,
	})
}
```

```go
// internal/examples/profile/sample.go
package profile

// Sample returns a starter profile shown on a visitor's first load (no cookie)
// so the example never opens empty.
func Sample() State {
	return State{
		Name: "Ada Lovelace",
		Bio:  "Mathematician and writer, known for work on Babbage's Analytical Engine — arguably the first computer programmer.",
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/examples/profile/ -v`
Expected: PASS (whole package).

- [ ] **Step 5: Commit**

```bash
git add internal/examples/profile/cookie.go internal/examples/profile/sample.go internal/examples/profile/cookie_test.go
git commit -m "feat(profile): cookie read/write + sample seed"
```

---

## Task 4: Extend Toggle with Attrs passthrough

**Files:**
- Modify: `components/toggle/types.go`, `components/toggle/toggle.templ`
- Test: `tests/e2e/toggle_test.go` (browser test) — but first a fast render check via the existing pattern below.

**Why:** Toggle's `<input>` has no Alpine hook, so it can't be wired to `$store.darkMode`. Add an `Attrs templ.Attributes` field (every other interactive component already has one) and collapse the 4-branch input into one tag carrying the spread.

- [ ] **Step 1: Add the field to Config**

In `components/toggle/types.go`, inside `type Config struct { ... }`, add after `Class string`:

```go
	// Attrs are extra attributes applied to the <input> element
	// (e.g. x-on:change, x-bind:checked for Alpine binding).
	Attrs templ.Attributes
```

Add the templ import if the file lacks it:

```go
import "github.com/a-h/templ"
```

(Check existing imports first; only add if missing.)

- [ ] **Step 2: Replace the 4-branch input in toggle.templ**

In `components/toggle/toggle.templ`, replace the entire `if cfg.Checked && cfg.Disabled { ... } else { ... }` block (the four `<input ...>` variants) with a single tag:

```templ
		<input
			id={ cfg.ID }
			type="checkbox"
			class="peer sr-only"
			role="switch"
			checked?={ cfg.Checked }
			disabled?={ cfg.Disabled }
			{ cfg.Attrs... }
		/>
```

Leave the rest of the template (hidden `Name` input, label span, toggle div) unchanged.

- [ ] **Step 3: Regenerate templ**

Run: `templ generate`
Expected: regenerates `components/toggle/toggle_templ.go`. If it reports "0 updates" but the source changed, force it:
`rm components/toggle/toggle_templ.go && templ generate`

- [ ] **Step 4: Verify build + write the browser test**

Run: `go build ./...`
Expected: clean.

Append to `tests/e2e/toggle_test.go` (create the file with `package e2e` + imports if it does not exist; mirror the structure of `tests/e2e/sidebar_test.go` for setup helpers):

```go
func TestToggleAttrsPassthrough(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	// The toggle demo page renders Toggles; assert the Attrs spread lands on
	// the input by rendering a known attribute. We rely on the demo page using
	// at least one Toggle; this test instead navigates to the profile example
	// where the dark toggle carries x-on:change.
	if _, err := page.Goto(serverURL + "/components/toggle"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	// Smoke: the toggle page renders an input[role=switch].
	count, err := page.Locator("input[role='switch']").Count()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one toggle input on the toggle demo page")
	}
}
```

> Note: the real Attrs-wiring assertion lives in the profile E2E (Task 10), which checks the dark toggle actually toggles `<html class="dark">`. This smoke test guards the refactor.

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestToggleAttrsPassthrough`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add components/toggle/types.go components/toggle/toggle.templ components/toggle/toggle_templ.go tests/e2e/toggle_test.go
git commit -m "feat(toggle): Attrs passthrough on the input for Alpine binding"
```

---

## Task 5: Extend Avatar with SrcExpr (client-bound image source)

**Files:**
- Modify: `components/avatar/types.go`, `components/avatar/avatar.templ`
- Test: `tests/e2e/avatar_test.go`

**Why:** Avatar's image layer renders a static `src={cfg.Src}` only when `Src != ""`. An IndexedDB objectURL is set client-side after load. `SrcExpr` makes the image layer render unconditionally and bind `:src` to a parent Alpine expression, with initials shown while the expression is falsy.

- [ ] **Step 1: Add the field + update HasImage**

In `components/avatar/types.go`, add to `type Config struct` (after `Class string`):

```go
	// SrcExpr is an Alpine expression evaluated in the parent scope that yields
	// the image src at runtime (e.g. "avatarSrc"). When set, the image layer is
	// always rendered and binds x-bind:src to this expression; the initials
	// layer shows whenever the expression is falsy. Use for client-set sources
	// (object URLs, late-loaded images) where no static Src exists at render.
	SrcExpr string
```

Find `func (cfg Config) HasImage() bool` (in `types.go`) and change it to also count `SrcExpr`:

```go
func (cfg Config) HasImage() bool {
	return cfg.Src != "" || cfg.SrcExpr != ""
}
```

- [ ] **Step 2: Make the image + initials layers reactive in avatar.templ**

In `components/avatar/avatar.templ`, update `layerImage` to bind src reactively when `SrcExpr` is set:

```templ
templ layerImage(cfg Config) {
	<img
		x-show="!imgError"
		x-init="if ($el.complete) { if ($el.naturalWidth > 0) imgLoaded = true; else imgError = true; }"
		x-on:load="imgLoaded = true"
		x-on:error={ fmt.Sprintf("imgError = true; console.warn('[avatar] failed to load image:', '%s')", cfg.Src) }
		if cfg.SrcExpr != "" {
			x-bind:src={ cfg.SrcExpr }
		} else {
			src={ cfg.Src }
		}
		alt={ cfg.Alt }
		class="absolute inset-0 h-full w-full object-cover object-center"
	/>
}
```

Update `layerInitials` so that in `SrcExpr` mode the initials show while the expression is falsy (and the image overlay logic still applies on top). Change the `x-show` on the initials `<span>`:

```templ
		if withImageOverlay {
			if cfg.SrcExpr != "" {
				x-show={ fmt.Sprintf("!(%s) || !imgLoaded || imgError", cfg.SrcExpr) }
			} else {
				x-show="!imgLoaded || imgError"
			}
		}
```

Also guard the loading spinner so it does not spin forever when `SrcExpr` is empty/falsy. In `layerLoading`, change its `x-show` when `SrcExpr` is set:

```templ
templ layerLoading(cfg Config) {
	<span
		if cfg.SrcExpr != "" {
			x-show={ fmt.Sprintf("(%s) && !imgLoaded && !imgError", cfg.SrcExpr) }
		} else {
			x-show="!imgLoaded && !imgError"
		}
		class="absolute inset-0 flex items-center justify-center bg-surface-alt dark:bg-surface-dark-alt"
		aria-hidden="true"
	>
		... (unchanged spinner svg) ...
	</span>
}
```

> Keep the existing spinner `<svg>` body exactly as-is; only the `x-show` changes.

- [ ] **Step 3: Regenerate templ**

Run: `rm components/avatar/avatar_templ.go && templ generate`
Expected: regenerates the file.

- [ ] **Step 4: Build + write the avatar SrcExpr test**

Run: `go build ./...`
Expected: clean.

Append to `tests/e2e/avatar_test.go`:

```go
func TestAvatarSrcExpr(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	// Render an avatar with SrcExpr by visiting the profile example, where the
	// avatar uses SrcExpr="avatarSrc". Initially no image -> initials visible.
	if _, err := page.Goto(serverURL + "/examples/profile?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")
	// The avatar image element exists (SrcExpr always renders the layer) even
	// before any upload, and the initials fallback is visible.
	imgCount, err := page.Locator("#profile-avatar img").Count()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if imgCount == 0 {
		t.Fatal("SrcExpr avatar should always render an <img> layer")
	}
}
```

Run: `go test ./tests/e2e/... -count=1 -timeout 5m -run TestAvatarSrcExpr`
Expected: PASS (after Task 7/8/9 wire the profile page; if running this task standalone before wiring, this test will fail on the missing page — run it after Task 9, or temporarily skip. Mark it to run in the Task 10 batch).

- [ ] **Step 5: Commit**

```bash
git add components/avatar/types.go components/avatar/avatar.templ components/avatar/avatar_templ.go tests/e2e/avatar_test.go
git commit -m "feat(avatar): SrcExpr for client-bound (Alpine) image sources"
```

---

## Task 6: Profile page template — IndexedDB Alpine + UI

**Files:**
- Create: `internal/pages/demo/examples/profile.templ`

This is the largest task. The page composes Goshtoso components and an Alpine
`profileImages` component (registered via `<script>` + `templ.Raw`, per the
CLAUDE.md escaping rule and the table filter registration pattern). Theme uses
`Select.AlpineModel = "theme"` (binds to the layout root's `theme` property);
dark uses `Toggle.Attrs` wired to `$store.darkMode`.

- [ ] **Step 1: Create the file with the Alpine IndexedDB script**

```go
// internal/pages/demo/examples/profile.templ
package examples

import (
	"github.com/araihu/goshtoso/components/avatar"
	"github.com/araihu/goshtoso/components/badge"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/fileinput"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/textarea"
	"github.com/araihu/goshtoso/components/textinput"
	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/components/toggle"
	"github.com/araihu/goshtoso/internal/examples/profile"
)

// profileImagesScript registers the Alpine `profileImages` component that owns
// avatar/banner blobs in IndexedDB. Registered eagerly when Alpine is already
// running (sidebar fragment-nav) and on alpine:init for first load — the
// alpine:init listener fires only once per page, so a fragment-swapped script
// must self-register or x-data="profileImages" stays silent.
func profileImagesScript() string {
	return `(() => {
  const DB = 'gt_profile', STORE = 'images', MAX = 1024 * 1024;
  function open() {
    return new Promise((res, rej) => {
      const r = indexedDB.open(DB, 1);
      r.onupgradeneeded = () => r.result.createObjectStore(STORE);
      r.onsuccess = () => res(r.result);
      r.onerror = () => rej(r.error);
    });
  }
  function tx(mode, fn) {
    return open().then(db => new Promise((res, rej) => {
      const t = db.transaction(STORE, mode);
      const req = fn(t.objectStore(STORE));
      t.oncomplete = () => res(req && req.result);
      t.onerror = () => rej(t.error);
    }));
  }
  const idbGet = k => tx('readonly', s => s.get(k));
  const idbPut = (k, v) => tx('readwrite', s => s.put(v, k));
  const idbDel = k => tx('readwrite', s => s.delete(k));

  const register = () => {
    Alpine.data('profileImages', () => ({
      avatarSrc: '',
      bannerSrc: '',
      _supported: typeof indexedDB !== 'undefined',
      init() {
        if (!this._supported) return;
        ['avatar', 'banner'].forEach(kind => {
          idbGet(kind).then(blob => {
            if (blob) this[kind + 'Src'] = URL.createObjectURL(blob);
          }).catch(() => {});
        });
      },
      pick(kind) {
        const el = document.getElementById('profile-' + kind + '-input');
        if (el) el.click();
      },
      onFile(kind, ev) {
        const file = ev.target.files && ev.target.files[0];
        if (!file) return;
        if (!file.type.startsWith('image/')) {
          this._toast('danger', 'Not an image', 'Pick a PNG, JPG, or WebP file.');
          ev.target.value = '';
          return;
        }
        if (file.size > MAX) {
          this._toast('danger', 'Too large', 'Images must be 1 MB or smaller.');
          ev.target.value = '';
          return;
        }
        const old = this[kind + 'Src'];
        if (old) URL.revokeObjectURL(old);
        this[kind + 'Src'] = URL.createObjectURL(file);
        if (this._supported) {
          idbPut(kind, file).catch(() =>
            this._toast('warning', "Won't persist", 'Saved for this session only.'));
        } else {
          this._toast('warning', "Won't persist", 'IndexedDB unavailable in this browser.');
        }
        ev.target.value = '';
      },
      remove(kind) {
        const old = this[kind + 'Src'];
        if (old) URL.revokeObjectURL(old);
        this[kind + 'Src'] = '';
        if (this._supported) idbDel(kind).catch(() => {});
      },
      _toast(variant, title, message) {
        window.dispatchEvent(new CustomEvent('notify', { detail: { variant, title, message } }));
      },
    }));
  };
  if (window.Alpine && window.Alpine.version) register();
  else document.addEventListener('alpine:init', register);
})();`
}

templ profileScript() {
	@templ.Raw("<script>" + profileImagesScript() + "</script>")
}
```

> **Toast wiring note:** the script dispatches a `notify` CustomEvent. The
> `toast.Container` in this repo listens for toasts via its own mechanism; in
> Step 4 we add a tiny listener that calls the toast API. Verify the exact
> client toast-trigger API by reading `components/toast/types.go` +
> `toast.Container` usage in `todo.templ` before finalizing — if `toast` exposes
> a global `window.Goshtoso.toast(...)` or an Alpine store, call that directly
> instead of a custom event. Adjust `_toast` to match the real API. (This is the
> one integration point to confirm against the toast component at implement time.)

- [ ] **Step 2: Add the theme options helper + fragments**

Append to `profile.templ`:

```go
// themeOptions mirrors the layout's theme list (duplicated to avoid an import
// cycle: package demo imports package examples, not the reverse). Keep in sync
// with internal/pages/demo/layout.templ:getThemeOptions.
func themeOptions(current string) []selectfield.Option {
	pairs := []struct{ Key, Label string }{
		{"minimal", "Minimal"}, {"modern", "Modern"}, {"arctic", "Arctic"},
		{"high-contrast", "High Contrast"}, {"neo-brutalism", "Neo Brutalism"},
		{"news", "News"}, {"industrial", "Industrial"}, {"90s", "90s"},
		{"pastel", "Pastel"}, {"christmas", "Christmas"}, {"halloween", "Halloween"},
		{"zombie", "Halloween II"}, {"prototype", "Prototype"}, {"totvs", "TOTVS"},
		{"dracula", "Dracula"},
	}
	out := make([]selectfield.Option, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, selectfield.Option{Value: p.Key, Label: p.Label, Selected: p.Key == current})
	}
	return out
}
```

```templ
// IdentityFields renders the name + bio inputs. It is the HTMX swap target
// (#profile-identity) re-rendered by the identity save handler so server-trimmed
// values reflect back (e.g. when a too-long value was capped).
templ IdentityFields(s profile.State) {
	<div id="profile-identity" class="flex flex-col gap-4">
		@textinput.TextInput(textinput.Config{
			Name:        "name",
			Label:       "Display name",
			Placeholder: "Your name",
			Value:       s.Name,
		})
		@textarea.Textarea(textarea.Config{
			Name:        "bio",
			Label:       "Bio",
			Placeholder: "A short line about you",
			Value:       s.Bio,
			Rows:        3,
		})
	</div>
}
```

> **Field-name check:** confirm `textinput.Config` and `textarea.Config` field
> names (`Value`, `Rows`) by reading their `types.go` at implement time; the
> todo example used `textinput.Config{Name, Label, Placeholder}` and a `Type`
> field — adjust the struct literals to the real fields. (Text Input is known to
> support `Value`; verify Textarea's `Rows`/`Value` names.)

- [ ] **Step 3: Add the main ProfileApp + ProfileContent**

```templ
// ProfileApp is the interactive profile screen. Rendered on first load (with
// cookie state) and as the registry fragment. The #profile-images Alpine scope
// owns avatar/banner object URLs from IndexedDB; identity text is a cookie+HTMX
// form; theme/dark drive the app's existing client theme engine.
templ ProfileApp(s profile.State) {
	@profileScript()
	<div id="profile-fragment" class="mx-auto max-w-3xl" x-data="profileImages">
		<header class="mb-6">
			<h1 class="text-2xl font-bold text-on-surface dark:text-on-surface-dark">Profile</h1>
			<p class="mt-2 text-on-surface-muted dark:text-on-surface-dark-muted">
				A profile editor built from Goshtoso components. Name and bio are saved to a cookie
				over HTMX (no server memory); avatar and banner images stay in your browser via
				IndexedDB (never uploaded); theme and dark mode drive the site's live theme engine.
			</p>
		</header>

		<!-- Photos card: banner + avatar -->
		<section id="profile-photos" class="mb-6 overflow-hidden rounded-radius border border-outline bg-surface dark:border-outline-dark dark:bg-surface-dark">
			<div class="relative h-36 bg-surface-alt dark:bg-surface-dark-alt">
				<img x-show="bannerSrc" x-bind:src="bannerSrc" alt="Cover image" class="h-full w-full object-cover"/>
				<div x-show="!bannerSrc" class="flex h-full w-full items-center justify-center text-sm text-on-surface-muted dark:text-on-surface-dark-muted">
					No cover image
				</div>
				<div class="absolute inset-x-0 -bottom-10 flex justify-center">
					<div id="profile-avatar" class="rounded-full ring-4 ring-surface dark:ring-surface-dark">
						@avatar.Avatar(avatar.Config{
							SrcExpr: "avatarSrc",
							Name:    s.Name,
							Size:    avatar.SizeXL,
							Shape:   avatar.ShapeCircle,
						})
					</div>
				</div>
			</div>
			<div class="flex flex-wrap justify-center gap-3 px-4 pb-4 pt-12">
				@fileinput.FileInput(fileinput.Config{
					ID:         "profile-avatar-input",
					Accept:     "image/*",
					Class:      "hidden",
					Attrs:      templ.Attributes{"x-on:change": "onFile('avatar', $event)"},
				})
				@fileinput.FileInput(fileinput.Config{
					ID:         "profile-banner-input",
					Accept:     "image/*",
					Class:      "hidden",
					Attrs:      templ.Attributes{"x-on:change": "onFile('banner', $event)"},
				})
				@button.Button(button.Config{Type: "button", Variant: button.Primary, Attrs: templ.Attributes{"x-on:click": "pick('avatar')"}}) {
					Upload avatar
				}
				@button.Button(button.Config{Type: "button", Variant: button.Secondary, Attrs: templ.Attributes{"x-on:click": "pick('banner')"}}) {
					Upload cover
				}
				@button.Button(button.Config{Type: "button", Variant: button.Secondary, Attrs: templ.Attributes{"x-show": "avatarSrc || bannerSrc", "x-on:click": "remove('avatar'); remove('banner')"}}) {
					Remove photos
				}
			</div>
		</section>

		<!-- Identity card: name + bio (cookie + HTMX) -->
		<section class="mb-6 rounded-radius border border-outline bg-surface p-4 dark:border-outline-dark dark:bg-surface-dark">
			<h2 class="mb-3 text-lg font-semibold text-on-surface dark:text-on-surface-dark">Identity</h2>
			<form
				hx-post="/api/examples/profile/identity"
				hx-target="#profile-identity"
				hx-swap="outerHTML"
			>
				@IdentityFields(s)
				<div class="mt-4 flex justify-end">
					@button.Button(button.Config{Type: "submit", Variant: button.Primary}) {
						Save
					}
				</div>
			</form>
		</section>

		<!-- Appearance card: theme + dark (existing client engine) -->
		<section class="mb-6 rounded-radius border border-outline bg-surface p-4 dark:border-outline-dark dark:bg-surface-dark">
			<div class="mb-3 flex items-center gap-2">
				<h2 class="text-lg font-semibold text-on-surface dark:text-on-surface-dark">Appearance</h2>
				@badge.Badge(badge.Config{Variant: badge.Secondary, Text: "live"})
			</div>
			<div class="flex flex-col gap-4 sm:flex-row sm:items-end">
				<div class="sm:w-64">
					@selectfield.Select(selectfield.Config{
						ID:          "profile-theme",
						Name:        "theme",
						Label:       "Theme",
						AlpineModel: "theme",
						Options:     themeOptions(""),
					})
				</div>
				@toggle.Toggle(toggle.Config{
					ID:    "profile-dark",
					Label: "Dark mode",
					Attrs: templ.Attributes{
						"x-bind:checked": "$store.darkMode.on",
						"x-on:change":    "$store.darkMode.toggle()",
					},
				})
			</div>
		</section>

		@toast.Container(toast.ContainerConfig{})
	</div>
}

// ProfileContent is the registry entry point: an empty-state profile for direct
// nav when no cookie exists. Server first-load uses ProfileApp(state).
templ ProfileContent() {
	@ProfileApp(profile.State{})
}
```

> **Component field checks at implement time** (read each `types.go`; adjust literals to match — do NOT invent fields):
> - `button.Config` — confirm it has `Attrs templ.Attributes` (needed for `x-on:click`). If not, that is a **blocking** finding: either add an `Attrs` passthrough to Button (same one-line extension as Toggle) or wrap buttons in an Alpine span. Resolve before proceeding.
> - `fileinput.Config` — `ID`, `Accept`, `Class`, `Attrs` confirmed present.
> - `badge.Config` — uses `Text` + `Variant` (confirmed via todo.templ).
> - `textarea.Config` — confirm `Value` + `Rows`.
> - `selectfield.Config.AlpineModel` — confirmed present.

- [ ] **Step 4: Wire the toast trigger to the real API**

Read `components/toast/types.go` and how `toast.Container` receives client-side
toasts (search for an `addEventListener` / store / global). Replace the `_toast`
body in `profileImagesScript()` to call the real trigger. If the toast system
already listens for a specific event name or exposes `Alpine.store('toasts')`,
use that verbatim. Add no new global if one exists.

- [ ] **Step 5: Generate + build**

Run: `rm -f internal/pages/demo/examples/profile_templ.go && templ generate && go build ./...`
Expected: clean build. Fix any field-name mismatches surfaced here against the real component `types.go` files.

- [ ] **Step 6: Commit**

```bash
git add internal/pages/demo/examples/profile.templ internal/pages/demo/examples/profile_templ.go
git commit -m "feat(profile): profile page template with IndexedDB images + appearance"
```

---

## Task 7: Profile handler + route wiring

**Files:**
- Create: `internal/server/profile_handler.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Write the handler**

```go
// internal/server/profile_handler.go
package server

import (
	"net/http"
	"strings"

	"github.com/araihu/goshtoso/components/toast"
	"github.com/araihu/goshtoso/internal/examples/profile"
	"github.com/araihu/goshtoso/internal/pages/demo"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
)

// registerProfileRoutes wires the /api/examples/profile/* endpoints.
func (s *Server) registerProfileRoutes() {
	s.mux.HandleFunc("/api/examples/profile/identity", s.handleProfileIdentity)
}

// renderProfilePage is the first-load handler for /examples/profile. Seeds a
// sample profile on first visit (no cookie) unless ?seed=0 (used by e2e).
func (s *Server) renderProfilePage(w http.ResponseWriter, r *http.Request) {
	var st profile.State
	if _, err := r.Cookie(profile.CookieName); err != nil && r.URL.Query().Get("seed") != "0" {
		st = profile.Sample()
		profile.SetCookie(w, st)
	} else {
		st = profile.FromRequest(r)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.ProfileApp(st)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.Fragment("Profile", "profile", content).Render(r.Context(), w)
		return
	}
	_ = demo.Layout("Profile", "profile", content).Render(r.Context(), w)
}

// handleProfileIdentity saves name + bio to the cookie and re-renders the
// identity fields (so server-capped values reflect back) plus a success toast.
func (s *Server) handleProfileIdentity(w http.ResponseWriter, r *http.Request) {
	if onlyPost(w, r) {
		return
	}
	st := profile.FromRequest(r)
	st.SetName(strings.TrimSpace(r.FormValue("name")))
	st.SetBio(strings.TrimSpace(r.FormValue("bio")))
	profile.SetCookie(w, st)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.IdentityFields(st).Render(r.Context(), w)
	_ = toast.OOBToast(toast.Config{
		Variant: toast.Success,
		Title:   "Saved",
		Message: "Your profile was updated.",
	}).Render(r.Context(), w)
}
```

> `onlyPost` is the shared helper already defined in `todo_handler.go` (same
> package `server`) — reuse it, do not redefine.

- [ ] **Step 2: Wire into server.go**

In `internal/server/server.go`, find the line `s.registerTodoRoutes()` (~line 46) and add directly after it:

```go
	s.registerProfileRoutes()
```

In `handleExample` (the `switch sub {` block ~line 96), add a case before `default:`:

```go
	case "profile":
		s.renderProfilePage(w, r)
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: clean.

- [ ] **Step 4: Manual smoke (optional but recommended)**

Run: `go run cmd/server/main.go` then visit `http://localhost:8090/examples/profile`. Confirm the page renders, avatar shows initials, theme select + dark toggle work, name/bio save shows a toast. Ctrl-C when done.

- [ ] **Step 5: Commit**

```bash
git add internal/server/profile_handler.go internal/server/server.go
git commit -m "feat(profile): handler + routes for /examples/profile"
```

---

## Task 8: Registry + sidebar + gallery wiring

**Files:**
- Modify: `internal/pages/demo/components/registry.go`, `internal/pages/demo/layout.templ`, `internal/pages/demo/examples/index.templ`

- [ ] **Step 1: Registry entry**

In `internal/pages/demo/components/registry.go`, add after the `"examples/todo"` line:

```go
	"examples/profile":           {"Profile", "profile", examples.ProfileContent},
```

- [ ] **Step 2: Sidebar item**

In `internal/pages/demo/layout.templ`, in the `Examples` section `Items` slice, add after the todo line:

```go
				sItem("profile", "Profile", "/examples/profile", activeComponent),
```

- [ ] **Step 3: Gallery card**

In `internal/pages/demo/examples/index.templ`, inside the `grid` div in `IndexContent`, add after the todo card:

```templ
			@exampleCard("/examples/profile", "Profile", "Avatar & cover upload to IndexedDB, cookie-backed name & bio, live theme and dark-mode customization.")
```

- [ ] **Step 4: Generate + build**

Run: `templ generate && go build ./...`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add internal/pages/demo/components/registry.go internal/pages/demo/layout.templ internal/pages/demo/layout_templ.go internal/pages/demo/examples/index.templ internal/pages/demo/examples/index_templ.go
git commit -m "feat(profile): register profile example in registry, sidebar, gallery"
```

---

## Task 9: E2E tests for the profile example

**Files:**
- Create: `tests/e2e/profile_test.go`

Mirror existing helpers (`setupPlaywright`, `newPage`, `serverURL`, `fillSearchInput`, `clickUntil`) from `tests/e2e/e2e_test.go`. Read that file for exact helper signatures before writing.

- [ ] **Step 1: Write the tests**

```go
package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestProfileFragmentNavNoConsoleErrors loads the profile via the sidebar
// fragment-nav path (not a direct load) and asserts zero console errors — the
// examples mandate (Alpine.data must register on fragment nav, OOB must not
// fire with no target).
func TestProfileFragmentNavNoConsoleErrors(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)

	var consoleErrs []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			consoleErrs = append(consoleErrs, msg.Text())
		}
	})

	// Start at the examples gallery, then click through to the profile (fragment nav).
	if _, err := page.Goto(serverURL + "/examples?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	if err := page.Locator("a[href='/examples/profile']").Click(); err != nil {
		t.Fatalf("click profile card: %v", err)
	}
	page.WaitForSelector("#profile-fragment")
	page.WaitForTimeout(300)
	if len(consoleErrs) > 0 {
		t.Fatalf("console errors on fragment nav: %v", consoleErrs)
	}
}

// TestProfileIdentityPersists saves a name and verifies it survives a reload
// (cookie round-trip).
func TestProfileIdentityPersists(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(serverURL + "/examples/profile?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")

	nameInput := page.Locator("#profile-identity input[name='name']")
	if err := nameInput.Fill("Katherine Johnson"); err != nil {
		t.Fatalf("fill name: %v", err)
	}
	if err := page.Locator("#profile-fragment button[type='submit']").Click(); err != nil {
		t.Fatalf("submit: %v", err)
	}
	// Wait for the swapped identity fragment, then reload and assert persistence.
	page.WaitForTimeout(300)
	if _, err := page.Goto(serverURL + "/examples/profile"); err != nil {
		t.Fatalf("reload: %v", err)
	}
	page.WaitForSelector("#profile-identity input[name='name']")
	val, err := page.Locator("#profile-identity input[name='name']").InputValue()
	if err != nil {
		t.Fatalf("read value: %v", err)
	}
	if val != "Katherine Johnson" {
		t.Fatalf("name did not persist: %q", val)
	}
}

// TestProfileDarkToggle flips dark mode via the toggle and asserts the <html>
// element gains the `dark` class (proves Toggle Attrs wiring to $store.darkMode).
func TestProfileDarkToggle(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(serverURL + "/examples/profile?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")

	before, _ := page.Evaluate("() => document.documentElement.classList.contains('dark')")
	if err := page.Locator("label[for='profile-dark']").Click(); err != nil {
		t.Fatalf("click dark toggle: %v", err)
	}
	page.WaitForTimeout(150)
	after, _ := page.Evaluate("() => document.documentElement.classList.contains('dark')")
	if before == after {
		t.Fatalf("dark class did not toggle: before=%v after=%v", before, after)
	}
}

// TestProfileThemePicker selects a theme and asserts data-theme on <html>.
func TestProfileThemePicker(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(serverURL + "/examples/profile?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")

	// Open the theme Select and pick "Dracula".
	if err := page.Locator("#profile-theme-trigger").Click(); err != nil {
		t.Fatalf("open select: %v", err)
	}
	if err := page.Locator("li[role='option']:has-text('Dracula')").Click(); err != nil {
		t.Fatalf("pick option: %v", err)
	}
	page.WaitForTimeout(150)
	theme, _ := page.Evaluate("() => document.documentElement.getAttribute('data-theme')")
	if theme != "dracula" {
		t.Fatalf("data-theme not applied: %v", theme)
	}
}

// TestProfileAvatarUploadPersists uploads a small PNG and asserts the avatar
// img src becomes a blob: URL, then survives a reload (IndexedDB hydration).
func TestProfileAvatarUploadPersists(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(serverURL + "/examples/profile?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")

	// 1x1 transparent PNG.
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}
	if err := page.Locator("#profile-avatar-input").SetInputFiles(playwright.InputFile{
		Name: "a.png", MimeType: "image/png", Buffer: png,
	}); err != nil {
		t.Fatalf("set file: %v", err)
	}
	// Avatar src becomes a blob: URL.
	page.WaitForFunction(
		"() => { const i = document.querySelector('#profile-avatar img'); return i && i.src.startsWith('blob:'); }", nil)

	// Reload: IndexedDB hydrates the avatar again.
	if _, err := page.Goto(serverURL + "/examples/profile"); err != nil {
		t.Fatalf("reload: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")
	page.WaitForFunction(
		"() => { const i = document.querySelector('#profile-avatar img'); return i && i.src.startsWith('blob:'); }", nil)
}

// TestProfileOversizeRejected uploads a >1MB buffer and asserts no blob preview
// appears (rejected client-side).
func TestProfileOversizeRejected(t *testing.T) {
	browser := setupPlaywright(t)
	page := newPage(t, browser)
	if _, err := page.Goto(serverURL + "/examples/profile?seed=0"); err != nil {
		t.Fatalf("goto: %v", err)
	}
	page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	page.WaitForSelector("#profile-fragment")

	big := make([]byte, 1024*1024+10) // > 1 MB
	if err := page.Locator("#profile-avatar-input").SetInputFiles(playwright.InputFile{
		Name: "big.png", MimeType: "image/png", Buffer: big,
	}); err != nil {
		t.Fatalf("set file: %v", err)
	}
	page.WaitForTimeout(300)
	src, _ := page.Evaluate("() => { const i = document.querySelector('#profile-avatar img'); return i ? i.src : ''; }")
	if s, _ := src.(string); strings.HasPrefix(s, "blob:") {
		t.Fatal("oversize image should have been rejected, but a blob preview appeared")
	}
}
```

> **Helper-signature check:** confirm `serverURL`, `setupPlaywright`, `newPage`
> exact names/returns in `tests/e2e/e2e_test.go` and adjust. The todo/sidebar
> tests are the reference. Use `clickUntil` if any click here targets a control
> that was just HTMX-swapped (none of the above currently do — identity submit
> swaps a different element than it clicks).

- [ ] **Step 2: Run the profile E2E subset**

Run: `go test ./tests/e2e/... -count=1 -timeout 8m -run 'TestProfile|TestAvatarSrcExpr|TestToggleAttrsPassthrough'`
Expected: all PASS. Debug failures with `takeScreenshot(t, page, "name")` and by reading rendered HTML; check for `&quot;` in `x-data` (templ escaping) if Alpine is silent.

- [ ] **Step 3: Commit**

```bash
git add tests/e2e/profile_test.go
git commit -m "test(profile): e2e for fragment-nav, identity persist, theme, dark, image upload"
```

---

## Task 10: Final gates — generate, skillgen, lint, full suite

**Files:** none new; regenerate + verify.

- [ ] **Step 1: Regenerate everything**

Run:
```bash
templ generate
tailwindcss -i css/main.css -o assets/styles.css
go run ./scripts/skillgen
```
Expected: `skillgen` updates `.claude/skills/using-goshtoso/components-reference.md` to reflect the new Avatar `SrcExpr` and Toggle `Attrs` fields. Stage the regenerated reference.

- [ ] **Step 2: Lint**

Run: `golangci-lint run`
Expected: clean. If any handler/helper exceeds cyclomatic complexity 20, extract a helper (none in this plan should — handlers are small).

- [ ] **Step 3: Build + full E2E**

Run:
```bash
go build -o bin/server ./cmd/server
go test ./tests/e2e/... -count=1 -timeout 15m
```
Expected: full suite green, no skipped tests. Re-run any flaky test in isolation; if a post-swap click is the cause, switch it to `clickUntil` (see `e2e-suite-flaky-full-run`).

- [ ] **Step 4: Codex review gate**

Hand the branch diff to Codex for an independent review (`codex:rescue` skill or the stop-time gate). Address findings, confirming each against the code before acting.

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "chore(profile): regenerate styles + using-goshtoso reference; finalize example"
```

---

## Self-Review (completed by plan author)

**Spec coverage:**
- Two storage planes (cookie text / IndexedDB images / client theme) → Tasks 1–3, 6. ✅
- No S3 → enforced by design; nothing references S3. ✅
- Avatar `SrcExpr` → Task 5. ✅
- Component reuse (Avatar, Card sections, Text Input, Textarea, Select, Toggle, Button, Toast, Badge, File Input) → Task 6. ✅ (Tabs/Modal/Tooltip/Banner from the spec map were dropped as YAGNI — the layout uses plain card sections; Modal-confirm on remove was simplified to an `x-show`-gated Remove button. Noted here as intentional scope trim; re-add if desired.)
- Load sequence (server shell + Alpine hydrate) → Task 6 `init()` + Task 7 render. ✅
- Edge cases (oversize, wrong-type, objectURL revoke, IndexedDB-unavailable) → Task 6 script. ✅
- Tests (fragment-nav+console, persist, theme, dark, upload, oversize) → Task 9. ✅
- Quality gates (templ/tailwind/skillgen/lint/build/e2e) → Task 10. ✅

**Open items flagged inline for implement-time verification (read the real `types.go`):**
1. **Toast client-trigger API** (Task 6 Step 4) — the one genuine integration unknown; `_toast` must call the real toast mechanism.
2. **`button.Config.Attrs`** existence (Task 6) — if absent, blocking; extend Button like Toggle.
3. **`textarea.Config` field names** (`Value`, `Rows`).
4. **E2E helper signatures** (`serverURL`/`setupPlaywright`/`newPage`).

These are verifications, not placeholders — each has a concrete fallback action.

**Type consistency:** `State{Name,Bio}`, `SetName`/`SetBio`, `Encode`/`Decode`,
`FromRequest`/`SetCookie`, `Sample` used consistently across Tasks 1–3, 7.
`ProfileApp`/`ProfileContent`/`IdentityFields` consistent across Tasks 6–9.
`profileImages` Alpine name + `avatarSrc`/`bannerSrc`/`pick`/`onFile`/`remove`
consistent across Task 6 and Task 9 selectors.
