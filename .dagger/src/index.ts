import {
  argument,
  CacheSharingMode,
  dag,
  Container,
  Directory,
  File,
  func,
  object,
  ReturnType,
  Secret,
} from "@dagger.io/dagger"

const GO_IMAGE =
  "golang:1.26.5-bookworm@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd"
const JQ_IMAGE =
  "ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486"
const NODE_IMAGE =
  "node:24.19.0-bookworm-slim@sha256:3638d9a6fe4030bd716be989438248074489337ba3275657f93595428be4fc03"
const TEMPL_VERSION = "v0.3.1020"
const PLAYWRIGHT_VERSION = "v0.6100.0"
const GOLANGCI_LINT_VERSION = "v2.12.2"
const GOLANGCI_LINT_AMD64_SHA256 = "8df580d2670fed8fa984aac0507099af8df275e665215f5c7a2ae3943893a553"
const GOLANGCI_LINT_ARM64_SHA256 = "44cd40a8c76c86755375adfeea52cfd3533cb43d7bd647771e0ae065e166df3a"
const LYCHEE_VERSION = "v0.24.2"
const LYCHEE_X86_64_SHA256 = "1f4e0ef7f6554a6ed33dd7ac144fb2e1bbed98598e7af973042fc5cd43951c9a"
const LYCHEE_AARCH64_SHA256 = "91a7bd65685da41b90ccb9bc867a3d649a7818042dae04ff405e55a25bddee4c"

const SOURCE_EXCLUDES = [
  ".dagger/node_modules",
  ".dagger/node_modules/**",
  ".dagger/sdk",
  ".dagger/sdk/**",
  ".git",
  ".git/**",
  ".cache",
  ".cache/**",
  ".coverage",
  ".coverage/**",
  "bin",
  "bin/**",
  "site/tests/e2e/test-results",
  "site/tests/e2e/test-results/**",
]

@object()
export class Goshtoso {
  /** Root+site lint, generated drift, dependency integrity, and current-source build. */
  @func()
  async lintBuild(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cachePartition: string,
  ): Promise<string> {
    return this.goProject(source, cachePartition)
      .withExec(["go", "tool", "muamba", "sync", "--strict", "--target", "linux/amd64", "--cache-dir", ".cache/muamba"])
      .withExec(["go", "tool", "muamba", "verify", "--strict", "--target", "linux/amd64", "--cache-dir", ".cache/muamba"])
      .withExec(["go", "tool", "muamba", "generate-go", "--strict", "--check", "--target", "linux/amd64", "--dir", "assets", "--output", "muamba_gen.go"])
      .withExec(["go", "run", "./cmd/runtimegen", "-check"])
      .withExec(["templ", "generate"])
      .withExec(["bash", "-euo", "pipefail", "-c", "git diff --exit-code -- '*_templ.go'"])
      .withExec(["go", "run", "./cmd/themegen"])
      .withExec([".tools/tailwindcss", "-i", "css/main.css", "-o", "assets/styles.css"])
      .withExec(["bash", "-euo", "pipefail", "-c", "git diff --exit-code -- assets/styles.css assets/goshtoso-theme.css"])
      .withExec(["go", "run", "./cmd/skillgen"])
      .withExec(["git", "diff", "--exit-code", "--", ".agents/skills/using-goshtoso/references/components-reference.md", ".claude/skills/using-goshtoso/components-reference.md"])
      .withExec(["go", "run", "./cmd/iconcatalog", "-catalog", "internal/iconcatalog/testdata/heroicons-catalog.json", "-namespace", "ui", "-product", "heroicons", "-sprite-url", "/assets/icons/heroicons.svg", "-package", "heroicons", "-const-prefix", "Icon", "-out", "components/icon/heroicons/names_gen.go", "-check"])
      .withExec(["go", "run", "./cmd/jslint"])
      .withExec(["go", "run", "./cmd/jsbuild", "-check"])
      .withExec(["go", "mod", "tidy"])
      .withExec(["git", "diff", "--exit-code", "--", "go.mod", "go.sum"])
      .withExec(["bash", "-euo", "pipefail", "-c", "test -z \"$(gofmt -l .)\""])
      .withExec(["go", "vet", "./..."])
      .withExec(["golangci-lint", "run", "--timeout", "5m"])
      .withExec(["bash", "-euo", "pipefail", "-c", "test -f go.work || go work init . ./site"])
      .withExec(["bash", "-euo", "pipefail", "-c", "cd site && go vet ./... && golangci-lint run --timeout 5m"])
      .withExec(["bash", "-euo", "pipefail", "-c", "v=$(cd site && GOWORK=off go list -m -f '{{.Version}}' github.com/araihu/goshtoso); go build -ldflags \"-X github.com/araihu/goshtoso/site/internal/buildinfo.goDocsVersion=$v\" -o /tmp/server ./site/cmd/server"])
      .stdout()
  }

  /** Unit, external-consumer, focused E2E, and merged partial coverage gates. */
  @func({ cache: "never" })
  tests(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    changes: File,
    cachePartition: string,
    runNonce: string,
  ): Directory {
    const script = `set -Eeuo pipefail
mkdir -p /out
finish() {
  rc=$?
  trap - EXIT
  set +e
  test -f /out/e2e-impact.json || printf '{"mode":"full","tags":["full"],"changed_paths":[],"reasons":["impact selection did not complete"]}\n' > /out/e2e-impact.json
  test ! -d .coverage || cp -a .coverage /out/coverage
  test ! -d site/tests/e2e/test-results || cp -a site/tests/e2e/test-results /out/e2e-results
  printf '%s\n' "$rc" > /out/status
  exit 0
}
trap finish EXIT
exec > >(tee /out/tests.log) 2>&1
test -f go.work || go work init . ./site
scripts/run-component-coverage.sh --phase units
(cd tests/external/runtime-manifest && GOWORK=off go test ./... -count=1)
go run ./cmd/e2eimpact --changes-file /tmp/e2e-changes > /out/e2e-impact.json
go tool muamba sync --strict --target linux/amd64 --cache-dir .cache/muamba tailwindcss/cli
templ generate
.tools/tailwindcss -i css/main.css -o assets/styles.css
scripts/run-focused-e2e.sh --current-source-theme-catalog
scripts/run-component-coverage.sh --phase e2e-merge --impact /out/e2e-impact.json`
    return this.browserProject(source, cachePartition)
      .withFile("/tmp/e2e-changes", changes)
      .withEnvVariable("GOSHTOSO_RUN_NONCE", runNonce)
      .withExec(["bash", "-c", script], { expect: ReturnType.Any })
      .directory("/out")
  }

  /** Standalone site/go.mod consumer contract (GOWORK=off). */
  @func()
  async required(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cachePartition: string,
  ): Promise<string> {
    return this.goProject(source, cachePartition)
      .withExec(["scripts/check-site-module", "pinned-dependency"])
      .stdout()
  }

  /** Markdown link gate using a checksum-verified lychee release. */
  @func()
  async docs(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cachePartition: string,
  ): Promise<string> {
    return this.base(source, cachePartition)
      .withExec(["bash", "-euo", "pipefail", "-c", `case "$(uname -m)" in x86_64) arch=x86_64; sha=${LYCHEE_X86_64_SHA256};; aarch64|arm64) arch=aarch64; sha=${LYCHEE_AARCH64_SHA256};; *) echo unsupported architecture >&2; exit 1;; esac; file=lychee-$arch-unknown-linux-gnu.tar.gz; dir=lychee-$arch-unknown-linux-gnu; curl -fsSL -o /tmp/lychee.tgz https://github.com/lycheeverse/lychee/releases/download/lychee-${LYCHEE_VERSION}/$file; echo "$sha  /tmp/lychee.tgz" | sha256sum -c -; tar -xzf /tmp/lychee.tgz -C /usr/local/bin --strip-components=1 "$dir/lychee"`])
      .withExec(["lychee", "--config", ".lychee.toml", "./**/*.md"])
      .stdout()
  }

  /** Full release-equivalent regeneration, both site contracts, E2E, and coverage. */
  @func({ cache: "never" })
  releaseVerify(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    cachePartition: string,
    runNonce: string,
  ): Directory {
    const script = `set -Eeuo pipefail
mkdir -p /out
finish() { rc=$?; trap - EXIT; set +e; test ! -d .coverage || cp -a .coverage/. /out/; test ! -d site/tests/e2e/test-results || cp -a site/tests/e2e/test-results /out/e2e-results; printf '%s\n' "$rc" > /out/status; exit 0; }
trap finish EXIT
exec > >(tee /out/release.log) 2>&1
for t in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do go tool muamba sync --strict --target "$t" --cache-dir .cache/muamba tailwindcss/cli; done
go tool muamba sync --strict --target linux/amd64 --cache-dir .cache/muamba
go tool muamba verify --strict --all-platforms --cache-dir .cache/muamba
go tool muamba generate-go --strict --check --target linux/amd64 --dir assets --output muamba_gen.go
go run ./cmd/runtimegen -check
test -f go.work || go work init . ./site
templ generate; go run ./cmd/themegen; go run ./cmd/jsbuild; go run ./cmd/skillgen
.tools/tailwindcss -i css/main.css -o assets/styles.css
git ls-files --error-unmatch assets/styles.css assets/goshtoso-theme.css >/dev/null
git diff --exit-code -- '*_templ.go' assets/styles.css assets/goshtoso-theme.css 'assets/js/*.js' .agents/skills/using-goshtoso/references/components-reference.md .claude/skills/using-goshtoso/components-reference.md
scripts/check-site-module current-source
scripts/check-site-module pinned-dependency
scripts/run-release-coverage.sh --local-dry-run`
    return this.browserProject(source, cachePartition)
      .withEnvVariable("GOSHTOSO_RUN_NONCE", runNonce)
      .withExec(["bash", "-c", script], { expect: ReturnType.Any })
      .directory("/out")
  }

  /** Publish the verified tag assets and both authoritative badge documents. */
  @func({ cache: "never" })
  async publishRelease(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    coverage: Directory,
    metadata: File,
    githubToken: Secret,
    gistToken: Secret,
    runNonce: string,
  ): Promise<string> {
    const release = await this.json(metadata, ["tag"])
    const tag = this.string(release, "tag")
    if (!/^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/.test(tag)) throw new Error(`invalid release tag: ${tag}`)
    const script = `set -euo pipefail
ver=$(sed -n '/^  tailwindcss:$/,/^  [^ ]/ s/^    version: "\\([^"]*\\)"/\\1/p' muamba.yaml)
test -n "$ver"
awk -v heading="## [$TAG]" 'index($0, heading) == 1 { found=1; next } found && /^## \\[/ { exit } found { print } END { if (!found) exit 1 }' CHANGELOG.md > /tmp/release-notes.md
printf '\\n## Build artifacts\\n\\nBuilt with Tailwind CSS %s.\\nAssets attached: styles.css and goshtoso-theme.css.\\n' "$ver" >> /tmp/release-notes.md
if gh release view "$TAG" --repo araihu/goshtoso >/dev/null 2>&1; then
  gh release edit "$TAG" --repo araihu/goshtoso --notes-file /tmp/release-notes.md
  gh release upload "$TAG" --repo araihu/goshtoso --clobber assets/styles.css assets/goshtoso-theme.css
else
  gh release create "$TAG" --repo araihu/goshtoso --verify-tag --notes-file /tmp/release-notes.md assets/styles.css assets/goshtoso-theme.css
fi
percent=$(cat /coverage/percentage.txt)
color=$(cat /coverage/color.txt)
jq -n --arg coverage "{\\"schemaVersion\\":1,\\"label\\":\\"authored coverage\\",\\"message\\":\\"$percent%\\",\\"color\\":\\"$color\\"}" --arg release "{\\"schemaVersion\\":1,\\"label\\":\\"release\\",\\"message\\":\\"$TAG\\",\\"color\\":\\"blue\\"}" '{files:{"coverage.json":{content:$coverage},"release.json":{content:$release}}}' > /tmp/gist.json
curl -fsS -X PATCH -H 'Accept: application/vnd.github+json' -H "Authorization: Bearer $GIST_TOKEN" -H 'X-GitHub-Api-Version: 2022-11-28' --data-binary @/tmp/gist.json https://api.github.com/gists/fb3843c3a13793eb6cc0af638bc00ad4 >/dev/null
echo "published $TAG and badge documents"`
    return this.base(source, "trusted-release")
      .withDirectory("/coverage", coverage)
      .withSecretVariable("GH_TOKEN", githubToken)
      .withSecretVariable("GIST_TOKEN", gistToken)
      .withEnvVariable("TAG", tag)
      .withEnvVariable("GOSHTOSO_RUN_NONCE", runNonce)
      .withExec(["bash", "-euo", "pipefail", "-c", script])
      .stdout()
  }

  /** Dispatch the verified main SHA to the central Fly deployment repository. */
  @func({ cache: "never" })
  async dispatchFly(metadata: File, token: Secret, runNonce: string): Promise<string> {
    const handoff = await this.json(metadata, ["source_repository", "source_run_id", "source_sha"])
    const sourceSha = this.string(handoff, "source_sha")
    const sourceRunId = this.string(handoff, "source_run_id")
    if (!/^[0-9a-f]{40}$/.test(sourceSha)) throw new Error("invalid Fly source SHA")
    if (!/^[1-9][0-9]*$/.test(sourceRunId)) throw new Error("invalid Fly source run ID")
    if (this.string(handoff, "source_repository") !== "araihu/goshtoso") throw new Error("invalid Fly source repository")
    const payload = JSON.stringify({ event_type: "goshtoso-main", client_payload: { goshtoso_ref: sourceSha, goshtoso_sha: sourceSha, goshtoso_run_id: sourceRunId, source_repository: "araihu/goshtoso" } })
    return dag.container().from(GO_IMAGE)
      .withExec(["apt-get", "update"])
      .withExec(["apt-get", "install", "-y", "--no-install-recommends", "ca-certificates", "curl"])
      .withExec(["rm", "-rf", "/var/lib/apt/lists/*"])
      .withSecretVariable("GH_TOKEN", token)
      .withNewFile("/tmp/payload.json", payload)
      .withEnvVariable("GOSHTOSO_RUN_NONCE", runNonce)
      .withExec(["bash", "-euo", "pipefail", "-c", "curl -fsS -X POST -H 'Accept: application/vnd.github+json' -H \"Authorization: Bearer $GH_TOKEN\" -H 'X-GitHub-Api-Version: 2022-11-28' --data-binary @/tmp/payload.json https://api.github.com/repos/araihu/fly-deploy/dispatches"])
      .withExec(["echo", "Fly dispatch accepted"])
      .stdout()
  }

  /** Validate immutable Assets handoff, update allowlisted files, and return the patch tree. */
  @func({ cache: "never" })
  async updateAraihuAssets(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    metadata: File,
    githubToken: Secret,
    cachePartition: string,
    runNonce: string,
  ): Promise<Directory> {
    const handoff = await this.json(metadata, ["assets_repository", "assets_revision", "release", "release_json_sha256", "release_sha256", "release_url"])
    const assetsRepository = this.string(handoff, "assets_repository")
    const assetsRevision = this.string(handoff, "assets_revision")
    const release = this.string(handoff, "release")
    const releaseUrl = this.string(handoff, "release_url")
    const releaseSha256 = this.string(handoff, "release_sha256")
    const releaseJsonSha256 = this.string(handoff, "release_json_sha256")
    if (assetsRepository !== "araihu/assets" || !/^[0-9a-f]{40}$/.test(assetsRevision) || !/^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/.test(release) || !/^[0-9a-f]{64}$/.test(releaseSha256) || !/^[0-9a-f]{64}$/.test(releaseJsonSha256) || releaseUrl !== `https://github.com/araihu/assets/releases/download/${release}/araihu-assets-${release}.tar.gz`) throw new Error("invalid immutable Assets handoff")
    const args = ["-release-dir", "/tmp/release", "-assets-repository", assetsRepository, "-assets-revision", assetsRevision, "-release", release, "-release-url", releaseUrl, "-release-sha256", releaseSha256, "-release-json-sha256", releaseJsonSha256]
    return this.goProject(source, cachePartition)
      .withSecretVariable("GH_TOKEN", githubToken)
      .withEnvVariable("ASSETS_REPOSITORY", assetsRepository).withEnvVariable("ASSETS_REVISION", assetsRevision)
      .withEnvVariable("RELEASE", release).withEnvVariable("RELEASE_URL", releaseUrl)
      .withEnvVariable("RELEASE_SHA256", releaseSha256).withEnvVariable("RELEASE_JSON_SHA256", releaseJsonSha256)
      .withExec(["bash", "-euo", "pipefail", "-c", "test \"$ASSETS_REPOSITORY\" = araihu/assets; [[ $ASSETS_REVISION =~ ^[0-9a-f]{40}$ ]]; [[ $RELEASE =~ ^v(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)$ ]]; [[ $RELEASE_SHA256 =~ ^[0-9a-f]{64}$ ]]; [[ $RELEASE_JSON_SHA256 =~ ^[0-9a-f]{64}$ ]]; test \"$RELEASE_URL\" = \"https://github.com/araihu/assets/releases/download/$RELEASE/araihu-assets-$RELEASE.tar.gz\""])
      .withEnvVariable("GOSHTOSO_RUN_NONCE", runNonce)
      .withExec(["bash", "-euo", "pipefail", "-c", "o=$(curl -fsS -H \"Authorization: Bearer $GH_TOKEN\" -H 'Accept: application/vnd.github+json' https://api.github.com/repos/araihu/assets/git/ref/tags/$RELEASE); type=$(printf '%s' \"$o\" | sed -n 's/.*\"type\": *\"\\([^\"]*\\)\".*/\\1/p'); sha=$(printf '%s' \"$o\" | sed -n 's/.*\"sha\": *\"\\([0-9a-f]*\\)\".*/\\1/p'); if test \"$type\" = tag; then o=$(curl -fsS -H \"Authorization: Bearer $GH_TOKEN\" -H 'Accept: application/vnd.github+json' https://api.github.com/repos/araihu/assets/git/tags/$sha); type=$(printf '%s' \"$o\" | sed -n 's/.*\"type\": *\"\\([^\"]*\\)\".*/\\1/p'); sha=$(printf '%s' \"$o\" | sed -n 's/.*\"sha\": *\"\\([0-9a-f]*\\)\".*/\\1/p'); fi; test \"$type\" = commit; test \"$sha\" = \"$ASSETS_REVISION\""])
      .withExec(["bash", "-euo", "pipefail", "-c", "curl -fsSL --retry 3 -o /tmp/release.tgz \"$RELEASE_URL\"; printf '%s  %s\n' \"$RELEASE_SHA256\" /tmp/release.tgz | sha256sum -c --strict; while IFS= read -r m; do case \"$m\" in /*|../*|*/../*|*/..) exit 1;; esac; done < <(tar -tzf /tmp/release.tgz); ! tar -tvzf /tmp/release.tgz | grep -Eq '^[lh]'; mkdir /tmp/release; tar -xzf /tmp/release.tgz -C /tmp/release --no-same-owner --no-same-permissions"])
      .withExec(["go", "run", "./cmd/araihu-assets-update", ...args])
      .withExec(["git", "diff", "--binary"], { redirectStdout: "/tmp/first.diff" })
      .withExec(["go", "run", "./cmd/araihu-assets-update", ...args])
      .withExec(["git", "diff", "--binary"], { redirectStdout: "/tmp/second.diff" })
      .withExec(["cmp", "/tmp/first.diff", "/tmp/second.diff"])
      .withExec(["go", "test", "./internal/araihuassets", "./cmd/araihu-assets-update", "-count=1"])
      .directory("/work")
  }

  /** Minimal runner-to-Dagger execution proof. */
  @func({ cache: "never" })
  async smoke(@argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory, identity: File, runNonce: string): Promise<string> {
    const metadata = await this.json(identity, ["revision"])
    const revision = this.string(metadata, "revision")
    if (!/^[0-9a-f]{40}$/.test(revision)) throw new Error("invalid provider Git revision")
    return dag.container().from(GO_IMAGE).withDirectory("/work", source).withWorkdir("/work")
      .withEnvVariable("GOSHTOSO_REVISION", revision)
      .withEnvVariable("GOSHTOSO_RUN_NONCE", runNonce)
      .withExec(["bash", "-euo", "pipefail", "-c", "printf '%s\\n' \"$GOSHTOSO_REVISION\"; go version; test -f dagger.json"])
      .stdout()
  }

  private base(source: Directory, partition: string): Container {
    const key = this.partition(partition)
    const cacheNamespace = this.cacheNamespace(key)
    const container = dag.container().from(GO_IMAGE)
      .withExec(["apt-get", "update"])
      .withExec(["apt-get", "install", "-y", "--no-install-recommends", "ca-certificates", "curl", "git", "gh"])
      .withExec(["rm", "-rf", "/var/lib/apt/lists/*"])
      .withFile("/usr/local/bin/jq", dag.container().from(JQ_IMAGE).file("/jq"), { permissions: 0o755 })
      .withDirectory("/work", source).withWorkdir("/work")
      .withExec(["bash", "-euo", "pipefail", "-c", "git init -q; git config user.name Dagger; git config user.email dagger@invalid; git add -A; git commit -qm source-baseline"])
    return container
      .withMountedCache("/root/.cache/go-build", dag.cacheVolume(`goshtoso-${cacheNamespace}-go-build-v1`))
      .withMountedCache("/go/pkg/mod", dag.cacheVolume(`goshtoso-${cacheNamespace}-go-mod-v1`))
      .withMountedCache("/work/.cache/muamba", dag.cacheVolume(`goshtoso-${cacheNamespace}-muamba-v1`))
  }

  private goProject(source: Directory, partition: string): Container {
    const key = this.partition(partition)
    const cacheNamespace = this.cacheNamespace(key)
    const node = dag.container().from(NODE_IMAGE)
    let project = this.base(source, key)
      .withFile("/usr/local/bin/node", node.file("/usr/local/bin/node"), { permissions: 0o755 })
      .withDirectory("/usr/local/lib/node_modules", node.directory("/usr/local/lib/node_modules"))
      .withExec(["ln", "-sf", "../lib/node_modules/npm/bin/npm-cli.js", "/usr/local/bin/npm"])
      .withExec(["ln", "-sf", "../lib/node_modules/npm/bin/npx-cli.js", "/usr/local/bin/npx"])
      .withExec(["node", "--version"])
      .withEnvVariable("GOBIN", "/tools/bin").withEnvVariable("PATH", "/tools/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
      .withExec(["mkdir", "-p", "/tools/bin"])
    if (cacheNamespace !== undefined) {
      project = project.withMountedCache("/tools", dag.cacheVolume(`goshtoso-${cacheNamespace}-go-tools-${TEMPL_VERSION}-${GOLANGCI_LINT_VERSION}-${PLAYWRIGHT_VERSION}`), { sharing: CacheSharingMode.Locked })
    }
    return project
      .withExec(["bash", "-euo", "pipefail", "-c", `test -x /tools/bin/templ || go install github.com/a-h/templ/cmd/templ@${TEMPL_VERSION}; test -x /tools/bin/golangci-lint || { case "$(uname -m)" in x86_64) arch=amd64; sha=${GOLANGCI_LINT_AMD64_SHA256};; aarch64|arm64) arch=arm64; sha=${GOLANGCI_LINT_ARM64_SHA256};; *) echo unsupported architecture >&2; exit 1;; esac; file=golangci-lint-2.12.2-linux-$arch.tar.gz; dir=\${file%.tar.gz}; curl -fsSL --retry 3 -o /tmp/golangci-lint.tgz https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}/$file; echo "$sha  /tmp/golangci-lint.tgz" | sha256sum -c -; tar -xzf /tmp/golangci-lint.tgz -C /tools/bin --strip-components=1 "$dir/golangci-lint"; }`])
  }

  private browserProject(source: Directory, partition: string): Container {
    const key = this.partition(partition)
    const cacheNamespace = this.cacheNamespace(key)
    let project = this.goProject(source, key)
      .withEnvVariable("PLAYWRIGHT_BROWSERS_PATH", "/playwright")
      .withExec(["mkdir", "-p", "/playwright"])
    if (cacheNamespace !== undefined) {
      project = project.withMountedCache("/playwright", dag.cacheVolume(`goshtoso-${cacheNamespace}-playwright-${PLAYWRIGHT_VERSION}`), { sharing: CacheSharingMode.Locked })
    }
    return project
      .withExec(["bash", "-euo", "pipefail", "-c", `test -x /tools/bin/playwright || go install github.com/mxschmitt/playwright-go/cmd/playwright@${PLAYWRIGHT_VERSION}; test -f /playwright/.chromium-ready || { playwright install --with-deps chromium; touch /playwright/.chromium-ready; }`])
  }

  private partition(value: string): string {
    if (!/^(trusted-(main|release|assets)|local(-[a-z0-9-]+)?|(untrusted|fork|internal)(-[a-z0-9-]+)?)$/.test(value)) throw new Error(`unsafe cache partition: ${value}`)
    return value
  }

  private isUntrustedPartition(value: string): boolean {
    return /^(untrusted|fork|internal)(-|$)/.test(value)
  }

  private cacheNamespace(value: string): string {
    // PR jobs now run on a proven host-owned isolated Engine. Never derive
    // that boundary from workflow input: every PR partition collapses to pr.
    if (this.isUntrustedPartition(value)) return "pr"
    return value
  }

  private async json(file: File, keys: string[]): Promise<Record<string, unknown>> {
    let value: unknown
    try { value = JSON.parse(await file.contents()) } catch { throw new Error("metadata is not valid JSON") }
    if (value === null || Array.isArray(value) || typeof value !== "object") throw new Error("metadata must be an object")
    const actual = Object.keys(value as Record<string, unknown>).sort()
    if (JSON.stringify(actual) !== JSON.stringify([...keys].sort())) throw new Error("metadata fields do not match the strict schema")
    return value as Record<string, unknown>
  }

  private string(value: Record<string, unknown>, key: string): string {
    if (typeof value[key] !== "string") throw new Error(`metadata field ${key} must be a string`)
    return value[key]
  }
}
