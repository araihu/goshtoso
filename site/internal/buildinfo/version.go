// Package buildinfo exposes immutable metadata injected into the demo binary.
package buildinfo

// goDocsVersion is replaced at build time with the Goshtoso module version
// pinned by site/go.mod. Plain local builds intentionally keep "development"
// so the site never presents an unversioned checkout as released API docs.
var goDocsVersion = "development"

// GoDocsVersion returns the Goshtoso version documented by this demo binary.
func GoDocsVersion() string {
	return goDocsVersion
}
