package fileinput

import (
	"github.com/a-h/templ"
	"strings"
	"testing"
)

func TestErrorStateUsesNativeValidationHook(t *testing.T) {
	for _, appearance := range []Appearance{AppearanceDropZone, AppearanceUpload} {
		html := renderConfig(t, Config{ID: "logo", Appearance: appearance, State: StateError, HelperText: "Image rejected", InputAttrs: templ.Attributes{"aria-describedby": "logo-error"}})
		for _, want := range []string{`aria-invalid="true"`, `aria-describedby="logo-error logo-helper"`, `group/fileinput`, `Image rejected`} {
			if !strings.Contains(html, want) {
				t.Fatalf("%s missing %s", appearance, want)
			}
		}
	}
}
