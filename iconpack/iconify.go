package iconpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/araihu/goshtoso/internal/iconcatalog"
)

type iconifyPack struct {
	Prefix string                 `json:"prefix"`
	Icons  map[string]iconifyIcon `json:"icons"`
}

type iconifyIcon struct {
	Body   string  `json:"body"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

func resolveIconifyPrefix(opts Options) (string, error) {
	prefix := opts.IconifyPrefix
	if prefix == "" {
		prefix = normalizeWebPath(opts.Package)
	}
	if !validResourceName(prefix) {
		return "", fmt.Errorf("iconify-prefix %q must be lower-kebab", prefix)
	}
	return prefix, nil
}

func buildIconifyPack(boundary releaseBoundary, prefix string, selected []selectedAsset) ([]byte, error) {
	pack := iconifyPack{Prefix: prefix, Icons: make(map[string]iconifyIcon, len(selected))}
	for _, asset := range selected {
		name := iconifyName(prefix, asset.SpriteSymbol)
		if !validGenericSymbol(name) {
			return nil, fmt.Errorf("icon %q has invalid Iconify name %q", asset.CanonicalName, name)
		}
		if _, exists := pack.Icons[name]; exists {
			return nil, fmt.Errorf("duplicate Iconify name %q", name)
		}
		raw, err := readVerifiedReleaseFile(boundary, asset.Path)
		if err != nil {
			return nil, fmt.Errorf("read source icon %q for Iconify JSON: %w", asset.CanonicalName, err)
		}
		body, err := standaloneSVGBody(raw, asset.Path, asset.Dimensions.ViewBox)
		if err != nil {
			return nil, err
		}
		width, height, err := iconifyDimensions(asset.Dimensions)
		if err != nil {
			return nil, fmt.Errorf("Iconify dimensions for %q: %w", asset.CanonicalName, err)
		}
		pack.Icons[name] = iconifyIcon{Body: string(body), Width: width, Height: height}
	}
	return marshalIconifyDocument(pack)
}

func marshalIconifyDocument(value any) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func iconifyName(prefix, symbol string) string {
	return strings.TrimPrefix(symbol, prefix+"-")
}

func iconifyDimensions(dimensions iconcatalog.Dimensions) (float64, float64, error) {
	parts := strings.Fields(dimensions.ViewBox)
	if len(parts) != 4 {
		return 0, 0, fmt.Errorf("viewBox %q must contain four numbers", dimensions.ViewBox)
	}
	width, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("width: %w", err)
	}
	height, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("height: %w", err)
	}
	return width, height, nil
}
