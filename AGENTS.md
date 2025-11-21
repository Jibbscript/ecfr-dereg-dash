# eCFR Deregulation Dashboard - Agent Guidelines

## Build, Lint, & Test
- **Backend (Go)**:
  - Build ETL: `go build -o etl ./cmd/etl`
  - Build API: `go build -o api ./cmd/api`
  - Test All: `go test ./...`
  - Test Single: `go test -v ./internal/path/to/pkg -run TestName`
- **Frontend (Web)**:
  - Directory: `cd web`
  - Install: `npm install` (Check for `yarn.lock` or `bun.lock` to prefer those)
  - Dev Server: `npm run dev`
  - Test (Unit): `npm run test` (Vitest)
  - Test (E2E): `npm run test:e2e` (Playwright)

## Architecture & Structure
- **Pattern**: Clean Architecture (`internal/{domain,usecase,adapter,delivery}`).
- **Components**:
  - `cmd/etl`: Ingests eCFR XML, calculates metrics (RSCS), saves to Parquet/SQLite.
  - `cmd/api`: HTTP API serving analytics via DuckDB querying Parquet.
  - `web`: Nuxt 3 + Vue 3 frontend using USWDS design system.
- **Data**: `data/` contains `ecfr.db` (metadata) and `yyyy-mm-dd/*.parquet` (snapshots).
- **Key Libs**: `chi` (router), `duckdb-go`, `parquet-go`, `vue-uswds`.

## Code Style & Conventions
- **Go**: Standard formatting (`gofmt`). Explicit error handling. Use `internal/platform` for logging/config.
- **Frontend**: Vue 3 Composition API (`<script setup lang="ts">`).
- **Naming**: CamelCase (Go exported), camelCase (TS/JS).
- **Configuration**: Environment variables loaded from `.env` via `godotenv`.
