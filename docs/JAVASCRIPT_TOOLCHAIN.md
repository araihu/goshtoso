# First-party JavaScript contract

Authored browser code lives only under `assets/js/src/`. Files next to that
directory are tracked generated output and must not be hand-edited. The embedded
asset handler deliberately does not publish `assets/js/src/`.

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

- `assets/js/goshtoso.min.js` from `src/combobox.js`, `src/action-group.js`,
  component globals (`structured-input`, `tooltip`), then demo Alpine providers
  (`action-group`, `avatar-showcase`, `log-feed`), in that fixed order;
- compatibility builds at `assets/js/combobox.js` and
  `assets/js/action-group.js`;
- standalone `assets/js/darkmode.js` and bootstrap
  `assets/js/dependency-loader.js`.

`head.Dependencies()` and `head.DependenciesMinimal()` load the combined bundle.
Runtime order is Alpine plugins, first-party bundle, Alpine core, then HTMX.
Component globals therefore exist before Alpine's first DOM scan, while demo
providers register during `alpine:init`. None of those initializers needs HTMX;
the log-feed provider consults `htmx.process` only on a later resume action.
The dependency loader remains standalone because it owns ordered third-party
loading, exact-version local fallback, readiness events and promise state, and
CSP nonce propagation to every child script. Passing `WithComboboxURL` or
`WithActionGroupURL` switches head rendering to both standalone compatibility
entries so either legacy override keeps its original behavior.

`assets.DefaultRuntimeManifest().Dependencies` remains the public inventory:
the bundle entry is enabled by default, while the standalone Combobox and
ActionGroup entries remain present with `Enabled == false` for compatibility.
Consumers should execute or cache enabled entries only.

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
