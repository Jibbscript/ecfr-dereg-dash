# ETL Pipeline Execution and Debugging Plan

## Goal Description
Execute, debug, and successfully complete the data ingestion ETL pipeline for the `ecfr-dereg-dash` project. The goal is to achieve a full successful run where data is validly ingested into a SQLite database and GCS Parquet files.

## User Review Required
None at this stage.

## Proposed Changes
### Configuration and Setup
- Identify `cleanup_instructions`, `gcs_configuration`, and `sqlite_db_path` from the codebase.
- Ensure the environment is correctly configured.

### Execution Loop
- **Cleanup**:
    - Remove SQLite DB: `rm -rf data/ecfr.db`
    - Remove Parquet Snapshots: `rm -rf data/20*`
    - Remove Raw Data: `rm -rf data/raw/*`
- **Execute**: Run `go run cmd/etl/main.go`.
- **Debug/Fix**: Iteratively fix any errors encountered.

## Verification Plan
### Automated Tests
- The primary verification is the successful completion of the ETL pipeline without errors.

### Manual Verification
- **SQLite**: Query the database to check for ingested data: `sqlite3 data/ecfr.db "SELECT count(*) FROM sections;"`
- **GCS**: Check the GCS bucket for generated Parquet files (if accessible) or verify local parquet generation.
