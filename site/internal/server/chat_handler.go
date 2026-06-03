package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/a-h/templ"
	"github.com/coder/websocket"

	"github.com/araihu/goshtoso/site/internal/examples/chat"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
	"github.com/araihu/goshtoso/site/internal/pages/demo/examples"
)

// chatHub is the single process-wide room. RAM-only: a server restart clears it.
var chatHub = chat.NewHub(50)

// chatSeq assigns a deterministic guest seed per new visitor (cookie-less hits).
var chatSeq atomic.Int64

const (
	chatMaxMessageLen = 500             // rune cap per message
	chatMaxNickLen    = 24              // rune cap per nick
	chatWriteTimeout  = 5 * time.Second // per-frame write deadline
)

// registerChatRoutes wires the chat example's endpoints.
func (s *Server) registerChatRoutes() {
	s.mux.HandleFunc("/api/examples/chat/ws", s.handleChatWS)
	s.mux.HandleFunc("/api/examples/chat/rename", s.handleChatRename)
}

// identityFromRequest reads the gt_chat cookie, falling back to a fresh guest.
func identityFromRequest(r *http.Request) chat.Identity {
	if c, err := r.Cookie(chat.CookieName); err == nil {
		if id, err := chat.Decode(c.Value); err == nil && id.Nick != "" {
			return id
		}
	}
	return chat.NewGuest(chatSeq.Add(1))
}

// renderChatPage is the first-load handler for /examples/chat. It ensures a
// gt_chat cookie exists (so the subsequent ws upgrade carries an identity), then
// renders either the full Layout or an HTMX Fragment.
func (s *Server) renderChatPage(w http.ResponseWriter, r *http.Request) {
	me := identityFromRequest(r)
	// Seed (or repair) the cookie when the request carries no usable identity —
	// missing OR undecodable — so the page HTML and the subsequent ws upgrade
	// resolve to the same guest instead of two different fallbacks.
	if c, err := r.Cookie(chat.CookieName); err != nil {
		setChatCookie(r, w, me)
	} else if _, derr := chat.Decode(c.Value); derr != nil {
		setChatCookie(r, w, me)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	content := examples.ChatApp(me)
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		_ = demo.Fragment("Chat", "chat", content).Render(r.Context(), w)
		return
	}
	_ = demo.Layout("Chat", "chat", content).Render(r.Context(), w)
}

// setChatCookie persists the visitor identity in the gt_chat cookie. HttpOnly is
// false so the cookie is also readable client-side if ever needed; SameSite=Lax.
func setChatCookie(r *http.Request, w http.ResponseWriter, me chat.Identity) {
	if !storageAllowed(r) {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     chat.CookieName,
		Value:    me.Encode(),
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

// handleChatRename sets a new nick on the gt_chat cookie and returns OOB fragments
// that update the visible nick and the composer's hidden identity inputs.
func (s *Server) handleChatRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	me := identityFromRequest(r)
	if nick := strings.TrimSpace(r.FormValue("nick")); nick != "" {
		me.Nick = truncateRunes(nick, chatMaxNickLen)
	}
	setChatCookie(r, w, me)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = examples.RenameResult(me).Render(r.Context(), w)
}

// wsFrame is the JSON htmx's ws extension sends on ws-send: the composer's form
// fields plus a HEADERS object we ignore.
type wsFrame struct {
	Message string `json:"message"`
	Nick    string `json:"nick"`
	Color   string `json:"color"`
	Eliza   string `json:"eliza"` // present+truthy when the toggle is on
}

// handleChatWS upgrades to a websocket, joins the room, replays backlog, then
// loops reading messages and broadcasting rendered bubbles.
func (s *Server) handleChatWS(w http.ResponseWriter, r *http.Request) {
	// Default same-origin check (no OriginPatterns override): this enforces the
	// browser's Origin header matching the host, the correct default that blocks
	// cross-site WebSocket hijacking. The demo page and this socket are
	// same-origin, so it works for anyone copying this example.
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.CloseNow() }()

	me := identityFromRequest(r)
	client := chat.NewClient()

	// Join atomically registers the client and replays the ring under one lock,
	// so a racing Broadcast is either fully in the replayed snapshot or delivered
	// live — never duplicated, never dropped.
	chatHub.Join(client)

	// Cancelable context shared by the read loop and the write pump: if the pump
	// exits (write error/timeout), it cancels ctx so conn.Read unblocks and the
	// deferred Unregister + leave-presence below fire promptly.
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Announce arrival to everyone (ephemeral: not retained in the ring, so it is
	// never replayed to future joiners). This client is already registered, so it
	// receives its own join notice with the correct online count.
	chatHub.BroadcastEphemeral(presenceEvent(me.Nick + " joined"))

	go chatWritePump(ctx, cancel, conn, client)
	defer func() {
		chatHub.Unregister(client)
		chatHub.BroadcastEphemeral(presenceEvent(me.Nick + " left"))
	}()

	conn.SetReadLimit(8 << 10) // 8 KiB inbound cap
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return // client closed, errored, or write pump canceled ctx
		}
		for _, ev := range buildChatEvents(data, me.Nick, client.ConnID) {
			chatHub.Broadcast(ev)
		}
	}
}

// presenceEvent builds a presence Event stamped with the current online count.
func presenceEvent(text string) chat.Event {
	return chat.Event{Kind: chat.EventPresence, SystemText: text, Count: chatHub.Count()}
}

// buildChatEvents parses one inbound ws frame into the domain Events to
// broadcast: the sender's message, plus an optional ELIZA reply right after it.
// Returns nil for a blank message. ConnID stamps the message with the sending
// connection so each recipient can decide "mine" at render time; the bot reply
// carries ConnID 0 (never anyone's own). fallbackNick covers a frame with no nick.
func buildChatEvents(data []byte, fallbackNick string, connID uint64) []chat.Event {
	var f wsFrame
	if err := json.Unmarshal(data, &f); err != nil {
		return nil
	}
	text := truncateRunes(strings.TrimSpace(f.Message), chatMaxMessageLen)
	if text == "" {
		return nil
	}
	nick := truncateRunes(strings.TrimSpace(f.Nick), chatMaxNickLen)
	if nick == "" {
		nick = fallbackNick
	}
	now := time.Now().Format("15:04")
	events := []chat.Event{{
		Kind: chat.EventMessage, ConnID: connID, Nick: nick, Color: f.Color, Text: text, Time: now,
	}}
	if isToggleOn(f.Eliza) {
		if reply, ok := chat.Reply(text); ok {
			events = append(events, chat.Event{
				Kind: chat.EventMessage, Nick: "ELIZA", Text: reply, Time: time.Now().Format("15:04"), IsBot: true,
			})
		}
	}
	return events
}

// renderChatEvent renders an Event into the HTML frame for one specific viewer.
// A message is "mine" (right-aligned, primary fill) only when its sending
// connection equals the viewer's — decided here, server-side, per connection, so
// no client JavaScript is needed to align own messages.
func renderChatEvent(viewerConnID uint64, ev chat.Event) []byte {
	if ev.Kind == chat.EventPresence {
		return renderFrame(examples.PresenceFrame(ev.SystemText, ev.Count))
	}
	mine := !ev.IsBot && ev.ConnID == viewerConnID
	return renderFrame(examples.MessageFrame(examples.Message{
		Nick: ev.Nick, Color: ev.Color, Text: ev.Text, Time: ev.Time, IsBot: ev.IsBot, Mine: mine,
	}))
}

// chatWritePump drains a client's send channel, rendering each Event for THIS
// client and writing it to the socket, until the channel closes or a write
// fails. On return it cancels the shared context so the read loop's blocked
// conn.Read unblocks and the handler tears the client down.
func chatWritePump(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, client *chat.Client) {
	defer cancel()
	for ev := range client.Send {
		wctx, wcancel := context.WithTimeout(ctx, chatWriteTimeout)
		err := conn.Write(wctx, websocket.MessageText, renderChatEvent(client.ConnID, ev))
		wcancel()
		if err != nil {
			return
		}
	}
}

// renderFrame renders a templ component to bytes for pushing over the socket.
func renderFrame(c templ.Component) []byte {
	var buf bytes.Buffer
	_ = c.Render(context.Background(), &buf)
	return buf.Bytes()
}

// isToggleOn reports whether the htmx-serialized toggle value is truthy.
func isToggleOn(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "off", "false", "0":
		return false
	default:
		return true
	}
}

// truncateRunes caps s to at most n runes.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
