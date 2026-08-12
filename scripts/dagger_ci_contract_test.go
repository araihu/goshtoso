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
