package head

import (
	"strings"
	"testing"
)

func TestMetadataRendersCompleteSocialContract(t *testing.T) {
	out := render(t, Metadata(MetadataConfig{
		Title:         "Muamba · Trust remote files once",
		Description:   "Language-agnostic TOFU vendoring and integrity verification.",
		CanonicalURL:  "https://muamba.araihu.com/docs",
		OpenGraphType: "article",
		SiteName:      "Muamba",
		Locale:        "en_US",
		Image: SocialImage{
			URL:      "https://muamba.araihu.com/og.jpg",
			MIMEType: "image/jpeg",
			Width:    1280,
			Height:   640,
			Alt:      "Muamba — Trust remote files once. Verify them forever.",
		},
		TwitterCard: TwitterCardSummaryLargeImage,
		TwitterSite: "@araihu",
	}))

	wants := []string{
		`<title>Muamba · Trust remote files once</title>`,
		`<meta name="description" content="Language-agnostic TOFU vendoring and integrity verification.">`,
		`<link rel="canonical" href="https://muamba.araihu.com/docs">`,
		`<meta property="og:url" content="https://muamba.araihu.com/docs">`,
		`<meta property="og:type" content="article">`,
		`<meta property="og:title" content="Muamba · Trust remote files once">`,
		`<meta property="og:description" content="Language-agnostic TOFU vendoring and integrity verification.">`,
		`<meta property="og:site_name" content="Muamba">`,
		`<meta property="og:locale" content="en_US">`,
		`<meta property="og:image" content="https://muamba.araihu.com/og.jpg">`,
		`<meta property="og:image:type" content="image/jpeg">`,
		`<meta property="og:image:width" content="1280">`,
		`<meta property="og:image:height" content="640">`,
		`<meta property="og:image:alt" content="Muamba — Trust remote files once. Verify them forever.">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="twitter:title" content="Muamba · Trust remote files once">`,
		`<meta name="twitter:description" content="Language-agnostic TOFU vendoring and integrity verification.">`,
		`<meta name="twitter:image" content="https://muamba.araihu.com/og.jpg">`,
		`<meta name="twitter:image:alt" content="Muamba — Trust remote files once. Verify them forever.">`,
		`<meta name="twitter:site" content="@araihu">`,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Metadata() missing %q\n%s", want, out)
		}
	}
}

func TestMetadataDefaultsAndEscapesValues(t *testing.T) {
	out := render(t, Metadata(MetadataConfig{
		Title:        `<unsafe> & title`,
		Description:  `quoted "description"`,
		CanonicalURL: "https://example.com/docs?a=1&b=2",
		Image: SocialImage{
			URL: "https://example.com/og.png?a=1&b=2",
			Alt: `Diagram "with labels"`,
		},
	}))

	wants := []string{
		`<title>&lt;unsafe&gt; &amp; title</title>`,
		`content="quoted &#34;description&#34;"`,
		`href="https://example.com/docs?a=1&amp;b=2"`,
		`<meta property="og:type" content="website">`,
		`content="https://example.com/og.png?a=1&amp;b=2"`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`content="Diagram &#34;with labels&#34;"`,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("Metadata() missing escaped/default value %q\n%s", want, out)
		}
	}
}

func TestMetadataOmitsUnavailableOptionalValues(t *testing.T) {
	out := render(t, Metadata(MetadataConfig{
		Title:        "Status",
		Description:  "Current service status.",
		CanonicalURL: "https://status.example.com/",
		TwitterCard:  TwitterCardSummary,
	}))

	for _, absent := range []string{
		`og:image`,
		`og:site_name`,
		`og:locale`,
		`twitter:image`,
		`twitter:site`,
	} {
		if strings.Contains(out, absent) {
			t.Errorf("Metadata() unexpectedly rendered %q\n%s", absent, out)
		}
	}
	if !strings.Contains(out, `<meta name="twitter:card" content="summary">`) {
		t.Fatalf("Metadata() missing explicit summary card\n%s", out)
	}
}
