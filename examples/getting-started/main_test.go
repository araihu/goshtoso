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

func TestAraiHuThemeIsServed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/araihu.css", nil)
	rec := httptest.NewRecorder()

	appMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("theme status = %d, want 200", rec.Code)
	}
	for _, want := range []string{`[data-theme="araihu"]`, `--color-primary: #173b72`, `--color-primary-dark: #c7ff4a`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("theme missing canonical token %q", want)
		}
	}
}

func TestBreedsPageRendersRoundDogPhotos(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	appMux().ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, want := range []string{`data-theme="araihu"`, `href="/araihu.css"`, `localStorage.getItem('theme') || 'araihu'`} {
		if !strings.Contains(body, want) {
			t.Fatalf("page missing Arai Hû default theme contract %q:\n%s", want, body)
		}
	}
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
