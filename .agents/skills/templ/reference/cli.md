# templ CLI reference

Verbatim from templ.guide. templ v0.3.x.

```
usage: templ <command> [<args>...]

commands:
  generate   Generates Go code from templ files
  fmt        Formats templ files
  lsp        Starts a language server for templ files
  info       Displays information about the templ environment
  version    Prints the version
```

## `templ generate`

Walks the current directory tree, turning every `*.templ` into `*_templ.go`. **Run after every `.templ` edit** — nothing changes until you do.

```
usage: templ generate [<args>...]

  -path <path>            Generate code for all files in path. (default .)
  -f <file>               Generate code for a single file, e.g. -f header.templ
  -source-map-visualisations  Emit HTML visualising templ↔Go mapping.
  -include-version        Include templ version in generated code. (default true)
  -include-timestamp      Include current time in generated code. (default false)
  -watch                  Watch path and regenerate on change.
  -cmd <cmd>              Command to run after generating code.
  -proxy                  URL to proxy after generating + running cmd.
  -proxyport              Proxy listen port. (default 7331)
  -proxybind              Proxy listen address. (default 127.0.0.1)
  -w                      Number of workers. (default NumCPU)
  -lazy                   Only regenerate when .templ is newer than .go.
  -pprof                  Port for pprof server.
  -keep-orphaned-files    Keep orphaned generated files. (default false)
  -v / -log-level         Verbosity ("debug"|"info"|"warn"|"error", default "info").
```

Single file: `templ generate -f header.templ`

## Live reload

```bash
templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
```
`--watch` regenerates on change; `--cmd` re-runs your server; `--proxy` injects browser auto-reload (default proxy URL `http://localhost:7331`). Watch-mode output is unoptimized — not for production builds.

## `templ fmt`

```bash
templ fmt .          # format dir tree
templ fmt            # stdin → stdout
templ fmt -fail .    # CI: exit 1 if any file needed formatting
```
If `prettierd`/`prettier`/`npx` is on PATH, formats `<script>`/`<style>` contents too.

## `templ lsp`

Language Server for IDE integrations (VS Code, Neovim, JetBrains, …). Not run directly by users. Starts its own `gopls`; `-gopls-remote` connects to a shared daemon. Requires the `templ` binary on PATH for IDE LSP to work. Flags: `-goplsLog`, `-goplsRPCTrace`, `-gopls-remote`, `-http <addr>`, `-log <file>`, `-pprof`.

## This repo (Goshtoso)

```bash
templ generate                                   # or: just gp-generate
tailwindcss -i css/main.css -o assets/styles.css # after CSS edits
go run cmd/server/main.go                         # or: just gp-dev (port 8090)
```

**Regeneration quirk:** `templ generate` sometimes reports "0 updates" though source changed. Force it:
```bash
rm components/<name>/<name>_templ.go && templ generate
```
Never hand-edit `*_templ.go`. On merge conflicts, resolve the `.templ` source then regenerate; never hand-resolve generated files.
