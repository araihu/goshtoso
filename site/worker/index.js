import { DurableObject } from "cloudflare:workers";
import { createRuntime } from "./runtime.js";
export { ChatRoom } from "./chat.js";

const assetPrefixes = ["/assets/", "/site-assets/", "/componentdocshell/assets/",
  "/consoleshell/assets/", "/landingshell/assets/", "/charts/assets/", "/dog-images/"];
const rootAssets = new Set(["/favicon.ico", "/favicon.svg", "/favicon-96x96.png",
  "/apple-touch-icon.png", "/site.webmanifest", "/web-app-manifest-192x192.png",
  "/web-app-manifest-512x512.png"]);

export default {
  async fetch(request, env) {
    const path = new URL(request.url).pathname;
    if (path === "/api/examples/chat/ws") {
      return env.CHAT_ROOM.get(env.CHAT_ROOM.idFromName("goshtoso-demo")).fetch(request);
    }
    if (rootAssets.has(path) || assetPrefixes.some((prefix) => path.startsWith(prefix))) {
      return env.ASSETS.fetch(request);
    }
    return env.SITE_SERVER.get(env.SITE_SERVER.idFromName("goshtoso-site")).fetch(request);
  },
};

// Go's scheduler (including regexp timeouts in syntax highlighting) needs an
// object-scoped I/O context. Ordinary Worker requests may end while a runtime
// timer is still scheduled, stranding later Go requests waiting on that timer.
export class SiteServer extends DurableObject {
  constructor(ctx, env) {
    super(ctx, env);
    this.runtime = createRuntime();
  }

  fetch(request) {
    return this.runtime.serveGo(request, this.ctx);
  }
}
