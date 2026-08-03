package head

import (
	"context"
	"strings"
	"testing"
)

func completeMetadataConfig() MetadataConfig {
	return MetadataConfig{
		Title:        "Example page",
		Description:  "A complete route-specific description.",
		CanonicalURL: "https://example.com/docs",
		Image: SocialImage{
			URL:      "https://example.com/social/docs.jpg",
			MIMEType: "image/jpeg",
			Width:    1280,
			Height:   640,
			Alt:      "Example documentation preview",
		},
	}
}

func renderMetadata(cfg MetadataConfig) (string, error) {
	var output strings.Builder
	err := Metadata(cfg).Render(context.Background(), &output)
	return output.String(), err
}

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
			URL:      "https://example.com/og.png?a=1&b=2",
			MIMEType: "image/png",
			Width:    1280,
			Height:   640,
			Alt:      `Diagram "with labels"`,
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
	cfg := completeMetadataConfig()
	cfg.TwitterCard = TwitterCardSummary
	out := render(t, Metadata(cfg))

	for _, absent := range []string{
		`og:site_name`,
		`og:locale`,
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

func TestMetadataRejectsInvalidURLsWithoutWritingPartialTags(t *testing.T) {
	invalidURLs := map[string]string{
		"relative":        "/docs",
		"scheme-relative": "//example.com/docs",
		"http":            "http://example.com/docs",
		"javascript":      "javascript:alert(1)",
		"data":            "data:text/plain,preview",
		"missing-host":    "https:///docs",
		"malformed":       "https://example.com/%zz",
	}

	for name, invalidURL := range invalidURLs {
		for _, field := range []string{"canonical", "image"} {
			t.Run(name+"/"+field, func(t *testing.T) {
				cfg := completeMetadataConfig()
				if field == "canonical" {
					cfg.CanonicalURL = invalidURL
				} else {
					cfg.Image.URL = invalidURL
				}

				out, err := renderMetadata(cfg)
				if err == nil {
					t.Fatalf("Metadata() accepted invalid %s URL %q", field, invalidURL)
				}
				if out != "" {
					t.Fatalf("Metadata() wrote partial tags before rejecting %s URL %q:\n%s", field, invalidURL, out)
				}
			})
		}
	}
}

func TestMetadataRejectsIncompleteSocialContractWithoutOutput(t *testing.T) {
	tests := map[string]MetadataConfig{
		"empty":      {},
		"title-only": {Title: "Only a title"},
		"url-only":   {CanonicalURL: "https://example.com/docs"},
		"image-only": {Image: completeMetadataConfig().Image},
		"missing-title": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Title = ""
			return cfg
		}(),
		"missing-description": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Description = ""
			return cfg
		}(),
		"missing-canonical": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.CanonicalURL = ""
			return cfg
		}(),
		"missing-image": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Image = SocialImage{}
			return cfg
		}(),
		"missing-image-type": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Image.MIMEType = ""
			return cfg
		}(),
		"missing-image-width": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Image.Width = 0
			return cfg
		}(),
		"missing-image-height": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Image.Height = 0
			return cfg
		}(),
		"missing-image-alt": func() MetadataConfig {
			cfg := completeMetadataConfig()
			cfg.Image.Alt = ""
			return cfg
		}(),
	}

	for name, cfg := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := renderMetadata(cfg)
			if err == nil {
				t.Fatal("Metadata() accepted an incomplete social contract")
			}
			if out != "" {
				t.Fatalf("Metadata() wrote partial tags before rejecting config:\n%s", out)
			}
		})
	}
}
