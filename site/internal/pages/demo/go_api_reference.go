package demo

import (
	"regexp"

	"github.com/araihu/goshtoso/site/internal/buildinfo"
	"github.com/araihu/goshtoso/site/internal/pages/catalog"
)

var goModuleVersionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`)

type goAPIReference struct {
	PackagePath string
	Version     string
	URL         string
	Versioned   bool
}

func componentGoAPIReferenceData(active string) goAPIReference {
	entry, ok := catalog.LookupActive(active)
	if !ok {
		return goAPIReference{}
	}

	version := buildinfo.GoDocsVersion()
	return newGoAPIReference(entry, version)
}

func newGoAPIReference(entry catalog.Entry, version string) goAPIReference {
	reference := goAPIReference{
		PackagePath: entry.GoPackagePath(),
		Version:     version,
	}
	if goModuleVersionPattern.MatchString(version) {
		reference.URL = entry.GoDocsURL(version)
		reference.Versioned = true
	}
	return reference
}
