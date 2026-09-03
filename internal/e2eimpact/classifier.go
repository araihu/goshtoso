package e2eimpact

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type classification struct {
	roots   []string
	direct  map[string]string
	full    []string
	skipped []string
}

var componentJavaScript = map[string]string{
	"action-group.js": "actiongroup", "carousel.js": "carousel", "code-block.js": "codeblock", "combobox-client.js": "combobox",
	"combobox.js": "combobox", "dropdown.js": "dropdown", "palette.js": "palette",
	"search.js": "search", "select.js": "select", "popover.js": "popover", "structured-input.js": "structuredinput",
	"table.js": "table", "tabs.js": "tabs", "tooltip.js": "tooltip",
}

func classifyChanges(changes []Change, graph packageGraph, manifest identityManifest) classification {
	result := classification{direct: make(map[string]string)}
	authoredTempl := false
	authoredJS := false
	for _, change := range changes {
		path := change.NewPath
		if strings.HasSuffix(path, ".templ") {
			authoredTempl = true
		}
		if strings.HasPrefix(path, "assets/js/src/") && strings.HasSuffix(path, ".js") {
			authoredJS = true
		}
	}

	for _, change := range changes {
		classifyChange(change, changes, authoredTempl, authoredJS, graph, manifest, &result)
	}
	slices.Sort(result.roots)
	result.roots = slices.Compact(result.roots)
	return result
}

func classifyChange(
	change Change,
	changes []Change,
	authoredTempl, authoredJS bool,
	graph packageGraph,
	manifest identityManifest,
	result *classification,
) {
	path := change.NewPath
	kind := change.Status[0]
	if kind == 'D' || kind == 'R' || kind == 'C' {
		result.full = append(result.full, fmt.Sprintf("%s requires base-side dependency analysis", change.Status))
		return
	}
	if derivedPath(path) {
		if derivedExplained(path, changes, authoredTempl, authoredJS) {
			result.skipped = append(result.skipped, path)
			return
		}
		result.full = append(result.full, fmt.Sprintf("generated-only change %s", path))
		return
	}
	if strings.HasPrefix(path, "site/internal/server/") {
		classifyServerPath(path, result)
		return
	}
	if strings.HasPrefix(path, "site/tests/e2e/") {
		classifyE2EPath(path, manifest, result)
		return
	}
	if strings.HasPrefix(path, "assets/js/src/") && strings.HasSuffix(path, ".js") {
		classifyJavaScriptPath(path, result)
		return
	}
	if globalPath(path) {
		result.full = append(result.full, fmt.Sprintf("global runtime or theme change %s", path))
		return
	}
	if pkg, ok := graph.packageForPath(path); ok && classifiablePackagePath(path) {
		result.roots = append(result.roots, pkg)
		if identity, identified := graph.identities[pkg]; identified {
			result.direct[identity] = fmt.Sprintf("%s: changed directly in %s", identity, path)
		}
		return
	}
	result.full = append(result.full, fmt.Sprintf("unclassified change %s", path))
}

func classifyServerPath(path string, result *classification) {
	identities, ok := identitiesForHandler(path)
	if !ok {
		result.full = append(result.full, fmt.Sprintf("shared or unclassified server change %s", path))
		return
	}
	for _, identity := range identities {
		result.direct[identity] = fmt.Sprintf("%s: changed handler %s", identity, path)
	}
}

func classifyE2EPath(path string, manifest identityManifest, result *classification) {
	identities := manifest.fileIdentities[filepath.Base(path)]
	if len(identities) == 0 {
		result.full = append(result.full, fmt.Sprintf("shared or full-only E2E change %s", path))
		return
	}
	for _, identity := range identities {
		result.direct[identity] = fmt.Sprintf("%s: changed E2E test %s", identity, path)
	}
}

func classifyJavaScriptPath(path string, result *classification) {
	identity, ok := componentJavaScript[filepath.Base(path)]
	if !ok {
		result.full = append(result.full, fmt.Sprintf("shared JavaScript change %s", path))
		return
	}
	result.direct[identity] = fmt.Sprintf("%s: changed JavaScript %s", identity, path)
}

func derivedPath(path string) bool {
	return strings.HasSuffix(path, "_templ.go") || path == "assets/styles.css" ||
		path == "assets/goshtoso-theme.css" || path == "assets/js/goshtoso.min.js" ||
		path == "assets/js/action-group.js" || path == "assets/js/code-block.js" || path == "assets/js/combobox.js"
}

func derivedExplained(path string, changes []Change, authoredTempl, authoredJS bool) bool {
	if before, ok := strings.CutSuffix(path, "_templ.go"); ok {
		source := before + ".templ"
		return slices.ContainsFunc(changes, func(change Change) bool { return change.NewPath == source })
	}
	if path == "assets/styles.css" {
		return authoredTempl
	}
	if strings.HasPrefix(path, "assets/js/") {
		return authoredJS
	}
	return path == "assets/goshtoso-theme.css" && slices.ContainsFunc(changes, func(change Change) bool {
		return change.NewPath == "assets/css/theme.css"
	})
}

func globalPath(path string) bool {
	return strings.HasPrefix(path, "assets/css/") || strings.HasPrefix(path, "assets/js/runtime/") ||
		strings.HasPrefix(path, "internal/runtimegen/") || strings.HasPrefix(path, "cmd/runtimegen/") || path == "muamba.yaml" ||
		path == "assets/runtime.overlay.yaml" || path == "assets/muamba_gen.go" ||
		path == "assets/embed.go" || path == "assets/runtime_manifest.go" || path == "assets/runtime_manifest_gen.go" || path == "assets/vendor_gen.go" ||
		strings.Contains(filepath.Base(path), "tailwind")
}

func classifiablePackagePath(path string) bool {
	return strings.HasPrefix(path, "components/") ||
		strings.HasPrefix(path, "site/internal/pages/demo/componentpages/") ||
		strings.HasPrefix(path, "site/internal/pages/demo/examplepages/") ||
		strings.HasPrefix(path, "site/internal/examples/")
}
