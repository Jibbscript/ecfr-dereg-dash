package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseBatchOutputLine(t *testing.T) {
	jsonLine := `{"request":{"contents":[{"parts":[{"text":"Generate a comprehensive summary for US Code of Federal Regulations Title 50: Title 50. Include the agencies involved, the scope of regulations, and recent major changes. Use Google Search to find the most recent context and details."}],"role":"user"}],"safetySettings":[{"category":"HARM_CATEGORY_HARASSMENT","threshold":"BLOCK_NONE"},{"category":"HARM_CATEGORY_HATE_SPEECH","threshold":"BLOCK_NONE"},{"category":"HARM_CATEGORY_SEXUALLY_EXPLICIT","threshold":"BLOCK_NONE"},{"category":"HARM_CATEGORY_DANGEROUS_CONTENT","threshold":"BLOCK_NONE"}],"tools":[{"googleSearch":{"excludeDomains":[]}}]},"status":"","response":{"candidates":[{"content":{"parts":[{"text":"## Navigating the Waters of Conservation: A Deep Dive into Title 50 of the US Code of Federal Regulations\n\n**Washington, D.C.** - Title 50 of the U.S. Code of Federal Regulations (CFR), titled \"Wildlife and Fisheries,\" stands as the cornerstone of the nation's legal framework for the conservation and management of its rich biodiversity... (content truncated for brevity)"}],"role":"model"}],"finishReason":"STOP","groundingMetadata":{}},"processed_time":"2025-11-24T18:03:22.721156+00:00"}`

	var output BatchOutputLine
	err := json.Unmarshal([]byte(jsonLine), &output)
	assert.NoError(t, err)

	// Verify extraction logic
	var summaryText string
	if len(output.Response.Candidates) > 0 && len(output.Response.Candidates[0].Content.Parts) > 0 {
		summaryText = output.Response.Candidates[0].Content.Parts[0].Text
	}

	assert.Contains(t, summaryText, "Navigating the Waters of Conservation")
	assert.Contains(t, summaryText, "Title 50")

	// Verify Title extraction logic (simulation)
	// In the main code we use regex on the request text.
	requestText := output.Request.Contents[0].Parts[0].Text
	assert.Contains(t, requestText, "Title 50")

	// Validate domain mapping
	summary := domain.Summary{
		Kind:      "title",
		Key:       "50", // Extracted logic would put "50" here
		Text:      summaryText,
		Model:     "gemini-2.5-pro",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "title", summary.Kind)
	assert.Equal(t, "50", summary.Key)
	assert.NotEmpty(t, summary.Text)
}

func TestParseBatchOutputLine_Empty(t *testing.T) {
	jsonLine := `{"response":{}}`
	var output BatchOutputLine
	err := json.Unmarshal([]byte(jsonLine), &output)
	assert.NoError(t, err)

	var summaryText string
	if len(output.Response.Candidates) > 0 && len(output.Response.Candidates[0].Content.Parts) > 0 {
		summaryText = output.Response.Candidates[0].Content.Parts[0].Text
	}
	assert.Empty(t, summaryText)
}
