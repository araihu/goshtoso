package demo

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso/assets"
	siteassets "github.com/araihu/goshtoso/site/assets"
)

func TestSiteRuntimeScriptsFollowLinkedManifestOrderWithExplicitOverrides(t *testing.T) {
	got := siteRuntimeScripts()
	roles := make([]assets.RuntimeAssetRole, 0, len(got))
	for _, script := range got {
		roles = append(roles, script.Role)
	}
	want := []assets.RuntimeAssetRole{"alpine-collapse", "alpine-focus", "alpine-mask", "first-party", "dark-mode", "site-demo", "alpine", "htmx", "htmx-ext-sse", "htmx-ext-ws"}
	if !reflect.DeepEqual(roles, want) {
		t.Fatalf("runtime roles = %v, want %v", roles, want)
	}
	if got[5].URL != siteassets.DemoBundleURL || got[7].Defer {
		t.Fatalf("site override semantics = %#v", got)
	}
	if siteRuntimeStylesheetURL() != assets.DefaultRuntimeManifest().Stylesheet.LocalURL {
		t.Fatal("site stylesheet does not come from linked manifest")
	}
}

func TestComponentDocsAdditionalRuntimeScriptsUseLinkedExtensionURLs(t *testing.T) {
	got := componentDocsAdditionalRuntimeScripts()
	want := []string{siteassets.DemoBundleURL, assets.HTMXExtSSEURL, assets.HTMXExtWSURL}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("component docs scripts = %v, want %v", got, want)
	}
}
