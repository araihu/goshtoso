package docs_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestConsumerButtonExamplesUseFunctionalOptions(t *testing.T) {
	channels := map[string]struct {
		path           string
		wantButtonType string
	}{
		"README": {
			path:           "../README.md",
			wantButtonType: `button.WithType("button")`,
		},
		"usage guide": {
			path:           "USAGE.md",
			wantButtonType: `button.WithType("submit")`,
		},
		"agent skill": {
			path:           "../.agents/skills/using-goshtoso/SKILL.md",
			wantButtonType: `button.WithType("button")`,
		},
	}

	for name, channel := range channels {
		t.Run(name, func(t *testing.T) {
			content := readDoc(t, channel.path)
			for _, removed := range []string{
				"button.Config",
				"button.Primary",
				"Variant:",
			} {
				if strings.Contains(content, removed) {
					t.Errorf("%s contains removed Button API %q", channel.path, removed)
				}
			}
			for _, current := range []string{
				"@button.Button(",
				"button.WithTone(button.TonePrimary)",
				channel.wantButtonType,
			} {
				if !strings.Contains(content, current) {
					t.Errorf("%s missing current Button API %q", channel.path, current)
				}
			}
		})
	}
}

func TestUsageCatalogDocumentsSeparatePrimitives(t *testing.T) {
	content := readDoc(t, "USAGE.md")

	for _, boundary := range []string{
		"separate `Banner` and `CookieBanner` primitives",
		"separate `Modal` and `AlertDialog` primitives; `Tone` belongs to `AlertDialog`",
		"separate `Toast` and `MessageToast` primitives; sender and avatar content belongs to `MessageToast`",
	} {
		if !strings.Contains(content, boundary) {
			t.Errorf("USAGE.md missing primitive boundary %q", boundary)
		}
	}

	for _, stale := range []string{
		"cookie consent variant",
		"Dialogs with info/danger/warning variants",
		"position, sender avatar",
	} {
		if strings.Contains(content, stale) {
			t.Errorf("USAGE.md contains stale primitive description %q", stale)
		}
	}
}

func TestConsumerChannelsAvoidObsoleteVariantVocabulary(t *testing.T) {
	variant := regexp.MustCompile(`(?i)\bvariants?\b`)
	for _, path := range []string{
		"../README.md",
		"USAGE.md",
		"../.agents/skills/using-goshtoso/SKILL.md",
	} {
		if matches := variant.FindAllString(readDoc(t, path), -1); len(matches) > 0 {
			t.Errorf("%s contains obsolete component Variant vocabulary: %v", path, matches)
		}
	}
}

func readDoc(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
