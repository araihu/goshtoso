import { execFileSync } from "node:child_process";
import { copyFileSync, mkdirSync, rmSync, readFileSync } from "node:fs";
import { gzipSync } from "node:zlib";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const site = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const dist = resolve(site, "dist");
rmSync(dist, { recursive: true, force: true });
mkdirSync(dist, { recursive: true });
const run = (args, options = {}) => execFileSync("go", args, { cwd: site, stdio: "inherit", ...options });
const version = process.env.GOSHTOSO_DOCS_VERSION || "development";
if (!/^(development|v\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?)$/.test(version)) {
  throw new Error("Invalid GOSHTOSO_DOCS_VERSION");
}
run(["build", "-trimpath", "-ldflags", `-s -w -X github.com/araihu/goshtoso/site/internal/buildinfo.goDocsVersion=${version}`,
  "-o", resolve(dist, "server.wasm"), "./cmd/worker"], {
  env: { ...process.env, GOOS: "js", GOARCH: "wasm", CGO_ENABLED: "0" },
});
const goroot = execFileSync("go", ["env", "GOROOT"], { cwd: site, encoding: "utf8" }).trim();
copyFileSync(resolve(goroot, "lib/wasm/wasm_exec.js"), resolve(dist, "wasm_exec.js"));
run(["run", "./cmd/worker-assets", "-out", resolve(dist, "public")]);

// Workers Free and Paid currently allow 64 MiB uncompressed, including WASM.
// Reserve 1 MiB for the JavaScript wrapper and metadata. gzip is informational.
const wasmBytes = readFileSync(resolve(dist, "server.wasm"));
const compressed = gzipSync(wasmBytes).byteLength;
if (wasmBytes.byteLength > 63 * 1024 * 1024) {
  throw new Error(`WASM too large for the 64 MiB Worker bundle limit: ${wasmBytes.byteLength} bytes`);
}
console.log(`Go WASM: ${(wasmBytes.byteLength / 1024 / 1024).toFixed(2)} MiB uncompressed, ${(compressed / 1024 / 1024).toFixed(2)} MiB gzip`);
