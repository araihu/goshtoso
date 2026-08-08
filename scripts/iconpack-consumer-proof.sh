#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 /path/to/araihu-assets-v0.2.0.tar.gz ARCHIVE_SHA256" >&2
  exit 2
fi

archive=$1
if [[ $archive != /* ]]; then
  archive=$(pwd)/$archive
fi
archive_sha256=$2
repo_root=$(git rev-parse --show-toplevel)
proof_root=$(mktemp -d /tmp/goshtoso-iconpack-consumer-proof.XXXXXX)
trap 'rm -rf "$proof_root"' EXIT

consumer_root=$proof_root/consumer
packs_root=$consumer_root/internal/appicons
mkdir -p "$packs_root" "$consumer_root/cmd/proof"
printf 'consumer-owned byte\n' > "$packs_root/owner.txt"
printf 'consumer-owned byte\n' > "$proof_root/owner.expected"

common_generator_args=(
  -release-archive "$archive"
  -archive-sha256 "$archive_sha256"
  -release v0.2.0
  -catalog-sha256 a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6
  -release-json-sha256 77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260
  -checksums-sha256 334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5
  -package appicons
  -const-prefix Icon
)

run_generator() {
  (
    cd "$repo_root"
    go run ./cmd/iconpack "${common_generator_args[@]}" "$@"
  )
}

assert_lock() {
  local lock_path=$1
  local mode
  if [[ $(uname -s) == Darwin ]]; then
    mode=$(stat -f '%Lp' "$lock_path")
  else
    mode=$(stat -c '%a' "$lock_path")
  fi
  if [[ $mode != 600 ]]; then
    echo "coordination file $lock_path has mode $mode, want 600" >&2
    exit 1
  fi
}

run_generator \
  -name brand-developer-icons-tRPC \
  -out "$packs_root/v1" \
  -sprite-url /assets/icons/app-v1.svg
run_generator \
  -name brand-developer-icons-tRPC \
  -out "$packs_root/v1" \
  -sprite-url /assets/icons/app-v1.svg \
  -check
assert_lock "$packs_root/.v1.goshtoso-iconpack.lock"
cp -R "$packs_root/v1" "$proof_root/v1.expected"

cd "$consumer_root"
go mod init example.com/goshtoso-iconpack-consumer
go mod edit -require=github.com/araihu/goshtoso@v0.0.0
go mod edit -replace="github.com/araihu/goshtoso=$repo_root"
cat > cmd/proof/main.go <<'EOF'
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	appicons "example.com/goshtoso-iconpack-consumer/internal/appicons/v1"
)

const generatedSpritePath = "internal/appicons/v1/sprite.svg"
const generatedSpriteURL = "/assets/icons/app-v1.svg"

func main() {
	sprite, err := os.ReadFile(generatedSpritePath)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+generatedSpriteURL, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(sprite)
	})
	request := httptest.NewRequest(http.MethodGet, generatedSpriteURL, nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), `id="devicon-trpc"`) {
		panic("v1 sprite route does not serve the selected symbol")
	}

	var output bytes.Buffer
	component := appicons.Icon(appicons.Config{
		Symbol: appicons.IconBrandDeveloperIconsTRPC,
		Label:  "tRPC",
	})
	if err := component.Render(context.Background(), &output); err != nil {
		panic(err)
	}
	if _, ok := appicons.Lookup(appicons.NameBrandDeveloperIconsTRPC); !ok {
		panic("literal canonical-name binding missing")
	}
	fmt.Print(output.String())
}
EOF
go mod tidy
go test ./...
rendered=$(go run ./cmd/proof)
case "$rendered" in
  *'/assets/icons/app-v1.svg#devicon-trpc'*) ;;
  *) echo "v1 consumer render did not preserve sprite symbol" >&2; exit 1 ;;
esac
case "$rendered" in
  *'aria-label="tRPC"'*) ;;
  *) echo "v1 consumer render did not preserve accessible label" >&2; exit 1 ;;
esac

run_generator \
  -name ui-hi-16-solid-check \
  -out "$packs_root/v2" \
  -sprite-url /assets/icons/app-v2.svg
run_generator \
  -name ui-hi-16-solid-check \
  -out "$packs_root/v2" \
  -sprite-url /assets/icons/app-v2.svg \
  -check
assert_lock "$packs_root/.v2.goshtoso-iconpack.lock"

cat > cmd/proof/main.go <<'EOF'
package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	appicons "example.com/goshtoso-iconpack-consumer/internal/appicons/v2"
)

const generatedSpritePath = "internal/appicons/v2/sprite.svg"
const generatedSpriteURL = "/assets/icons/app-v2.svg"

func main() {
	sprite, err := os.ReadFile(generatedSpritePath)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+generatedSpriteURL, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(sprite)
	})
	request := httptest.NewRequest(http.MethodGet, generatedSpriteURL, nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), `id="hi-16-solid-check"`) {
		panic("v2 sprite route does not serve the selected symbol")
	}

	var output bytes.Buffer
	component := appicons.Icon(appicons.Config{
		Symbol: appicons.IconUiHi16SolidCheck,
		Label:  "Done",
	})
	if err := component.Render(context.Background(), &output); err != nil {
		panic(err)
	}
	if _, ok := appicons.Lookup(appicons.NameUiHi16SolidCheck); !ok {
		panic("updated canonical-name binding missing")
	}
	fmt.Print(output.String())
}
EOF

go test ./...
rendered=$(go run ./cmd/proof)
case "$rendered" in
  *'/assets/icons/app-v2.svg#hi-16-solid-check'*) ;;
  *) echo "v2 consumer render did not use the migrated sprite URL and symbol" >&2; exit 1 ;;
esac
case "$rendered" in
  *'aria-label="Done"'*) ;;
  *) echo "v2 consumer render did not preserve accessible label" >&2; exit 1 ;;
esac

if ! diff -r "$packs_root/v1" "$proof_root/v1.expected" >/dev/null; then
  echo "v1 rollback pack changed during migration" >&2
  exit 1
fi
if ! cmp -s "$packs_root/owner.txt" "$proof_root/owner.expected"; then
  echo "migration changed an unrelated consumer-owned byte" >&2
  exit 1
fi

actual_children=$(find "$packs_root" -mindepth 1 -maxdepth 1 -exec basename {} \; | LC_ALL=C sort)
expected_children=$(printf '%s\n' \
  .v1.goshtoso-iconpack.lock \
  .v2.goshtoso-iconpack.lock \
  owner.txt \
  v1 \
  v2)
if [[ $actual_children != "$expected_children" ]]; then
  echo "migration left undeclared persistent files:" >&2
  printf '%s\n' "$actual_children" >&2
  exit 1
fi

echo "iconpack consumer update proof: PASS"
