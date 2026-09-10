FROM golang:1.27.1-alpine AS builder

WORKDIR /src

# Two modules: the library at the repo root and the site/examples under site/.
# Copy both manifests first so module downloads cache independently of source.
COPY go.mod go.sum ./
COPY site/go.mod site/go.sum ./site/

# A build-time workspace ties the site to the in-repo library, so the image
# always reflects the library at this commit (go.work is gitignored; it exists
# only inside the build). The site's pinned require is for fresh-clone builds.
RUN go work init . ./site && go mod download

COPY . .
RUN GOSHTOSO_DOCS_VERSION="$(cd site && GOWORK=off go list -m -f '{{.Version}}' github.com/araihu/goshtoso)" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags "-X github.com/araihu/goshtoso/site/internal/buildinfo.goDocsVersion=${GOSHTOSO_DOCS_VERSION}" \
      -o /out/server ./site/cmd/server
# Drop the build-only workspace so it is never baked into the runtime image
# (the final stage copies all of /src for its assets/).
RUN rm -f go.work go.work.sum

FROM alpine:3.24

WORKDIR /app

# cwd /app holds assets/ (copied from the repo root), so the server's
# resolveProjectRoot() finds them via its cwd branch.
COPY --from=builder /src /app
COPY --from=builder /out/server /app/server

EXPOSE 8090

CMD ["/app/server", "-port", "8090"]
