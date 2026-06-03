#!/bin/bash
set -euo pipefail

# Only run in remote (Claude Code on the web) environments
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Go version required by the project
GO_VERSION="1.26.1"
REQUIRED_MAJOR=1
REQUIRED_MINOR=26

# Check if Go upgrade is needed (use GOTOOLCHAIN=local to prevent auto-download)
CURRENT_GO_VERSION=$(GOTOOLCHAIN=local go version 2>/dev/null | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
CURRENT_MAJOR=$(echo "$CURRENT_GO_VERSION" | cut -d. -f1)
CURRENT_MINOR=$(echo "$CURRENT_GO_VERSION" | cut -d. -f2)

if [ "$CURRENT_MAJOR" -lt "$REQUIRED_MAJOR" ] || { [ "$CURRENT_MAJOR" -eq "$REQUIRED_MAJOR" ] && [ "$CURRENT_MINOR" -lt "$REQUIRED_MINOR" ]; }; then
  echo "==> Installing Go ${GO_VERSION} (current: ${CURRENT_GO_VERSION})..."

  # Try go.dev/dl first, fall back to proxy.golang.org toolchain module
  if curl -sfL --connect-timeout 10 "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz 2>/dev/null; then
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
  else
    echo "==> go.dev unreachable, downloading from proxy.golang.org..."
    TOOLCHAIN_URL="https://proxy.golang.org/golang.org/toolchain/@v/v0.0.1-go${GO_VERSION}.linux-amd64.zip"
    curl -sfL "$TOOLCHAIN_URL" -o /tmp/go-toolchain.zip
    rm -rf /usr/local/go
    python3 -c "
import zipfile, os, shutil
prefix = 'golang.org/toolchain@v0.0.1-go${GO_VERSION}.linux-amd64/'
dest = '/usr/local/go'
os.makedirs(dest, exist_ok=True)
with zipfile.ZipFile('/tmp/go-toolchain.zip', 'r') as z:
    for info in z.infolist():
        if not info.filename.startswith(prefix):
            continue
        relpath = info.filename[len(prefix):]
        if not relpath:
            continue
        target = os.path.join(dest, relpath)
        if info.is_dir():
            os.makedirs(target, exist_ok=True)
        else:
            os.makedirs(os.path.dirname(target), exist_ok=True)
            with z.open(info) as src, open(target, 'wb') as dst:
                shutil.copyfileobj(src, dst)
            if '/bin/' in relpath or relpath.startswith('bin/'):
                os.chmod(target, 0o755)
    "
    # Ensure all tool binaries are executable
    chmod -R +x /usr/local/go/bin/ /usr/local/go/pkg/tool/*/
    rm /tmp/go-toolchain.zip
  fi

  export PATH="/usr/local/go/bin:$PATH"
  echo "==> Go installed: $(go version)"
fi

echo "==> Installing Go module dependencies..."
cd "$PROJECT_DIR"
go mod download

echo "==> Installing templ CLI..."
go install github.com/a-h/templ/cmd/templ@latest

echo "==> Installing Tailwind CSS CLI (standalone)..."
if ! command -v tailwindcss &>/dev/null; then
  curl -sL "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.4/tailwindcss-linux-x64" -o /usr/local/bin/tailwindcss
  chmod +x /usr/local/bin/tailwindcss
fi

echo "==> Installing Playwright driver and Chromium..."
export PATH="$PATH:$(go env GOPATH)/bin"
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install --with-deps chromium 2>/dev/null || \
  echo "==> WARNING: Playwright install failed (CDN may be blocked). E2E tests will not run."

echo "==> Generating templ files..."
cd "$PROJECT_DIR"
templ generate

echo "==> Building CSS..."
tailwindcss -i css/main.css -o assets/styles.css

echo "==> Persisting PATH for session..."
echo "export PATH=\"/usr/local/go/bin:\$PATH:$(go env GOPATH)/bin\"" >> "$CLAUDE_ENV_FILE"

echo "==> Session start hook completed successfully."
