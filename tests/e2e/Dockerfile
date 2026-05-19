FROM golang:1.26-bookworm

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    git \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1

RUN playwright install --with-deps chromium

WORKDIR /workspace/goshtoso

CMD ["bash", "-lc", "go test ./tests/e2e/..."]
