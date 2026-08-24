package scripts_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDaggerWorkflowSecurityAndArtifactContracts(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		t.Helper()
		data, readErr := os.ReadFile(filepath.Join(root, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		return string(data)
	}

	all := read(".github/workflows/araihu-assets.yml") + read(".github/workflows/ci.yml") +
		read(".github/workflows/deploy.yml") + read(".github/workflows/docs.yml") +
		read(".github/workflows/release.yml") + read(".github/workflows/required.yml") +
		read(".github/workflows/runner-smoke-test.yml")
	if strings.Contains(strings.ToLower(all), "coderabbit") {
		t.Fatal("CodeRabbit must be absent from authored pipelines")
	}
	unsafeArg := regexp.MustCompile(`args:.*\$\{\{[^}]*(client_payload|workflow_run|ref_name)`) // external strings must cross a file boundary
	if unsafeArg.MatchString(all) {
		t.Fatal("external event data is interpolated into dagger-for-github args")
	}

	for _, expected := range []string{"runner.environment == 'self-hosted'", "dagger version | awk", "v0.21.8"} {
		if !strings.Contains(all, expected) {
			t.Fatalf("missing embedded-runner contract %q", expected)
		}
	}
	for _, expected := range []string{".dagger-output/status", ".dagger-output/tests.log", ".coverage/status", ".coverage/release.log"} {
		if !strings.Contains(all, expected) {
			t.Fatalf("missing failure-artifact contract %q", expected)
		}
	}

	module := read(".dagger/src/index.ts")
	for _, expected := range []string{"LYCHEE_X86_64_SHA256", "LYCHEE_AARCH64_SHA256", "expect: ReturnType.Any", "private async json", "ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486", "node:24.19.0-bookworm-slim@sha256:3638d9a6fe4030bd716be989438248074489337ba3275657f93595428be4fc03"} {
		if !strings.Contains(module, expected) {
			t.Fatalf("missing Dagger module contract %q", expected)
		}
	}
	if strings.Contains(module, `"gh", "jq"`) {
		t.Fatal("jq must come from the pinned image, not the mutable apt repository")
	}
	if strings.Contains(module, `"npm", "audit", "--prefix", ".dagger"`) || strings.Contains(module, `"npm", "ls"`) {
		t.Fatal("Dagger module must not claim its skeletal local-SDK lock proves the generated runtime graph")
	}
	for _, authored := range []string{read("dagger.json"), module} {
		if strings.Contains(strings.ToLower(authored), "coderabbit") {
			t.Fatal("CodeRabbit must be absent from the authored Dagger module")
		}
	}
}

func TestReleaseCoverageHandoffIncludesHiddenFiles(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/workflows/release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	release := string(data)
	if count := strings.Count(release, "include-hidden-files: true"); count != 2 {
		t.Fatalf("release hidden-artifact inclusions = %d, want 2", count)
	}
	for _, artifact := range []string{"release-e2e-results", "release-coverage"} {
		name := strings.Index(release, "name: "+artifact)
		if name < 0 {
			t.Fatalf("release workflow missing artifact %q", artifact)
		}
		block := release[name:]
		nextStep := strings.Index(block, "\n      - name:")
		if nextStep >= 0 {
			block = block[:nextStep]
		}
		if !strings.Contains(block, "include-hidden-files: true") {
			t.Fatalf("release artifact %q excludes hidden files", artifact)
		}
	}
}

func TestAssetsHandoffSurvivesCheckoutAndPrecedesWriteToken(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/workflows/araihu-assets.yml"))
	if err != nil {
		t.Fatal(err)
	}
	assets := string(data)
	checkout := strings.Index(assets, "Check out Goshtoso read-only")
	handoff := strings.Index(assets, "Validate and materialize immutable handoff")
	token := strings.Index(assets, "Create selected-repository App token")
	daggerCall := strings.Index(assets, "Validate and update fallback assets through Dagger")
	if checkout < 0 || handoff <= checkout || token <= handoff || daggerCall <= token {
		t.Fatalf("Assets provider boundary/order is unsafe: checkout=%d handoff=%d token=%d dagger=%d", checkout, handoff, token, daggerCall)
	}
	if strings.Count(assets, "actions/checkout@") != 1 || !strings.Contains(assets[checkout:handoff], "persist-credentials: false") {
		t.Fatal("Assets workflow must materialize the handoff after one read-only checkout")
	}
}

func TestDaggerBootstrapAuditPrecedesRealRuntimeSmoke(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/workflows/runner-smoke-test.yml"))
	if err != nil {
		t.Fatal(err)
	}
	smoke := string(data)
	bootstrapAudit := strings.Index(smoke, "Audit committed Dagger bootstrap lock")
	runtimeProof := strings.Index(smoke, "Runner-to-Dagger smoke proof")
	if bootstrapAudit < 0 || runtimeProof <= bootstrapAudit || !strings.Contains(smoke, "npm audit --prefix .dagger --package-lock-only --omit=dev --audit-level=high") {
		t.Fatal("runner smoke must audit only the committed bootstrap lock before the real generated-runtime Dagger proof")
	}
}

func TestDaggerGeneratedTypescriptPackageContract(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	pkg := readJSON(t, root, ".dagger/package.json")
	dependencies, ok := pkg["dependencies"].(map[string]any)
	if !ok {
		t.Fatal("Dagger package dependencies missing")
	}
	if dependencies["@dagger.io/dagger"] != "./sdk" {
		t.Fatal("Dagger TypeScript package must consume the Engine-generated local SDK")
	}
	if dependencies["typescript"] != "^6.0.3" {
		t.Fatal("TypeScript must be a runtime dependency compatible with the generated SDK")
	}
	if _, exists := pkg["devDependencies"]; exists {
		t.Fatal("Dagger runtime dependencies must not be hidden in devDependencies")
	}
	if _, exists := pkg["overrides"]; exists {
		t.Fatal("committed bootstrap package must not override the Engine-generated SDK graph")
	}
}

func TestDaggerGeneratedTypescriptLockContract(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	lock := readJSON(t, root, ".dagger/package-lock.json")
	packages, ok := lock["packages"].(map[string]any)
	if !ok {
		t.Fatal("Dagger lock packages missing")
	}
	rootPackage, ok := packages[""].(map[string]any)
	if !ok {
		t.Fatal("Dagger lock root package missing")
	}
	lockedDependencies, ok := rootPackage["dependencies"].(map[string]any)
	if !ok || lockedDependencies["@dagger.io/dagger"] != "./sdk" || lockedDependencies["typescript"] != "^6.0.3" {
		t.Fatal("Dagger lock must preserve the local generated SDK and runtime TypeScript contract")
	}
	if _, exists := rootPackage["overrides"]; exists {
		t.Fatal("committed bootstrap lock must not encode overrides for the generated SDK graph")
	}
	linkedSDK, ok := packages["node_modules/@dagger.io/dagger"].(map[string]any)
	if !ok || linkedSDK["resolved"] != "sdk" || linkedSDK["link"] != true {
		t.Fatal("Dagger lock must link node_modules/@dagger.io/dagger to sdk")
	}
}

func TestDaggerGeneratedSDKIsRuntimeOnly(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	tracked := exec.Command("git", "ls-files", "--", ".dagger/sdk")
	tracked.Dir = root
	output, err := tracked.Output()
	if err != nil {
		t.Fatalf("inspect tracked generated SDK: %v", err)
	}
	if strings.TrimSpace(string(output)) != "" {
		t.Fatal("generated Dagger SDK bytes must not be committed")
	}
	moduleData, err := os.ReadFile(filepath.Join(root, ".dagger/src/index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	module := string(moduleData)
	for _, excluded := range []string{`".dagger/sdk"`, `".dagger/sdk/**"`} {
		if !strings.Contains(module, excluded) {
			t.Fatalf("generated Dagger SDK must be excluded from source snapshots: %s", excluded)
		}
	}
	workflowData, err := os.ReadFile(filepath.Join(root, ".github/workflows/runner-smoke-test.yml"))
	if err != nil {
		t.Fatal(err)
	}
	authored := strings.ToLower(module + "\n" + string(workflowData))
	for _, forbidden := range []string{"npm audit --prefix .dagger/sdk", `"npm", "audit", "--prefix", ".dagger/sdk"`, "npm ls --prefix .dagger/sdk", `"npm", "ls", "--prefix", ".dagger/sdk"`} {
		if strings.Contains(authored, forbidden) {
			t.Fatalf("authored CI falsely audits the Engine-generated SDK graph: %s", forbidden)
		}
	}
}

func readJSON(t *testing.T, root, name string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return value
}

func TestPullRequestsUseHostIsolatedPersistentDaggerCaches(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		t.Helper()
		data, readErr := os.ReadFile(filepath.Join(root, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		return string(data)
	}

	for _, workflow := range []string{".github/workflows/ci.yml", ".github/workflows/docs.yml", ".github/workflows/required.yml"} {
		content := read(workflow)
		if strings.Contains(content, "'trusted-hostinger' || 'untrusted-github'") {
			t.Fatalf("%s lets an internal PR select a persistent cache domain", workflow)
		}
		if !strings.Contains(content, "github.event_name == 'push' && 'trusted-main' ||\n") || !strings.Contains(content, "'untrusted-github'") {
			t.Fatalf("%s does not map every PR to the untrusted provider domain", workflow)
		}
		if !strings.Contains(content, `"hostinger-vps-pr"`) || strings.Contains(content, `"hostinger-vps"`) {
			t.Fatalf("%s does not select the exact host-owned PR lane", workflow)
		}
	}

	module := read(".dagger/src/index.ts")
	if !strings.Contains(module, `/^(untrusted|fork|internal)(-|$)/`) {
		t.Fatal("Dagger module must classify fork and internal PR domains as untrusted")
	}
	for _, bounds := range [][2]string{
		{"private base(", "private goProject("},
		{"private goProject(", "private browserProject("},
		{"private browserProject(", "private partition("},
	} {
		start := strings.Index(module, bounds[0])
		end := strings.Index(module, bounds[1])
		if start < 0 || end <= start {
			t.Fatalf("cannot locate cache builder %q", bounds[0])
		}
		body := module[start:end]
		guard := strings.Index(body, "cacheNamespace")
		mount := strings.Index(body, "withMountedCache")
		if guard < 0 || mount < 0 || guard > mount {
			t.Fatalf("%s must guard before its first persistent CacheVolume mount", bounds[0])
		}
	}
	if strings.Count(module, "withMountedCache") < 5 {
		t.Fatal("trusted main, release, assets, and local domains must retain effective caches")
	}
	if !strings.Contains(module, "if (this.isUntrustedPartition(value)) return \"pr\"") {
		t.Fatal("PR cache namespace must stay centralized as constant pr behind the host-owned Engine boundary")
	}
}

func TestDaggerTestsAcquireTailwindBeforeCSSBuild(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".dagger/src/index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	module := string(data)
	start := strings.Index(module, "  tests(")
	end := strings.Index(module, "  /** Standalone site/go.mod consumer contract")
	if start < 0 || end <= start {
		t.Fatal("cannot locate Dagger tests pipeline")
	}
	body := module[start:end]
	acquire := strings.Index(body, "go tool muamba sync --strict --target linux/amd64 --cache-dir .cache/muamba tailwindcss/cli")
	build := strings.Index(body, ".tools/tailwindcss -i css/main.css -o assets/styles.css")
	if acquire < 0 || build <= acquire {
		t.Fatalf("Dagger tests must acquire locked Tailwind before CSS build: acquire=%d build=%d", acquire, build)
	}
}

func TestDaggerDocsUsesPinnedStaticLycheeArchive(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".dagger/src/index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	module := string(data)
	start := strings.Index(module, "  async docs(")
	end := strings.Index(module, "  /** Full release-equivalent regeneration")
	if start < 0 || end <= start {
		t.Fatal("cannot locate Dagger docs pipeline")
	}
	body := module[start:end]
	for _, expected := range []string{
		`const LYCHEE_VERSION = "v0.24.2"`,
		`const LYCHEE_X86_64_SHA256 = "73657a111819a30c47c08352896796f23d64e4eb2b3ed39b6d32149241566fc5"`,
		`const LYCHEE_AARCH64_SHA256 = "5d0b0e3aeab240f41920c633a6eaf97599be6eedda034b36e858ede7dba5e535"`,
		"file=lychee-$arch-unknown-linux-musl.tar.gz",
		"dir=lychee-$arch-unknown-linux-musl",
		`releases/download/lychee-${LYCHEE_VERSION}/$file`,
		`--strip-components=1 "$dir/lychee"`,
	} {
		if !strings.Contains(module, expected) {
			t.Fatalf("Dagger docs must use the pinned static lychee release contract %q", expected)
		}
	}
	if strings.Contains(body, "unknown-linux-gnu") {
		t.Fatal("Dagger docs must not use the glibc-dependent lychee archive")
	}
}

func TestDaggerSerializesSharedToolCacheWriters(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".dagger/src/index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	module := string(data)
	for _, expected := range []string{
		"CacheSharingMode,",
		`withMountedCache("/tools", dag.cacheVolume(`,
		`withMountedCache("/playwright", dag.cacheVolume(`,
	} {
		if !strings.Contains(module, expected) {
			t.Fatalf("missing shared tool cache contract %q", expected)
		}
	}
	if strings.Count(module, "{ sharing: CacheSharingMode.Locked }") != 2 {
		t.Fatal("tool and Playwright caches must both serialize check-then-install writers")
	}
}

func TestDaggerCachesMatchingPlaywrightDriverAndBrowsers(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".dagger/src/index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	module := string(data)
	for _, expected := range []string{
		`const PLAYWRIGHT_VERSION = "v0.6201.0"`,
		`const PLAYWRIGHT_DRIVER_VERSION = "1.62.1"`,
		`.withEnvVariable("PLAYWRIGHT_DRIVER_PATH", "/playwright/driver")`,
		`ready=/playwright/.chromium-${PLAYWRIGHT_DRIVER_VERSION}-ready`,
		`! test -f "$PLAYWRIGHT_DRIVER_PATH/package/cli.js"`,
		`! test -x "$PLAYWRIGHT_DRIVER_PATH/node"`,
		`rm -rf -- "$PLAYWRIGHT_DRIVER_PATH"; rm -f -- "$ready"`,
		`playwright install-deps chromium; ready=/playwright/.chromium-${PLAYWRIGHT_DRIVER_VERSION}-ready`,
		`playwright install chromium; test -f "$PLAYWRIGHT_DRIVER_PATH/package/cli.js"; test -x "$PLAYWRIGHT_DRIVER_PATH/node"`,
		`marker_tmp="$ready.tmp.$$"; : > "$marker_tmp"; mv -f -- "$marker_tmp" "$ready"`,
		`goshtoso-${cacheNamespace}-playwright-${PLAYWRIGHT_VERSION}`,
	} {
		if !strings.Contains(module, expected) {
			t.Fatalf("missing Playwright driver/browser cache contract %q", expected)
		}
	}
}

func TestPlaywrightWarmMarkerWithoutDriverReinstalls(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	driver := filepath.Join(root, "playwright", "driver")
	ready := filepath.Join(root, "playwright", ".chromium-1.62.1-ready")
	installLog := filepath.Join(root, "install.log")
	depsLog := filepath.Join(root, "deps.log")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(ready), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(driver, "package"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(driver, "package", "cli.js"), []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ready, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	fake := `#!/bin/sh
set -eu
if test "$*" = "install-deps chromium"; then
  : > "$PLAYWRIGHT_DEPS_LOG"
  exit 0
fi
test "$*" = "install chromium"
test ! -e "$PLAYWRIGHT_DRIVER_PATH"
test ! -e "$PLAYWRIGHT_READY"
mkdir -p "$PLAYWRIGHT_DRIVER_PATH/package"
: > "$PLAYWRIGHT_DRIVER_PATH/package/cli.js"
printf '#!/bin/sh\n' > "$PLAYWRIGHT_DRIVER_PATH/node"
chmod +x "$PLAYWRIGHT_DRIVER_PATH/node"
: > "$PLAYWRIGHT_INSTALL_LOG"
`
	if err := os.WriteFile(filepath.Join(bin, "playwright"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}

	script := `set -euo pipefail
playwright install-deps chromium
ready="$PLAYWRIGHT_READY"
if ! test -f "$ready" || ! test -f "$PLAYWRIGHT_DRIVER_PATH/package/cli.js" || ! test -x "$PLAYWRIGHT_DRIVER_PATH/node"; then
  rm -rf -- "$PLAYWRIGHT_DRIVER_PATH"
  rm -f -- "$ready"
  playwright install chromium
  test -f "$PLAYWRIGHT_DRIVER_PATH/package/cli.js"
  test -x "$PLAYWRIGHT_DRIVER_PATH/node"
  marker_tmp="$ready.tmp.$$"
  : > "$marker_tmp"
  mv -f -- "$marker_tmp" "$ready"
fi`
	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(),
		"PATH="+bin+":"+os.Getenv("PATH"),
		"PLAYWRIGHT_DRIVER_PATH="+driver,
		"PLAYWRIGHT_READY="+ready,
		"PLAYWRIGHT_INSTALL_LOG="+installLog,
		"PLAYWRIGHT_DEPS_LOG="+depsLog,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("warm marker recovery failed: %v\n%s", err, output)
	}
	for _, path := range []string{depsLog, installLog, filepath.Join(driver, "package", "cli.js"), filepath.Join(driver, "node"), ready} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected recovered Playwright artifact %s: %v", path, err)
		}
	}
}

func TestDaggerInstallsVerifiedGolangCILintArchives(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".dagger/src/index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	module := string(data)
	if strings.Contains(module, "raw.githubusercontent.com/golangci/golangci-lint") || strings.Contains(module, "install.sh | sh") {
		t.Fatal("golangci-lint must not execute a remote installer script")
	}
	for _, expected := range []string{
		`const GOLANGCI_LINT_AMD64_SHA256 = "b17bfbc9d4aaa48be7f4f1ce3240bc3d8200c870c072bacf15c26219e2cfb9cc"`,
		`const GOLANGCI_LINT_ARM64_SHA256 = "908317c23db18448f924e853b3d8a659fd919614cd438f224810a4053daa2607"`,
		`file=golangci-lint-2.13.1-linux-$arch.tar.gz`,
		`dir=\${file%.tar.gz}`,
		`echo "$sha  /tmp/golangci-lint.tgz" | sha256sum -c -`,
		`--strip-components=1 "$dir/golangci-lint"`,
	} {
		if !strings.Contains(module, expected) {
			t.Fatalf("missing verified golangci-lint archive contract %q", expected)
		}
	}
}

func TestLocalE2EChangeProviderPreservesCleanCommittedRange(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	provider := filepath.Join(root, "scripts/materialize-e2e-changes")
	repo := newGitRepo(t)
	base := strings.TrimSpace(runGitOutput(t, repo, "rev-parse", "HEAD"))
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("committed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "tracked.txt")
	runGit(t, repo, "commit", "-qm", "change")
	output := filepath.Join(repo, ".e2e-changes")
	command := exec.Command(provider, base, "HEAD", output)
	command.Dir = repo
	if combined, err := command.CombinedOutput(); err != nil {
		t.Fatalf("materialize clean E2E changes: %v: %s", err, combined)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "M\x00tracked.txt\x00" {
		t.Fatalf("clean committed range changed bytes: %q", data)
	}
}

func TestAssetsWorkflowPreservesDataFileModes(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/workflows/araihu-assets.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	if !strings.Contains(workflow, `install -D -m 0644 ".dagger-assets/$path" "$path"`) {
		t.Fatal("asset materialization must preserve non-executable 0644 modes")
	}
}

func TestLocalE2EChangeProviderForcesFullForDirtySource(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	provider := filepath.Join(root, "scripts/materialize-e2e-changes")

	for _, test := range []struct {
		name  string
		dirty func(t *testing.T, repo string)
	}{
		{name: "unstaged", dirty: func(t *testing.T, repo string) {
			t.Helper()
			if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("unstaged\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "staged", dirty: func(t *testing.T, repo string) {
			t.Helper()
			if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("staged\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			runGit(t, repo, "add", "tracked.txt")
		}},
		{name: "untracked", dirty: func(t *testing.T, repo string) {
			t.Helper()
			if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("untracked\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newGitRepo(t)
			test.dirty(t, repo)
			output := filepath.Join(repo, ".e2e-changes")
			command := exec.Command(provider, "HEAD", "HEAD", output)
			command.Dir = repo
			if combined, err := command.CombinedOutput(); err != nil {
				t.Fatalf("materialize E2E changes: %v: %s", err, combined)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "M\x00.dagger-dirty-worktree\x00" {
				t.Fatalf("dirty source did not force full E2E selection: %q", data)
			}
		})
	}
}

func newGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.name", "Dagger contract")
	runGit(t, repo, "config", "user.email", "dagger-contract@invalid")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("clean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "tracked.txt")
	runGit(t, repo, "commit", "-qm", "baseline")
	return repo
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = repo
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
}

func runGitOutput(t *testing.T, repo string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = repo
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
