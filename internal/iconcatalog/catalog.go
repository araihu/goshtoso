// Package iconcatalog loads schema-v1 asset catalogs and generates typed sprite
// bindings for a selected namespace and product.
package iconcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

const schemaVersion = 1

// Catalog is the schema-v1 assets catalog. Hash is the SHA-256 digest of the
// exact source bytes and is intentionally excluded from JSON.
type Catalog struct {
	SchemaVersion    int     `json:"schemaVersion"`
	Release          string  `json:"release"`
	IdentityRevision int     `json:"identityRevision"`
	Assets           []Asset `json:"assets"`
	Hash             string  `json:"-"`
}

// Asset is one entry from the assets schema-v1 catalog.
type Asset struct {
	CanonicalName string     `json:"canonicalName"`
	Namespace     string     `json:"namespace"`
	Path          string     `json:"path"`
	Product       string     `json:"product"`
	Artwork       string     `json:"artwork"`
	Appearance    string     `json:"appearance"`
	Surface       string     `json:"surface"`
	Framing       string     `json:"framing"`
	Format        string     `json:"format"`
	Dimensions    Dimensions `json:"dimensions"`
	SpriteSymbol  string     `json:"spriteSymbol"`
	ColorBehavior string     `json:"colorBehavior"`
	License       string     `json:"license"`
	Source        string     `json:"source"`
	SHA256        string     `json:"sha256"`
}

// Dimensions is the schema-v1 dimensions object.
type Dimensions struct {
	ViewBox string `json:"viewBox"`
}

// Load decodes a schema-v1 assets catalog and records its exact source hash.
// Empty sprite symbols are allowed because catalogs also contain non-sprite
// assets; Generate validates the selected sprite binding set.
func Load(r io.Reader) (Catalog, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Catalog{}, fmt.Errorf("read catalog: %w", err)
	}
	var catalog Catalog
	if err := json.Unmarshal(b, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("decode catalog: %w", err)
	}
	if catalog.SchemaVersion != schemaVersion {
		return Catalog{}, fmt.Errorf("unsupported schemaVersion %d: want %d", catalog.SchemaVersion, schemaVersion)
	}

	names := make(map[string]struct{}, len(catalog.Assets))
	symbols := make(map[string]struct{}, len(catalog.Assets))
	for _, asset := range catalog.Assets {
		if asset.CanonicalName == "" {
			return Catalog{}, fmt.Errorf("empty canonicalName")
		}
		if _, exists := names[asset.CanonicalName]; exists {
			return Catalog{}, fmt.Errorf("duplicate canonicalName %q", asset.CanonicalName)
		}
		names[asset.CanonicalName] = struct{}{}
		if asset.SpriteSymbol == "" {
			continue
		}
		if _, exists := symbols[asset.SpriteSymbol]; exists {
			return Catalog{}, fmt.Errorf("duplicate spriteSymbol %q", asset.SpriteSymbol)
		}
		symbols[asset.SpriteSymbol] = struct{}{}
	}

	sum := sha256.Sum256(b)
	catalog.Hash = hex.EncodeToString(sum[:])
	return catalog, nil
}
