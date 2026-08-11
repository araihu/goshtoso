# Arai Hu fallback assets

Goshtoso keeps four repository-owned Arai Hû theme and brand fallback files
synchronized with immutable
[`araihu/assets`](https://github.com/araihu/assets) releases. The root
[`araihu-assets.json`](../araihu-assets.json) manifest pins the release tag,
release commit, archive URL and SHA-256, exact `release.json` SHA-256, and every
allowed source-to-destination mapping.

Brand mappings additionally pin the catalog canonical name and semantic roles.
The updater rejects a release when a canonical name resolves to a different
path, role, or checksum. Theme CSS is a release-inventory file without a brand
catalog identity. Goshtoso's bundled Heroicons sprite remains a separately
curated, typed surface and is not replaced by Arai Hû release fan-out.

## Local update

Download and verify the immutable tar archive separately, then extract it into
a new directory. The Go updater deliberately has no download mode:

```bash
go run ./cmd/araihu-assets-update -release-dir /path/to/verified-release
```

To advance the manifest, provide the complete identity as one unit:

```bash
go run ./cmd/araihu-assets-update \
  -release-dir /path/to/verified-release \
  -assets-repository araihu/assets \
  -assets-revision 0123456789abcdef0123456789abcdef01234567 \
  -release v1.2.3 \
  -release-url https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz \
  -release-sha256 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
  -release-json-sha256 abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789
```

Only stable `vX.Y.Z` tags and the exact `araihu/assets` release URL shape are
accepted. Source and destination traversal, symlinks, duplicate destinations,
unknown catalog collisions, and checksum mismatches fail before any copy. Run
the same command twice; the second run must report that fallbacks are current.
All changed outputs are staged before replacement. A mid-apply failure restores
every already-replaced file, and the manifest is replaced last. Transaction
errors report the failed path, paths applied before failure, and any incomplete
rollback operations requiring manual recovery.

Focused verification:

```bash
go test ./internal/araihuassets ./cmd/araihu-assets-update -count=1
```

## Automation

`.github/workflows/araihu-assets.yml` accepts `repository_dispatch` event
`araihu-assets-released` and guarded manual dispatch. Both use the same fields:

- `assets_repository`
- `assets_revision`
- `release`
- `release_url`
- `release_sha256`
- `release_json_sha256`

The workflow validates all fields and resolves the release tag to the dispatched
commit before download, downloads and verifies the archive once, runs the
offline updater twice to prove idempotence, and opens or updates
`automation/araihu-assets-vX.Y.Z`. It uses the selected-repository
GitHub App secrets `ARAIHU_ASSETS_APP_ID` and
`ARAIHU_ASSETS_APP_PRIVATE_KEY`. Existing `dependencies` and `assets` labels
are applied when present. No label is created, and no PR is auto-merged.

## Known hardening debt

Archive SHA-256 verification makes the published archive immutable, and the
workflow currently rejects traversal and link members before extraction. A
future focused hardening slice should also reject duplicate member names,
case-folded member collisions, and every non-regular archive member type before
extraction. This does not change the offline updater's release inventory,
catalog, checksum, path, symlink, or transactional replacement checks.
