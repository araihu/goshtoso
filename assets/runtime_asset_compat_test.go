package assets_test

import "github.com/araihu/goshtoso/assets"

// Keep the original positional layout source-compatible for consumers that
// used an unkeyed RuntimeAsset literal before runtime metadata was added.
var _ = assets.RuntimeAsset{
	assets.RuntimeRoleStylesheet,
	assets.RuntimeAssetStylesheet,
	assets.StylesURL,
	assets.StylesURL,
	"",
	true,
	true,
	false,
	false,
}
