package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/adapter/sqlite"
	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
)

// JSONL structures matching Vertex AI batch prediction output
type batchOutputLine struct {
	Request  *batchRequest  `json:"request"`
	Response *batchResponse `json:"response"`
	Status   string         `json:"status"`
}

type batchRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type batchResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func main() {
	// Configuration
	dataDir := getEnv("DATA_DIR", "./data")
	summariesDir := filepath.Join(dataDir, "summaries")
	dbPath := filepath.Join(dataDir, "ecfr.db")

	fmt.Printf("Importing summaries from: %s\n", summariesDir)
	fmt.Printf("Database path: %s\n", dbPath)

	// Initialize SQLite repo (this will create the summaries table if needed)
	repo, err := sqlite.NewRepo(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}

	// Find all JSONL files recursively
	var jsonlFiles []string
	err = filepath.Walk(summariesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jsonl") {
			jsonlFiles = append(jsonlFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to walk directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d JSONL files\n", len(jsonlFiles))

	// Regex to extract Title number from prompt or summary text
	titleRegex := regexp.MustCompile(`Title (\d+)`)

	// Track unique titles to avoid duplicates (use the latest one)
	titleSummaries := make(map[string]domain.Summary)
	var parseErrors int

	for _, file := range jsonlFiles {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open file %s: %v\n", file, err)
			continue
		}

		scanner := bufio.NewScanner(f)
		// Increase buffer size for large JSONL lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			var line batchOutputLine
			if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing line %d in %s: %v\n", lineNum, file, err)
				parseErrors++
				continue
			}

			// Extract summary text from response
			var summaryText string
			if line.Response != nil && len(line.Response.Candidates) > 0 {
				if len(line.Response.Candidates[0].Content.Parts) > 0 {
					summaryText = line.Response.Candidates[0].Content.Parts[0].Text
				}
			}

			if summaryText == "" {
				fmt.Fprintf(os.Stderr, "No summary text in line %d of %s\n", lineNum, file)
				parseErrors++
				continue
			}

			// Extract Title number from request prompt first, then from summary text
			var titleKey string
			if line.Request != nil && len(line.Request.Contents) > 0 {
				if len(line.Request.Contents[0].Parts) > 0 {
					prompt := line.Request.Contents[0].Parts[0].Text
					if matches := titleRegex.FindStringSubmatch(prompt); len(matches) > 1 {
						titleKey = matches[1]
					}
				}
			}

			// Fallback: extract from summary text
			if titleKey == "" {
				if matches := titleRegex.FindStringSubmatch(summaryText); len(matches) > 1 {
					titleKey = matches[1]
				}
			}

			if titleKey == "" {
				fmt.Fprintf(os.Stderr, "Could not extract Title number from line %d in %s\n", lineNum, file)
				parseErrors++
				continue
			}

			// Create summary object
			summary := domain.Summary{
				Kind:      "title",
				Key:       titleKey,
				Text:      summaryText,
				Model:     "gemini-2.5-pro",
				CreatedAt: time.Now(),
			}

			// Store in map (later files will overwrite earlier ones)
			titleSummaries[titleKey] = summary
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
		}

		f.Close()
	}

	fmt.Printf("Extracted %d unique title summaries\n", len(titleSummaries))
	if parseErrors > 0 {
		fmt.Printf("Parse errors: %d\n", parseErrors)
	}

	// Convert map to slice
	var summaries []domain.Summary
	for _, s := range titleSummaries {
		summaries = append(summaries, s)
	}

	// Insert into database
	if len(summaries) > 0 {
		if err := repo.InsertSummaries(summaries); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to insert summaries: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully imported %d summaries into database\n", len(summaries))
	} else {
		fmt.Println("No summaries to import")
	}

	// Verify by listing inserted summaries
	inserted, err := repo.GetAllSummaries()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to verify summaries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSummaries in database: %d\n", len(inserted))
	for _, s := range inserted {
		textPreview := s.Text
		if len(textPreview) > 100 {
			textPreview = textPreview[:100] + "..."
		}
		fmt.Printf("  Title %s: %s\n", s.Key, textPreview)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
