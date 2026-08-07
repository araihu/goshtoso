# Release Checklist

Use this checklist before pushing a `v*` tag. The release workflow validates
many of these steps, but the checklist keeps the public release story coherent.

## Before Tagging

- Confirm `main` is green in GitHub Actions.
- Run the local gates relevant to the release:
  - `templ generate`
  - `just css`
  - `go run ./cmd/skillgen`
  - `go tool muamba verify --strict`
  - `go tool muamba generate-go --strict --check --dir assets --output muamba_gen.go`
  - `go run ./cmd/runtimegen -check`
  - `npx --yes skills add . --list`
  - `npx --yes skills use . --skill using-goshtoso`
  - `just site-current-source-integration`
  - `just site-pinned-dependency-deployability`
  - `scripts/run-release-coverage.sh --local-dry-run`
- Check that generated files have no drift:
  - `*_templ.go`
  - `assets/styles.css`
  - `assets/goshtoso-theme.css`
  - `assets/vendor_gen.go`
  - `assets/runtime_manifest_gen.go`
  - `assets/muamba_gen.go`
  - `docs/RUNTIME_DEPENDENCIES.md`
  - `site/internal/pages/demo/contentpages/legal/runtime_attributions_gen.go`
  - `.claude/skills/using-goshtoso/components-reference.md`
  - `.agents/skills/using-goshtoso/references/components-reference.md`
  - `.agents/skills/using-goshtoso/references/application-patterns.md`
  - `.agents/skills/using-goshtoso/references/visual-acceptance.md`
  - `.agents/skills/using-goshtoso/references/adversarial-acceptance.md`
- Update `VERSIONS.md` when the release uses a new Goshtoso tag or Tailwind
  version from `muamba.yaml`.
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

- Confirm the `Full release verification + coverage` job passed before the
  publishing job began. A manual workflow dispatch runs this same gate without
  creating a release or updating either badge.
- Confirm the GitHub release exists and includes `assets/styles.css` and
  `assets/goshtoso-theme.css`.
- Confirm the release badge endpoint has the new tag.
- Confirm the coverage badge reports the authoritative full-suite percentage
  for authored Go source from this release; focused PR/main runs must never
  update it.
- Confirm Codecov received `.coverage/coverage-authored.out` and exposes the
  current release report. The `CODECOV_TOKEN` repository secret must be present.
- Confirm the release coverage artifact retains both authored-source and full
  generated-inclusive profiles, function summaries, and HTML reports.
- Open a follow-up PR that pins `site/go.mod` to the new tag and updates any
  version-aware documentation links. Never push this follow-up directly to the
  protected `main` branch.
- Confirm `VERSIONS.md` has a row for the released tag.
- Confirm the documentation site deploy completed or was intentionally skipped.
- Confirm `npx skills add araihu/goshtoso --list` discovers the released
  consumer-agent skill.

## Historical v0.1.0 icon catalog evidence

This section is retained as historical evidence for the v0.1.0 catalog work. It
does not describe the v0.2.0 release boundary and does not authorize a tag,
push, deployment, or a `site/go.mod` dependency update.

- Assets release candidate: P1 source `246cb28`, integrated `80d43a3`; UI
  sprite correction `613335f`.
- Upstream immutable Assets `dist/catalog.json`: 302 records, schema `1`,
  SHA-256 `d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af`.
- Local exact Heroicons generator subset:
  `internal/iconcatalog/testdata/heroicons-catalog.json`, 67 records selected
  as namespace `ui` and product `heroicons`; every selected record is the exact
  upstream object. SHA-256
  `0a420ad65e2fe7db3e2cc5dbb6c87167fcd6e85f64a3ebc409e2a58c9bd111ef`.
- Assets UI sprite correction `613335f`: SHA-256
  `75e282de7a19efba9cf0285b44af0641c1527361f921b7d7f8020efc1f1f0fb7`.
- Goshtoso approved functional head: `ab05821`; bundled default sprite path:
  `/assets/icons/heroicons.svg`.
- Bundled binding command:

  ```bash
  go run ./cmd/iconcatalog -catalog internal/iconcatalog/testdata/heroicons-catalog.json -namespace ui -product heroicons -sprite-url /assets/icons/heroicons.svg -package heroicons -const-prefix Icon -out components/icon/heroicons/names_gen.go
  ```

- Current-source site integration and full icon E2E pass. Pinned-dependency
  deployability is intentionally deferred: `site/go.mod` pins public
  `github.com/araihu/goshtoso@v0.0.14-0.20260729070831-8863d6b7d0e8`. All
  observed old-pin failures are recorded: missing
  `github.com/araihu/goshtoso/components/icon` at
  `site/internal/pages/demo/componentpages/icon/icon_templ.go:12:2`; missing
  `github.com/araihu/goshtoso/components/icon/heroicons` at `:13:2`; and
  `undefined: components.KindIcon` at
  `site/internal/pages/catalog/catalog.go:221`.
- Do not hide this with build tags, duplicated component code, a `replace`, or
  a local workspace. Acceptance after approval: tag the root release, update
  the site pin in its follow-up pull request, then rerun pinned-dependency
  deployability successfully before merge or deployment.

## v0.2.0 iconpack release gate

- [x] Muamba `v0.0.4` is publicly released and the Assets release no longer
  depends on an unpublished pseudo-version.
- [x] Arai Hû Assets `v0.2.0` is publicly released with the recorded catalog,
  release, checksums, and archive digests reverified from published bytes.
  Published SHA-256 values: catalog
  `a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6`, release
  JSON `77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260`,
  checksums `334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5`,
  and archive `5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5`.
- [x] `docs/ICONPACK.md` uses the immutable Assets release URL, distinguishes
  Assets hashes from Goshtoso's version, and records the published archive,
  release, and checksums digests.
- [x] Root and site documentation tests assert the pinned command, archive
  hashes, clickable guide, release notes, and complete route metadata.
- [ ] After the root tag exists, `site/go.mod` is updated to
  `github.com/araihu/goshtoso v0.2.0` in a separate follow-up PR with no
  `replace`, and both site module contracts pass.

## Support Notes

The latest supported pre-1.0 release line receives routine fixes. Older tags
remain available for reproducibility but are not maintained.
