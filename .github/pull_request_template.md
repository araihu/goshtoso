## Summary

<!-- What does this PR do and why? -->

## Related issue

<!-- e.g. Closes #123 -->

## Type of change

- [ ] Bug fix
- [ ] New component / variant
- [ ] Visual-parity fix
- [ ] Docs
- [ ] Tests / CI
- [ ] Refactor / chore

## Checklist

- [ ] Ran `templ generate` after editing `.templ` files
- [ ] Rebuilt Tailwind (`tailwindcss -i css/main.css -o assets/styles.css`) if CSS changed
- [ ] `golangci-lint run` is clean (cyclomatic-complexity ceiling 20)
- [ ] `go fix ./...` applied (pre-commit hook enabled via `git config core.hooksPath .githooks`)
- [ ] Unit + E2E tests pass (`go test ./... -count=1` and `go test ./site/tests/e2e/... -count=1 -timeout 15m`)
- [ ] New/changed component demo page follows the docs-page pattern and has an API reference
- [ ] Regenerated the component reference (`go run ./cmd/skillgen`) if a component API changed
- [ ] Tested in light **and** dark mode across themes (especially Minimal — no border-radius)
- [ ] Did **not** hand-edit generated files (`*_templ.go`, `assets/styles.css`)

## Screenshots

<!-- For visual changes: before/after, light + dark. -->
