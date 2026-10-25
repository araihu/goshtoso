# Goshtoso documentation site

The site can run as a native Go HTTP server or as Go WebAssembly inside
Cloudflare Workers. The Go component library and the native server remain usable
independently of the Worker tooling.

## Cloudflare Workers

Prerequisites: Go matching `site/go.mod`, Node.js 22 or newer, and a Cloudflare
account. The bundled mode uses features available on Workers Free, including
SQLite-backed Durable Objects. The current bundle is about 39 MiB uncompressed.
[Workers limits](https://developers.cloudflare.com/workers/platform/limits/#worker-size)
allow 64 MiB uncompressed on both Free and Paid; gzip size is informational.
The build reserves 1 MiB for its JavaScript wrapper and metadata.

From this directory:

```sh
npm ci
npm run dev                  # builds Go WASM and starts local workerd
npm test                     # HTTP, assets, SSE, and WebSocket integration tests
npm run check                # build and Wrangler deployment dry run
npm run deploy               # publish to your authenticated Cloudflare account
```

Wrangler's custom build runs `scripts/build-worker.mjs`. It compiles
`cmd/worker` with `GOOS=js GOARCH=wasm`, copies `wasm_exec.js` from that exact Go
toolchain, and exports the native server's static bytes into `dist/public`.
No Node.js service, TCP listener, or Go runtime in the browser is needed.
`dist/` and `.wrangler/` are local build/runtime state.

`wrangler.jsonc` declares the Worker, the `ASSETS` binding, and migrations for
`SITE_SERVER` and `CHAT_ROOM`. Deployment uses the configured Worker name on
workers.dev; add a Cloudflare route or custom domain when moving production
traffic. The existing container publication workflow does not publish Workers.

For a tagged documentation build, set `GOSHTOSO_DOCS_VERSION=vX.Y.Z` when running
Wrangler. This injects the same build-time version used by the native server;
unversioned builds show `development`.

## Staying on Workers Free

Use `wrangler.jsonc` (bundled WASM), with SQLite-backed `SiteServer` and `ChatRoom`.
The build includes the Go WASM module in the deployed Worker. Go initialization
and request handling execute in the Cloudflare Durable Objects.

Static assets are served first (`run_worker_first: false`), so matching assets
bypass Worker execution and its request quota. Dynamic HTTP requests still pass
through the lightweight JavaScript router into the Go Durable Object. Parent
Workers and Durable Objects have separate quotas; consult their respective
limits rather than applying the parent Worker's CPU budget to the Go object.

Relevant Free allowances include 100,000 Worker script requests/day per account
(shared by its Workers) and 10 ms of CPU per parent invocation. SQLite Durable
Objects include 100,000 requests/day, 13,000 GB-s/day of active duration, and 5 GB total stored data, with additional
row-operation quotas. Go initialization and rendering occur inside the Durable
Object; that object's documented default CPU limit is 30 seconds per invocation.
Chat hibernates when idle; open SSE streams and pending timers keep the server
object active and consume the duration allowance. Reaching a Free quota causes
operations to fail until reset. Remaining on the Free account plan prevents
usage from automatically becoming a paid overage.

Sources: [Worker pricing](https://developers.cloudflare.com/workers/platform/pricing/),
[Durable Object pricing](https://developers.cloudflare.com/durable-objects/platform/pricing/),
[Durable Object CPU limits](https://developers.cloudflare.com/durable-objects/platform/limits/),
[Static Asset billing](https://developers.cloudflare.com/workers/static-assets/billing-and-limitations/).

The local workerd tests and deployment dry run validate code and bindings, not
live account billing, daily usage, or Cloudflare's production startup check.

## Request handling

- `worker/index.js` dispatches static URLs to `ASSETS`. Files retain the Go
  server's cache policy and bytes, including runtime licenses, chart assets,
  app-shell assets, icon sprites, favicons, and starter dog images.
- Other HTTP requests go to one named `SiteServer` Durable Object and cross
  `internal/workerhttp` into the existing Go handlers and templ pages. This keeps
  Go scheduler timers (including syntax-highlighter timeouts) in an object-scoped
  I/O context, rather than attaching them to a completed Worker request.
  Methods, query strings, form bodies, headers, redirects, cookies, and HTMX
  fragments use the same application code. Request bodies are bounded at 1 MiB
  before copying them into Go memory.
- SSE uses `ReadableStream`; Go `Flush` forwards each event. Disconnecting or
  canceling the response cancels the Go context and stops request-owned timers.
  Slow SSE consumers are disconnected before their queue grows without bound.
- The ticker demo uses a simulator per stream on Workers. Native Go retains its
  process-wide shared simulator. Workers do not run a background ticker between
  requests.
- The chat endpoint routes every viewer to one named Durable Object. Cloudflare
  owns sockets and hibernation; Go owns validation, identity, ELIZA, and HTML
  rendering. Socket attachments preserve each connection's identity across
  hibernation. SQLite stores only the last 50 messages and the connection
  sequence, so replay survives object eviction. This is a public demo room.
- The Getting Started walkthrough includes its source files as Wrangler text
  modules, avoiding runtime reads from a local filesystem.

The Go runtime initializes lazily inside each Durable Object's first request.
The HTTP server and chat own independent Go instances; no HTTP requests or timer
references are shared with a hibernating chat object. The HTTP backend currently
uses one named object and does not store application data in SQLite. Cookies
remain the per-user example storage.

Lazy initialization keeps clock/randomness use out of Worker module evaluation
and avoids spending the platform's module-startup budget on Go initialization.
