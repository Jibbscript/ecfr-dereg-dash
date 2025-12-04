# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

eCFR Deregulation Dashboard - Analytics app over eCFR (Electronic Code of Federal Regulations) corpus with RSCS (Regulatory Restrictions and Complexity Score) metric, LSA (Legislation, Standards, and Amendments) changes, and AI-generated summaries.

**Tech stack:** Go backend, Nuxt 3 + USWDS frontend, Parquet + SQLite persistence, DuckDB analytics, Vertex AI (Gemini) for summaries.

## Build & Run Commands

### Local Development
```bash
# Full stack with Docker
docker-compose up

# Build Go binaries
go build -o api ./cmd/api
go build -o etl ./cmd/etl
go build -o import-summaries ./cmd/import-summaries
go build -o etl-summary-parse ./cmd/etl-summary-parse

# Run API server (requires SQLite DB and Parquet files in data/)
./api  # or: go run ./cmd/api

# Run ETL pipeline
./etl  # or: go run ./cmd/etl
./etl -skip-summary  # Skip AI summary generation

# Import pre-computed summaries into SQLite
./import-summaries

# Parse Vertex AI batch output into Parquet
./etl-summary-parse

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
  api/              → HTTP server entry point
  etl/              → ETL pipeline entry point
  import-summaries/ → Import pre-computed summaries to SQLite
  etl-summary-parse/→ Parse Vertex AI batch output to Parquet
internal/
  domain/    → Entities (Agency, Title, Section, Summary, LSAActivity, Diff)
  usecase/   → Business logic (Ingest, Snapshot, Metrics, Summaries)
  adapter/   → External integrations
    anthropic/ → Anthropic Claude API client (fallback)
    ecfr/      → eCFR API client
    govinfo/   → GovInfo API + GCS raw XML storage
    parquet/   → Parquet file I/O (local or GCS)
    sqlite/    → SQLite repository
    duck/      → DuckDB query helper
    vertexai/  → Vertex AI (Gemini) batch & online predictions
    lsa/       → LSA activity collector
  delivery/
    http/    → Chi router handlers, DTOs, validation
  platform/ → Config loading, logging (zap)
web/         → Nuxt 3 frontend with USWDS
```

### Data Flow
1. **ETL** fetches titles from GovInfo → parses XML → computes RSCS metrics → generates AI summaries → writes Parquet + SQLite
2. **API** reads from DuckDB (querying Parquet) for fast aggregates, SQLite for lookups and summaries
3. **Frontend** consumes REST API, renders USWDS-compliant pages with modals for RSCS explainer and AI summaries

### Key Domain Concepts
- **RSCS (Regulatory Restrictions and Complexity Score):** Computed from definitions, cross-references, and modal verb counts
- **LSA Activity:** Per-agency regulatory activity (proposed rules, final rules, notices) fetched from Federal Register API
- **AgencyLSA:** Stores per-agency document counts from Federal Register API (last 30 days)
- **Snapshots:** Daily Parquet partitions by `snapshot_date/title/`
- **Summaries:** AI-generated title-level summaries stored in SQLite, generated via Vertex AI batch API

### ETL Pipeline
- **Concurrent processing:** Up to 4 titles processed in parallel
- **Dedicated SQLite writer:** Buffered channel (100-capacity) for concurrency safety
- **Regex worker pool:** CPU-bound metrics computed in parallel (NumCPU workers)
- **GCS-native:** Raw XML stored in GCS, local filesystem only for SQLite DB
- **Agency LSA collection:** Fetches per-agency regulatory activity from Federal Register API (faceted batch queries)

## Configuration

Copy `.env.template` to `.env`. Key variables:
- `ENV`: `local`/`dev`/`prod` (controls GCS vs local filesystem)
- `DATA_DIR`: Local data directory (default: `./data`)
- `VERTEX_PROJECT_ID`, `VERTEX_LOCATION`: Vertex AI config
- `VERTEX_MODEL_ID`: AI model (default: `gemini-2.5-pro`)
- `GCS_BUCKET`: Batch prediction job I/O (default: `ecfr-dereg-dash-batch`)
- `PARQUET_BUCKET_NAME`, `RAW_XML_BUCKET_NAME`: GCS buckets for prod
- `RAW_XML_PREFIX`: Prefix for raw XML in GCS (default: `raw`)
- `DUCKDB_UI`: Set to `1` to enable DuckDB web UI on port 4213

## API Endpoints

- `GET /api/agencies` - List agencies with metrics
  - `?title={number}` - Filter by specific CFR Title
  - `?include_checksum=true` - Include content checksums (expensive)
- `GET /api/titles/{id}` - Title details and summary
- `GET /api/sections/{id}` - Section text and RSCS score
- `GET /api/summaries` - All AI-generated title summaries

## Database Schema

### SQLite Tables
- `agencies` - Agency metadata (id=slug, name, short_name, parent_id)
- `agency_cfr_references` - Agency-to-title mappings (agency_id, title, chapter)
- `agency_lsa` - Per-agency LSA data from Federal Register API:
  - `agency_id`, `agency_name`, `proposed_rules`, `final_rules`, `notices`, `total_documents`
  - `snapshot_date`, `captured_at`, `source_hint`
  - Unique constraint on `(agency_id, snapshot_date)`
- `lsa_activity` - Legacy per-title LSA table (no longer populated, kept for schema compatibility)
- `summaries` - AI-generated summaries with columns:
  - `kind` (title/agency/section), `key`, `text`, `model`, `created_at`
  - Unique constraint on `(kind, key)`

## Frontend Components

- **AiSummariesModal** - Browse all 50 title summaries with accordion UI
- **RscsExplainerModal** - Centralized RSCS methodology explainer
- **useRscsExplainer composable** - Global modal state management (provide/inject pattern)
