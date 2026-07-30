// Command araihu-assets-update refreshes repository-owned fallback assets from
// an already downloaded and extracted araihu/assets release. It never fetches.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/araihu/goshtoso/internal/araihuassets"
)

func main() {
	root := flag.String("root", ".", "repository root")
	releaseDir := flag.String("release-dir", "", "verified, extracted araihu/assets release directory")
	manifest := flag.String("manifest", araihuassets.DefaultManifestPath, "repository-relative manifest path")
	repository := flag.String("assets-repository", "", "release repository identity")
	revision := flag.String("assets-revision", "", "release commit identity")
	release := flag.String("release", "", "stable release tag")
	releaseURL := flag.String("release-url", "", "immutable release archive URL")
	releaseSHA256 := flag.String("release-sha256", "", "release archive SHA-256")
	releaseJSONSHA256 := flag.String("release-json-sha256", "", "release.json SHA-256")
	flag.Parse()

	if *releaseDir == "" {
		fail("-release-dir is required")
	}
	identityValues := []string{*repository, *revision, *release, *releaseURL, *releaseSHA256, *releaseJSONSHA256}
	provided := 0
	for _, value := range identityValues {
		if value != "" {
			provided++
		}
	}
	var identity *araihuassets.ReleaseIdentity
	if provided != 0 {
		if provided != len(identityValues) {
			fail("release identity flags must be provided together")
		}
		identity = &araihuassets.ReleaseIdentity{
			AssetsRepository:  *repository,
			AssetsRevision:    *revision,
			Release:           *release,
			ReleaseURL:        *releaseURL,
			ReleaseSHA256:     *releaseSHA256,
			ReleaseJSONSHA256: *releaseJSONSHA256,
		}
	}

	result, err := araihuassets.Update(araihuassets.Options{
		RepoRoot:     *root,
		ReleaseRoot:  *releaseDir,
		ManifestPath: *manifest,
		Identity:     identity,
	})
	if err != nil {
		fail(err.Error())
	}
	if len(result.Changed) == 0 {
		fmt.Println("Arai Hu fallback assets already current")
		return
	}
	for _, changed := range result.Changed {
		fmt.Println(changed)
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, "araihu-assets-update:", message)
	os.Exit(1)
}
