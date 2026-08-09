package iconpack

import (
	"fmt"
	"sort"

	"github.com/araihu/goshtoso/internal/iconcatalog"
)

type sourceFamily struct {
	Namespace      string
	Product        string
	SpritePath     string
	ProvenancePath string
	LicensePath    string
	LicenseOutput  string
	Generic        bool
}

var allowedFamilies = map[string]sourceFamily{
	"ui/heroicons": {
		Namespace:      "ui",
		Product:        "heroicons",
		SpritePath:     "icons/ui/sprite.svg",
		ProvenancePath: "icons/ui/heroicons/provenance.json",
		LicensePath:    "licenses/heroicons-MIT.txt",
		LicenseOutput:  "LICENSES/heroicons-MIT.txt",
	},
	"brand/developer-icons": {
		Namespace:      "brand",
		Product:        "developer-icons",
		SpritePath:     "icons/brand/developer-icons/sprite.svg",
		ProvenancePath: "icons/brand/developer-icons/provenance.json",
		LicensePath:    "licenses/developer-icons-MIT.txt",
		LicenseOutput:  "LICENSES/developer-icons-MIT.txt",
	},
}

type selectedAsset struct {
	iconcatalog.Asset
	family sourceFamily
	goName string
}

func selectAssets(boundary releaseBoundary, names []string, prefix string) ([]selectedAsset, []sourceFamily, error) {
	byName := make(map[string]iconcatalog.Asset, len(boundary.catalog.Assets))
	for _, asset := range boundary.catalog.Assets {
		byName[asset.CanonicalName] = asset
	}
	selected := make([]selectedAsset, 0, len(names))
	usedFamilies := map[string]sourceFamily{}
	for _, name := range names {
		asset, ok := byName[name]
		if !ok {
			return nil, nil, fmt.Errorf("canonical name %q is not present in catalog", name)
		}
		familyKey := asset.Namespace + "/" + asset.Product
		family, allowed := allowedFamilies[familyKey]
		if boundary.generic != nil {
			familyKey = "generic/" + boundary.generic.manifest.Name
			family = boundary.generic.family
			allowed = true
		}
		if !allowed {
			return nil, nil, fmt.Errorf("canonical name %q resolves to unsupported namespace/product %s", name, familyKey)
		}
		if asset.Format != "svg" || asset.SpriteSymbol == "" {
			return nil, nil, fmt.Errorf("canonical name %q is not an SVG sprite asset", name)
		}
		artifact, err := readVerifiedReleaseFile(boundary, asset.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("verify selected artifact %q: %w", name, err)
		}
		got := hashBytes(artifact)
		if got != asset.SHA256 || boundary.checksums[asset.Path] != asset.SHA256 {
			return nil, nil, fmt.Errorf("selected artifact %q SHA-256 does not match catalog and release checksums", name)
		}
		goName, err := goIdentifier(prefix, asset.CanonicalName)
		if err != nil {
			return nil, nil, err
		}
		selected = append(selected, selectedAsset{Asset: asset, family: family, goName: goName})
		usedFamilies[familyKey] = family
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].CanonicalName < selected[j].CanonicalName })
	if err := validateGeneratedNamespace(selected); err != nil {
		return nil, nil, err
	}
	families := make([]sourceFamily, 0, len(usedFamilies))
	for _, family := range usedFamilies {
		families = append(families, family)
	}
	sort.Slice(families, func(i, j int) bool {
		return families[i].Namespace+"/"+families[i].Product < families[j].Namespace+"/"+families[j].Product
	})
	return selected, families, nil
}
