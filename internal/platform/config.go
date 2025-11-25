package platform

import (
	"os"
)

type Config struct {
	Env             string
	DataDir         string
	VertexProjectID string
	VertexLocation  string
	VertexModelID   string
	GCSBucket       string
	DuckDBUI        bool

	// New GCS config
	ParquetBucket string
	ParquetPrefix string
	RawXMLBucket  string
	RawXMLPrefix  string
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func LoadConfig() Config {
	return Config{
		Env:             os.Getenv("ENV"),
		DataDir:         getEnv("DATA_DIR", "data"),
		VertexProjectID: os.Getenv("VERTEX_PROJECT_ID"),
		VertexLocation:  getEnv("VERTEX_LOCATION", "us-central1"),
		VertexModelID:   getEnv("VERTEX_MODEL_ID", "gemini-2.5-pro"),
		GCSBucket:       getEnv("GCS_BUCKET", "ecfr-dereg-dash-batch"),
		DuckDBUI:        os.Getenv("DUCKDB_UI") == "1",

		ParquetBucket: getEnv("PARQUET_BUCKET_NAME", "ecfr-parquet"),
		ParquetPrefix: getEnv("PARQUET_PREFIX", "parquet"),
		RawXMLBucket:  getEnv("RAW_XML_BUCKET_NAME", "ecfr-raw-xml"),
		RawXMLPrefix:  getEnv("RAW_XML_PREFIX", "raw"),
	}
}
