package form

import (
	"encoding/json"
	"html"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidationValuesStaticIdentity(t *testing.T) {
	for _, name := range []string{"", "email", "quote\"'&<\nfield"} {
		t.Run(name, func(t *testing.T) {
			cfg := FieldGroupConfig{Validation: &ValidationConfig{Endpoint: "/validate"}}
			if name != "" {
				cfg.Meta = &FieldMeta{FieldName: name}
			}
			markup := render(t, FieldGroup(cfg))
			match := regexp.MustCompile(`hx-vals="([^"]*)"`).FindStringSubmatch(markup)
			require.Len(t, match, 2)
			var values map[string]string
			require.NoError(t, json.Unmarshal([]byte(html.UnescapeString(match[1])), &values))
			require.Equal(t, map[string]string{"X-Goshtoso-Validation": "field", "X-Goshtoso-Field": name}, values)
			require.NotContains(t, markup, "js:")
			require.NotContains(t, markup, " name=")
		})
	}
}
