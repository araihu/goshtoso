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
  - confirm the installed `using-goshtoso` copy contains `SKILL.md`,
    `agents/openai.yaml`, and every `references/*.md` file
  - confirm streamed `skills use` output contains the complete required
    discovery pass without depending on bundled references
  - `just site-current-source-integration`
  - `just published-consumer`
  - `scripts/run-release-coverage.sh --local-dry-run`
- Check that generated and public distribution files have no drift:
  - `*_templ.go`
  - `assets/styles.css`
  - `assets/goshtoso-theme.css`
  - `assets/vendor_gen.go`
  - `assets/runtime_manifest_gen.go`
  - `assets/muamba_gen.go`
  - `docs/RUNTIME_DEPENDENCIES.md`
  - `site/internal/pages/demo/contentpages/legal/runtime_attributions_gen.go`
  - `.claude/skills/using-goshtoso/components-reference.md`
  - `.agents/skills/using-goshtoso/SKILL.md`
  - `.agents/skills/using-goshtoso/agents/openai.yaml`
  - `.agents/skills/using-goshtoso/references/components-reference.md`
  - `.agents/skills/using-goshtoso/references/design-intelligence.md`
  - `.agents/skills/using-goshtoso/references/ecosystem-discovery.md`
  - `.agents/skills/using-goshtoso/references/runtime-integration.md`
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
- Confirm the publish job verified the new Go tag with the isolated consumer
  fixture before creating the release. The site uses the tagged checkout and
  requires no dependency-pin follow-up.
- Confirm `VERSIONS.md` has a row for the released tag.
- Confirm the documentation site deploy completed or was intentionally skipped.
- Confirm the release image job passed, the version tag and `latest` resolve to
  the same GHCR digest for the latest stable release, and the
  `image-goshtoso-release` artifact contains that digest. Image publication does
  not perform a homelab rollout.
- Confirm `npx skills add araihu/goshtoso --list` discovers the released
  consumer-agent skill.

## Support Notes

During alpha, only the latest `v0.x.y` tag receives fixes. Older tags remain
available for reproducibility but are not maintained.
