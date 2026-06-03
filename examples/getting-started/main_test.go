package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBreedsAPIUpdatesPaginatorForOnePageFilter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/breeds?search=USA&order_by=breed&order_dir=asc&page=1&per_page=5", nil)
	rec := httptest.NewRecorder()

	appMux().ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `hx-swap-oob="true"`) {
		t.Fatalf("filtered response must include paginator OOB replacement:\n%s", body)
	}
	if !strings.Contains(body, "Page 1 of 1") {
		t.Fatalf("filtered response paginator = %q, want Page 1 of 1", body)
	}
	if strings.Contains(body, "page 2") || strings.Contains(body, `page=2`) {
		t.Fatalf("one-page filtered response must not keep stale page 2 controls:\n%s", body)
	}
}
