package assets

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesEveryMuambaRuntimeDownload(t *testing.T) {
	server := httptest.NewServer(Handler())
	t.Cleanup(server.Close)
	for _, resource := range MuambaResources() {
		for _, download := range resource.Downloads {
			if !strings.HasPrefix(download.Path, "assets/") {
				continue
			}
			local := "/assets/" + strings.TrimPrefix(download.Path, "assets/")
			response, err := http.Get(server.URL + local)
			if err != nil {
				t.Fatalf("GET %s: %v", local, err)
			}
			got, readErr := io.ReadAll(response.Body)
			_ = response.Body.Close()
			file, openErr := MuambaOpen(resource.Name, download.Name)
			if openErr != nil {
				t.Fatal(openErr)
			}
			want, wantErr := io.ReadAll(file)
			_ = file.Close()
			if readErr != nil || wantErr != nil || response.StatusCode != http.StatusOK || !bytes.Equal(got, want) {
				t.Fatalf("served %s differs: status=%d read=%v want=%v", local, response.StatusCode, readErr, wantErr)
			}
			if strings.Contains(local, "/js/runtime/") && response.Header.Get("Cache-Control") != "public, max-age=31536000, immutable" {
				t.Errorf("GET %s Cache-Control = %q", local, response.Header.Get("Cache-Control"))
			}
		}
	}
}

func TestFirstPartyEmbedDoesNotContainMuambaRuntimeTree(t *testing.T) {
	if _, err := files.Open("js/runtime/alpinejs/3.14.9/alpine.min.js"); err == nil {
		t.Fatal("first-party embed contains Muamba bytes")
	}
}

func TestHandlerDoesNotPublishTailwindExecutable(t *testing.T) {
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/.tools/tailwindcss", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}
