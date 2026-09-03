# Embedded JavaScript runtime

`muamba.yaml` is the acquisition and SHA-384 integrity source of truth. `assets/runtime.overlay.yaml` owns Goshtoso runtime order, loading defaults, and attribution metadata. Generated Go, site, and documentation files must not be edited by hand.

Inspect the exact runtime linked into a Go binary with `assets.DefaultRuntimeManifest()`. Inspect normalized vendored hashes for cache busting with `assets.RuntimeHash(role)` or the complete acquisition registry with `assets.MuambaResources()`.

Inspect generated identity and retained-license metadata with `assets.DefaultRuntimeMetadata()`. Runtime metadata does not replace `RuntimeAsset`: the latter remains the supported loading, override, SRI, enablement, and ordering contract.

Pinned versions are the tested combination. Replacing URLs or versions configures loading; it does not guarantee compatibility with arbitrary combinations.

| Order | Asset | Version | Role | Default | Minimal | Embedded URL | Primary URL | License |
| ---: | --- | --- | --- | :---: | :---: | --- | --- | --- |
| 0 | Goshtoso dependency loader | `Goshtoso release` | `dependency-loader` | true | true | `/assets/js/dependency-loader.js` | `/assets/js/dependency-loader.js` | Goshtoso |
| 1 | @alpinejs/collapse | `3.14.9` | `alpine-collapse` | true | false | `/assets/js/runtime/alpinejs-collapse/3.14.9/alpine-collapse.min.js` | `https://unpkg.com/@alpinejs/collapse@3.14.9/dist/cdn.min.js` | [MIT](../assets/js/runtime/alpinejs-collapse/3.14.9/LICENSE.txt) |
| 2 | @alpinejs/focus | `3.14.9` | `alpine-focus` | true | false | `/assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js` | `https://unpkg.com/@alpinejs/focus@3.14.9/dist/cdn.min.js` | [MIT](../assets/js/runtime/alpinejs-focus/3.14.9/LICENSE.txt) |
| 3 | @alpinejs/mask | `3.14.9` | `alpine-mask` | true | false | `/assets/js/runtime/alpinejs-mask/3.14.9/alpine-mask.min.js` | `https://unpkg.com/@alpinejs/mask@3.14.9/dist/cdn.min.js` | [MIT](../assets/js/runtime/alpinejs-mask/3.14.9/LICENSE.txt) |
| 4 | Goshtoso component runtime | `Goshtoso release` | `first-party` | true | true | `/assets/js/goshtoso.min.js` | `/assets/js/goshtoso.min.js` | Goshtoso |
| 5 | Goshtoso dark-mode runtime | `Goshtoso release` | `dark-mode` | false | true | `/assets/js/darkmode.js` | `/assets/js/darkmode.js` | Goshtoso |
| 6 | Alpine.js | `3.14.9` | `alpine` | true | true | `/assets/js/runtime/alpinejs/3.14.9/alpine.min.js` | `https://unpkg.com/alpinejs@3.14.9/dist/cdn.min.js` | [MIT](../assets/js/runtime/alpinejs/3.14.9/LICENSE.txt) |
| 7 | htmx | `2.0.8` | `htmx` | true | true | `/assets/js/runtime/htmx.org/2.0.8/htmx.min.js` | `https://unpkg.com/htmx.org@2.0.8/dist/htmx.min.js` | [Zero-Clause BSD](../assets/js/runtime/htmx.org/2.0.8/LICENSE.txt) |
| 8 | htmx SSE extension | `2.2.3` | `htmx-ext-sse` | false | true | `/assets/js/runtime/htmx-ext-sse/2.2.3/htmx-ext-sse.min.js` | `https://unpkg.com/htmx-ext-sse@2.2.3/dist/sse.min.js` | [Zero-Clause BSD](../assets/js/runtime/htmx-ext-sse/2.2.3/LICENSE.txt) |
| 9 | htmx WebSocket extension | `2.0.3` | `htmx-ext-ws` | false | true | `/assets/js/runtime/htmx-ext-ws/2.0.3/htmx-ext-ws.js` | `https://unpkg.com/htmx-ext-ws@2.0.3/ws.js` | [Zero-Clause BSD](../assets/js/runtime/htmx-ext-ws/2.0.3/LICENSE.txt) |
| 10 | Goshtoso combobox compatibility runtime | `Goshtoso release` | `combobox` | false | true | `/assets/js/combobox.js` | `/assets/js/combobox.js` | Goshtoso |
| 11 | Goshtoso action-group runtime | `Goshtoso release` | `action-group` | false | true | `/assets/js/action-group.js` | `/assets/js/action-group.js` | Goshtoso |
| 12 | Goshtoso CodeBlock runtime | `Goshtoso release` | `code-block` | false | true | `/assets/js/code-block.js` | `/assets/js/code-block.js` | Goshtoso |

Regenerate all consumers with `go run ./cmd/runtimegen`; verify acquisition integrity and generated drift with `go tool muamba verify --strict`, `go tool muamba generate-go --strict --check --dir assets --output muamba_gen.go`, and `go run ./cmd/runtimegen -check`.
