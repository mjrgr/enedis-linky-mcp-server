# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache \
    musl-dev \
    openssl-dev \
    openssl-libs-static \
    ca-certificates

WORKDIR /build

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /enedis-linky-mcp-server ./cmd/server

# ── Final stage ──────────────────────────────────────────────────────────────
FROM scratch

LABEL org.opencontainers.image.base.name="scratch"
LABEL org.opencontainers.image.title="Enedis Linky MCP Server"
LABEL org.opencontainers.image.description="A production-ready Model Context Protocol (MCP) server in Go for Enedis Linky, enabling AI assistants to access and monitor your smart meter data. Includes an optional web dashboard for visualization."
LABEL org.opencontainers.image.ref.name="enedis-linky-mcp-server"
LABEL org.opencontainers.image.authors="Mehdi Jr-Gr <contact@mehdi.dev>"
LABEL org.opencontainers.image.vendor="Mehdi Jr-Gr"
LABEL org.opencontainers.image.source="https://github.com/mjrgr/enedis-linky-mcp-server"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Copy essential files for networking and TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER 1000:1000

COPY --from=builder /enedis-linky-mcp-server /enedis-linky-mcp-server

ENV PORT=8080 \
    MCP_TRANSPORT=stdio \
    LOG_LEVEL=info

EXPOSE 8080

ENTRYPOINT ["/enedis-linky-mcp-server"]
