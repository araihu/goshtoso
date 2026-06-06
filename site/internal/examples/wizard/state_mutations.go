// internal/examples/wizard/state_mutations.go
package wizard

import "strings"

// capField trims surrounding whitespace and caps a stored field to maxFieldLen
// runes without splitting a multibyte rune.
func capField(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxFieldLen {
		return s
	}
	r := []rune(s)
	if len(r) <= maxFieldLen {
		return s
	}
	return string(r[:maxFieldLen])
}

// SetAccount stores the step-1 fields. The password is trimmed only of trailing
// newlines a browser might send; internal spaces are preserved, but it is still
// length-capped. Values are stored as entered; validation is separate.
func (s *WizardState) SetAccount(name, email, password string) {
	s.Account = Account{
		Name:     capField(name),
		Email:    capField(email),
		Password: capField(password),
	}
}

// SetAddress stores the step-2 fields.
func (s *WizardState) SetAddress(line1, city, country, postal string) {
	s.Address = Address{
		Line1:   capField(line1),
		City:    capField(city),
		Country: capField(country),
		Postal:  capField(postal),
	}
}

// SetPlan stores the step-3 selection only when it names a real tier; an unknown
// value leaves the previous selection untouched.
func (s *WizardState) SetPlan(plan string) {
	if ValidPlan(plan) {
		s.Plan = plan
	}
}

// Advance moves to the next step when below LastStep. It does not validate; the
// handler validates the current step before calling Advance. Advancing past the
// budget is refused (no-op) so a giant input cannot produce an unstorable cookie.
func (s *WizardState) Advance() {
	if s.Step < FirstStep {
		s.Step = FirstStep
	}
	if s.Step >= LastStep {
		return
	}
	candidate := *s
	candidate.Step++
	if len(Encode(candidate)) > maxCookieBytes {
		return
	}
	s.Step = candidate.Step
}

// Back moves to the previous step, never below FirstStep. It clears Done so the
// success screen does not "stick" if the user navigates back after confirming.
func (s *WizardState) Back() {
	s.Done = false
	if s.Step <= FirstStep {
		s.Step = FirstStep
		return
	}
	s.Step--
}

// Confirm marks the flow complete. It is the handler's responsibility to have
// validated every step first.
func (s *WizardState) Confirm() {
	s.Done = true
}

// Reset returns the wizard to a fresh first step with no entered data.
func (s *WizardState) Reset() {
	*s = WizardState{Step: FirstStep}
}
