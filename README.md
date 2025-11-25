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

## Architecture
The system follows clean architecture in Go, with domain entities, use cases, adapters for external services (eCFR, GovInfo, Anthropic), and HTTP delivery. Data flows from ETL (fetches eCFR titles/sections, computes metrics, generates summaries) to Parquet partitions (by date/title) and SQLite mirror. DuckDB queries Parquet for fast aggregates in API. Frontend consumes API, rendering USWDS-compliant pages for agencies, titles, sections.
