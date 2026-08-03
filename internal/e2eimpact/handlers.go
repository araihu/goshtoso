package e2eimpact

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var handlerIdentities = map[string][]string{
	"accordion_handler.go":       {"accordion"},
	"banner_handler.go":          {"banner"},
	"button_handler.go":          {"button"},
	"carousel_handler.go":        {"carousel"},
	"chat_handler.go":            {"example_chat"},
	"dropdown_handler.go":        {"dropdown"},
	"expense_handler.go":         {"example_expense"},
	"form_handler.go":            {"form"},
	"form_validation_handler.go": {"form", "textinput"},
	"logs_handler.go":            {"example_logs"},
	"profile_handler.go":         {"example_profile"},
	"radio_handler.go":           {"radio"},
	"search_handler.go":          {"search"},
	"steps_handler.go":           {"steps"},
	"table_handler.go":           {"table"},
	"table_fragments.templ":      {"table"},
	"table_fragments_templ.go":   {"table"},
	"tabs_handler.go":            {"tabs"},
	"ticker_handler.go":          {"example_ticker"},
	"toast_handler.go":           {"toast"},
	"todo_handler.go":            {"example_todo"},
	"wizard_handler.go":          {"example_wizard"},
}

var sharedServerFiles = map[string]bool{
	"getting_started_handler.go": true,
	"server.go":                  true,
	"storage_consent.go":         true,
}

func validateHandlerOwnership(repoRoot string, known map[string]bool) error {
	directory := filepath.Join(repoRoot, "site/internal/server")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		identities, mapped := handlerIdentities[name]
		if !mapped && !sharedServerFiles[name] {
			return fmt.Errorf("site/internal/server/%s is neither identity-owned nor shared", name)
		}
		for _, identity := range identities {
			if !known[identity] {
				return fmt.Errorf("site/internal/server/%s maps unknown identity %q", name, identity)
			}
		}
	}
	return nil
}

func identitiesForHandler(path string) ([]string, bool) {
	identities, ok := handlerIdentities[filepath.Base(path)]
	return slices.Clone(identities), ok
}
