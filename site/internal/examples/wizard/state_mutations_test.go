// internal/examples/wizard/state_mutations_test.go
package wizard

import (
	"strings"
	"testing"
)

func TestValidateAccount(t *testing.T) {
	cases := []struct {
		name     string
		acct     Account
		wantKeys []string
	}{
		{"all valid", Account{Name: "Ada", Email: "ada@example.com", Password: "hunter2hunter"}, nil},
		{"empty name", Account{Email: "ada@example.com", Password: "hunter2hunter"}, []string{"name"}},
		{"bad email", Account{Name: "Ada", Email: "nope", Password: "hunter2hunter"}, []string{"email"}},
		{"empty email", Account{Name: "Ada", Password: "hunter2hunter"}, []string{"email"}},
		{"short password", Account{Name: "Ada", Email: "ada@example.com", Password: "short"}, []string{"password"}},
		{"all bad", Account{}, []string{"name", "email", "password"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateAccount(tc.acct)
			if len(errs) != len(tc.wantKeys) {
				t.Fatalf("got errs %v, want keys %v", errs, tc.wantKeys)
			}
			for _, k := range tc.wantKeys {
				if errs[k] == "" {
					t.Errorf("expected error for %q, got none (errs=%v)", k, errs)
				}
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	full := Address{Line1: "1 Main", City: "London", Country: "UK", Postal: "EC1"}
	if errs := ValidateAddress(full); len(errs) != 0 {
		t.Fatalf("valid address should pass, got %v", errs)
	}
	if errs := ValidateAddress(Address{}); len(errs) != 4 {
		t.Fatalf("empty address should fail all 4 fields, got %v", errs)
	}
}

func TestValidatePlan(t *testing.T) {
	if errs := ValidatePlan("pro"); len(errs) != 0 {
		t.Fatalf("pro should be valid, got %v", errs)
	}
	if errs := ValidatePlan(""); errs["plan"] == "" {
		t.Fatalf("empty plan should error")
	}
	if errs := ValidatePlan("enterprise"); errs["plan"] == "" {
		t.Fatalf("unknown plan should error")
	}
}

func TestSetPlanRejectsUnknown(t *testing.T) {
	s := WizardState{Plan: "pro"}
	s.SetPlan("enterprise")
	if s.Plan != "pro" {
		t.Fatalf("unknown plan should not overwrite, got %q", s.Plan)
	}
	s.SetPlan("team")
	if s.Plan != "team" {
		t.Fatalf("valid plan should set, got %q", s.Plan)
	}
}

func TestSetAccountTrimsAndCaps(t *testing.T) {
	s := WizardState{}
	s.SetAccount("  Ada  ", "  ada@example.com ", "hunter2hunter")
	if s.Account.Name != "Ada" || s.Account.Email != "ada@example.com" {
		t.Fatalf("fields not trimmed: %+v", s.Account)
	}
	long := strings.Repeat("x", maxFieldLen+50)
	s.SetAccount(long, "a@b.co", "pwpwpwpw")
	if len([]rune(s.Account.Name)) != maxFieldLen {
		t.Fatalf("name not capped to %d runes, got %d", maxFieldLen, len([]rune(s.Account.Name)))
	}
}

func TestAdvanceAndBack(t *testing.T) {
	s := WizardState{Step: 1}
	s.Advance()
	if s.Step != 2 {
		t.Fatalf("advance 1->2, got %d", s.Step)
	}
	s.Step = LastStep
	s.Advance()
	if s.Step != LastStep {
		t.Fatalf("advance past last should be no-op, got %d", s.Step)
	}
	s.Done = true
	s.Back()
	if s.Step != LastStep-1 || s.Done {
		t.Fatalf("back should decrement and clear Done, got step=%d done=%v", s.Step, s.Done)
	}
	s.Step = FirstStep
	s.Back()
	if s.Step != FirstStep {
		t.Fatalf("back below first should clamp, got %d", s.Step)
	}
}

func TestConfirmAndReset(t *testing.T) {
	s := WizardState{Step: 4}
	s.SetAccount("Ada", "ada@example.com", "hunter2hunter")
	s.Confirm()
	if !s.Done {
		t.Fatalf("confirm should set Done")
	}
	s.Reset()
	if s.Step != FirstStep || s.Done || s.Account.Name != "" || s.Plan != "" {
		t.Fatalf("reset should clear everything, got %+v", s)
	}
}

func TestValidateStepDispatch(t *testing.T) {
	s := WizardState{
		Account: Account{Name: "Ada", Email: "ada@example.com", Password: "hunter2hunter"},
		Address: Address{Line1: "1 Main", City: "London", Country: "UK", Postal: "EC1"},
		Plan:    "pro",
	}
	for step := 1; step <= 4; step++ {
		if errs := ValidateStep(step, s); len(errs) != 0 {
			t.Errorf("step %d on full valid state should pass, got %v", step, errs)
		}
	}
	if errs := ValidateStep(1, WizardState{}); len(errs) == 0 {
		t.Fatalf("step 1 on empty state should fail")
	}
}
