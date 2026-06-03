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

func TestBreedsPageRendersRoundDogPhotos(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	appMux().ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `/dog-images/australian-shepherd.webp`) {
		t.Fatalf("page must render dog image cells:\n%s", body)
	}
	if !strings.Contains(body, `rounded-full`) {
		t.Fatalf("dog photos must use round thumbnail styling:\n%s", body)
	}
}

func TestDogImagesAreServed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dog-images/labrador-retriever.webp", nil)
	rec := httptest.NewRecorder()

	appMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("dog image status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "image/webp") {
		t.Fatalf("dog image content-type = %q, want image/webp", ct)
	}
	if size := rec.Body.Len(); size == 0 {
		t.Fatal("dog image body is empty")
	}
}
