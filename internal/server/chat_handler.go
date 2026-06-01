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

	"github.com/araihu/goshtoso/internal/examples/chat"
	"github.com/araihu/goshtoso/internal/pages/demo"
	"github.com/araihu/goshtoso/internal/pages/demo/examples"
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
		setChatCookie(w, me)
	} else if _, derr := chat.Decode(c.Value); derr != nil {
		setChatCookie(w, me)
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
func setChatCookie(w http.ResponseWriter, me chat.Identity) {
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
	setChatCookie(w, me)
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
	// receives its own join frame with the correct online count.
	chatHub.BroadcastEphemeral(renderFrame(examples.PresenceFrame(me.Nick+" joined", chatHub.Count())))

	go chatWritePump(ctx, cancel, conn, client)
	defer func() {
		chatHub.Unregister(client)
		chatHub.BroadcastEphemeral(renderFrame(examples.PresenceFrame(me.Nick+" left", chatHub.Count())))
	}()

	conn.SetReadLimit(8 << 10) // 8 KiB inbound cap
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return // client closed, errored, or write pump canceled ctx
		}
		now, bot := buildMessageFrames(data, me.Nick)
		if now != nil {
			chatHub.Broadcast(now)
		}
		if bot != nil {
			chatHub.Broadcast(bot)
		}
	}
}

// buildMessageFrames parses one inbound ws frame and renders the message bubble
// to broadcast now, plus an optional ELIZA reply bubble to broadcast immediately
// after it. Either return may be nil (blank message → now nil; bot off or no
// match → bot nil). fallbackNick is used when the frame omits a nick.
func buildMessageFrames(data []byte, fallbackNick string) (now, bot []byte) {
	var f wsFrame
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, nil
	}
	text := truncateRunes(strings.TrimSpace(f.Message), chatMaxMessageLen)
	if text == "" {
		return nil, nil
	}
	nick := truncateRunes(strings.TrimSpace(f.Nick), chatMaxNickLen)
	if nick == "" {
		nick = fallbackNick
	}
	now = renderFrame(examples.MessageFrame(examples.Message{
		Nick: nick, Color: f.Color, Text: text, Time: time.Now().Format("15:04"),
	}))
	if isToggleOn(f.Eliza) {
		if reply, ok := chat.Reply(text); ok {
			bot = renderFrame(examples.MessageFrame(examples.Message{
				Nick: "ELIZA", Text: reply, Time: time.Now().Format("15:04"), IsBot: true,
			}))
		}
	}
	return now, bot
}

// chatWritePump drains a client's send channel to the socket until it closes or
// a write fails. On return it cancels the shared context so the read loop's
// blocked conn.Read unblocks and the handler tears the client down.
func chatWritePump(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, client *chat.Client) {
	defer cancel()
	for frame := range client.Send {
		wctx, wcancel := context.WithTimeout(ctx, chatWriteTimeout)
		err := conn.Write(wctx, websocket.MessageText, frame)
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
