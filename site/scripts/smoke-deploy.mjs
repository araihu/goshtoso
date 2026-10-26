import assert from "node:assert/strict";

const origin = new URL(process.argv[2] || "https://goshtoso.araihu.com");
assert.equal(origin.protocol, "https:", "Production smoke requires HTTPS");
const request = (path, init = {}) => fetch(new URL(path, origin), {
  ...init, signal: AbortSignal.timeout(20000),
});

for (const path of ["/", "/getting-started", "/components/button"]) {
  const response = await request(path);
  assert.equal(response.status, 200, path);
  const html = await response.text();
  assert.match(html, /<!doctype html>/i, path);
  assert.doesNotMatch(html, /Source file not found/, path);
}
const partial = await request("/components/button", { headers: { "HX-Request-Type": "partial" } });
assert.equal(partial.status, 200);
assert.doesNotMatch(await partial.text(), /<!doctype html>/i);

const runtime = await request("/assets/js/runtime/htmx.org/4.0.0/htmx.min.js");
assert.equal(runtime.status, 200);
assert.match(runtime.headers.get("cache-control"), /immutable/);
assert.ok((await runtime.arrayBuffer()).byteLength > 0);

const post = await request("/api/examples/todo/add", {
  method: "POST",
  headers: { "Content-Type": "application/x-www-form-urlencoded" },
  body: "title=Ol%C3%A1+Cloudflare",
});
assert.equal(post.status, 200);
assert.match(await post.text(), /Olá Cloudflare/);
const cookie = post.headers.get("set-cookie");
assert.ok(cookie);
const restored = await request("/api/examples/todo/filter", {
  method: "POST", headers: { Cookie: cookie.split(";")[0] },
});
assert.equal(restored.status, 200);
assert.match(await restored.text(), /Olá Cloudflare/);

const stream = await request("/api/examples/logs/stream?interval=10ms&max=2");
assert.equal(stream.status, 200);
assert.match(stream.headers.get("content-type"), /text\/event-stream/);
assert.equal(((await stream.text()).match(/: log message/g) || []).length, 2);
const delayed = await request("/api/hello?delay_ms=20");
assert.equal(delayed.status, 200);
assert.match(await delayed.text(), /Hello from HTMX/);

const socketURL = new URL("/api/examples/chat/ws", origin);
socketURL.protocol = "wss:";
const socket = new WebSocket(socketURL);
await new Promise((resolve, reject) => {
  const timer = setTimeout(() => {
    socket.close();
    reject(new Error("WebSocket presence timed out"));
  }, 20000);
  socket.addEventListener("error", () => {
    clearTimeout(timer);
    reject(new Error("WebSocket handshake failed"));
  }, { once: true });
  socket.addEventListener("message", (event) => {
    clearTimeout(timer);
    socket.close(1000, "Deployment smoke complete");
    try {
      assert.match(event.data, /chat-/);
      resolve();
    } catch (error) {
      reject(error);
    }
  }, { once: true });
});
console.log(`Production smoke passed: ${origin.origin} (pages, assets, HTMX, cookies, SSE, WebSocket)`);
