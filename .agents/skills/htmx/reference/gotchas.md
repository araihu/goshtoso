# htmx 4 gotchas

Sources: [migration guide](https://four.htmx.org/docs/whats-new-in-htmx-4), [morphing guide](https://four.htmx.org/docs/morphing-swaps-guide), [Alpine integration](https://four.htmx.org/extensions/hx-alpine-compat).

- **Inheritance:** inspect parent/child usage before adding `:inherited`. Shared targets, CSRF headers, indicators and boosts are common migration points.
- **Renames:** migrate old `hx-disable` (ignore subtree) to `hx-ignore` before changing `hx-disabled-elt` to the new `hx-disable` (disable controls during a request).
- **Form values:** GET and DELETE do not automatically include the enclosing form. Add `hx-include="closest form"` when needed; verify actual payloads.
- **Error responses:** 4xx/5xx swap by default. `noSwap: [204, 304, "4xx", "5xx"]` restores old status behavior when deliberately chosen.
- **OOB:** main content now swaps first. Keep targets and HTML parsing context valid. First-render fragments should not carry update-only OOB instructions. OOB-only responses no longer clear the main target unless configured.
- **History:** core no longer uses the old localStorage snapshot cache. Verify back/forward and full-page server responses. Add `hx-history-cache` only if snapshot restoration is desired, and test it together with Alpine compatibility.
- **Post-processing:** use `htmx:after:process` with `event.target` to enhance inserted subtrees, including secondary partials. `htmx:after:init` covers powered elements only; `htmx:after:swap` fires on the request source and carries `detail.ctx`.
- **Dynamic nodes:** htmx processes its own swaps. Call `htmx.process(subtree)` for manually or Alpine-inserted htmx markup when required; avoid repeated whole-document processing.
- **Events:** use colon-separated 4.x names, including inside Alpine listeners. Migrate old camelCase/kebab aliases; Goshtoso does not support htmx 2 bridges.
- **Morphing:** preserve stable identity where state should survive. Do not expect a morph to reset edited input values. Test intentional component teardown separately from preservation.

## templ and Alpine

Follow Goshtoso's component-provider conventions for complex expressions and inspect the live DOM attribute value, not just serialized HTML. HTML entity escaping in source is normal; distinguish it from double encoding or malformed JavaScript. Do not bypass escaping on user data. Read [Alpine gotchas](../../alpinejs/reference/gotchas.md) for registration and nil-array handling.

## Security settings

Use server-side escaping for untrusted data. In 4.x, `hx-ignore` excludes a subtree from htmx; `hx-disable` no longer does. Cross-origin request behavior uses Fetch `mode` (default `same-origin`). Old `allowEval`, `allowScriptTags`, and `selfRequestsOnly` configuration is removed; do not claim those switches protect a 4.x page. Review actual script execution and CSP requirements when migrating custom security configuration.
