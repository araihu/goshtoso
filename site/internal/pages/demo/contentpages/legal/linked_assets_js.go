//go:build js && wasm

package legalpages

import "github.com/araihu/goshtoso/assets"

// The Worker exports the locked runtime inventory to ASSETS. Inspect metadata
// instead of linking an HTTP file server and all library artwork into WASM.
func linkedAssetServes(assetURL string) bool {
	for _, resource := range assets.MuambaResources() {
		for _, download := range resource.Downloads {
			if "/"+download.Path == assetURL {
				return true
			}
		}
	}
	return false
}
