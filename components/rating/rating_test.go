package rating

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRatingDefaultRendersFiveStarRadios(t *testing.T) {
	var buf bytes.Buffer
	err := Rating(Config{ID: "product-rating", Name: "rating", Value: 3}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render rating: %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="product-rating"`,
		`role="radiogroup"`,
		`aria-label="Rating"`,
		`x-data="{ currentVal: 3 }"`,
		`name="rating"`,
		`id="product-rating-5"`,
		`x-model.number="currentVal"`,
		`text-warning`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered rating missing %q:\n%s", want, html)
		}
	}
}

func TestRatingDisplayClampsValue(t *testing.T) {
	var buf bytes.Buffer
	err := RatingDisplay(DisplayConfig{ID: "rating-display", Value: 99}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render rating display: %v", err)
	}

	html := buf.String()
	if strings.Contains(html, `type="radio"`) {
		t.Fatalf("rating display must not render radio inputs:\n%s", html)
	}
	if !strings.Contains(html, `aria-label="5 stars"`) {
		t.Fatalf("rating display should clamp aria label to max value:\n%s", html)
	}
}

func TestEmojiRatingRendersSentimentLabels(t *testing.T) {
	var buf bytes.Buffer
	err := Rating(Config{ID: "sentiment", Appearance: AppearanceEmoji, Value: 4}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("render emoji rating: %v", err)
	}

	html := buf.String()
	for _, want := range []string{"very dissatisfied", "neutral", "very satisfied", "😍"} {
		if !strings.Contains(html, want) {
			t.Fatalf("emoji rating missing %q:\n%s", want, html)
		}
	}
}

func TestEmojiRatingUsesExactSelectionExpression(t *testing.T) {
	cfg := Config{Appearance: AppearanceEmoji, Value: 4}

	if got := cfg.BindClass(3); !strings.Contains(got, "currentVal === 3") {
		t.Fatalf("emoji ratings should bind exact selection, got %q", got)
	}
	if cfg.IsActive(3) {
		t.Fatal("emoji rating should not render lower values active")
	}
	if !cfg.IsActive(4) {
		t.Fatal("emoji rating should render only the selected value active")
	}
}

func TestStarRatingUsesCumulativeSelectionExpression(t *testing.T) {
	cfg := Config{Appearance: AppearanceStars, Value: 4}

	if got := cfg.BindClass(3); !strings.Contains(got, "currentVal >= 3") {
		t.Fatalf("star ratings should bind cumulative selection, got %q", got)
	}
	if !cfg.IsActive(3) {
		t.Fatal("star rating should render lower values active")
	}
}
