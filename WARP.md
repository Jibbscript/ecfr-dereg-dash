# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Commands

### Backend (Go)
- Build binaries
  - API: `go build -o api ./cmd/api`
  - ETL: `go build -o etl ./cmd/etl`
- Run locally (without Docker)
  - API: `go run ./cmd/api`
  - ETL (full pipeline): `go run ./cmd/etl`
  - ETL (skip summaries): `go run ./cmd/etl -- -skip-summary`
- Tests
  - All: `go test ./...`
  - Single package: `go test -v ./internal/usecase`
  - Single test: `go test -v ./cmd/etl -run TestName`
- Lint (static analysis)
  - `go vet ./...`

Environment needed for local runs
- Copy env: `cp .env.template .env`, then edit as needed.
- Common vars (see `internal/platform/config.go`): `ENV=local`, `DATA_DIR=./data`, `PARQUET_PREFIX=parquet`, optionally `DUCKDB_UI=1`.

### Frontend (web)
- Install deps: `cd web && npm install`
- Dev server: `npm run dev` (serves on http://localhost:3000)
- Build: `npm run build`
- Unit tests (Vitest): `npm run test`
  - Single test: `npm run test -- -t "partial name or regex"`
- E2E tests (Playwright): `npm run test:e2e`
  - Single spec/title: `npx playwright test tests/<file>.spec.ts -g "title"`

### Full stack via Docker Compose
- Bring up API, ETL, and Web: `docker-compose up --build`
  - Web: http://localhost:3000
  - API: http://localhost:8080 (handlers mounted under `/api`)
  - DuckDB UI: http://localhost:4213 (if enabled in code/env)
- Services and Dockerfiles: `Dockerfile.api`, `Dockerfile.etl`, `Dockerfile.web`; compose file: `docker-compose.yml`.

### Useful data inspection
- SQLite snapshot: `sqlite3 data/ecfr.db "SELECT count(*) FROM sections;"`
- Parquet snapshots (local mode): under `data/parquet/YYYY-MM-DD/`.

## Architecture and structure (big picture)

This repo implements an analytics dashboard over the eCFR corpus. It follows a Clean Architecture layout in Go with a Nuxt 3 frontend.

High-level components
- `cmd/api`: HTTP server (Go + chi) exposing analytics endpoints under `/api`.
- `cmd/etl`: Parallel ETL pipeline that ingests eCFR, computes metrics, snapshots diffs, and generates title-level summaries.
- `web/`: Nuxt 3 + Vue 3 frontend (USWDS) consuming the API.

Core Go layers (under `internal/`)
- `domain/`: Core entities (Title, Section, AgencyMetric, Diff, Summary, etc.).
- `usecase/`: Business logic orchestrators
  - `Ingest`: downloads/parse XML, computes per-section metrics, writes Parquet + queues SQLite inserts.
  - `Snapshot`: reads Parquet to compute per-title diffs between dates.
  - `Metrics`: serves aggregates (agency totals) sourced from SQLite (and wired for DuckDB).
  - `Summaries`: generates title-level summaries (Vertex AI) and writes to Parquet.
- `adapter/`: External and data-access integrations
  - `govinfo`: streams Title XML into GCS (or local FS in `ENV=local`).
  - `parquet`: reads/writes Parquet snapshots (`YYYY-MM-DD/<title>.parquet`, plus `_diffs` and `summaries.parquet`).
  - `sqlite`: local mirror and quick aggregates (e.g., agency totals).
  - `duck`: DuckDB helper (prepared for Parquet/SQLite queries; UI optional).
  - `ecfr`, `lsa`, `vertexai`, `anthropic`: data/API sources for catalog, LSA activity, and summaries.
- `delivery/http/`: chi router + handlers and DTOs
  - Implements `/agencies` (real aggregate from SQLite) and dummy-backed `/titles/{id}` and `/sections/{id}` used for E2E scaffolding.
- `platform/`: config loading (env-based) and logging (zap).

Data flow (ETL)
1) Ingest
- `govinfo` saves latest Title XML to GCS or local FS.
- `Ingest` parses XML into Sections, normalizes text, computes metrics (word count, RSCS proxies), and writes:
  - Parquet per title at `parquet/<YYYY-MM-DD>/<title>.parquet`.
  - Sections into SQLite via a dedicated writer goroutine to avoid lock contention.
2) Snapshot
- `Snapshot` compares current vs previous snapshot (from Parquet) and persists `<title>_diffs.parquet`.
3) LSA activity
- `lsa` scrapes ReaderAids to estimate monthly “sections affected” counts, persisted with the snapshot.
4) Summaries
- `Summaries` calls Vertex AI to produce one summary per Title (stored in `summaries.parquet`).

Runtime/API
- API wires adapters based on env (local FS vs GCS) and exposes routes under `/api`.
- See `openapi.yaml` for shapes; note the runtime prefix (`/api`) when calling locally.

## Notes and pointers
- Quick start (local): `cp .env.template .env && docker-compose up --build`, then open http://localhost:3000.
- Important references: `README.md` (setup + live demo), `ETL_GUIDE.md` (pipeline details), `API.md` and `openapi.yaml` (endpoints), `AGENTS.md` (concise build/test cheatsheet and architecture bullets).
