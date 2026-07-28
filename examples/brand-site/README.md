# Goshtoso static brand site

Copyable starter for a public organization, product, or publication. It
generates `public/index.html` plus Goshtoso's compiled CSS and a product-owned
`brand.css`; deploy `public/` to any static host.

This is deliberately not a generic landing-page kit. Keep identity, copy,
typography, imagery, and content hierarchy in `home.templ` and `brand.css`.
Goshtoso provides the token contract and remains available for real controls,
forms, feedback, and navigation when the public site needs them.

## Use

Inside this repository:

```bash
templ generate
GOWORK=off go test ./...
GOWORK=off go run .
```

Or create a fresh copy from a released Goshtoso module:

```bash
go run github.com/araihu/goshtoso/cmd/goshtoso@latest -init-brand-site=./my-site
cd my-site
go mod edit -dropreplace=github.com/araihu/goshtoso
go get github.com/araihu/goshtoso@latest
go mod tidy
templ generate
go test ./...
go run .
```
