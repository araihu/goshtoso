package carousel

import (
	"encoding/json"
	"testing"
)

func TestSlidesJSONPreservesControlCharactersAsData(t *testing.T) {
	payload := "value'\n\r\t\\\u2028\u2029</script>"
	out := slidesJSON([]Slide{{
		ImgSrc:      payload,
		ImgAlt:      payload,
		Title:       payload,
		Description: payload,
		CTAHref:     payload,
		CTALabel:    payload,
	}})

	var data []slideData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("slidesJSON must return valid data JSON: %v", err)
	}
	if len(data) != 1 || data[0].ImgSrc != payload || data[0].CTAURL != payload {
		t.Fatalf("slidesJSON changed configured data: %#v", data)
	}
}

func TestSlidesJSONSanitizesExecutableCTAHref(t *testing.T) {
	out := slidesJSON([]Slide{{
		ImgSrc:   "/safe.webp",
		ImgAlt:   "Safe",
		CTAHref:  "javascript:alert(1)",
		CTALabel: "Open",
	}})

	var data []slideData
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatal(err)
	}
	if data[0].CTAURL != "about:invalid#TemplFailedSanitizationURL" {
		t.Fatalf("slidesJSON missing inert CTA fallback: %#v", data)
	}
}
