package wizard

import (
	"net/mail"
	"strings"
)

// minPasswordLen is the smallest accepted password length in runes.
const minPasswordLen = 8

// ValidateAccount checks step-1 fields and returns a field->error map. An empty
// map means the step is valid. Keys are the form field names ("name", "email",
// "password") so the templ layer can attach each message to its input.
func ValidateAccount(a Account) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(a.Name) == "" {
		errs["name"] = "Name is required."
	}
	email := strings.TrimSpace(a.Email)
	if email == "" {
		errs["email"] = "Email is required."
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "Enter a valid email address."
	}
	if len([]rune(a.Password)) < minPasswordLen {
		errs["password"] = "Password must be at least 8 characters."
	}
	return errs
}

// ValidateAddress checks step-2 fields and returns a field->error map.
func ValidateAddress(a Address) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(a.Line1) == "" {
		errs["line1"] = "Address line is required."
	}
	if strings.TrimSpace(a.City) == "" {
		errs["city"] = "City is required."
	}
	if strings.TrimSpace(a.Country) == "" {
		errs["country"] = "Country is required."
	}
	if strings.TrimSpace(a.Postal) == "" {
		errs["postal"] = "Postal code is required."
	}
	return errs
}

// ValidatePlan checks step-3 and returns a field->error map.
func ValidatePlan(plan string) map[string]string {
	errs := map[string]string{}
	if !ValidPlan(plan) {
		errs["plan"] = "Choose a plan to continue."
	}
	return errs
}

// ValidateStep runs the validator for the given step number against the state.
// Steps with no input (e.g. the read-only review) return an empty map.
func ValidateStep(step int, s WizardState) map[string]string {
	switch step {
	case 1:
		return ValidateAccount(s.Account)
	case 2:
		return ValidateAddress(s.Address)
	case 3:
		return ValidatePlan(s.Plan)
	default:
		return map[string]string{}
	}
}
