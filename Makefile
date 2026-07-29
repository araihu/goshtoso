.PHONY: all build test test-e2e clean dev dev-watch dev-air install install-templ install-playwright css css-watch generate js js-check

# Default target
all: css build

# Build CSS and server
build: css
	go build -o bin/server ./site/cmd/server

# Run development server (builds CSS first)
dev: css
	go run ./site/cmd/server

# Development with CSS watching (requires GNU parallel or manual two terminals)
dev-watch: css
	@echo "Starting dev server with CSS watch..."
	@echo "Run these in separate terminals:"
	@echo "  Terminal 1: make css-watch"
	@echo "  Terminal 2: make dev"

# Run with Air for live reloading (recommended for development)
dev-air: css
	@echo "Starting server with Air (live reload)..."
	@echo "Server will be available at http://localhost:8090"
	@echo "Air will watch for changes and automatically rebuild"
	@which air > /dev/null || (echo "Air not installed. Run: go install github.com/air-verse/air@latest" && exit 1)
	air

# Install Air if not already installed
install-air:
	go install github.com/air-verse/air@latest

# Build Tailwind CSS
css:
	@echo "Building Tailwind CSS..."
	tailwindcss -i css/main.css -o assets/styles.css

# Watch Tailwind CSS for changes
css-watch:
	@echo "Watching CSS for changes..."
	tailwindcss -i css/main.css -o assets/styles.css --watch

# Run Go tests
test:
	go test ./...
	cd site && go test $$(go list ./... | grep -v /tests/e2e)

# Run E2E tests (builds CSS first)
test-e2e: css
	cd site && go test ./tests/e2e/... -v

# Run specific E2E test
test-e2e-one:
	cd site && go test ./tests/e2e/... -v -run $(TEST)

# Install dependencies
install: install-templ install-playwright install-air
	go mod download

# Install templ CLI
install-templ:
	go install github.com/a-h/templ/cmd/templ@latest

# Install Playwright browsers
install-playwright:
	go install github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1
	playwright install chromium

# Generate templ files
generate:
	templ generate

js:
	go run ./cmd/jsbuild

js-check:
	go run ./cmd/jslint
	go run ./cmd/jsbuild -check

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf assets/styles.css
	rm -rf site/tests/e2e/test-results/

# Format code
fmt:
	go fmt ./...
	cd site && go fmt ./...

# Lint
lint:
	go vet ./...
	cd site && go vet ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build CSS and server binary"
	@echo "  make dev            - Build CSS and run dev server"
	@echo "  make dev-watch      - Show instructions for dev with CSS watching"
	@echo "  make dev-air        - Run with Air (live reload) - RECOMMENDED"
	@echo "  make css            - Build Tailwind CSS"
	@echo "  make css-watch      - Watch and rebuild CSS on changes"
	@echo "  make test           - Run Go tests"
	@echo "  make test-e2e       - Run E2E tests"
	@echo "  make test-e2e-one   - Run specific E2E test (TEST=TestName)"
	@echo "  make install        - Install all dependencies"
	@echo "  make install-air    - Install Air live reload tool"
	@echo "  make generate       - Generate templ files"
	@echo "  make js             - Build minified first-party JavaScript"
	@echo "  make js-check       - Lint JavaScript and verify generated drift"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
