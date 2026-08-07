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

cd "$repo_root"
mkdir -p "$proof_root/consumer"
go run ./cmd/iconpack \
  -release-archive "$archive" \
  -archive-sha256 "$archive_sha256" \
  -release v0.2.0 \
  -catalog-sha256 a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6 \
  -release-json-sha256 0650e51dd2b7ec7797622b3cdd9ff75dfd53cb1914155931014223bbd1684fa6 \
  -checksums-sha256 86dac118901d423117e20bd14ce6ed30717fca9c2a8244909c2be4b926ce1c4e \
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
