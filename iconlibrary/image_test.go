package iconlibrary

import "testing"

func TestRejectActiveSVG(t *testing.T) {
	for _, body := range []string{`<script/>`, `<image href="https://example.com/a.png"/>`, `<path onclick="alert(1)"/>`, `<path style="fill:red"/>`, `<use href="https://example.com/a.svg#x"/>`} {
		if _, _, _, err := ValidateImage([]byte(`<svg xmlns="http://www.w3.org/2000/svg">` + body + `</svg>`)); err == nil {
			t.Errorf("accepted %s", body)
		}
	}
	if mime, _, _, err := ValidateImage([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#fff" d="M0 0"/></svg>`)); err != nil || mime != "image/svg+xml" {
		t.Fatalf("%s %v", mime, err)
	}
}
