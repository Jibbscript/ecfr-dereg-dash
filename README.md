# eCFR Deregulation Dashboard

## Live Demo

🚀 **Access the live dashboard here:** [https://ecfr-web-369420849740.us-central1.run.app](https://ecfr-web-369420849740.us-central1.run.app)

## Overview

An analytics dashboard for exploring the regulatory complexity of the Code of Federal Regulations (eCFR). The system calculates the **RSCS (Regulatory Complexity Score)** metric per 1,000 words, tracks **LSA (List of Sections Affected)** activity from the Federal Register, and provides AI-generated summaries for titles and sections.

**Tech Stack:**
- **Backend**: Go 1.24+ with Chi router, clean architecture
- **Frontend**: Nuxt 3 + Vue 3, USWDS design system
- **Storage**: SQLite (metadata), Parquet (daily snapshots)
- **Analytics**: DuckDB for fast aggregations
- **AI**: Vertex AI / Anthropic for summaries

## Prerequisites

- **Go**: 1.24+
- **Node.js**: 18+ with npm (or bun)
- **Docker** (optional): For containerized deployment

## Quick Start

### 1. Environment Setup
```bash
cp .env.template .env
```

For local development, the defaults work out of the box:
- `ENV=local` — Uses local filesystem instead of GCS
- `DATA_DIR=./data` — Location of SQLite DB and Parquet files
- `DUCKDB_UI=1` — Enables DuckDB Web UI at `:4213`

### 2. Data Requirements
The `data/` directory must contain:
- `ecfr.db` — SQLite database with metadata
- Date-partitioned Parquet files (e.g., `parquet/2025-01-15/*.parquet`)

To populate data, run the ETL pipeline (see [ETL_GUIDE.md](ETL_GUIDE.md)).

### 3a. Run Without Docker (Recommended for Development)

**Terminal 1 — API Server:**
```bash
go run ./cmd/api
```
API runs on `http://localhost:8080`

**Terminal 2 — Frontend Dev Server:**
```bash
cd web
npm install
npm run dev
```
UI runs on `http://localhost:3000` with hot reload.

### 3b. Run With Docker
```bash
docker-compose up
```
- **UI**: `http://localhost:3000`
- **API**: `http://localhost:8080`
- **DuckDB UI**: `http://localhost:4213` (if `DUCKDB_UI=1`)

## User Guide

### Navigation
The dashboard provides an intuitive interface to explore federal regulations:

1. **Agency Overview**: Start here to see a list of all federal agencies.
   - **Metrics**: View total word counts, average RSCS scores, and LSA activity for each agency.
   - **Sort & Filter**: Use table headers to sort by complexity or volume; filter by CFR Title.
   - **Hierarchy**: Expand departments to see their sub-agencies.

2. **Key Metrics Explained**:
   - **RSCS (Regulatory Complexity Score)**: Measures complexity based on word count, definitions, cross-references, and modal verbs ("shall", "must", etc.) per 1,000 words.
   - **LSA Activity**: Counts of proposed rules, final rules, and notices from the Federal Register API (last 30 days).

3. **AI Summaries**: Click the "AI Summaries" button to view machine-generated summaries of titles and sections.

## Usage

### ETL Pipeline
Refresh the regulatory data by running the ETL pipeline:
```bash
go run ./cmd/etl
```

The pipeline:
1. Fetches changed titles from the eCFR API
2. Downloads and parses XML from GovInfo
3. Computes RSCS metrics for each section
4. Collects agency-level LSA data from Federal Register
5. Writes snapshots to Parquet and SQLite

For full details, see [ETL_GUIDE.md](ETL_GUIDE.md).

### Development Commands

**Backend:**
```bash
go build -o api ./cmd/api      # Build API
go build -o etl ./cmd/etl      # Build ETL
go test ./...                   # Run all tests
```

**Frontend:**
```bash
cd web
npm install                     # Install dependencies
npm run dev                     # Dev server with hot reload
npm run build                   # Production build
npm run test                    # Unit tests (Vitest)
npm run test:e2e               # E2E tests (Playwright)
```

## DuckDB Local Analytics

The DuckDB web UI provides an interactive SQL interface for analysts and developers to explore eCFR regulatory data directly.

### Prerequisites
1. Ensure `DUCKDB_UI=1` is set in your `.env` file (enabled by default)
2. Run the ETL pipeline at least once to populate Parquet data: `go run ./cmd/etl`
3. Ensure `data/ecfr.db` exists (SQLite database)

### Starting the DuckDB UI
```bash
# Start the API server (DuckDB UI starts automatically)
go run ./cmd/api

# Or with Docker
docker-compose up api
```

Access the UI at: **http://localhost:4213**

### Dashboard-Equivalent Queries

These queries replicate the backend API data shown on the dashboard.

> **Note:** SQLite tables are attached as `ecfr.*` (e.g., `ecfr.agencies`). Parquet columns use PascalCase (e.g., `SnapshotDate`, `RSCSPer1K`).

#### Agency Metrics (matches /api/agencies)
```sql
SELECT
    a.id, a.name, a.parent_id,
    COALESCE(SUM(s.word_count), 0) as total_words,
    COALESCE(AVG(s.rscs_per_1k), 0) as avg_rscs,
    COALESCE(lsa.total_documents, 0) as lsa_counts
FROM ecfr.agencies a
LEFT JOIN ecfr.agency_cfr_references acr ON acr.agency_id = a.id
LEFT JOIN ecfr.sections s ON s.title = CAST(acr.title AS TEXT) AND s.agency_id = acr.chapter
LEFT JOIN (SELECT * FROM ecfr.agency_lsa WHERE snapshot_date = (SELECT MAX(snapshot_date) FROM ecfr.agency_lsa)) lsa ON lsa.agency_id = a.id
GROUP BY a.id, a.name, a.parent_id, lsa.total_documents
ORDER BY total_words DESC;
```

#### Sections from Parquet Snapshots
```sql
-- Use filename filter to exclude _diffs files, union_by_name for schema compatibility
SELECT * FROM read_parquet('data/parquet/*/*.parquet', filename=true, union_by_name=true)
WHERE filename NOT LIKE '%_diffs%'
LIMIT 100;
```

#### AI Summaries
```sql
SELECT kind, key, text, model, created_at FROM ecfr.summaries ORDER BY key;
```

### Example Analytical Queries

#### 1. Regulatory Complexity Trend Analysis
```sql
-- Compare RSCS scores across snapshot dates by CFR Title
SELECT
    "SnapshotDate" as snapshot_date,
    "Title" as title,
    COUNT(*) as section_count,
    ROUND(AVG("RSCSPer1K"), 2) as avg_complexity,
    SUM("WordCount") as total_words
FROM read_parquet('data/parquet/*/*.parquet', filename=true, union_by_name=true)
WHERE filename NOT LIKE '%_diffs%'
    AND "SnapshotDate" IS NOT NULL
GROUP BY "SnapshotDate", "Title"
ORDER BY "SnapshotDate" DESC, avg_complexity DESC
LIMIT 15;
```

#### 2. Top 10 Most Complex Agencies
```sql
SELECT
    a.name,
    COUNT(DISTINCT s.id) as section_count,
    SUM(s.word_count) as total_words,
    ROUND(AVG(s.rscs_per_1k), 2) as avg_rscs,
    ROUND(MAX(s.rscs_per_1k), 2) as max_rscs
FROM ecfr.agencies a
JOIN ecfr.agency_cfr_references acr ON acr.agency_id = a.id
JOIN ecfr.sections s ON s.title = CAST(acr.title AS TEXT)
GROUP BY a.name
ORDER BY avg_rscs DESC
LIMIT 10;
```

#### 3. Recent Regulatory Activity by Agency
```sql
SELECT
    agency_name,
    proposed_rules,
    final_rules,
    notices,
    total_documents,
    snapshot_date
FROM ecfr.agency_lsa
WHERE CAST(snapshot_date AS DATE) >= current_date - INTERVAL 30 DAY
ORDER BY total_documents DESC
LIMIT 10;
```

## Architecture

The system follows **Clean Architecture** in Go:

```
internal/
├── domain/       # Core entities (Agency, Section, Summary, etc.)
├── usecase/      # Business logic (Ingest, Metrics, Snapshot, Summaries)
├── adapter/      # External integrations
│   ├── ecfr/     # eCFR API client
│   ├── govinfo/  # GovInfo XML/GCS client
│   ├── parquet/  # Local & GCS Parquet storage
│   ├── sqlite/   # SQLite repository
│   ├── duck/     # DuckDB analytics helper
│   ├── lsa/      # Federal Register API collector
│   ├── anthropic/# Anthropic Claude client (Deprecated)
│   └── vertexai/ # Google Vertex AI client
└── delivery/
    └── http/     # Chi router, handlers, DTOs
```

**Data Flow:**
1. **ETL** fetches eCFR titles/sections from GovInfo, computes metrics, generates summaries
2. Results stored in Parquet (date-partitioned snapshots) and SQLite (metadata mirror)
3. **API** queries DuckDB over Parquet for fast aggregates; SQLite for metadata
4. **Frontend** (Nuxt/Vue) renders USWDS-compliant pages

## Project Structure

```
├── cmd/
│   ├── api/              # HTTP API server
│   └── etl/              # ETL pipeline
├── internal/             # Go application code (clean architecture)
├── web/                  # Nuxt 3 frontend
│   ├── components/       # Vue components
│   ├── composables/      # Vue composables
│   ├── pages/            # Route pages
│   └── tests/            # Vitest & Playwright tests
├── data/                 # SQLite DB & Parquet files (gitignored)
├── infra/                # Terraform IaC
├── openapi.yaml          # API specification
├── docker-compose.yml    # Local multi-service setup
└── Dockerfile.*          # Container definitions
```

## API Reference

| Endpoint | Description |
|----------|-------------|
| `GET /api/agencies` | List agencies with word counts, RSCS, LSA activity |
| `GET /api/agencies?title=12` | Filter agencies by CFR title |
| `GET /api/agencies?include_checksum=true` | Include content checksums |
| `GET /api/titles/{id}` | Title metrics and summary |
| `GET /api/sections/{id}` | Section text, RSCS, and summary |
| `GET /api/summaries` | All AI-generated summaries |

See [API.md](API.md) and [openapi.yaml](openapi.yaml) for full specification.

## Testing

**Backend (Go):**
```bash
go test ./...                           # All tests
go test -v ./internal/usecase -run TestMetrics  # Single test
```

**Frontend (Vitest + Playwright):**
```bash
cd web
npm run test        # Unit tests
npm run test:e2e    # E2E tests
```

## Documentation

- [API.md](API.md) — API endpoints reference
- [DEPLOY.md](DEPLOY.md) — Deployment guide (local & GCP)
- [ETL_GUIDE.md](ETL_GUIDE.md) — ETL pipeline details
- [SCHEMA.md](SCHEMA.md) — Database schema

## License

MIT License — see [LICENSE](LICENSE) for details.
