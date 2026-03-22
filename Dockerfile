# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /build

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /enedis-linky-mcp-server ./cmd/server

# ── Final stage ──────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /enedis-linky-mcp-server /enedis-linky-mcp-server

# Default: SSE transport on port 8080. Override with MCP_TRANSPORT=stdio.
ENV PORT=8080 \
    MCP_TRANSPORT=sse \
    LOG_LEVEL=info

EXPOSE 8080

ENTRYPOINT ["/enedis-linky-mcp-server"]
