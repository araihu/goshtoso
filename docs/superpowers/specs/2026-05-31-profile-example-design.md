# Profile Example — Design Spec

**Date:** 2026-05-31
**Branch:** `worktree-profile-example`
**Status:** Approved (brainstorm), pending spec review

## Amendments (planning, 2026-05-31)

Decided while writing the implementation plan; these override the body below where they conflict:

1. **Cookie holds only `Name` + `Bio`.** No `Theme`/`Dark`/`Seq`. `Seq` was unused (no IDs); theme/dark are client-side.
2. **Theme + dark are NOT in the cookie.** They reuse the app's existing client engine (`theme` on the layout root x-data + `localStorage`, and `$store.darkMode`) — one source of truth, applies live. The repo already persists both in `localStorage`; a cookie copy would be a second, conflicting source.
3. **Select needs no extension.** Its existing `AlpineModel` field bound to the root `theme` property is the onChange hook. Only **Avatar** (`SrcExpr`) and **Toggle** (`Attrs`) are extended.
4. **Upload uses the existing File Input component** (`components/fileinput`, has `Accept` + `Attrs`), not a raw `<input>`.
5. **Tabs / Modal / Tooltip / Banner dropped (YAGNI).** Layout is plain card sections; remove-photo is an `x-show`-gated button, not a Modal. Re-add later if desired.

## Goal

Add a second example app to Goshtoso: a **user profile page** with avatar +
banner image upload, editable identity text, and theme/accent customization.
Like the todo example it flows through the shared demo registry + `renderDemo`
(theme, dark mode, fragment-swap nav for free) and gets an "Examples" sidebar
entry. It showcases real Goshtoso components in a realistic screen.

It adds one dimension the todo example does not: **client-side binary storage**.
Images live in the browser (IndexedDB), the server never sees the bytes.

## Two storage planes (never mixed)

| Data | Plane | Mechanism |
|------|-------|-----------|
| Display name, bio, theme name, dark flag | **Cookie** (server-owned) | `gt_profile` base64url-JSON, HTMX mutations → server re-renders fragments. Same pattern as todo. |
| Avatar image, banner image | **IndexedDB** (browser-owned) | Alpine reads `File` → `createObjectURL` preview → stores `Blob`. Server byte-blind. |

**Why no S3 / no server upload:** examples are stateless by design (no server
memory, no external service at runtime, hermetic E2E with no skipped tests).
S3 would add a hard external dep unreachable from CI and force a fallback fake.
IndexedDB keeps the example shippable everywhere with zero infra while still
teaching a real, non-trivial pattern (client-side binary persistence + an Alpine
↔ component image binding). Decision recorded; S3 explicitly out of scope.

**Constraints on images:** `image/*` only, **≤ 1 MB**. Reject → Toast, no write.

## Load sequence (no flash, fragment-nav safe)

1. Server renders the page shell from cookie `gt_profile` — name, bio, theme,
   dark are already correct in the HTML (no client flash).
2. Alpine `profileImages` component `x-init` opens IndexedDB, reads the `avatar`
   and `banner` blobs, `createObjectURL`s them, and sets the bound srcs.
3. The Avatar shows initials until step 2 hydrates a real photo.

Register `Alpine.data('profileImages', …)` **immediately if Alpine is already
running**, plus on `alpine:init` — otherwise it is undefined when the page
arrives via sidebar fragment-nav (CLAUDE.md examples gotcha).

## Architecture (mirror `internal/examples/todo/`)

```
internal/examples/profile/            # HTTP-free domain, unit-tested
  state.go            State{Name, Bio, Theme string; Dark bool; Seq int}
                      Encode/Decode (base64url-JSON); maxCookieBytes budget
  state_mutations.go  SetName / SetBio / SetAppearance(theme,dark)
                      each refuses if Encode(candidate) > maxCookieBytes
  cookie.go           CookieName = "gt_profile"
  sample.go           default seed (empty name, theme "Minimal", dark false)
  state_test.go / state_mutations_test.go   incl. cookie-size-budget test

internal/pages/demo/examples/profile.templ   # exported; rendered by shell AND handlers
internal/server/profile_handler.go           # thin: read cookie → mutate → write → fragment
```

Domain logic carries no HTTP types. Handlers read the cookie, call a mutation,
write the cookie, render the relevant fragment. Theme/bio/name length is capped
so the encoded cookie stays under budget (mutations enforce it).

## Component map (maximize Goshtoso)

| Page piece | Component |
|------------|-----------|
| Profile photo | **Avatar** (circle, lg) — via new `SrcExpr` (see below) |
| Section containers | **Card** ×3 (Identity / Photos / Appearance) |
| Tabbed layout | **Tabs** (Profile · Appearance) |
| Display name | **Text Input** |
| Bio | **Textarea** |
| Theme picker (13 themes) | **Select** |
| Dark mode | **Toggle** |
| Save / Upload / Remove | **Button** |
| Save confirm + reject errors | **Toast** |
| Active theme label | **Badge** |
| "Data stays in your browser" notice | **Banner** (or Alert) |
| Remove-photo confirm | **Modal** |
| "Max 1 MB" hint | **Tooltip** |

**Stays raw (no component hook exists — documented like todo's checkbox/button):**
- Hidden `<input type="file">` — no file-input component in the library.
- Banner cover `<img :src>` — a wide cover image, not an avatar shape; bound
  client-side from IndexedDB the same way as the avatar.

## Avatar component extension — `SrcExpr`

Avatar's image layer today renders `src={cfg.Src}` **statically** at
server-render and only when `Src` is non-empty (`HasImage()`). An IndexedDB
objectURL is produced **client-side after load**, so there is no static `Src`.

Add a `SrcExpr string` field to `avatar.Config`, mirroring the existing
`Reactive` pattern (parent Alpine scope drives a value via `x-bind`):

- When `SrcExpr != ""`, the image layer is always rendered; its `src` comes from
  `x-bind:src="<expr>"` instead of the static `src=`.
- The initials/loading layers gate on the same expression being truthy
  (initials show when `<expr>` is falsy → empty/objectURL-less).
- `Src == "" && SrcExpr != ""` must count as "has image" for layer selection.

Parent scope (profile page) exposes `avatarSrc` (objectURL or `''`) and renders
`@avatar.Avatar(avatar.Config{SrcExpr: "avatarSrc", Name: name, …})`.

This is a genuine library improvement (Avatar usable with a client-bound source)
and keeps the headline element on the real component. Touches: `avatar/types.go`,
`avatar/avatar.templ`, regen `_templ.go`, `skillgen` (components changed), and an
Avatar E2E covering the new mode.

## Image subsystem — `profileImages` Alpine component

Registered via `<script>` + `templ.Raw` + `Alpine.data()` (CLAUDE.md escaping
rule — no JSON in attributes). Exposes:

```
avatarSrc, bannerSrc        // objectURLs; bound into Avatar SrcExpr / banner :src
pick(kind)                  // clicks the hidden <input type="file"> for kind
onFile(kind, event):
  - validate file.type.startsWith('image/') && file.size <= 1*1024*1024
    else  $dispatch toast (danger: ">1 MB" or "images only");  return
  - revoke previous objectURL for kind (leak guard)
  - URL.createObjectURL(file) → set {kind}Src   (instant preview)
  - idbPut(kind, file)                            (persist Blob)
remove(kind)                // Modal-confirmed: idbDelete(kind) + revoke + clear src
x-init:  idbGet('avatar') / idbGet('banner') → objectURL → hydrate
```

**IndexedDB:** DB `gt_profile`, object store `images`, keys `avatar` / `banner`,
value = `Blob`. Promise-wrapped open/get/put/delete helpers.

## Handlers + endpoints (cookie plane)

Registered in `internal/server` routing alongside the todo handlers:

- `POST /api/examples/profile/name` → `SetName` → identity fragment + Toast "Saved"
- `POST /api/examples/profile/bio` → `SetBio` → bio fragment + Toast "Saved"
- `POST /api/examples/profile/appearance` → `SetAppearance(theme, dark)` →
  appearance fragment + **OOB** theme Badge; drives the live `data-theme` / dark
  via the existing Alpine theme store.

Each: read cookie → mutate (refuse if `> maxCookieBytes`) → write cookie →
return fragment. **OOB gotcha:** gate every `hx-swap-oob` to update-only (an
`oob bool` like todo's `CountBadge`/`ClearButton`) so first paint via sidebar
fragment-nav does not raise `htmx:oobErrorNoTarget`.

## Registration checklist

- `Demos` registry entry (so `renderDemo` + fragment nav work).
- Sidebar item in the **Examples** section (`getSidebarItems`).
- Routes for the three endpoints.

## Edge cases

- Oversize / wrong-type file → Toast, no IndexedDB write, no preview change.
- Replacing an image revokes the previous objectURL (no memory leak).
- IndexedDB unavailable (private mode / old browser) → keep the session
  objectURL so preview still works, Toast "won't persist across reloads".
- Name / bio length capped so the encoded cookie respects `maxCookieBytes`
  (mutation refuses over-budget input).
- Empty name allowed → Avatar shows initials fallback.

## Testing

`tests/e2e/profile_test.go` (+ extend `avatar_test.go` for `SrcExpr` mode):

1. **Sidebar fragment-nav load + zero console errors** (examples mandate — not
   just direct load).
2. Edit name → reload → persists (cookie roundtrip).
3. Pick theme + toggle dark → applies live → reload → persists.
4. Upload via Playwright `SetInputFiles` (in-memory buffer) → Avatar `src`
   becomes `blob:` → reload → IndexedDB re-hydrates the photo.
5. Oversize buffer (>1 MB) → Toast shown, preview unchanged, no write.
6. Remove photo → **Modal** confirm → Avatar returns to initials.
7. Use `clickUntil` for any click on a just-swapped control (full-suite flake
   rule — see `e2e-suite-flaky-full-run`).

## Quality gates

- `templ generate` after `.templ` edits; rebuild Tailwind for any new utility
  class (CSS is embedded).
- `go run ./scripts/skillgen` after the Avatar change (CI fails if stale).
- `golangci-lint run` clean; keep new functions under cyclomatic complexity 20
  (extract helpers in handlers / Alpine glue rather than suppress).
- `go build -o bin/server ./cmd/server` + full E2E green, no skipped tests.

## Out of scope (YAGNI)

- S3 / any server-side image storage or presigned URLs.
- Image cropping / resizing / rotation.
- Multiple profiles or auth (single stateless demo user per browser).
- Drag-and-drop upload (todo removed native DnD as unreliable; same call here).
