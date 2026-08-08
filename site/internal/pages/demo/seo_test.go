package demo

import (
	"bytes"
	"context"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestSharedSocialImageAltDescribesTrackedCard(t *testing.T) {
	const want = "Goshtoso wordmark beside a sunglasses-wearing Go gopher in Rio, with Button, Input, Alert, and Card UI previews."
	for _, meta := range []PageMeta{HomeMeta(), DefaultMeta("Icon"), DefaultMeta("Accordion")} {
		if got := meta.OGImageAlt(); got != want {
			t.Fatalf("OGImageAlt() = %q, want shared-card description %q", got, want)
		}
	}
}

func TestSharedSocialImageMatchesDeclaredPNGContract(t *testing.T) {
	filename := filepath.Join("..", "..", "..", "..", "assets", "images", "goshtoso-social-card.png")
	contents, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, "image/png", http.DetectContentType(contents))

	image, err := png.DecodeConfig(bytes.NewReader(contents))
	require.NoError(t, err)
	require.Equal(t, 1200, image.Width)
	require.Equal(t, 630, image.Height)
}

func TestComponentHeadMetadataPairsSharedImageWithSharedAlt(t *testing.T) {
	const imageURL = "https://goshtoso.araihu.com/assets/images/goshtoso-social-card.png"
	const imageAlt = "Goshtoso wordmark beside a sunglasses-wearing Go gopher in Rio, with Button, Input, Alert, and Card UI previews."
	for _, meta := range []PageMeta{
		{Title: "Icon Component - Goshtoso UI Library for Go", Description: "Icon docs", Path: "/components/icon", Type: "TechArticle"},
		{Title: "Accordion Component - Goshtoso UI Library for Go", Description: "Accordion docs", Path: "/components/accordion", Type: "TechArticle"},
	} {
		var output bytes.Buffer
		require.NoError(t, HeadMeta(meta).Render(context.Background(), &output))
		html := output.String()

		require.Equal(t, 1, strings.Count(html, `<link rel="canonical" href="`+meta.CanonicalURL()+`">`))
		require.Equal(t, 1, strings.Count(html, `<meta property="og:image" content="`+imageURL+`">`))
		require.Equal(t, 1, strings.Count(html, `<meta property="og:image:alt" content="`+imageAlt+`">`))
		require.Equal(t, 1, strings.Count(html, `<meta name="twitter:image" content="`+imageURL+`">`))
		require.Equal(t, 1, strings.Count(html, `<meta name="twitter:image:alt" content="`+imageAlt+`">`))
	}
}
