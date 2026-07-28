package assets

import (
	"runtime/debug"
	"strings"
)

// GoshtosoModulePath is the Go module path whose build identity
// GoshtosoVersion resolves.
const GoshtosoModulePath = "github.com/araihu/goshtoso"

// VersionStatus describes whether a consumer binary has an exact,
// release-safe Goshtoso module identity.
type VersionStatus string

const (
	// VersionExact means Version identifies the unreplaced module recorded by Go
	// build information. Version can be a release or immutable pseudo-version;
	// Sum records its checksum when Go build information provides one.
	VersionExact VersionStatus = "exact"
	// VersionDevelopment means Goshtoso is the unversioned main module or has a
	// development marker instead of an exact module version.
	VersionDevelopment VersionStatus = "development"
	// VersionReplaced means Go build information records a replacement. The
	// requested version does not identify the replacement bytes.
	VersionReplaced VersionStatus = "replaced"
	// VersionUnavailable means build information is absent or does not contain
	// the Goshtoso module.
	VersionUnavailable VersionStatus = "unavailable"
)

// VersionReference is one module record from Go build information.
type VersionReference struct {
	Path    string
	Version string
	Sum     string
}

// VersionInfo reports the Goshtoso module identity linked into the running
// consumer. Version is populated only when Status is VersionExact; Sum carries
// the Go module checksum when build information provides one.
// For VersionReplaced, Requested records the dependency requested by the
// consumer and Replacement records the module/path actually selected; neither
// is promoted to Version because replacement bytes may be local or otherwise
// differ from the requested release. Consumers that bind caches or offline
// readiness to exact library bytes must fail closed unless Status is
// VersionExact.
type VersionInfo struct {
	ModulePath  string
	Status      VersionStatus
	Version     string
	Sum         string
	Requested   VersionReference
	Replacement VersionReference
}

// GoshtosoVersion reads immutable Go build information for the running
// consumer and returns its linked Goshtoso identity. Development, replaced, and
// unavailable builds deliberately return an empty Version so they cannot be
// mislabeled as a release.
func GoshtosoVersion() VersionInfo {
	info, ok := debug.ReadBuildInfo()
	return resolveGoshtosoVersion(info, ok)
}

func resolveGoshtosoVersion(info *debug.BuildInfo, ok bool) VersionInfo {
	result := VersionInfo{
		ModulePath: GoshtosoModulePath,
		Status:     VersionUnavailable,
	}
	if !ok || info == nil {
		return result
	}

	module := findGoshtosoModule(info)
	if module == nil {
		return result
	}
	if module.Replace != nil {
		result.Status = VersionReplaced
		result.Requested = versionReference(*module)
		result.Replacement = versionReference(*module.Replace)
		return result
	}
	if isDevelopmentVersion(module.Version) {
		result.Status = VersionDevelopment
		return result
	}

	result.Status = VersionExact
	result.Version = module.Version
	result.Sum = module.Sum
	return result
}

func findGoshtosoModule(info *debug.BuildInfo) *debug.Module {
	if info.Main.Path == GoshtosoModulePath {
		return &info.Main
	}
	for _, dependency := range info.Deps {
		if dependency != nil && dependency.Path == GoshtosoModulePath {
			return dependency
		}
	}
	return nil
}

func versionReference(module debug.Module) VersionReference {
	return VersionReference{
		Path:    module.Path,
		Version: module.Version,
		Sum:     module.Sum,
	}
}

func isDevelopmentVersion(version string) bool {
	switch strings.ToLower(strings.TrimSpace(version)) {
	case "", "(devel)", "devel", "development", "(development)":
		return true
	default:
		return false
	}
}
