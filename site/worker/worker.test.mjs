import assert from "node:assert/strict";
import { test } from "node:test";
import { readFileSync } from "node:fs";
import { createTestHarness } from "wrangler";

// These execute the compiled Go module in workerd with real asset and Durable
// Object bindings; no transport mocks or browser-side Go runtime are involved.
test("Go WASM site in Cloudflare Workers", { timeout: 120000 }, async (t) => {
  const { $schema, ...config } = JSON.parse(readFileSync("./wrangler.jsonc", "utf8"));
  const harness = createTestHarness({ workers: [{ config: { ...config, build: {} } }] });
  t.after(async () => {
    const errors = harness.getLogs().filter((log) => log.level === "error");
    await harness.close();
    assert.deepEqual(errors, [], "Worker runtime errors");
  });
  const { url } = await harness.listen();
  const worker = harness.getWorker();

  await t.test("pages, partial navigation, JSON, status, and HEAD", async () => {
    for (const path of ["/", "/getting-started", "/components/table", "/examples/deployments", "/attributions", "/playground/extensions/charts", "/modules/app-shells/live/consoleshell/overview"]) {
      t.diagnostic(path);
      const response = await worker.fetch(path);
      assert.equal(response.status, 200, path);
      const html = await response.text();
      assert.match(html, /<!doctype html>/i, path);
      assert.doesNotMatch(html, /Source file not found/, path);
    }
    const partial = await worker.fetch("/components/button", { headers: { "HX-Request-Type": "partial" } });
    assert.equal(partial.status, 200);
    assert.doesNotMatch(await partial.text(), /<!doctype html>/i);
    const search = await worker.fetch("/api/components/search/items");
    assert.equal(search.status, 200);
    assert.equal((await search.json())[0].id, "search-remote-table");
    assert.equal((await worker.fetch("/missing-page")).status, 404);
    const redirect = await worker.fetch("/components/", { redirect: "manual" });
    assert.equal(redirect.status, 301);
    assert.equal(redirect.headers.get("Location"), "/components/accordion");
    const head = await worker.fetch("/", { method: "HEAD" });
    assert.equal(head.status, 200);
    assert.equal(await head.text(), "");
  });

  await t.test("static bytes and cache policy", async () => {
    for (const path of ["/assets/images/goshtoso-art.png", "/assets/styles.css", "/site-assets/js/goshtoso-demo.min.js",
      "/favicon.svg", "/site.webmanifest", "/assets/icons/heroicons/sprite.svg", "/dog-images/beagle.webp",
      "/componentdocshell/assets/shell.js", "/consoleshell/assets/shell.css", "/landingshell/assets/shell.js",
      "/charts/assets/js/runtime/echarts/5.6.0/echarts.min.js"]) {
      const response = await worker.fetch(path);
      assert.equal(response.status, 200, path);
      const bytes = Buffer.from(await response.arrayBuffer());
      assert.deepEqual(bytes, readFileSync(`dist/public${path}`), path);
    }
    const alias = await worker.fetch("/assets/styles.css");
    assert.equal(alias.headers.get("Cache-Control"), "public, max-age=0, must-revalidate");
    const runtime = await worker.fetch("/assets/js/runtime/htmx.org/4.0.0/htmx.min.js");
    assert.match(runtime.headers.get("Cache-Control"), /immutable/);
    assert.equal((await worker.fetch("/assets/missing.png")).status, 404);
  });

  await t.test("POST, UTF-8, cookies, consent, and method guards", async () => {
    const post = (body, extra = {}) => ({ method: "POST", headers: { "Content-Type": "application/x-www-form-urlencoded", ...extra }, body });
    const response = await worker.fetch("/api/examples/todo/add", post("title=Ol%C3%A1+Cloudflare"));
    assert.equal(response.status, 200);
    assert.match(await response.text(), /Olá Cloudflare/);
    const cookie = response.headers.get("Set-Cookie");
    assert.ok(cookie);
    const restored = await worker.fetch("/api/examples/todo/filter", { method: "POST", headers: { Cookie: cookie.split(";")[0] } });
    assert.match(await restored.text(), /Olá Cloudflare/);
    const denied = await worker.fetch("/api/examples/todo/add", post("title=Private", { Cookie: "gt_storage=denied" }));
    assert.equal(denied.headers.get("Set-Cookie"), null);
    const form = await worker.fetch("/api/components/form/external-submit", post("version=1.2.3"));
    assert.match(await form.text(), /v1\.2\.3/);
    assert.equal((await worker.fetch("/api/examples/todo/add")).status, 405);
    const oversized = await worker.fetch("/api/examples/todo/add", post("x".repeat(1024 * 1024 + 1)));
    assert.equal(oversized.status, 413);
  });

  await t.test("SSE streams flush, finish, and cancel independently", async () => {
    const logs = await worker.fetch("/api/examples/logs/stream?interval=10ms&max=2");
    assert.equal(logs.status, 200, logs.status === 200 ? "" : await logs.text());
    assert.equal(logs.headers.get("Content-Type"), "text/event-stream");
    const text = await logs.text();
    assert.equal((text.match(/: log message/g) || []).length, 2);
    const ticker = await worker.fetch("/api/examples/ticker/stream");
    const reader = ticker.body.getReader();
    const first = await reader.read();
    assert.match(new TextDecoder().decode(first.value), /ticker-cell-AAPL/);
    await reader.cancel();
    // A fresh timed handler must still run after the stream's timer is canceled.
    const delayed = await worker.fetch("/api/hello?delay_ms=20");
    assert.equal(delayed.status, 200);
    assert.match(await delayed.text(), /Hello from HTMX/);
    const concurrent = await Promise.all([10, 30, 50].map((delay) =>
      worker.fetch(`/api/hello?delay_ms=${delay}`).then((response) => response.text())));
    for (const html of concurrent) assert.match(html, /Hello from HTMX/);
  });

  await t.test("WebSocket chat shares messages and restores history after eviction", async () => {
    const open = async () => {
      const response = await worker.fetch("/api/examples/chat/ws", { headers: { Upgrade: "websocket", Origin: url.origin } });
      assert.equal(response.status, 101);
      response.webSocket.accept();
      return response.webSocket;
    };
    const one = await open();
    const two = await open();
    const message = waitForFrame(two, "Olá sala");
    one.send(JSON.stringify({ message: "Olá sala", nick: "Ada" }));
    assert.match(await message, /Ada/);
    await worker.evictDurableObject("ChatRoom", { name: "goshtoso-demo", webSockets: "hibernate" });
    const reply = waitForFrame(one, "depois");
    two.send(JSON.stringify({ message: "depois", nick: "Lin" }));
    await reply;
    const three = await open();
    // Message delivery starts when accepted, so collect replay immediately.
    await waitForFrame(three, "Olá sala");
    one.close(); two.close(); three.close();
    assert.equal((await worker.fetch("/api/examples/chat/ws", { headers: { Upgrade: "websocket", Origin: "https://foreign.invalid" } })).status, 403);
    assert.equal((await worker.fetch("/api/examples/chat/ws")).status, 426);
  });
});

function waitForFrame(socket, text) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => { socket.removeEventListener("message", listener); reject(new Error(`Missing frame: ${text}`)); }, 10000);
    const listener = ({ data }) => {
      if (!data.includes(text)) return;
      clearTimeout(timer);
      socket.removeEventListener("message", listener);
      resolve(data);
    };
    socket.addEventListener("message", listener);
  });
}
