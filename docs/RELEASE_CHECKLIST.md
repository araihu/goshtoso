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
  - `.agents/skills/using-goshtoso/LICENSE`
  - `.agents/skills/using-goshtoso/SKILL.md`
  - `.agents/skills/using-goshtoso/agents/openai.yaml`
  - `.agents/skills/using-goshtoso/references/components-reference.md`
  - `.agents/skills/using-goshtoso/references/design-intelligence.md`
  - `.agents/skills/using-goshtoso/references/ecosystem-discovery.md`
  - `.agents/skills/using-goshtoso/references/migration-v0.3.0.md`
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
- Confirm the new Go tag passed the isolated published-consumer fixture before
  creating the release. The fixture needs no site dependency-pin follow-up.
- Confirm `VERSIONS.md` has a row for the released tag.
- Confirm the documentation site deployment for the release commit completed
  through the successful `Required CI` run on main, or was intentionally skipped.
  Creating a GitHub Release does not trigger a site deployment.
- Release-image publication is retired and does not apply to current releases.
- Confirm `npx skills add araihu/goshtoso --list` discovers the released
  consumer-agent skill.

## v0.4.4 publication recovery

The immutable tag `v0.4.4` points to
`234eaaf97a7058d137dfed942423dcfb77f2ace6`. In
[run 40219694440](https://github.com/araihu/goshtoso/actions/runs/40219694440),
full release verification passed, but publication stopped while extracting the
missing changelog section. Add that section and the compatibility row on main;
do not move the tag or change the library or assets in it.

Complete the publication outputs using that exact tag:

1. Create the GitHub Release with the corrected notes and Tailwind 4.3.3 artifact
   section. Upload the tag's unmodified `styles.css` and `goshtoso-theme.css`.
2. Patch both documents in the existing badge Gist in one request. Use
   `release.json` message `v0.4.4`, color `blue`, and `coverage.json` message
   `94.3%`, color `brightgreen`, from this run's successful `release-coverage`
   artifact. Keep existing badge labels and schema version. Use authenticated
   tooling without writing credentials to files or output.
3. Download the published assets and compare their SHA-256 values to the tag:
   `styles.css` is `18b633cad74debfdfa972b324ec4c99ce31716301e66cdba84ba617a5d9f3d7b`;
   `goshtoso-theme.css` is `e31cdba9e67f19658f6fce2994ab0a0828d006b8e26b59077433255563c9125d`.
   Verify both public badge values and retain the original workflow outcome.
   Manual recovery does not make the failed publication job successful.
4. Verify the existing main-based site deployment for the exact tag commit,
   [run 40220673930](https://github.com/araihu/goshtoso/actions/runs/40220673930).
   Manual GitHub Release creation triggers no new workflow. Release-image
   publication is retired and is not a recovery requirement.

## Support Notes

During alpha, only the latest `v0.x.y` tag receives fixes. Older tags remain
available for reproducibility but are not maintained.
