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
	DuckDBUI        bool
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
		VertexLocation:  getEnv("VERTEX_LOCATION", "global"),
		VertexModelID:   getEnv("VERTEX_MODEL_ID", "gemini-3-pro"),
		DuckDBUI:        os.Getenv("DUCKDB_UI") == "1",
	}
}
