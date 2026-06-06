// Package expense holds the pure, HTTP-free domain model for the /examples/expense
// app: a small expense tracker. State is serialized into a cookie, so the server
// keeps no per-user memory (the same stateless pattern as the todo example).
package expense

import (
	"encoding/base64"
	"encoding/json"
)

const (
	// MaxItems is a coarse upper bound on the list length. The real size guarantee
	// is enforced by maxCookieBytes: Add refuses to append when the resulting
	// encoded state would exceed that budget.
	MaxItems = 100
	// MaxDescLen bounds a single description's stored length in runes.
	MaxDescLen = 120
	// PerPage is the number of visible rows shown per page.
	PerPage = 8
	// maxCookieBytes bounds the encoded cookie value so the browser never silently
	// drops it (browsers cap a cookie near 4KB).
	maxCookieBytes = 3800
)

// Categories is the fixed set of spend categories. The first paint, the add-form
// combobox, and the filter select all read from this single source of truth.
var Categories = []string{
	"Food",
	"Transport",
	"Housing",
	"Health",
	"Entertainment",
	"Other",
}

// Expense is a single line item. ID is assigned from State.Seq (deterministic,
// no rand). AmountCents stores money as an integer count of cents so the domain
// never touches floating point.
type Expense struct {
	ID          int    `json:"i"`
	Desc        string `json:"d"`
	AmountCents int    `json:"a"`
	Category    string `json:"c"`
	Date        string `json:"t"` // "2006-01-02" or ""
	Order       int    `json:"o"` // explicit sort index (newest-first on add)
}

// State is the whole per-user expense list plus the active filters, the current
// page, and the next ID counter.
type State struct {
	Items    []Expense `json:"items"`
	Search   string    `json:"q"`
	Category string    `json:"cat"` // "" means "all categories"
	Page     int       `json:"pg"`  // 1-indexed; 0 is normalized to 1
	Seq      int       `json:"seq"`
}

// Encode serializes State to a base64url(JSON) string for cookie storage.
// State is always serializable; a marshal error is a programmer error and panics.
func Encode(s State) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic("expense.Encode: " + err.Error())
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
