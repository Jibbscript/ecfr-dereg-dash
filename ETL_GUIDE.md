# eCFR ETL Pipeline Guide

This guide details how to run the initial ETL pipeline to ingest the complete eCFR dataset from GovInfo, process it, and store the results in Google Cloud Storage (GCS) and a local SQLite database.

## Prerequisites

1.  **Google Cloud Platform (GCP) Project**: You need a GCP project with Vertex AI API enabled.
2.  **GCS Bucket**: A bucket to store the Parquet files (e.g., `ecfr-parquet`).
3.  **GCP Credentials**:
    -   Install the [gcloud CLI](https://cloud.google.com/sdk/docs/install).
    -   Authenticate locally:
        ```bash
        gcloud auth application-default login
        ```
    -   Or set `GOOGLE_APPLICATION_CREDENTIALS` to your service account key path.
4.  **Go 1.21+**: Installed locally.

## Configuration

1.  Copy the template environment file:
    ```bash
    cp .env.template .env
    ```

2.  Edit `.env` and set the following variables:
    ```env
    # GCP Project ID for Vertex AI
    VERTEX_PROJECT_ID=your-gcp-project-id
    VERTEX_LOCATION=us-central1
    VERTEX_MODEL_ID=gemini-1.5-pro-001 # or similar

    # GCS Bucket for Parquet output
    GCS_BUCKET=your-gcs-bucket-name

    # Local data directory for temporary files and SQLite DB
    DATA_DIR=./data

    # (Optional) Anthropic Key if used elsewhere, but ETL uses Vertex
    # ANTHROPIC_API_KEY=...
    ```

## Running the Pipeline

### Option 1: Run Locally (Recommended for Initial Seed)

1.  Ensure your `DATA_DIR` exists:
    ```bash
    mkdir -p data/raw
    ```

2.  Run the ETL command:
    ```bash
    go run cmd/etl/main.go
    ```

    **What happens:**
    -   The pipeline fetches the list of eCFR titles (currently hardcoded/simulated in MVP).
    -   It downloads the XML bulk data for each title from GovInfo.
    -   It parses the XML into sections.
    -   It computes metrics (Word Count, RSCS score, etc.).
    -   It generates summaries using Vertex AI (this may take time and incur costs).
    -   It writes the processed data to:
        -   **Parquet**: `gs://<GCS_BUCKET>/<date>/<title>/sections.parquet` (and diffs/summaries)
        -   **SQLite**: `./data/ecfr.db`

### Option 2: Run via Docker

1.  Build the ETL image:
    ```bash
    docker build -f Dockerfile.etl -t ecfr-etl .
    ```

2.  Run the container (mounting credentials and data):
    ```bash
    docker run --env-file .env \
      -v $(pwd)/data:/app/data \
      -v $HOME/.config/gcloud:/root/.config/gcloud \
      ecfr-etl
    ```

## Verification

### Check GCS
Verify that Parquet files are created in your bucket:
```bash
gcloud storage ls -r gs://your-gcs-bucket-name/
```

### Check SQLite
Verify the local database:
```bash
sqlite3 data/ecfr.db "SELECT count(*) FROM sections;"
```

## Troubleshooting

-   **GovInfo Download Fails**: Check your internet connection. The files are large.
-   **Vertex AI Errors**: Ensure the API is enabled in your project and your credentials have `aiplatform.user` role.
-   **Parquet Write Errors**: Ensure the GCS bucket exists and you have write permissions.
