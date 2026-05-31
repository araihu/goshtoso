// Package todo holds the pure, HTTP-free domain model for the /examples/todo
// app. State is serialized into a cookie, so the server keeps no per-user memory.
package todo

import (
	"encoding/base64"
	"encoding/json"
)

const (
	// MaxTodos is a coarse upper bound on the list length. The real size guarantee
	// is enforced by maxCookieBytes: Add refuses to append when the resulting
	// encoded state would exceed that budget.
	MaxTodos = 50
	// MaxTitleLen bounds a single title's stored length in runes.
	MaxTitleLen = 200
	// maxCookieBytes bounds the encoded cookie value so the browser never silently
	// drops it (browsers cap a cookie near 4KB).
	maxCookieBytes = 3800
)

// Todo is a single task. ID is assigned from State.Seq (deterministic, no rand).
type Todo struct {
	ID       int    `json:"i"`
	Title    string `json:"t"`
	Done     bool   `json:"d"`
	Priority string `json:"p"` // "low" | "med" | "high"
	Due      string `json:"u"` // "2006-01-02" or ""
	Order    int    `json:"o"` // explicit sort index
}

// State is the whole per-user todo list plus view filter and the next ID counter.
type State struct {
	Todos  []Todo `json:"todos"`
	Filter string `json:"filter"` // "all" | "active" | "done"
	Seq    int    `json:"seq"`
}

// Encode serializes State to a base64url(JSON) string for cookie storage.
// State is always serializable; a marshal error is a programmer error and panics.
func Encode(s State) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic("todo.Encode: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a cookie value back into State. Any error yields the zero State
// so a corrupt/absent cookie degrades gracefully to "empty list".
func Decode(raw []byte) (State, error) {
	var s State
	if len(raw) == 0 {
		return s, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(string(raw))
	if err != nil {
		return State{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	return s, nil
}
