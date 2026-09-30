FROM golang:1.27.1-bookworm

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    git \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/mxschmitt/playwright-go/cmd/playwright@v0.6100.0

RUN /go/bin/playwright install --with-deps chromium

WORKDIR /workspace/goshtoso

CMD ["bash", "-lc", "go test -tags=e2e,full ./tests/e2e/..."]
