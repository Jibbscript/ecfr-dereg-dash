package main

import (
	"bufio"
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/xai/ecfr-dereg-dashboard/internal/adapter/parquet"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"github.com/xai/ecfr-dereg-dashboard/internal/platform"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

// BatchOutputLine represents a single line in the JSONL output file
type BatchOutputLine struct {
	Request struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	} `json:"request"`
	Response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	} `json:"response"`
}

func main() {
	_ = godotenv.Load()
	config := platform.LoadConfig()
	logger := platform.NewLogger(config.Env)
	defer logger.Sync()

	ctx := context.Background()

	logger.Info("Starting Summary Parsing Job", zap.String("bucket", config.GCSBucket))

	// Initialize GCS Client
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Fatal("Failed to create storage client", zap.Error(err))
	}
	defer client.Close()

	// Initialize Parquet Repo for saving results
	parquetRepo, err := parquet.NewRepo(ctx, config.ParquetBucket, config.ParquetPrefix)
	if err != nil {
		logger.Fatal("Failed to create Parquet repo", zap.Error(err))
	}

	bucket := client.Bucket(config.GCSBucket)
	prefix := "batch-outputs/"
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	var summaries []domain.Summary
	// Regex to extract Title number from the prompt text in the request
	titleRegex := regexp.MustCompile(`Title (\d+):`)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Fatal("Failed to list objects", zap.Error(err))
		}

		if !strings.HasSuffix(attrs.Name, ".jsonl") {
			continue
		}

		logger.Info("Processing file", zap.String("file", attrs.Name))

		rc, err := bucket.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			logger.Error("Failed to read object", zap.String("file", attrs.Name), zap.Error(err))
			continue
		}

		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			line := scanner.Bytes()
			var output BatchOutputLine
			if err := json.Unmarshal(line, &output); err != nil {
				logger.Error("Failed to unmarshal line", zap.Error(err))
				continue
			}

			// Extract Summary Text
			var summaryText string
			if len(output.Response.Candidates) > 0 && len(output.Response.Candidates[0].Content.Parts) > 0 {
				summaryText = output.Response.Candidates[0].Content.Parts[0].Text
			}

			if summaryText == "" {
				logger.Warn("Empty summary text found in line")
				continue
			}

			// Extract Title from Request
			var titleID string
			if len(output.Request.Contents) > 0 && len(output.Request.Contents[0].Parts) > 0 {
				prompt := output.Request.Contents[0].Parts[0].Text
				matches := titleRegex.FindStringSubmatch(prompt)
				if len(matches) > 1 {
					titleID = matches[1]
				}
			}

			if titleID == "" {
				// This is likely an old section-level summary batch, skipping.
				continue
			}

			summary := domain.Summary{
				Kind:      "title",
				Key:       titleID,
				Text:      summaryText,
				Model:     "gemini-2.5-pro", // As inferred from previous context
				CreatedAt: time.Now(),
			}
			summaries = append(summaries, summary)
		}
		rc.Close()
	}

	logger.Info("Parsing complete", zap.Int("summaries_found", len(summaries)))

	if len(summaries) > 0 {
		// Save to Parquet
		// We use the current date as the snapshot date for this repair job
		snapshotDate := time.Now().Format("2006-01-02")
		if err := parquetRepo.WriteSummaries(ctx, snapshotDate, summaries); err != nil {
			logger.Fatal("Failed to write summaries to Parquet", zap.Error(err))
		}
		logger.Info("Successfully saved summaries to Parquet")
	}
}
