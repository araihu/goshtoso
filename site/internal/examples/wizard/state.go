// Package wizard holds the pure, HTTP-free domain model for the /examples/wizard
// onboarding/checkout app. State is serialized into a cookie, so the server keeps
// no per-user memory.
package wizard

import (
	"encoding/base64"
	"encoding/json"
)

const (
	// FirstStep and LastStep bound the wizard's step range (1..4).
	FirstStep = 1
	LastStep  = 4
	// maxFieldLen caps a single stored field's length in runes so a hostile or
	// accidental large input cannot blow the cookie budget.
	maxFieldLen = 200
	// maxCookieBytes bounds the encoded cookie value so the browser never silently
	// drops it (browsers cap a cookie near 4KB).
	maxCookieBytes = 3800
)

// Account is step 1: the user's identity.
type Account struct {
	Name     string `json:"n"`
	Email    string `json:"e"`
	Password string `json:"pw"`
}

// Address is step 2: where the user lives.
type Address struct {
	Line1   string `json:"l"`
	City    string `json:"c"`
	Country string `json:"co"`
	Postal  string `json:"z"`
}

// WizardState is the whole per-user onboarding flow: current step, the data
// entered so far, the chosen plan, and whether the flow was confirmed.
type WizardState struct {
	Step    int     `json:"s"` // current step 1..4
	Account Account `json:"a"`
	Address Address `json:"ad"`
	Plan    string  `json:"p"` // "free" | "pro" | "team"
	Done    bool    `json:"d"`
}

// Encode serializes WizardState to a base64url(JSON) string for cookie storage.
// State is always serializable; a marshal error is a programmer error and panics.
func Encode(s WizardState) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic("wizard.Encode: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a cookie value back into WizardState. Any error yields the zero
// state so a corrupt/absent cookie degrades gracefully to "fresh wizard".
func Decode(raw []byte) (WizardState, error) {
	var s WizardState
	if len(raw) == 0 {
		return s, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(string(raw))
	if err != nil {
		return WizardState{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return WizardState{}, err
	}
	return s, nil
}

// Normalized returns a copy with Step clamped into the valid 1..4 range. A zero
// (or out-of-range) Step from a fresh/corrupt cookie becomes FirstStep.
func (s WizardState) Normalized() WizardState {
	if s.Step < FirstStep {
		s.Step = FirstStep
	}
	if s.Step > LastStep {
		s.Step = LastStep
	}
	return s
}
