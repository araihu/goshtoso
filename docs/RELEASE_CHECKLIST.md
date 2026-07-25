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
- Update `VERSIONS.md` when the release uses a new Goshtoso tag or Tailwind
  version.
- Review `README.md`, `docs/USAGE.md`, and `ROADMAP.md` for stale version,
  component, or stability language.
- Review `/docs/agents` and `.agents/skills/using-goshtoso/SKILL.md` for stale
  consumer-agent installation guidance.
- Prepare release notes that call out:
  - breaking API changes,
  - new components or variants,
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

## Support Notes

During alpha, only the latest `v0.0.x` tag receives fixes. Older tags remain
available for reproducibility but are not maintained.
