# Site and Published-Package Contracts

The documentation site demonstrates the library in the same checkout. A separate
consumer fixture verifies the published Go module. An API change and its site
examples can be reviewed and merged in one PR.

## Site integration

The root module is the publishable library. The `site/` module contains the
website, example applications, and browser tests. Build them together with:

```bash
go work init . ./site
go build -o bin/server ./site/cmd/server
just site-current-source-integration
```

`go.work` remains local and gitignored. The integration check creates its own
workspace, runs every non-E2E site package, including theme agreement tests,
and builds the server. Code CI also runs browser tests against this checkout.

The site's `go.mod` replaces `github.com/araihu/goshtoso` with `..`. This also
makes commands such as `cd site && go mod tidy` use the checkout, since tidy
operates on an individual module. Go still records a required version for its
module graph, but the replacement selects the checkout. That version does not
determine the documentation version and does not need release updates. The site
requires the repository checkout; it is not a standalone consumer module.

## Published-package compatibility

`tests/external/published-consumer` is a small standalone application pinned to
a released Goshtoso version, with no replacement directive. It exercises module
identity, component rendering, runtime markup, and bundled CSS and JavaScript.

```bash
just published-consumer
# Check another published tag without changing any tracked file:
scripts/check-published-consumer v0.3.3
```

The script copies only the fixture to a temporary directory, forces `GOWORK=off`,
rejects a Goshtoso replacement, and checks that the fixture actually imports the
library. Tests and builds run with module files read-only. An explicit tag is
resolved and its checksums written only in the temporary copy before testing.

The protected `Required CI` check runs both site integration and this contract. The fixture's checked-in
version is a stable compatibility baseline; update it when the fixture needs a
newly released API, not after every release. Site changes never depend on this
pin. Existing external-consumer tests with local replacements continue to test
the current library's integration contracts separately.

## Build and release identity

Local and main-branch site builds show `dev` and omit versioned API links.
Release images build the root and site from the tagged commit and receive the
exact tag through `GOSHTOSO_DOCS_VERSION`. Docker defaults to `development` when
that build argument is absent.

Before publishing a GitHub release, the release workflow checks the new public
Go tag with the isolated consumer fixture. It then publishes the release and
versioned image. No library-first/site-second PR sequence, temporary public
pseudo-version, or post-release site-pin PR is required.

Browser tests use development metadata by default. Set `GOSHTOSO_DOCS_VERSION`
to a release tag to exercise the versioned badge and API-link assertions.
