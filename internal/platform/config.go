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

func LoadConfig() Config {
	return Config{
		Env:             os.Getenv("ENV"),
		DataDir:         os.Getenv("DATA_DIR"),
		VertexProjectID: os.Getenv("VERTEX_PROJECT_ID"),
		VertexLocation:  os.Getenv("VERTEX_LOCATION"),
		VertexModelID:   os.Getenv("VERTEX_MODEL_ID"),
		DuckDBUI:        os.Getenv("DUCKDB_UI") == "1",
	}
}
