# eCFR Deregulation Dashboard MVP

## Live Demo

🚀 **Access the live dashboard here:** [https://ecfr-web-369420849740.us-central1.run.app](https://ecfr-web-369420849740.us-central1.run.app)

## Overview
Analytics app over eCFR corpus with RSCS metric, LSA changes, AI summaries. Backend in Go, frontend in Nuxt+USWDS, persistence in Parquet+SQLite, analytics via DuckDB.

## User Guide

### Navigation
The dashboard provides an intuitive interface to explore federal regulations:

1.  **Agency Overview**: Start here to see a list of all federal agencies.
    *   **Metrics**: View total word counts and average Regulatory Restrictions (RSCS) scores for each agency.
    *   **Sort & Filter**: Use the table headers to sort agencies by complexity or volume.
    *   **Drill Down**: Click on an Agency name to explore its specific Titles and regulations.

2.  **Title Explorer**:
    *   View detailed metrics for specific Titles of the Code of Federal Regulations.
    *   Read AI-generated summaries that distill complex regulatory text into key points, agencies involved, and scope.
    *   Navigate through the hierarchy of Parts and Sections.

3.  **Section View**:
    *   **Full Text**: Access the complete, official text of individual regulations.
    *   **Complexity Analysis**: See the RSCS score per 1,000 words to understand the regulatory burden.
    *   **AI Summary**: Quickly grasp the intent and requirements of a section without reading the full legalese.

## Setup
1. Clone repo
2. Copy `.env.template` to `.env`, fill keys
3. Local: `docker-compose up`
4. Access UI at `localhost:3000`, API at `8080`

## Usage
- ETL: Run `cmd/etl` for refresh
- Dev: DuckDB UI at `localhost:4213` (if enabled)

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
The system follows clean architecture in Go, with domain entities, use cases, adapters for external services (eCFR, GovInfo, Anthropic), and HTTP delivery. Data flows from ETL (fetches eCFR titles/sections, computes metrics, generates summaries) to Parquet partitions (by date/title) and SQLite mirror. DuckDB queries Parquet for fast aggregates in API. Frontend consumes API, rendering USWDS-compliant pages for agencies, titles, sections.
