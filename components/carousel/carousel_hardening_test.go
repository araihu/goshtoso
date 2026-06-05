package carousel

import (
	"strings"
	"testing"
)

func TestSlidesToJSONEscapesControlCharactersInAlpineStrings(t *testing.T) {
	payload := "value'\n\r\t\\\u2028\u2029</script>"
	out := slidesToJSON([]Slide{{
		ImgSrc:      payload,
		ImgAlt:      payload,
		Title:       payload,
		Description: payload,
		CTAHref:     payload,
		CTALabel:    payload,
	}})

	want := `value\'\n\r\t\\\u2028\u2029</script>`
	for _, field := range []string{"imgSrc", "imgAlt", "title", "description", "ctaUrl", "ctaText"} {
		if !strings.Contains(out, field+":'"+want+"'") {
			t.Fatalf("slidesToJSON missing escaped %s payload in:\n%s", field, out)
		}
	}
	if strings.Contains(out, "value'\n") || strings.Contains(out, "\r") || strings.Contains(out, "\u2028") || strings.Contains(out, "\u2029") {
		t.Fatalf("slidesToJSON emitted raw JS line terminator:\n%s", out)
	}
}

func TestSlidesToJSONSanitizesExecutableCTAHref(t *testing.T) {
	out := slidesToJSON([]Slide{{
		ImgSrc:   "/safe.webp",
		ImgAlt:   "Safe",
		CTAHref:  "javascript:alert(1)",
		CTALabel: "Open",
	}})

	if strings.Contains(out, "javascript:alert(1)") {
		t.Fatalf("slidesToJSON emitted executable CTA href:\n%s", out)
	}
	if !strings.Contains(out, "ctaUrl:'about:invalid#TemplFailedSanitizationURL'") {
		t.Fatalf("slidesToJSON missing inert CTA fallback:\n%s", out)
	}
}
