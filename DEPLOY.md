# Deployment Guide

## Prerequisites
- **Go**: 1.24+ (for API/ETL)
- **Node.js**: 18+ with npm (for web frontend)
- **Docker** (optional): For containerized local dev

## Local Development

### 1. Environment Setup
```bash
cp .env.template .env
```

For local development, the default values in `.env.template` work out of the box. Key variables:
- `ENV=local` or `ENV=dev` - Uses local filesystem instead of GCS, mocks Vertex AI
- `DATA_DIR=./data` - Location of SQLite DB and Parquet files
- `DUCKDB_UI=1` - Enables DuckDB Web UI at `:4213`

### 2. Data Requirements
The `data/` directory must contain:
- `ecfr.db` - SQLite database with metadata
- Date-partitioned Parquet files (e.g., `2025-11-19/*.parquet`)

To populate data, run the ETL process (see Usage section in README).

### 3a. Run Without Docker (Recommended for Development)

**Terminal 1 - API Server:**
```bash
ENV=local go run ./cmd/api
```
API runs on `http://localhost:8080`

**Terminal 2 - Frontend Dev Server:**
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
- UI: `http://localhost:3000`
- API: `http://localhost:8080`
- DuckDB UI: `http://localhost:4213` (if `DUCKDB_UI=1`)

## GCP
1. `terraform init`
2. `terraform apply -var="project_id=..." -var="gcs_bucket=..."`
3. Build/push images to GCR
4. Apply k8s manifests: deployment.yaml, cronjob.yaml, service.yaml, ingress.yaml
  - Deployment: api image, mount GCS via gcsfuse
  - CronJob: etl image, daily
  - Secrets: from Secret Manager
5. Access via ingress URL
