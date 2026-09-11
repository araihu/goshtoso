package combobox

import "strings"

// Mutation buttons are disabled until the response settles. Search is separately
// made read-only so a focused query keeps its focus and caret.
// The root request barrier explicitly cancels stale reads before mutating selection.
func mutationDisableSelector(cfg Config) string {
	id := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\a `, "\r", `\d `).Replace(cfg.ID)
	return `[id="` + id + `"] button:not(:disabled)`
}
