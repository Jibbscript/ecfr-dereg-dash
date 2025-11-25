# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

eCFR Deregulation Dashboard - Analytics app over eCFR (Electronic Code of Federal Regulations) corpus with RSCS (Regulatory Restrictions and Complexity Score) metric, LSA (Legislation, Standards, and Amendments) changes, and AI-generated summaries.

**Tech stack:** Go backend, Nuxt 3 + USWDS frontend, Parquet + SQLite persistence, DuckDB analytics.

## Build & Run Commands

### Local Development
```bash
# Full stack with Docker
docker-compose up

# Build Go binaries
go build -o api ./cmd/api
go build -o etl ./cmd/etl

# Run API server (requires SQLite DB and Parquet files in data/)
./api  # or: go run ./cmd/api

# Run ETL pipeline
./etl  # or: go run ./cmd/etl
./etl -skip-summary  # Skip AI summary generation

# Web frontend (from web/ directory)
npm run dev      # Development server
npm run build    # Production build
npm run preview  # Preview production build
```

### Testing
```bash
# Go tests
go test ./...                          # All tests
go test ./internal/usecase/...         # Specific package
go test -v -run TestMetrics ./...      # Single test by name

# Web tests (from web/ directory)
npm run test              # Unit tests (vitest)
npm run test:e2e          # E2E tests (playwright)
```

### Ports
- API: `8080`
- Web: `3000` (proxies `/api/**` to API)
- DuckDB UI: `4213` (when `DUCKDB_UI=1`)

## Architecture

### Clean Architecture Layers
```
cmd/
  api/       → HTTP server entry point
  etl/       → ETL pipeline entry point
internal/
  domain/    → Entities (Agency, Title, Section, Summary, LSAActivity, Diff)
  usecase/   → Business logic (Ingest, Snapshot, Metrics, Summaries)
  adapter/   → External integrations
    ecfr/      → eCFR API client
    govinfo/   → GovInfo API + GCS raw XML storage
    parquet/   → Parquet file I/O (local or GCS)
    sqlite/    → SQLite repository
    duck/      → DuckDB query helper
    vertexai/  → Vertex AI (Gemini) for summaries
    lsa/       → LSA activity collector
  delivery/
    http/    → Chi router handlers, DTOs, validation
  platform/ → Config loading, logging (zap)
web/         → Nuxt 3 frontend with USWDS
```

### Data Flow
1. **ETL** fetches titles from GovInfo → parses XML → computes RSCS metrics → writes Parquet + SQLite
2. **API** reads from DuckDB (querying Parquet) for fast aggregates, SQLite for lookups
3. **Frontend** consumes REST API, renders USWDS-compliant pages

### Key Domain Concepts
- **RSCS (Regulatory Restrictions and Complexity Score):** Computed from definitions, cross-references, and modal verb counts
- **LSA Activity:** Tracks proposals, amendments, and finals from recent regulatory activity
- **Snapshots:** Daily Parquet partitions by `snapshot_date/title/`

## Configuration

Copy `.env.template` to `.env`. Key variables:
- `ENV`: `local`/`dev`/`prod` (controls GCS vs local filesystem)
- `DATA_DIR`: Local data directory (default: `./data`)
- `VERTEX_PROJECT_ID`, `VERTEX_LOCATION`, `VERTEX_MODEL_ID`: Gemini AI config
- `PARQUET_BUCKET_NAME`, `RAW_XML_BUCKET_NAME`: GCS buckets for prod
- `DUCKDB_UI`: Set to `1` to enable DuckDB web UI on port 4213

## API Endpoints

- `GET /api/agencies` - List agencies with metrics
- `GET /api/titles/{id}` - Title details and summary
- `GET /api/sections/{id}` - Section text and RSCS score
