#!/usr/bin/env bash
set -euo pipefail

if [[ $# -eq 2 ]]; then
  release=v0.2.0
  archive=$1
  archive_sha256=$2
elif [[ $# -eq 3 ]]; then
  release=$1
  archive=$2
  archive_sha256=$3
else
  echo "usage: $0 [v0.2.0|v0.2.1] /path/to/araihu-assets-RELEASE.{tar.gz,zip} ARCHIVE_SHA256" >&2
  exit 2
fi

if [[ $archive != /* ]]; then
  archive=$(pwd)/$archive
fi
case "$release" in
  v0.2.0)
    catalog_sha256=a0e8e5c8928e37de979ce9a60f3d66fad1aa1b4c7d2904f9275f0be9932a33d6
    release_json_sha256=77c696ae5eceb5e7bc11d19affb7c2c7b7e8afc6414882b9b059239e315f2260
    checksums_sha256=334005c77622250a1e827b9472161cd6e56c82d487fc0d44023d49261f8dbee5
    tar_sha256=5d7d691e22d4071507b0bf2248713d7008adf57c18840cfd46e20901db0b78e5
    zip_sha256=881094d3d161b79904fcfad320c26d947c9a1e526ee0b69ce8a2d04c3ff4b1b0
    ;;
  v0.2.1)
    catalog_sha256=b2b3ab2ac7e87e2eb333725195c394ebcfb1edc5f89542d89d375f2675a2aead
    release_json_sha256=1e071ba6d88efa862b6166820bdc759c7edb917c8566ce7111358c5c3dc2714e
    checksums_sha256=05cf07d924827f1eb306323a3bae5591b2a8d3b6255211c354952c0ac8dc190f
    tar_sha256=818a32246c040871c8f28bb085269b6b9f21c579b18dc4c3c1f20d70716eaf70
    zip_sha256=700212506c8c3a44c877d10a7a6696d73c561b321724aecec8e6ee51c4cdb099
    ;;
  *) echo "unsupported trusted Assets release: $release" >&2; exit 1 ;;
esac
case "$archive" in
  *.tar.gz|*.tgz) expected_archive_sha256=$tar_sha256; archive_kind=tar ;;
  *.zip) expected_archive_sha256=$zip_sha256; archive_kind=zip ;;
  *) echo "unsupported trusted archive kind" >&2; exit 1 ;;
esac
if [[ $archive_sha256 != "$expected_archive_sha256" ]]; then
  echo "archive SHA-256 is not the trusted Arai Hû Assets $release boundary" >&2
  exit 1
fi
repo_root=$(git rev-parse --show-toplevel)
proof_root=$(mktemp -d /tmp/goshtoso-iconpack-consumer-proof.XXXXXX)
trap 'rm -rf "$proof_root"' EXIT

cd "$repo_root"
mkdir -p "$proof_root/consumer"
# Extract only the authenticated catalog for an independent generator drift
# check. The source archive is never modified.
if [[ $archive_kind == tar ]]; then
  tar -xOf "$archive" catalog.json > "$proof_root/catalog.json"
else
  unzip -p "$archive" catalog.json > "$proof_root/catalog.json"
fi
if [[ $(shasum -a 256 "$proof_root/catalog.json" | awk '{print $1}') != "$catalog_sha256" ]]; then
  echo "extracted catalog SHA-256 does not match the trusted $release boundary" >&2
  exit 1
fi
mkdir -p "$proof_root/catalogbindings"
GOWORK=off go run ./cmd/iconcatalog \
  -catalog "$proof_root/catalog.json" -namespace brand -product developer-icons \
  -out "$proof_root/catalogbindings/icons_gen.go" -package catalogbindings \
  -const-prefix Icon -sprite-url /assets/icons/app.svg
GOWORK=off go run ./cmd/iconcatalog \
  -catalog "$proof_root/catalog.json" -namespace brand -product developer-icons \
  -out "$proof_root/catalogbindings/icons_gen.go" -package catalogbindings \
  -const-prefix Icon -sprite-url /assets/icons/app.svg -check

GOWORK=off go run ./cmd/iconpack \
  -release-archive "$archive" \
  -archive-sha256 "$archive_sha256" \
  -release "$release" \
  -catalog-sha256 "$catalog_sha256" \
  -release-json-sha256 "$release_json_sha256" \
  -checksums-sha256 "$checksums_sha256" \
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
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	"example.com/goshtoso-iconpack-consumer/appicons"
)

func main() {
	sprite, err := os.ReadFile("appicons/sprite.svg")
	if err != nil {
		panic(err)
	}
	page := renderPage()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/icons/app.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write(sprite)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	verifyHTTP(server.URL, sprite)
	if _, ok := appicons.Lookup(appicons.NameBrandDeveloperIconsTRPC); !ok {
		panic("literal canonical-name binding missing")
	}
	fmt.Println("real HTTP consumer proof: PASS")
}

func renderPage() []byte {
	var output bytes.Buffer
	output.WriteString("<main>")
	labelled := appicons.Icon(appicons.Config{
		Symbol: appicons.IconBrandDeveloperIconsTRPC,
		Label:  "tRPC",
	})
	if err := labelled.Render(context.Background(), &output); err != nil {
		panic(err)
	}
	decorative := appicons.Icon(appicons.Config{
		Symbol:     appicons.IconBrandDeveloperIconsTRPC,
		Label:      "must be hidden",
		Decorative: true,
	})
	if err := decorative.Render(context.Background(), &output); err != nil {
		panic(err)
	}
	output.WriteString("</main>")
	return output.Bytes()
}

func verifyHTTP(baseURL string, sprite []byte) {
	page, pageType := get(baseURL + "/")
	if pageType != "text/html" {
		panic("page content type mismatch: " + pageType)
	}
	verifyAccessibility(page)

	servedSprite, spriteType := get(baseURL + "/assets/icons/app.svg")
	if spriteType != "image/svg+xml" {
		panic("sprite content type mismatch: " + spriteType)
	}
	if !bytes.Equal(servedSprite, sprite) {
		panic("served sprite bytes differ from generated sprite")
	}
	verifySpriteGeometry(servedSprite, "devicon-trpc")
}

func get(url string) ([]byte, string) {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("GET %s status = %d", url, response.StatusCode))
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		panic(err)
	}
	contents, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		panic(err)
	}
	return contents, mediaType
}

func verifyAccessibility(page []byte) {
	decoder := xml.NewDecoder(bytes.NewReader(page))
	labelled, decorative := false, false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "svg" {
			continue
		}
		attrs := attributes(start.Attr)
		if attrs["aria-label"] == "tRPC" && attrs["role"] == "img" && attrs["aria-hidden"] == "" {
			labelled = true
		}
		if attrs["aria-hidden"] == "true" && attrs["aria-label"] == "" && attrs["role"] == "" {
			decorative = true
		}
	}
	if !labelled || !decorative {
		panic(fmt.Sprintf("accessibility contract failed: labelled=%t decorative=%t", labelled, decorative))
	}
}

func verifySpriteGeometry(sprite []byte, symbolID string) {
	decoder := xml.NewDecoder(bytes.NewReader(sprite))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			panic("sprite symbol missing: " + symbolID)
		}
		if err != nil {
			panic(err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "symbol" {
			continue
		}
		attrs := attributes(start.Attr)
		if attrs["id"] != symbolID {
			continue
		}
		parts := strings.Fields(attrs["viewBox"])
		if len(parts) != 4 {
			panic("invalid symbol viewBox")
		}
		width, widthErr := strconv.ParseFloat(parts[2], 64)
		height, heightErr := strconv.ParseFloat(parts[3], 64)
		if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
			panic("non-positive symbol geometry")
		}
		return
	}
}

func attributes(input []xml.Attr) map[string]string {
	result := make(map[string]string, len(input))
	for _, attr := range input {
		result[attr.Name.Local] = attr.Value
	}
	return result
}
EOF
GOWORK=off go mod tidy
GOWORK=off go test ./...
GOWORK=off go run ./cmd/proof

echo "iconpack consumer proof: PASS"
