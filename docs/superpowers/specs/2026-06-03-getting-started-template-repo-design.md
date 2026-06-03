# Getting Started Template Repo Design

## Context

The current Getting Started page teaches users how to build the dog-breeds
table, but it still asks too much of a new user. The page contains large inline
source blocks and explains Goshtoso runtime assets through `assets.Handler()`,
yet the complete beginner project is not presented as a clean, copyable starter.

Recent Getting Started work also introduced dog images conceptually: the example
should demonstrate table image cells and a polished expected result. Those app
images are different from Goshtoso's bundled runtime assets. Runtime assets
belong to the library and are served by `assets.Handler()` from `/assets/`; dog
photos belong to the starter app and should be embedded or served by the app
itself.

## Decision

Create `araihu/goshtoso-getting-started` as the canonical beginner starter repo.
The docs page should make this the primary path for users who want the complete
runnable project, including app images, while preserving the manual walkthrough
for users who want to understand the implementation.

## Starter Repository

The starter repo should contain the full dog-breeds app:

- `README.md` with a short overview, prerequisites, run commands, and a note on
  runtime assets versus app images.
- `go.mod` and `go.sum` pinned to compatible Goshtoso and templ versions.
- `main.go` with the HTTP server, `assets.Handler()` mount, HTMX endpoint, dog
  data, filtering, sorting, pagination, and app image serving.
- `page.templ` and generated `page_templ.go` so users can run the app
  immediately after cloning, while still seeing the templ source.
- `assets/dogs/*.webp` with the dog images used by the table.
- Optional `.gitignore` and license/readme metadata suitable for a public GitHub
  template repository.

The repository should be marked as a GitHub template if possible, but ordinary
`git clone https://github.com/araihu/goshtoso-getting-started` must remain the
lowest-friction documented path.

## Docs Page

Update `site/internal/pages/demo/components/getting_started.templ` so the page
answers "what am I building?" before the code wall.

The top of the page should include a live expected outcome for the dog-breeds
app. This preview should be rendered on the same docs page by the site from the
same example data and components, with search/filter/sort/pagination working
through the site server. A link to the starter repository can sit beside the
preview, but it does not replace the same-page outcome.

Below the outcome, add a primary "Use the starter repo" path with clone/template
commands:

```bash
git clone https://github.com/araihu/goshtoso-getting-started dog-breeds
cd dog-breeds
go run .
```

The manual build path stays below as an educational section. It should continue
to show `page.templ` and `main.go`, but the surrounding copy should make clear
that copying source manually is optional.

## Asset Story

The docs must distinguish two asset sources:

- Goshtoso runtime assets: CSS, Alpine.js, HTMX, fonts, and library images served
  from `/assets/` by `github.com/araihu/goshtoso/assets.Handler()`.
- Starter app assets: dog images stored under `assets/dogs` in the starter repo,
  embedded with Go's `embed` package and served by the starter app under a route
  such as `/dog-images/`.

This distinction should appear in the starter README and in the docs page near
the setup commands, because it is the main source of confusion.

## Data Flow

The starter app remains server-rendered:

1. The root route renders a full templ page with `table.Table`.
2. Filter, sort, and pagination controls issue HTMX requests to the app endpoint.
3. The endpoint filters/sorts/paginates Go data and returns rendered table rows.
4. Table OOB fragments update pagination and header state as needed.
5. Dog image URLs point at the starter app image route and are rendered through
   Goshtoso's native table image cell API.

No client-side JSON API, bundler, or CDN is introduced.

## Testing

Implementation should verify:

- The starter repo builds from a fresh clone with `go run .`.
- The docs page builds after `templ generate`.
- `go test ./...` passes in the library module.
- `cd site && go test ./...` passes with a local `go.work` overlay.
- At least one browser/E2E check confirms the Getting Started page shows the
  expected outcome section and the starter repo commands.
