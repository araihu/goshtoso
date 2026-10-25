package server

import (
	"encoding/json"
	"net/http"

	"github.com/araihu/goshtoso/site/internal/examples/chat"
)

// WorkerChatRequest carries transport-independent chat operations from a
// Durable Object. Go owns identity, validation, ELIZA, escaping, and templates;
// the object owns sockets, connection IDs, and the bounded room history.
type WorkerChatRequest struct {
	Operation string          `json:"operation"`
	Cookie    string          `json:"cookie"`
	Data      json.RawMessage `json:"data"`
	Nick      string          `json:"nick"`
	ConnID    uint64          `json:"connID"`
	Event     chat.Event      `json:"event"`
	Count     int             `json:"count"`
}

// WorkerChat executes one synchronous operation without opening a TCP socket.
func WorkerChat(input WorkerChatRequest) any {
	switch input.Operation {
	case "identity":
		r, _ := http.NewRequest(http.MethodGet, "https://chat.invalid/", nil)
		r.Header.Set("Cookie", input.Cookie)
		return identityFromRequest(r)
	case "events":
		return buildChatEvents(input.Data, input.Nick, input.ConnID)
	case "render":
		return string(renderChatEvent(input.ConnID, input.Event))
	case "presence":
		return string(renderChatEvent(0, chat.Event{Kind: chat.EventPresence, SystemText: input.Nick, Count: input.Count}))
	default:
		return nil
	}
}
