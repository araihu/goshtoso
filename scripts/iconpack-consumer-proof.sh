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
expected_archive_sha256=5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5
if [[ $archive_sha256 != "$expected_archive_sha256" ]]; then
  echo "archive SHA-256 is not the published Arai Hu Assets v0.2.0 boundary" >&2
  exit 1
fi
repo_root=$(git rev-parse --show-toplevel)
proof_root=$(mktemp -d /tmp/goshtoso-iconpack-consumer-proof.XXXXXX)
trap 'rm -rf "$proof_root"' EXIT

cd "$repo_root"
mkdir -p "$proof_root/consumer"
# Frozen from exact Assets v0.2.0 archive bytes with outer SHA-256
# 5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5.
# release.json SHA-256: 77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260
# checksums.txt SHA-256: 334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5
go run ./cmd/iconpack \
  -release-archive "$archive" \
  -archive-sha256 "$archive_sha256" \
  -release v0.2.0 \
  -catalog-sha256 a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6 \
  -release-json-sha256 77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260 \
  -checksums-sha256 334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5 \
  -name brand-developer-icons-tRPC \
  -name ui-hi-16-solid-check \
  -out "$proof_root/consumer/appicons" \
  -package appicons \
  -const-prefix Icon \
  -sprite-url /assets/icons/app.svg

mkdir -p "$proof_root/consumer/cmd/proof"
cd "$proof_root/consumer"
go mod init example.com/goshtoso-iconpack-consumer
go mod edit -require=github.com/araihu/goshtoso@v0.0.0
go mod edit -replace="github.com/araihu/goshtoso=$repo_root"
cat > cmd/proof/main.go <<'EOF'
package main

import (
	"bytes"
	"context"
	"fmt"

	"example.com/goshtoso-iconpack-consumer/appicons"
)

func main() {
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
  *'/assets/icons/app.svg#devicon-trpc'*) ;;
  *) echo "consumer render did not preserve sprite symbol" >&2; exit 1 ;;
esac
case "$rendered" in
  *'aria-label="tRPC"'*) ;;
  *) echo "consumer render did not preserve accessible label" >&2; exit 1 ;;
esac

echo "iconpack consumer proof: PASS"
