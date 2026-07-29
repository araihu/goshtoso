# First-party JavaScript contract

Reusable library browser code lives under `assets/js/src/`. Demo-site providers
live under `site/assets/js/src/`; they are site-owned and are not published by
the library asset handler. Files next to those source directories are tracked
generated output and must not be hand-edited. Neither embedded asset handler
publishes an authored `src/` directory.

The toolchain uses the Go API from exactly pinned `github.com/evanw/esbuild`
module version in `go.mod`; it does not require Node.js or a package manager.

## Commands

```bash
just js        # minify and write tracked generated output
just js-check  # syntax lint, similarity report, inline policy, drift check

# inspect all current extraction candidates with file and line evidence
go run ./cmd/jslint -inventory
```

`just js` generates:

- `assets/js/goshtoso.min.js` from `src/combobox.js`, client-mode Combobox,
  `src/action-group.js`, component globals (`structured-input`, `tooltip`),
  shared component data decoding, shared navigation validation, Search and
  Table runtimes, then component factories (`carousel`, `dropdown`, `palette`,
  `select`, `tabs`), in that fixed order;
- `site/assets/js/goshtoso-demo.min.js` from the site bootstrap and providers
  `site-bootstrap`, `demo-layout`, `select-demo`, `tab-view`, `action-group`,
  `avatar-showcase`, `log-feed`, `chat`, `profile-images`, `ticker-pane`, and
  `theme-page`, in that fixed order;
- compatibility builds at `assets/js/combobox.js` (keyboard plus client-mode
  selection) and `assets/js/action-group.js`;
- standalone `assets/js/darkmode.js` and bootstrap
  `assets/js/dependency-loader.js`.

`head.Dependencies()` and `head.DependenciesMinimal()` load only the reusable
component bundle. Runtime order is Alpine plugins, component bundle, Alpine
core, then HTMX. The Goshtoso demo site's layouts separately load
`/site-assets/js/goshtoso-demo.min.js`; that URL is embedded and served by the
site module and never enters the public runtime manifest. The legacy site layout
uses deferred component-bundle, demo-bundle, Alpine order. Componentdocshell
emits the site bundle as a synchronous site runtime after its deferred dependency
tags, so the providers subscribe to `alpine:init` before Alpine executes. None
of those initializers needs HTMX; the log-feed provider consults `htmx.process`
only on a later resume action.
The dependency loader remains standalone because it owns ordered third-party
loading, exact-version local fallback, readiness events and promise state, and
CSP nonce propagation to every child script. Passing `WithComboboxURL` or
`WithActionGroupURL` switches head rendering to both standalone compatibility
entries so either legacy override keeps its original behavior.

`assets.DefaultRuntimeManifest().Dependencies` remains the public library
inventory: the component bundle entry is enabled by default, while the
standalone Combobox and ActionGroup entries remain present with
`Enabled == false` for compatibility. Consumers should execute or cache enabled
entries only. The site bundle is deliberately absent.

## Inline extraction formula

`go run ./cmd/jslint` scans production `.templ`, `.go`, and `.html` sources,
excluding generated templ Go, tests, and fixtures. It reports:

- every executable inline `<script>` body, including one-line bodies;
- `<script>` bodies emitted through `templ.Raw`;
- Alpine and DOM event attribute expressions that contain a newline or exceed
  80 bytes after trimming;
- multiline Go functions whose string literals contain repeated browser and
  JavaScript syntax markers, covering concatenated script builders.

JSON data scripts (`application/ld+json` and `application/json`), external
`src` scripts, empty script tags, and one-line attribute expressions of at most
80 bytes are the only automatic inline allowances. This keeps the allowance
small and mechanical.

Existing candidates are content-fingerprinted in
`tools/javascript/inline-baseline.txt`. The normal command fails only for new
candidates, allowing extraction to land in independent component lanes. Inspect
the full inventory before deliberately refreshing it with:

```bash
go run ./cmd/jslint -inventory
go run ./cmd/jslint -update-inline-baseline
```

Never refresh the baseline merely to make CI green. Extraction should remove
entries.

## Similarity report

Authored `.js` is syntax-parsed by esbuild. Named function declarations,
function expressions, and block-bodied arrow assignments are tokenized with
identifiers and literals normalized. The report computes a multiset Dice score
over five-token shingles, prints pairs at or above `0.82`, and sorts by score
then source location. Change the discovery threshold with
`-similarity-threshold`; similarities are evidence for review, not automatic
deletions.
