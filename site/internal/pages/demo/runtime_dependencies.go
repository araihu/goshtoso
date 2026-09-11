package demo

import (
	"github.com/araihu/goshtoso/assets"
	siteassets "github.com/araihu/goshtoso/site/assets"
)

type siteRuntimeScript struct {
	Role  assets.RuntimeAssetRole
	URL   string
	Defer bool
}

var siteEnabledRuntimeRoles = map[assets.RuntimeAssetRole]bool{
	assets.RuntimeAssetRole("alpine-collapse"):    true,
	assets.RuntimeAssetRole("alpine-focus"):       true,
	assets.RuntimeAssetRole("alpine-mask"):        true,
	assets.RuntimeAssetRole("first-party"):        true,
	assets.RuntimeAssetRole("dark-mode"):          true,
	assets.RuntimeAssetRole("alpine"):             true,
	assets.RuntimeAssetRole("htmx-alpine-compat"): true,
	assets.RuntimeAssetRole("htmx"):               true,
	assets.RuntimeAssetRole("htmx-ext-sse"):       true,
	assets.RuntimeAssetRole("htmx-ext-ws"):        true,
}

// siteRuntimeScripts derives execution order and local URLs from the Goshtoso
// module linked into the site. The map above is the site's only enablement
// override. The site bundle is inserted immediately before Alpine so its
// alpine:init providers register before Alpine starts.
func siteRuntimeScripts() []siteRuntimeScript {
	manifest := assets.DefaultRuntimeManifest()
	scripts := make([]siteRuntimeScript, 0, len(manifest.Dependencies)+1)
	for _, dependency := range manifest.Dependencies {
		if !siteEnabledRuntimeRoles[dependency.Role] {
			continue
		}
		if dependency.Role == assets.RuntimeAssetRole("alpine") {
			scripts = append(scripts, siteRuntimeScript{Role: "site-demo", URL: siteassets.DemoBundleURL, Defer: true})
		}
		scripts = append(scripts, siteRuntimeScript{Role: dependency.Role, URL: dependency.LocalURL, Defer: dependency.Defer})
	}

	return scripts
}

func siteRuntimeStylesheetURL() string {
	return assets.DefaultRuntimeManifest().Stylesheet.LocalURL
}

func componentDocsAdditionalRuntimeScripts() []string {
	urls := []string{siteassets.DemoBundleURL}
	for _, script := range siteRuntimeScripts() {
		if script.Role == assets.RuntimeAssetRole("htmx-ext-sse") || script.Role == assets.RuntimeAssetRole("htmx-ext-ws") {
			urls = append(urls, script.URL)
		}
	}
	return urls
}
