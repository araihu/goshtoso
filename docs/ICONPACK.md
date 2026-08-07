# Consumer-local icon packs

`iconpack` generates an attributed SVG sprite and typed Go component package in
a consumer-owned directory. It does not add selected icons to Goshtoso's
embedded Heroicons package. Consumers choose assets by exact Arai Hu Assets
catalog canonical name, including literal mixed-case names such as
`brand-developer-icons-tRPC`.

The input must be either an extracted release root or a release archive. A
Goshtoso checkout, GitHub source archive, `internal/acquisition/vendor` tree, or
other source directory is not a release boundary and is rejected. Generation
requires separately pinned SHA-256 values for `catalog.json`, `release.json`,
and `checksums.txt`; archive mode also requires the archive SHA-256. The command
verifies every checksums record, release inventory entry, selected artifact,
namespace/product allowlist, and exact source-sprite symbol before writing.

## Exact-name selection

This local candidate example uses the frozen Arai Hu Assets v0.2.0 archive. It
does not claim that v0.2.0, its Muamba prerequisite, or this Goshtoso change is
published.

```bash
go run ./cmd/iconpack \
  -release-archive /path/to/araihu-assets-v0.2.0.tar.gz \
  -archive-sha256 dcb97bbbbf98fb2e3c0e96b63eefb17b9b60eb2b3d8097fa6b4e2876f3f19271 \
  -release v0.2.0 \
  -catalog-sha256 a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6 \
  -release-json-sha256 0650e51dd2b7ec7797622b3cdd9ff75dfd53cb1914155931014223bbd1684fa6 \
  -checksums-sha256 86dac118901d423117e20bd14ce6ed30717fca9c2a8244909c2be4b926ce1c4e \
  -name brand-developer-icons-tRPC \
  -name ui-hi-16-solid-check \
  -out ./internal/appicons \
  -package appicons \
  -const-prefix Icon \
  -sprite-url /assets/icons/app.svg
```

Use `-release-root /path/to/extracted-release` instead of `-release-archive`
and `-archive-sha256` for a verified extracted root. The output parent must
already exist.

## Selection manifest

JSON is the canonical manifest form:

```json
{
  "schemaVersion": 1,
  "names": [
    "brand-developer-icons-tRPC",
    "ui-hi-16-solid-check"
  ]
}
```

Pass it with `-manifest ./icons.json` instead of repeated `-name` flags. YAML
with identical field names is accepted for compatibility:

```yaml
schemaVersion: 1
names:
  - brand-developer-icons-tRPC
  - ui-hi-16-solid-check
```

## Output contract

One successful publication creates the requested directory atomically with:

- `sprite.svg`: only selected `<symbol>` elements, using catalog
  `spriteSymbol` values literally.
- `icons_gen.go`: typed canonical names, `icon.Symbol` bindings, lookup data,
  and a consumer-local `Icon` component bound to the selected sprite URL.
- `manifest.json`: release, catalog, selected-asset, generated-identifier, and
  output-file identities.
- `provenance.json` plus `PROVENANCE/*.json`: selected and release-authored
  upstream provenance.
- `LICENSES/*.txt` and `NOTICE`: exact verified attribution bytes.

The public and generated API uses generic names such as `Name`, `Glyph`,
`Config`, `Lookup`, and `Icon`. Only Go identifiers are normalized;
`CanonicalName` and manifest values preserve catalog bytes exactly.

Publication uses a sibling lock and same-parent staged directory. An absent
destination is published with one directory rename. An identical owned output
is idempotent. A changed, symlinked, non-owned, or unrelated destination is
never replaced. Use `-check` in CI to verify an existing output byte-for-byte.

## External consumer proof

The proof script extracts the verified release archive into a private temporary
boundary, generates a separate Go module, builds it, renders the generated
component, and checks the literal `devicon-trpc` reference:

```bash
./scripts/iconpack-consumer-proof.sh \
  /path/to/araihu-assets-v0.2.0.tar.gz \
  dcb97bbbbf98fb2e3c0e96b63eefb17b9b60eb2b3d8097fa6b4e2876f3f19271
```

The temporary consumer replaces Goshtoso only as its Go dependency so an
uncommitted producer can be tested. All icon inputs still come exclusively from
the verified release archive; no Goshtoso checkout or vendored asset tree is an
icon source.
