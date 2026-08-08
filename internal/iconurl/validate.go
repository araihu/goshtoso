package iconurl

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateSpriteURL accepts a relative sprite path or an absolute HTTP(S)
// URL, but never a fragment-only, protocol-relative, credential-bearing, or
// non-HTTP URL.
func ValidateSpriteURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("SpriteURL is required in external mode")
	}
	if raw != strings.TrimSpace(raw) || strings.ContainsAny(raw, "\\\\\r\n") {
		return fmt.Errorf("invalid SpriteURL %q", raw)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid SpriteURL %q: %w", raw, err)
	}
	if parsed.Fragment != "" || parsed.Path == "" {
		return fmt.Errorf("SpriteURL must identify a sprite document")
	}
	if parsed.Scheme == "" {
		if parsed.Host != "" {
			return fmt.Errorf("protocol-relative SpriteURL is not allowed")
		}
		return nil
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("SpriteURL must use relative, http, or https URL syntax")
	}
	return nil
}
