//go:build js && wasm

package startpages

import "syscall/js"

// Wrangler includes these source texts as build inputs, so the walkthrough
// matches the starter without needing a filesystem in the Worker.
func readExampleFile(name string) string {
	value := js.Global().Get("goshtosoExampleSources").Get(name)
	if value.Type() != js.TypeString {
		return "// Source file not found: " + name
	}
	return value.String()
}
