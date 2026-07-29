# Release Checklist

Use this checklist before pushing a `v*` tag. The release workflow validates
many of these steps, but the checklist keeps the public release story coherent.

## Before Tagging

- Confirm `main` is green in GitHub Actions.
- Run the local gates relevant to the release:
  - `templ generate`
  - `just css`
  - `go run ./cmd/skillgen`
  - `go run ./cmd/vendorgen -check`
  - `npx --yes skills add . --list`
  - `npx --yes skills use . --skill using-goshtoso`
  - `just site-current-source-integration`
  - `just site-pinned-dependency-deployability`
  - `go test ./... -count=1`
  - `cd site && go test $(go list ./... | grep -v /tests/e2e) -count=1`
  - `go test ./site/tests/e2e/... -count=1 -timeout 15m`
- Check that generated files have no drift:
  - `*_templ.go`
  - `assets/styles.css`
  - `assets/goshtoso-theme.css`
  - `assets/vendor_gen.go`
  - `.claude/skills/using-goshtoso/components-reference.md`
  - `.agents/skills/using-goshtoso/references/components-reference.md`
  - `.agents/skills/using-goshtoso/references/application-patterns.md`
  - `.agents/skills/using-goshtoso/references/visual-acceptance.md`
  - `.agents/skills/using-goshtoso/references/adversarial-acceptance.md`
- Update `VERSIONS.md` when the release uses a new Goshtoso tag or Tailwind
  version.
- Review `README.md`, `docs/USAGE.md`, and `ROADMAP.md` for stale version,
  component, or stability language.
- Review `/docs/agents`, application recipes, and
  `.agents/skills/using-goshtoso/SKILL.md` for stale consumer-agent guidance.
- Prepare release notes that call out:
  - breaking API changes,
  - new components or configuration choices,
  - asset/runtime version changes,
  - migration steps for consumers.
- Give the release a dated `CHANGELOG.md` heading matching the tag. The release
  workflow publishes that section as the GitHub release notes.

## After the Tag Workflow

- Confirm the GitHub release exists and includes `assets/styles.css` and
  `assets/goshtoso-theme.css`.
- Confirm the release badge endpoint has the new tag.
- Confirm the coverage badge endpoint still renders.
- Open a follow-up PR that pins `site/go.mod` to the new tag and updates any
  version-aware documentation links. Never push this follow-up directly to the
  protected `main` branch.
- Confirm `VERSIONS.md` has a row for the released tag.
- Confirm the documentation site deploy completed or was intentionally skipped.
- Confirm `npx skills add araihu/goshtoso --list` discovers the released
  consumer-agent skill.

## Pending v0.1.0 icon catalog evidence

This is release-candidate evidence only. It does not authorize a tag, push,
deployment, or a `site/go.mod` dependency update.

- Assets release candidate: P1 source `246cb28`, integrated `80d43a3`.
- Assets catalog: schema `1`, SHA-256
  `d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af`.
- Assets UI sprite: SHA-256
  `6b312ee2cf9f0e91c4621bd4eec348ecaf39cfdc0a00c8ddefdf4d7f8e9f32a5`.
- Goshtoso approved functional head: `ab05821`; bundled default sprite path:
  `/assets/icons/heroicons.svg`.
- Bundled binding command:

  ```bash
  go run ./cmd/iconcatalog -catalog internal/iconcatalog/testdata/heroicons-catalog.json -namespace ui -product heroicons -sprite-url /assets/icons/heroicons.svg -package heroicons -const-prefix Icon -out components/icon/heroicons/names_gen.go
  ```

- Current-source site integration and full icon E2E pass. Pinned-dependency
  deployability is intentionally deferred: `site/go.mod` pins public
  `github.com/araihu/goshtoso@v0.0.14-0.20260729070831-8863d6b7d0e8`, which
  does not provide `github.com/araihu/goshtoso/components/icon` or
  `github.com/araihu/goshtoso/components/icon/heroicons`. The standalone
  command fails with `internal/pages/demo/components/icon_templ.go:12:2: no
  required module provides package github.com/araihu/goshtoso/components/icon`
  and the corresponding Heroicons import at `:13:2`.
- Do not hide this with build tags, duplicated component code, a `replace`, or
  a local workspace. After approval: tag the root release, update the site pin
  in its follow-up pull request, then rerun pinned-dependency deployability
  before merge or deployment.

## Support Notes

During alpha, only the latest `v0.0.x` tag receives fixes. Older tags remain
available for reproducibility but are not maintained.
