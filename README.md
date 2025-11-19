# eCFR Deregulation Dashboard MVP

## Overview
Analytics app over eCFR corpus with RSCS metric, LSA changes, AI summaries. Backend in Go, frontend in Nuxt+USWDS, persistence in Parquet+SQLite, analytics via DuckDB.

## Setup
1. Clone repo
2. Copy `.env.template` to `.env`, fill keys
3. Local: `docker-compose up`
4. Access UI at `localhost:3000`, API at `8080`

## Usage
- ETL: Run `cmd/etl` for refresh
- Dev: DuckDB UI at `localhost:4213` (if enabled)

## Architecture
The system follows clean architecture in Go, with domain entities, use cases, adapters for external services (eCFR, GovInfo, Anthropic), and HTTP delivery. Data flows from ETL (fetches eCFR titles/sections, computes metrics, generates summaries) to Parquet partitions (by date/title) and SQLite mirror. DuckDB queries Parquet for fast aggregates in API. Frontend consumes API, rendering USWDS-compliant pages for agencies, titles, sections.
