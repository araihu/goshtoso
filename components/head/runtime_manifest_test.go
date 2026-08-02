package head

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/assets"
)

func TestWithRuntimeManifestSnapshotsBaselineAndAppliesOrdinaryOptionsAfterward(t *testing.T) {
	manifest := assets.DefaultRuntimeManifest()
	manifest.Stylesheet.PrimaryURL = "/baseline/styles.css"
	manifest.Stylesheet.LocalURL = "/baseline/styles.css"
	manifest.Loader.PrimaryURL = "/baseline/loader.js"
	manifest.Loader.LocalURL = "/inventory/loader.js"

	baseline := WithRuntimeManifest(manifest)
	manifest.Stylesheet.PrimaryURL = "/mutated/styles.css"
	manifest.Loader.PrimaryURL = "/mutated/loader.js"

	component := Dependencies(
		WithDependencyCDNURL(DependencyHTMX, "https://override.example/htmx.js"),
		baseline,
		WithDependencyIntegrity(DependencyHTMX, "sha384-override"),
	)
	manifest.Dependencies[0].PrimaryURL = "/mutated/collapse.js"

	output := render(t, component)
	if !strings.Contains(output, `href="/baseline/styles.css"`) {
		t.Fatalf("manifest was not snapshotted when WithRuntimeManifest was called:\n%s", output)
	}
	if !strings.Contains(output, `src="/baseline/loader.js"`) {
		t.Fatalf("loader primary did not come from manifest snapshot:\n%s", output)
	}
	loader := parseLoaderConfig(t, output)
	if got := loaderEntry(t, loader, "alpine-collapse").PrimaryURL; got != assets.AlpineCollapseCDNURL {
		t.Fatalf("caller mutation changed manifest snapshot: collapse URL = %q", got)
	}
	htmx := loaderEntry(t, loader, "htmx")
	if htmx.PrimaryURL != "https://override.example/htmx.js" || htmx.Integrity != "sha384-override" {
		t.Fatalf("ordinary options did not apply after baseline selection: %#v", htmx)
	}
}

func TestDependenciesRejectsMultipleRuntimeManifestBaselinesWithoutOutput(t *testing.T) {
	var output strings.Builder
	err := Dependencies(
		WithRuntimeManifest(assets.DefaultRuntimeManifest()),
		WithRuntimeManifest(assets.DefaultRuntimeManifest()),
	).Render(context.Background(), &output)
	if err == nil || !strings.Contains(err.Error(), "WithRuntimeManifest may be used once") {
		t.Fatalf("render error = %v, want duplicate manifest error", err)
	}
	if output.Len() != 0 {
		t.Fatalf("invalid manifest wrote partial HTML: %q", output.String())
	}
}

func TestCustomRuntimeManifestHonorsTopLevelFields(t *testing.T) {
	manifest := publicTestManifest()
	manifest.Stylesheet = assets.RuntimeAsset{
		Role: assets.RuntimeRoleStylesheet, Kind: assets.RuntimeAssetStylesheet,
		PrimaryURL: "/custom/styles.css", LocalURL: "/inventory/styles.css",
		Integrity: "sha384-styles", Enabled: true, IncludeInMinimal: true,
	}
	manifest.Loader = assets.RuntimeAsset{
		Role: assets.RuntimeRoleDependencyLoader, Kind: assets.RuntimeAssetScript,
		PrimaryURL: "/custom/loader.js", LocalURL: "/inventory/loader.js",
		Integrity: "sha384-loader", Enabled: true, IncludeInMinimal: true,
	}
	for index := range manifest.Dependencies {
		manifest.Dependencies[index].Enabled = false
	}

	output := render(t, Dependencies(WithRuntimeManifest(manifest)))
	for _, want := range []string{
		`<link rel="stylesheet" href="/custom/styles.css"`,
		`integrity="sha384-styles"`,
		`<script src="/custom/loader.js"`,
		`integrity="sha384-loader"`,
		`crossorigin="anonymous"`,
	} {
		if !strings.Contains(output, want) {
			t.Errorf("custom top-level asset missing %q:\n%s", want, output)
		}
	}
	for _, unwanted := range []string{"/inventory/styles.css", "/inventory/loader.js", `<script defer src="/custom/loader.js"`} {
		if strings.Contains(output, unwanted) {
			t.Errorf("custom top-level field ignored; found %q:\n%s", unwanted, output)
		}
	}
}

func TestCustomRuntimeManifestHonorsTopLevelEnabledAndMinimalMembership(t *testing.T) {
	manifest := publicTestManifest()
	manifest.Stylesheet.Enabled = false
	manifest.Loader.Enabled = false
	for index := range manifest.Dependencies {
		manifest.Dependencies[index].Enabled = false
	}
	if output := render(t, Dependencies(WithRuntimeManifest(manifest))); output != "" {
		t.Fatalf("disabled top-level assets rendered HTML: %q", output)
	}

	manifest = publicTestManifest()
	manifest.Stylesheet.IncludeInMinimal = false
	manifest.Loader.IncludeInMinimal = false
	for index := range manifest.Dependencies {
		manifest.Dependencies[index].IncludeInMinimal = false
	}
	if output := render(t, DependenciesMinimal(WithRuntimeManifest(manifest))); output != "" {
		t.Fatalf("non-minimal top-level assets rendered in minimal set: %q", output)
	}
}

func TestCustomRuntimeManifestAddsFiltersAndPreservesDependencyOrder(t *testing.T) {
	manifest := publicTestManifest()
	manifest.Dependencies = []assets.RuntimeAsset{
		{Role: "custom-second", Kind: assets.RuntimeAssetScript, PrimaryURL: "/runtime/second.js", LocalURL: "/fallback/second.js", Enabled: true, WaitForWindowLoaded: true},
		{Role: "custom-disabled", Kind: assets.RuntimeAssetScript, PrimaryURL: "/runtime/disabled.js", LocalURL: "/fallback/disabled.js"},
		{Role: "custom-first", Kind: assets.RuntimeAssetScript, PrimaryURL: "/runtime/first.js", LocalURL: "/fallback/first.js", Enabled: true, IncludeInMinimal: true},
	}

	full := parseLoaderConfig(t, render(t, Dependencies(WithRuntimeManifest(manifest))))
	minimal := parseLoaderConfig(t, render(t, DependenciesMinimal(WithRuntimeManifest(manifest))))
	if got, want := loaderNames(full), []string{"custom-second", "custom-first"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("full dependency order = %v, want %v", got, want)
	}
	if got, want := loaderNames(minimal), []string{"custom-first"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("minimal dependency order = %v, want %v", got, want)
	}
	if !loaderEntry(t, full, "custom-second").WaitForWindowLoaded {
		t.Fatal("custom WaitForWindowLoaded was not passed to loader")
	}
}

func TestCustomRuntimeManifestControlsEveryDeclaredDependencyRole(t *testing.T) {
	for _, declared := range assets.DefaultRuntimeManifest().Dependencies {
		role := declared.Role
		t.Run(string(role), func(t *testing.T) {
			manifest := publicTestManifest()
			for index := range manifest.Dependencies {
				dependency := &manifest.Dependencies[index]
				dependency.Enabled = dependency.Role == role
				if dependency.Role == role {
					dependency.PrimaryURL = "/primary/" + string(role) + ".js"
					dependency.LocalURL = "/fallback/" + string(role) + ".js"
					dependency.Integrity = "sha384-" + string(role)
					dependency.IncludeInMinimal = true
					dependency.WaitForWindowLoaded = true
				}
			}

			full := parseLoaderConfig(t, render(t, Dependencies(WithRuntimeManifest(manifest))))
			minimal := parseLoaderConfig(t, render(t, DependenciesMinimal(WithRuntimeManifest(manifest))))
			for _, config := range []testLoaderConfig{full, minimal} {
				if got := loaderNames(config); !reflect.DeepEqual(got, []string{string(role)}) {
					t.Fatalf("selected roles = %v, want %s", got, role)
				}
				entry := loaderEntry(t, config, string(role))
				if entry.PrimaryURL != "/primary/"+string(role)+".js" || entry.FallbackURL != "/fallback/"+string(role)+".js" || entry.Integrity != "sha384-"+string(role) || !entry.WaitForWindowLoaded {
					t.Fatalf("custom role entry = %#v", entry)
				}
			}

			for index := range manifest.Dependencies {
				manifest.Dependencies[index].Enabled = false
			}
			if got := loaderNames(parseLoaderConfig(t, render(t, Dependencies(WithRuntimeManifest(manifest))))); len(got) != 0 {
				t.Fatalf("disabled role remained in loader: %v", got)
			}
		})
	}
}

func TestCustomRuntimeManifestValidationFailsBeforeWritingHTML(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*assets.RuntimeManifest)
		minimal bool
		want    string
	}{
		{name: "stylesheet role", mutate: func(m *assets.RuntimeManifest) { m.Stylesheet.Role = "css" }, want: "stylesheet role"},
		{name: "stylesheet kind", mutate: func(m *assets.RuntimeManifest) { m.Stylesheet.Kind = assets.RuntimeAssetScript }, want: "stylesheet kind"},
		{name: "stylesheet defer", mutate: func(m *assets.RuntimeManifest) { m.Stylesheet.Defer = true }, want: "stylesheet Defer is unsupported"},
		{name: "stylesheet wait for window loaded", mutate: func(m *assets.RuntimeManifest) { m.Stylesheet.WaitForWindowLoaded = true }, want: "stylesheet WaitForWindowLoaded is unsupported"},
		{name: "stylesheet URL", mutate: func(m *assets.RuntimeManifest) { m.Stylesheet.PrimaryURL = "javascript:alert(1)" }, want: "stylesheet primary URL"},
		{name: "loader role", mutate: func(m *assets.RuntimeManifest) { m.Loader.Role = "bootstrap" }, want: "loader role"},
		{name: "loader kind", mutate: func(m *assets.RuntimeManifest) { m.Loader.Kind = assets.RuntimeAssetStylesheet }, want: "loader kind"},
		{name: "loader wait for window loaded", mutate: func(m *assets.RuntimeManifest) { m.Loader.WaitForWindowLoaded = true }, want: "loader WaitForWindowLoaded is unsupported"},
		{name: "loader URL", mutate: func(m *assets.RuntimeManifest) { m.Loader.PrimaryURL = "//cdn.example/loader.js" }, want: "loader primary URL"},
		{name: "loader disabled", mutate: func(m *assets.RuntimeManifest) { m.Loader.Enabled = false }, want: "enabled dependencies require an enabled loader"},
		{name: "loader excluded from minimal", minimal: true, mutate: func(m *assets.RuntimeManifest) { m.Loader.IncludeInMinimal = false }, want: "enabled dependencies require an enabled loader"},
		{name: "duplicate role", mutate: func(m *assets.RuntimeManifest) { m.Dependencies[1].Role = m.Dependencies[0].Role }, want: "duplicate dependency role"},
		{name: "empty role", mutate: func(m *assets.RuntimeManifest) { m.Dependencies[0].Role = "" }, want: "dependency role is empty"},
		{name: "unsafe role", mutate: func(m *assets.RuntimeManifest) { m.Dependencies[0].Role = `bad role"` }, want: "dependency role"},
		{name: "dependency kind", mutate: func(m *assets.RuntimeManifest) { m.Dependencies[0].Kind = assets.RuntimeAssetStylesheet }, want: "dependency alpine-collapse kind"},
		{name: "dependency primary URL", mutate: func(m *assets.RuntimeManifest) { m.Dependencies[0].PrimaryURL = "data:text/javascript,alert(1)" }, want: "dependency alpine-collapse primary URL"},
		{name: "dependency fallback URL", mutate: func(m *assets.RuntimeManifest) { m.Dependencies[0].LocalURL = "https:/missing-host.js" }, want: "dependency alpine-collapse local URL"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := publicTestManifest()
			test.mutate(&manifest)
			var output strings.Builder
			var err error
			if test.minimal {
				err = DependenciesMinimal(WithRuntimeManifest(manifest)).Render(context.Background(), &output)
			} else {
				err = Dependencies(WithRuntimeManifest(manifest)).Render(context.Background(), &output)
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("render error = %v, want substring %q", err, test.want)
			}
			if output.Len() != 0 {
				t.Fatalf("invalid manifest wrote partial HTML: %q", output.String())
			}
		})
	}
}

func TestCustomRuntimeManifestRejectsUnsafeURLForms(t *testing.T) {
	for _, rawURL := range []string{"", " javascript:alert(1)", "ftp://example.com/runtime.js", "//example.com/runtime.js", "https:/example.com/runtime.js", "https://user@example.com/runtime.js", "runtime\\script.js"} {
		t.Run(rawURL, func(t *testing.T) {
			manifest := publicTestManifest()
			manifest.Dependencies[0].PrimaryURL = rawURL
			var output strings.Builder
			err := Dependencies(WithRuntimeManifest(manifest)).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), "primary URL") {
				t.Fatalf("render error = %v, want unsafe primary URL rejection", err)
			}
			if output.Len() != 0 {
				t.Fatalf("unsafe URL wrote partial HTML: %q", output.String())
			}
		})
	}
}

func TestCustomRuntimeManifestRejectsMissingOptionTargets(t *testing.T) {
	tests := []struct {
		name   string
		remove assets.RuntimeAssetRole
		option Option
	}{
		{name: "CDN URL", remove: assets.RuntimeRoleHTMX, option: WithDependencyCDNURL(DependencyHTMX, "/custom/htmx.js")},
		{name: "local URL", remove: assets.RuntimeRoleHTMX, option: WithDependencyLocalURL(DependencyHTMX, "/custom/htmx.js")},
		{name: "integrity", remove: assets.RuntimeRoleHTMX, option: WithDependencyIntegrity(DependencyHTMX, "sha384-custom")},
		{name: "omit", remove: assets.RuntimeRoleHTMX, option: WithoutDependency(DependencyHTMX)},
		{name: "combobox compatibility", remove: assets.RuntimeRoleCombobox, option: WithComboboxURL("/custom/combobox.js")},
		{name: "action group compatibility", remove: assets.RuntimeRoleActionGroup, option: WithActionGroupURL("/custom/action-group.js")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := publicTestManifest()
			manifest.Dependencies = withoutRole(manifest.Dependencies, test.remove)
			var output strings.Builder
			err := Dependencies(WithRuntimeManifest(manifest), test.option).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), "missing dependency role") {
				t.Fatalf("render error = %v, want missing target rejection", err)
			}
			if output.Len() != 0 {
				t.Fatalf("missing option target wrote partial HTML: %q", output.String())
			}
		})
	}

	if err := Dependencies(WithDependencyCDNURL(Dependency("future"), "/future.js")).Render(context.Background(), &strings.Builder{}); err != nil {
		t.Fatalf("default-manifest unknown dependency lost historical no-op behavior: %v", err)
	}
}

func TestCustomRuntimeManifestRejectsUnsafeKnownOrderAndFirstPartyConflicts(t *testing.T) {
	tests := []struct {
		name   string
		before assets.RuntimeAssetRole
		after  assets.RuntimeAssetRole
		want   string
	}{
		{name: "Alpine collapse after core", before: assets.RuntimeRoleAlpineJS, after: assets.RuntimeRoleAlpineCollapse, want: "alpine-collapse must precede alpine"},
		{name: "Alpine focus after core", before: assets.RuntimeRoleAlpineJS, after: assets.RuntimeRoleAlpineFocus, want: "alpine-focus must precede alpine"},
		{name: "Alpine mask after core", before: assets.RuntimeRoleAlpineJS, after: assets.RuntimeRoleAlpineMask, want: "alpine-mask must precede alpine"},
		{name: "first party after Alpine", before: assets.RuntimeRoleAlpineJS, after: assets.RuntimeRoleFirstParty, want: "first-party must precede alpine"},
		{name: "dark mode after Alpine", before: assets.RuntimeRoleAlpineJS, after: assets.RuntimeRoleDarkMode, want: "dark-mode must precede alpine"},
		{name: "SSE before HTMX", before: assets.RuntimeRoleHTMXExtSSE, after: assets.RuntimeRoleHTMX, want: "htmx must precede htmx-ext-sse"},
		{name: "WS before HTMX", before: assets.RuntimeRoleHTMXExtWS, after: assets.RuntimeRoleHTMX, want: "htmx must precede htmx-ext-ws"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := publicTestManifest()
			disableKnownOrderRoles(&manifest)
			enableRoles(&manifest, test.before, test.after)
			putRoleImmediatelyBefore(&manifest, test.before, test.after)
			assertRenderFailure(t, Dependencies(WithRuntimeManifest(manifest)), test.want)
		})
	}

	for _, role := range []assets.RuntimeAssetRole{assets.RuntimeRoleCombobox, assets.RuntimeRoleActionGroup} {
		t.Run("bundle with "+string(role), func(t *testing.T) {
			manifest := publicTestManifest()
			enableRoles(&manifest, assets.RuntimeRoleFirstParty, role)
			assertRenderFailure(t, Dependencies(WithRuntimeManifest(manifest)), "first-party bundle cannot be combined with standalone")
		})
	}
}

func disableKnownOrderRoles(manifest *assets.RuntimeManifest) {
	known := map[assets.RuntimeAssetRole]struct{}{
		assets.RuntimeRoleAlpineCollapse: {},
		assets.RuntimeRoleAlpineFocus:    {},
		assets.RuntimeRoleAlpineMask:     {},
		assets.RuntimeRoleFirstParty:     {},
		assets.RuntimeRoleDarkMode:       {},
		assets.RuntimeRoleAlpineJS:       {},
		assets.RuntimeRoleHTMX:           {},
		assets.RuntimeRoleHTMXExtSSE:     {},
		assets.RuntimeRoleHTMXExtWS:      {},
	}
	for index := range manifest.Dependencies {
		if _, ok := known[manifest.Dependencies[index].Role]; ok {
			manifest.Dependencies[index].Enabled = false
		}
	}
}

func TestCustomRuntimeManifestRejectsWithLocalRuntimeButSupportsExplicitLocalOnlyRecipe(t *testing.T) {
	manifest := publicTestManifest()
	assertRenderFailure(t, Dependencies(WithRuntimeManifest(manifest), WithLocalRuntime()), "custom RuntimeManifest cannot be combined with WithLocalRuntime")

	manifest.Stylesheet.PrimaryURL = manifest.Stylesheet.LocalURL
	manifest.Loader.PrimaryURL = manifest.Loader.LocalURL
	for index := range manifest.Dependencies {
		manifest.Dependencies[index].PrimaryURL = manifest.Dependencies[index].LocalURL
	}
	output := render(t, Dependencies(WithRuntimeManifest(manifest), WithoutLocalFallback()))
	loader := parseLoaderConfig(t, output)
	for _, dependency := range loader.Dependencies {
		if dependency.FallbackURL != "" || !strings.HasPrefix(dependency.PrimaryURL, "/") {
			t.Errorf("explicit local-only recipe entry = %#v", dependency)
		}
	}
}

func publicTestManifest() assets.RuntimeManifest {
	manifest := assets.DefaultRuntimeManifest()
	manifest.Stylesheet.PrimaryURL = "/runtime/styles.css"
	manifest.Stylesheet.LocalURL = "/inventory/styles.css"
	manifest.Loader.PrimaryURL = "/runtime/loader.js"
	manifest.Loader.LocalURL = "/inventory/loader.js"
	for index := range manifest.Dependencies {
		dependency := &manifest.Dependencies[index]
		dependency.PrimaryURL = "/runtime/" + string(dependency.Role) + ".js"
		dependency.LocalURL = "/fallback/" + string(dependency.Role) + ".js"
		dependency.Integrity = ""
	}
	return manifest
}

func loaderNames(config testLoaderConfig) []string {
	names := make([]string, 0, len(config.Dependencies))
	for _, dependency := range config.Dependencies {
		names = append(names, dependency.Name)
	}
	return names
}

func withoutRole(dependencies []assets.RuntimeAsset, role assets.RuntimeAssetRole) []assets.RuntimeAsset {
	filtered := make([]assets.RuntimeAsset, 0, len(dependencies)-1)
	for _, dependency := range dependencies {
		if dependency.Role != role {
			filtered = append(filtered, dependency)
		}
	}
	return filtered
}

func enableRoles(manifest *assets.RuntimeManifest, roles ...assets.RuntimeAssetRole) {
	for index := range manifest.Dependencies {
		for _, role := range roles {
			if manifest.Dependencies[index].Role == role {
				manifest.Dependencies[index].Enabled = true
			}
		}
	}
}

func putRoleImmediatelyBefore(manifest *assets.RuntimeManifest, before, after assets.RuntimeAssetRole) {
	var beforeAsset, afterAsset assets.RuntimeAsset
	remaining := make([]assets.RuntimeAsset, 0, len(manifest.Dependencies)-2)
	for _, dependency := range manifest.Dependencies {
		switch dependency.Role {
		case before:
			beforeAsset = dependency
		case after:
			afterAsset = dependency
		default:
			remaining = append(remaining, dependency)
		}
	}
	manifest.Dependencies = append([]assets.RuntimeAsset{beforeAsset, afterAsset}, remaining...)
}

func assertRenderFailure(t *testing.T, component interface {
	Render(context.Context, io.Writer) error
}, want string) {
	t.Helper()
	var output strings.Builder
	err := component.Render(context.Background(), &output)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("render error = %v, want substring %q", err, want)
	}
	if output.Len() != 0 {
		t.Fatalf("invalid manifest wrote partial HTML: %q", output.String())
	}
}
