package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	chartassets "github.com/araihu/goshtoso-charts/assets"
	libraryassets "github.com/araihu/goshtoso/assets"
)

func TestDemoAssetHandlersUseSharedCachePolicy(t *testing.T) {
	server := newAssetTestServer(t)
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "versioned Goshtoso runtime", path: "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js", want: libraryassets.ImmutableCacheControl},
		{name: "versioned charts runtime", path: chartassets.RuntimeURL, want: libraryassets.ImmutableCacheControl},
		{name: "versioned charts control", path: chartassets.ControlRuntimeURL, want: libraryassets.ImmutableCacheControl},
		{name: "compiled CSS alias", path: "/assets/styles.css", want: libraryassets.RevalidateCacheControl},
		{name: "first-party JavaScript alias", path: "/assets/js/goshtoso.min.js", want: libraryassets.RevalidateCacheControl},
		{name: "icon alias", path: "/assets/icons/heroicons.svg", want: libraryassets.RevalidateCacheControl},
		{name: "logo alias", path: "/assets/images/goshtoso-logo.svg", want: libraryassets.RevalidateCacheControl},
		{name: "site JavaScript alias", path: "/site-assets/js/goshtoso-demo.min.js", want: libraryassets.RevalidateCacheControl},
		{name: "component docs content query", path: "/componentdocshell/assets/goshtoso-logo.svg?v=3c91e915370d", want: libraryassets.RevalidateCacheControl},
		{name: "root favicon alias", path: "/favicon.svg", want: libraryassets.RevalidateCacheControl},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", test.path, recorder.Code)
			}
			if got := recorder.Header().Get("Cache-Control"); got != test.want {
				t.Fatalf("GET %s Cache-Control = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestPackageAndDemoHandlersReturnIdenticalCachePolicy(t *testing.T) {
	server := newAssetTestServer(t)
	for _, path := range []string{
		"/assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
		"/assets/styles.css",
		"/assets/js/goshtoso.min.js",
		"/assets/icons/heroicons.svg",
		"/assets/images/goshtoso-logo.svg",
	} {
		packageResponse := httptest.NewRecorder()
		libraryassets.Handler().ServeHTTP(packageResponse, httptest.NewRequest(http.MethodGet, path, nil))
		demoResponse := httptest.NewRecorder()
		server.ServeHTTP(demoResponse, httptest.NewRequest(http.MethodGet, path, nil))
		if packageResponse.Code != http.StatusOK || demoResponse.Code != http.StatusOK {
			t.Fatalf("GET %s statuses: package=%d demo=%d", path, packageResponse.Code, demoResponse.Code)
		}
		if got, want := demoResponse.Header().Get("Cache-Control"), packageResponse.Header().Get("Cache-Control"); got != want {
			t.Errorf("GET %s demo Cache-Control = %q, package = %q", path, got, want)
		}
	}
}

func TestRootAssetAliasesUseFixedFileNames(t *testing.T) {
	server := newAssetTestServer(t)

	manifest := httptest.NewRecorder()
	server.ServeHTTP(manifest, httptest.NewRequest(http.MethodGet, "/site.webmanifest", nil))
	if manifest.Code != http.StatusOK {
		t.Fatalf("GET /site.webmanifest status = %d, want 200", manifest.Code)
	}
	if got := manifest.Header().Get("Content-Type"); got != "application/manifest+json" {
		t.Fatalf("GET /site.webmanifest Content-Type = %q, want application/manifest+json", got)
	}

	for _, path := range []string{
		"/favicon.svg%2f..%2fsite.webmanifest",
		"/favicon.svg/../../site.webmanifest",
	} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code == http.StatusOK {
				t.Fatalf("GET %s unexpectedly served a root asset", path)
			}
		})
	}
}

func newAssetTestServer(t *testing.T) *Server {
	t.Helper()
	server := &Server{
		projectRoot: serverProjectRoot(t),
		mux:         http.NewServeMux(),
	}
	server.setupAssetRoutes()
	return server
}

func serverProjectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}
