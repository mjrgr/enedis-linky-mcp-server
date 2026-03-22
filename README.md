# enedis-linky-mcp-server

A production-ready **Model Context Protocol (MCP) server** written in Go that exposes [Conso API](https://conso.boris.sh) — a free, open-source proxy for Enedis Linky smart meter data — as structured tools usable by any MCP-compatible LLM client (Claude Desktop, Continue, etc.).

---

## Features

| MCP Tool | Description |
|---|---|
| `get_daily_consumption` | Daily electricity consumption (Wh) for a date range |
| `get_load_curve` | 30-minute average power readings (W) |
| `get_max_power` | Maximum power reached each day (VA) |
| `get_daily_production` | Daily solar production (Wh) for solar installations |
| `get_production_load_curve` | 30-minute average production power readings (W) |
| `get_consumption_summary` | Aggregated stats: total, average/day, peak day |
| `health_check` | Verify Conso API reachability |

**Additional capabilities:**
- Bearer token authentication
- Automatic retry with exponential backoff
- Rate-limit awareness (5 req/s, 10k req/h)
- Dual MCP transports: **stdio** (Claude Desktop) and **SSE** (HTTP server)
- Structured JSON logging via `log/slog`
- Multi-platform Docker image (amd64 / arm64)

---

## Architecture

```
enedis-linky-mcp-server/
├── cmd/
│   └── server/
│       └── main.go            # Entrypoint — wires config, client, service, MCP server
│
├── internal/
│   ├── config/
│   │   └── config.go          # ENV-based configuration with validation
│   ├── client/
│   │   └── client.go          # Typed HTTP client — retry, backoff, User-Agent
│   ├── service/
│   │   └── service.go         # Business logic — validation, aggregation
│   └── mcp/
│       ├── server.go          # MCP server lifecycle (stdio / SSE transport)
│       └── tools.go           # Tool definitions & handlers
│
├── .github/
│   └── workflows/
│       ├── ci.yml             # lint → test → build
│       └── release.yml        # GoReleaser + Docker (triggered on tags)
│
├── Dockerfile                 # Multi-stage, distroless final image
├── Makefile                   # Developer shortcuts
├── .env.example               # Configuration template
└── .golangci.yml              # Linter configuration
```

---

## Prerequisites

- A free **Conso API token**: register at [conso.boris.sh](https://conso.boris.sh)
- On your Enedis account, enable:
  - *Enregistrement de la consommation horaire*
  - *Collecte de la consommation horaire*
- Your Linky meter's **PRM** number (14-digit identifier on your electricity bill)

---

## Quick Start

### Claude Desktop (stdio transport)

1. **Build the binary**

   ```bash
   make build
   # Binary is at ./bin/enedis-linky-mcp-server
   ```

2. **Configure Claude Desktop** — edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

   ```json
   {
     "mcpServers": {
       "linky": {
         "command": "/absolute/path/to/bin/enedis-linky-mcp-server",
         "env": {
           "CONSO_API_TOKEN": "your_token_here"
         }
       }
     }
   }
   ```

3. **Restart Claude Desktop** — the Linky tools will appear in the tool list.

---

### HTTP/SSE transport

```bash
export CONSO_API_TOKEN=your_token_here
export MCP_TRANSPORT=sse
export PORT=8080
./bin/enedis-linky-mcp-server
# Listening on :8080
```

Then point your MCP client at `http://localhost:8080/sse`.

---

### Docker

```bash
docker run --rm \
  -e CONSO_API_TOKEN=your_token_here \
  -e MCP_TRANSPORT=sse \
  -p 8080:8080 \
  ghcr.io/mjrgr/enedis-linky-mcp-server:latest
```

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `CONSO_API_TOKEN` | **Yes** | — | Conso API bearer token |
| `MCP_TRANSPORT` | No | `stdio` | `stdio` or `sse` |
| `PORT` | No | `8080` | HTTP port (SSE transport only) |
| `LOG_LEVEL` | No | `info` | `debug` \| `info` \| `warn` \| `error` |
| `CONSO_API_BASE_URL` | No | `https://conso.boris.sh/api` | Override for testing |

---

## Example Prompts

Once connected to Claude Desktop, try:

- *"What was my electricity consumption last month?"*
  → Claude calls `get_daily_consumption` and summarises the data.

- *"Show me my peak consumption day in January 2025."*
  → Claude calls `get_consumption_summary`.

- *"Plot my load curve for 2025-01-15."*
  → Claude calls `get_load_curve` and describes the usage pattern.

- *"Is the Linky API working?"*
  → Claude calls `health_check`.

---

## Tool Reference

### `get_daily_consumption`
```json
{
  "prm":   "12345678901234",
  "start": "2025-01-01",
  "end":   "2025-01-31"
}
```
Returns:
```json
{
  "prm": "12345678901234",
  "start_date": "2025-01-01",
  "end_date": "2025-01-31",
  "unit": "Wh",
  "measuring_period": "P1D",
  "count": 31,
  "readings": [
    { "date": "2025-01-01", "value": "12340" },
    ...
  ]
}
```

### `get_consumption_summary`
```json
{
  "prm":   "12345678901234",
  "start": "2025-01-01",
  "end":   "2025-01-31"
}
```
Returns:
```json
{
  "prm": "12345678901234",
  "start_date": "2025-01-01",
  "end_date": "2025-01-31",
  "unit": "Wh",
  "total_wh": 382540,
  "total_kwh": 382.54,
  "average_per_day_wh": 12340.0,
  "peak_day": "2025-01-15",
  "peak_value_wh": 18750,
  "reading_count": 31
}
```

### `health_check`
No input required. Returns:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-31T10:00:00Z",
  "message": "Conso API is reachable"
}
```

---

## Development

```bash
# Install dependencies
go mod download

# Run locally (stdio)
cp .env.example .env && $EDITOR .env
make run

# Run tests
make test

# Lint
make lint

# Build Docker image
make docker-build
```

---

## Rate Limits

The Conso API enforces Enedis-imposed quotas shared across all users:
- **5 requests per second**
- **10,000 requests per hour**

The client automatically retries transient failures with exponential backoff. Make requests only when necessary; data is refreshed once daily around 8 AM.

---

## License

[MIT](LICENSE)

---

## Acknowledgements

- [Conso API](https://github.com/bokub/conso-api) by [Boris K](https://github.com/bokub) — the underlying free Linky data proxy (GPL-3.0)
- [mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP SDK
