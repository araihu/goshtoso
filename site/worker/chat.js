import { DurableObject } from "cloudflare:workers";
import { createRuntime } from "./runtime.js";

// Hibernation retains sockets and attachments; SQLite retains the last 50
// messages and the connection sequence across object eviction.
export class ChatRoom extends DurableObject {
  constructor(ctx, env) {
    super(ctx, env);
    // Keep Go's scheduler and JS references independent of HTTP/SSE requests.
    this.runtime = createRuntime();
  }
  async fetch(request) {
    if (request.method !== "GET" || request.headers.get("Upgrade")?.toLowerCase() !== "websocket") {
      return new Response("Expected WebSocket upgrade", { status: 426 });
    }
    const origin = request.headers.get("Origin");
    if (origin && origin !== new URL(request.url).origin) {
      return new Response("Forbidden origin", { status: 403 });
    }
    const identity = await this.runtime.chat({ operation: "identity", cookie: request.headers.get("Cookie") || "" });
    const connID = await this.ctx.storage.transaction(async (storage) => {
      const next = (await storage.get("sequence") || 0) + 1;
      await storage.put("sequence", next);
      return next;
    });
    const pair = new WebSocketPair();
    const [client, server] = Object.values(pair);
    server.serializeAttachment({ connID, nick: identity.n });
    this.ctx.acceptWebSocket(server);
    for (const event of await this.ctx.storage.get("history") || []) {
      server.send(await this.runtime.chat({ operation: "render", connID, event }));
    }
    await this.presence(`${identity.n} joined`);
    return new Response(null, { status: 101, webSocket: client });
  }

  async webSocketMessage(socket, message) {
    if (typeof message !== "string" || new TextEncoder().encode(message).byteLength > 8192) {
      socket.close(1009, "Message too large or not text");
      return;
    }
    let data;
    try { data = JSON.parse(message); } catch { return; }
    const { connID, nick } = socket.deserializeAttachment();
    const events = await this.runtime.chat({ operation: "events", connID, nick, data });
    if (!events?.length) return;
    await this.ctx.storage.transaction(async (storage) => {
      const history = [...await storage.get("history") || [], ...events].slice(-50);
      await storage.put("history", history);
    });
    for (const peer of this.ctx.getWebSockets()) {
      const viewer = peer.deserializeAttachment().connID;
      for (const event of events) {
        this.send(peer, await this.runtime.chat({ operation: "render", connID: viewer, event }));
      }
    }
  }

  async webSocketClose(socket) {
    socket.close(1000, "Closed");
    await this.presence(`${socket.deserializeAttachment().nick} left`, socket);
  }

  async webSocketError(socket) {
    socket.close(1011, "WebSocket error");
  }

  async presence(nick, excluded) {
    const sockets = this.ctx.getWebSockets().filter((socket) => socket !== excluded);
    const frame = await this.runtime.chat({ operation: "presence", nick, count: sockets.length });
    for (const socket of sockets) this.send(socket, frame);
  }

  send(socket, frame) {
    try { socket.send(frame); } catch { socket.close(1011, "Send failed"); }
  }
}
