# Site Module Contracts

Goshtoso has two Go modules with different dependency contracts. Pull requests
must keep both green.

## Current-source integration

`site/` must integrate with the root library from the same checkout. This mode
uses a temporary `go.work`; it proves an atomic root-plus-site change works
together without committing workspace state.

Run:

```bash
just site-current-source-integration
```

Code CI retains this contract for site vet, lint, tests, coverage, E2E, and the
server build.

## Pinned-dependency deployability

`site/` must also build as a standalone module from the Goshtoso version in
`site/go.mod`. This mode forces `GOWORK=off`, discovers every site package, runs
all non-E2E site tests, and builds `site/cmd/server`.

Run:

```bash
just site-pinned-dependency-deployability
```

The protected `Required CI` check runs this contract on every pull request. A
failure ends with a message such as:

```text
site pinned-dependency deployability failed during package discovery
site/go.mod must pin a public Goshtoso version containing every API imported by site/.
Do not mask this contract with go.work or a replace directive; publish or merge root changes first, then pin a reachable tag or pseudo-version.
```

The gate rejects a `replace` for `github.com/araihu/goshtoso`. A checked
`replace github.com/araihu/goshtoso => ..` would make the standalone check use
the sibling checkout, hide stale pins, and make `site/go.mod` non-portable. Pin
a public tag or exact public pseudo-version instead.

## API-change sequencing

A pull request cannot pin the future squash commit that will result from its
own merge. When `site/` needs a new root API, use two phases unless the repository
deliberately changes its module architecture:

1. Merge the root API without making the standalone site depend on it, then
   publish a tag or wait until the merged commit resolves as a public
   pseudo-version.
2. In a follow-up pull request, update `site/go.mod` and `site/go.sum` to that
   reachable version, adopt the API in `site/`, and run both contracts.

A release tag is preferred for stable documentation. A public pseudo-version is
acceptable between releases; after the next semantic tag, follow the release
checklist and pin the site to that tag.
