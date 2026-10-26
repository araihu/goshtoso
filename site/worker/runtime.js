import starterMain from "../../examples/getting-started/main.go";
import starterPage from "../../examples/getting-started/page.templ";
import "../dist/wasm_exec.js";
import module from "../dist/server.wasm";

globalThis.goshtosoExampleSources = { "main.go": starterMain, "page.templ": starterPage };

export function createRuntime() {
  let ready;

  // Instantiate lazily inside a request: Go initialization can use clocks and
  // randomness, which Workers disallows during module evaluation.
  function initialize() {
    return ready ??= (async () => {
      const go = new globalThis.Go();
      const instance = await WebAssembly.instantiate(module, go.importObject);
      go.run(instance).catch((error) => console.error("Go runtime exited", error));
      if (!globalThis.goshtosoHandle) throw new Error("Go server did not initialize");
      return { handle: globalThis.goshtosoHandle, cancel: globalThis.goshtosoCancel, chat: globalThis.goshtosoChat };
    })();
  }

  const maxBodyBytes = 1024 * 1024;
  const maxQueuedBytes = 1024 * 1024;

  async function readBody(request) {
    if (!request.body) return new Uint8Array();
    const reader = request.body.getReader();
    const chunks = [];
    let length = 0;
    try {
      for (;;) {
        const { value, done } = await reader.read();
        if (done) break;
        length += value.byteLength;
        if (length > maxBodyBytes) {
          await reader.cancel();
          return null;
        }
        chunks.push(value);
      }
    } finally {
      reader.releaseLock();
    }
    const body = new Uint8Array(length);
    let offset = 0;
    for (const chunk of chunks) {
      body.set(chunk, offset);
      offset += chunk.byteLength;
    }
    return body;
  }

  async function serveGo(request, ctx) {
    const body = await readBody(request);
    if (body === null) return new Response("Request body too large", { status: 413 });
    const runtime = await initialize();
    const id = crypto.randomUUID();
    let complete;
    const completion = new Promise((resolve) => { complete = resolve; });
    ctx?.waitUntil(completion);
    return new Promise((resolve, reject) => {
      let controller;
      let finished = false;
      let started = false;
      let streaming = false;
      const cancel = () => {
        finish(new DOMException("Request aborted", "AbortError"));
        runtime.cancel(id);
      };
      const finish = (error) => {
        if (finished) return;
        finished = true;
        request.signal.removeEventListener("abort", cancel);
        if (error) {
          if (started) controller.error(error);
          else reject(error);
        } else controller.close();
      };
      const stream = new ReadableStream({
        start(value) { controller = value; },
        cancel() {
          finished = true;
          request.signal.removeEventListener("abort", cancel);
          runtime.cancel(id);
          return completion;
        },
      }, { highWaterMark: 65536, size: (chunk) => chunk.byteLength });
      const sink = {
        start(status, pairs, isStreaming) {
          streaming = isStreaming && pairs.some(([key, value]) =>
            key.toLowerCase() === "content-type" && value.startsWith("text/event-stream"));
          if (finished) return;
          const headers = new Headers(pairs);
          headers.delete("Connection");
          headers.delete("Transfer-Encoding");
          const noBody = request.method === "HEAD" || [204, 205, 304].includes(status);
          started = true;
          resolve(new Response(noBody ? null : stream, { status, headers }));
        },
        write(chunk) {
          if (finished || request.method === "HEAD") return;
          controller.enqueue(chunk);
          if (streaming && controller.desiredSize < -maxQueuedBytes) {
            finish(new Error("Response consumer is too slow"));
            runtime.cancel(id);
          }
        },
        end() { finish(); complete(); },
        fail(message) {
          console.error("Go handler failed", message);
          finish(new Error("Go handler failed"));
          complete();
        },
      };
      request.signal.addEventListener("abort", cancel, { once: true });
      runtime.handle(id, {
        method: request.method,
        url: request.url,
        headers: [...request.headers],
      }, body, sink);
      if (request.signal.aborted) cancel();
    });
  }

  async function chat(input) {
    const runtime = await initialize();
    return JSON.parse(runtime.chat(JSON.stringify(input)));
  }
  return { serveGo, chat };
}

