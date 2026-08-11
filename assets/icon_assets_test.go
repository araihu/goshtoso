package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesHeroiconsSpriteAndLicense(t *testing.T) {
	tests := []struct {
		path        string
		contentType string
		sha256      string
	}{
		{
			path:        "/assets/icons/heroicons.svg",
			contentType: "image/svg+xml",
			sha256:      "65cdb814125787460b548428dd49edd8e29250ee9eba5e6f27f4eb1b746fc3ca",
		},
		{
			path:        "/assets/icons/HEROICONS_LICENSE.txt",
			contentType: "text/plain; charset=utf-8",
			sha256:      "60e0b68c0f35c078eef3a5d29419d0b03ff84ec1df9c3f9d6e39a519a5ae7985",
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", test.path, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, test.contentType) {
				t.Errorf("GET %s Content-Type = %q, want prefix %q", test.path, got, test.contentType)
			}
			if got := recorder.Header().Get("Cache-Control"); got != RevalidateCacheControl {
				t.Errorf("GET %s Cache-Control = %q, want revalidation policy", test.path, got)
			}
			sum := sha256.Sum256(recorder.Body.Bytes())
			if got := hex.EncodeToString(sum[:]); got != test.sha256 {
				t.Errorf("GET %s SHA-256 = %s, want %s", test.path, got, test.sha256)
			}
		})
	}
}
