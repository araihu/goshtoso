// Command custommanifest renders the current root module's custom runtime
// fixture for the site module's browser contract.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/components/head"
)

const customRuntimePrefix = "/custom-runtime/"

type probeOutput struct {
	HeadHTML      string            `json:"head_html"`
	Rewrites      map[string]string `json:"rewrites"`
	FailedPrimary string            `json:"failed_primary"`
}

func main() {
	output, err := renderProbe()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func renderProbe() (probeOutput, error) {
	manifest := assets.DefaultRuntimeManifest()
	rewrites := make(map[string]string)

	stylesheetURL := customRuntimePrefix + "primary/styles.css"
	rewrites[stylesheetURL] = manifest.Stylesheet.LocalURL
	manifest.Stylesheet.PrimaryURL = stylesheetURL
	manifest.Stylesheet.LocalURL = customRuntimePrefix + "inventory/styles.css"

	loaderURL := customRuntimePrefix + "primary/loader.js"
	rewrites[loaderURL] = manifest.Loader.LocalURL
	manifest.Loader.PrimaryURL = loaderURL
	manifest.Loader.LocalURL = customRuntimePrefix + "inventory/loader.js"

	for index := range manifest.Dependencies {
		dependency := &manifest.Dependencies[index]
		primaryURL := customRuntimePrefix + "primary/" + string(dependency.Role) + ".js"
		fallbackURL := customRuntimePrefix + "fallback/" + string(dependency.Role) + ".js"
		rewrites[primaryURL] = dependency.LocalURL
		rewrites[fallbackURL] = dependency.LocalURL
		dependency.PrimaryURL = primaryURL
		dependency.LocalURL = fallbackURL
		if dependency.Role == assets.RuntimeRoleDarkMode ||
			dependency.Role == assets.RuntimeRoleHTMXExtSSE ||
			dependency.Role == assets.RuntimeRoleHTMXExtWS {
			dependency.Enabled = true
		}
	}

	var headHTML strings.Builder
	if err := head.Dependencies(head.WithRuntimeManifest(manifest)).Render(context.Background(), &headHTML); err != nil {
		return probeOutput{}, fmt.Errorf("render custom runtime manifest: %w", err)
	}
	return probeOutput{
		HeadHTML:      headHTML.String(),
		Rewrites:      rewrites,
		FailedPrimary: customRuntimePrefix + "primary/htmx-ext-sse.js",
	}, nil
}
