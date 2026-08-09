package assets

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCacheControlClassifiesAssetIdentity(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "Goshtoso runtime", path: "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js", want: ImmutableCacheControl},
		{name: "prerelease runtime", path: "/assets/js/runtime/example/v1.2.3-rc.1+linux-arm64/example.js", want: ImmutableCacheControl},
		{name: "charts runtime", path: "/charts/assets/js/runtime/echarts/5.6.0/echarts.min.js", want: ImmutableCacheControl},
		{name: "charts control generation", path: "/charts/assets/js/controls/5/controls.js", want: ImmutableCacheControl},
		{name: "versioned license", path: "/assets/licenses/tailwindcss/4.3.3/LICENSE.txt", want: ImmutableCacheControl},
		{name: "content hash filename", path: "/assets/app.3c91e915370d.js", want: ImmutableCacheControl},
		{name: "content hash directory", path: "/assets/3c91e915370d3c91e915370d3c91e915/app.js", want: ImmutableCacheControl},
		{name: "compiled CSS alias", path: "/assets/styles.css", want: RevalidateCacheControl},
		{name: "first-party JavaScript alias", path: "/assets/js/goshtoso.min.js", want: RevalidateCacheControl},
		{name: "icon alias", path: "/assets/icons/heroicons.svg", want: RevalidateCacheControl},
		{name: "logo alias", path: "/assets/images/goshtoso-logo.svg", want: RevalidateCacheControl},
		{name: "fallback alias", path: "/assets/images/avatars/avatar-1.webp", want: RevalidateCacheControl},
		{name: "site JavaScript alias", path: "/site-assets/js/goshtoso-demo.min.js", want: RevalidateCacheControl},
		{name: "root favicon alias", path: "/favicon.svg", want: RevalidateCacheControl},
		{name: "latest runtime alias", path: "/assets/js/runtime/alpinejs/latest/alpine.min.js", want: RevalidateCacheControl},
		{name: "partial runtime version", path: "/assets/js/runtime/alpinejs/3.14/alpine.min.js", want: RevalidateCacheControl},
		{name: "invalid leading-zero runtime version", path: "/assets/js/runtime/alpinejs/03.14.9/alpine.min.js", want: RevalidateCacheControl},
		{name: "unversioned control alias", path: "/charts/assets/js/controls/latest/controls.js", want: RevalidateCacheControl},
		{name: "query token does not establish identity", path: "/assets/styles.css?v=3c91e915370d", want: RevalidateCacheControl},
		{name: "long numeric directory is not content hash", path: "/assets/20260808235959/app.js", want: RevalidateCacheControl},
		{name: "traversal rejected", path: "/assets/js/runtime/alpinejs/3.14.9/../latest.js", want: RevalidateCacheControl},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CacheControl(test.path); got != test.want {
				t.Fatalf("CacheControl(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestWithCacheControlEnforcesClassifierOverWrappedHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		downstream string
		want       string
	}{
		{
			name:       "mutable overrides immutable",
			path:       "/assets/styles.css",
			downstream: ImmutableCacheControl,
			want:       RevalidateCacheControl,
		},
		{
			name:       "versioned overrides revalidate",
			path:       "/assets/js/runtime/alpinejs/3.14.9/alpine.min.js",
			downstream: RevalidateCacheControl,
			want:       ImmutableCacheControl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := WithCacheControl(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Cache-Control", test.downstream)
				writer.WriteHeader(http.StatusOK)
			}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if got := recorder.Header().Get("Cache-Control"); got != test.want {
				t.Fatalf("Cache-Control = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWithCacheControlDoesNotMakeMissingVersionedPathImmutable(t *testing.T) {
	handler := WithCacheControl(http.NotFoundHandler())
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/js/runtime/example/1.2.3/missing.js", nil))
	if got := recorder.Header().Get("Cache-Control"); got != RevalidateCacheControl {
		t.Fatalf("404 Cache-Control = %q, want %q", got, RevalidateCacheControl)
	}
}

func TestHandlerDoesNotMakeMissingVersionedPathImmutable(t *testing.T) {
	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/js/runtime/example/1.2.3/missing.js", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != RevalidateCacheControl {
		t.Fatalf("404 Cache-Control = %q, want %q", got, RevalidateCacheControl)
	}
}
