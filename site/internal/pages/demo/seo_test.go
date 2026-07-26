package demo

import "testing"

func TestHomeMetaUsesLandscapeSocialCard(t *testing.T) {
	got := HomeMeta().OGImageURL()
	want := SiteBaseURL + HomeOGImagePath
	if got != want {
		t.Fatalf("HomeMeta().OGImageURL() = %q, want %q", got, want)
	}
}

func TestDefaultMetaKeepsDefaultSocialImage(t *testing.T) {
	got := DefaultMeta("Button").OGImageURL()
	want := SiteBaseURL + OGImagePath
	if got != want {
		t.Fatalf("DefaultMeta().OGImageURL() = %q, want %q", got, want)
	}
}
