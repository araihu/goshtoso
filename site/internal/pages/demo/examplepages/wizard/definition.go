// Package wizardpage owns the Onboarding Wizard runnable example page.
package wizardpage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Onboarding Wizard example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/wizard",
	Title:   "Onboarding Wizard",
	Active:  "wizard",
	Type:    "SoftwareSourceCode",
	Content: WizardContent,
}
