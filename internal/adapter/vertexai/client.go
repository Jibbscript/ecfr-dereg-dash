package vertexai

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"cloud.google.com/go/storage"
	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/genai"
)

type Client struct {
	projectID     string
	location      string
	modelID       string
	gcsBucket     string
	client        *genai.Client
	jobClient     *aiplatform.JobClient
	storageClient *storage.Client
}

func NewClient(ctx context.Context, projectID, location, modelID, gcsBucket string) (*Client, error) {
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// For Batch Prediction, we need the aiplatform JobClient
	endpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", location)
	jobClient, err := aiplatform.NewJobClient(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create aiplatform job client: %w", err)
	}

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &Client{
		projectID:     projectID,
		location:      location,
		modelID:       modelID,
		gcsBucket:     gcsBucket,
		client:        genaiClient,
		jobClient:     jobClient,
		storageClient: storageClient,
	}, nil
}

func NewMockClient() *Client {
	return &Client{
		projectID: "mock",
		location:  "mock",
		modelID:   "mock",
		gcsBucket: "mock-bucket",
		client:    nil,
	}
}

// GenerateSummary performs online prediction (for fallbacks or small batches)
func (c *Client) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	if c.client == nil {
		return "Mock summary for testing", nil
	}

	// Configure safety settings
	safetySettings := []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockThresholdBlockNone},
	}

	// Add Grounding with Google Search
	tools := []*genai.Tool{
		{GoogleSearch: &genai.GoogleSearch{}},
	}

	config := &genai.GenerateContentConfig{
		SafetySettings: safetySettings,
		Tools:          tools,
	}

	// Retry logic... (same as before)
	// ... omitted for brevity as we are moving to batch, but good to keep if we need online fallback
	// For now, simple call:
	resp, err := c.client.Models.GenerateContent(ctx, c.modelID, genai.Text(prompt), config)
	if err != nil {
		return "", err
	}

	if text := resp.Text(); text != "" {
		return text, nil
	}
	return "", domain.ErrInvalidData
}

// BatchGenerateSummaries orchestrates the batch prediction flow
func (c *Client) BatchGenerateSummaries(ctx context.Context, prompts []string) ([]string, error) {
	if c.client == nil {
		// Mock behavior for tests
		results := make([]string, len(prompts))
		for i := range prompts {
			results[i] = "Mock batch summary"
		}
		return results, nil
	}

	if len(prompts) == 0 {
		return nil, nil
	}

	// 1. Create JSONL input
	timestamp := time.Now().Format("20060102-150405")
	uid := make([]byte, 4)
	rand.Read(uid)
	uniqueID := hex.EncodeToString(uid)

	inputFileName := fmt.Sprintf("batch-inputs/%s-%s.jsonl", timestamp, uniqueID)
	inputURI := fmt.Sprintf("gs://%s/%s", c.gcsBucket, inputFileName)

	if err := c.uploadBatchInput(ctx, prompts, inputFileName); err != nil {
		return nil, fmt.Errorf("failed to upload batch input: %w", err)
	}

	// 2. Submit Batch Job
	outputPrefix := fmt.Sprintf("gs://%s/batch-outputs/%s-%s", c.gcsBucket, timestamp, uniqueID)
	jobName, err := c.submitBatchJob(ctx, inputURI, outputPrefix, uniqueID)
	if err != nil {
		return nil, fmt.Errorf("failed to submit batch job: %w", err)
	}
	fmt.Printf("Batch job submitted: %s\n", jobName)

	// 3. Poll for completion
	job, err := c.pollBatchJob(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("batch job failed: %w", err)
	}

	// 4. Download and parse results from the *actual* output dir
	gcsOutputDir := job.GetOutputInfo().GetGcsOutputDirectory()
	if gcsOutputDir == "" {
		return nil, fmt.Errorf("batch job succeeded but gcsOutputDirectory is empty")
	}

	// Retry loop for eventual consistency of GCS listing
	var results []string
	for i := 0; i < 5; i++ {
		results, err = c.downloadBatchResults(ctx, gcsOutputDir, len(prompts))
		if err == nil && len(results) > 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return results, err
}

type batchRequest struct {
	Request *batchRequestBody `json:"request"`
}

type batchRequestBody struct {
	Contents       []*genai.Content       `json:"contents"`
	SafetySettings []*genai.SafetySetting `json:"safetySettings,omitempty"`
	Tools          []*batchTool           `json:"tools,omitempty"`
}

type batchTool struct {
	GoogleSearch *batchGoogleSearch `json:"googleSearch,omitempty"`
}

type batchGoogleSearch struct {
	// Using ExcludeDomains ensures the struct isn't empty, preventing BigQuery schema errors.
	ExcludeDomains []string `json:"excludeDomains"`
}

func (c *Client) uploadBatchInput(ctx context.Context, prompts []string, fileName string) error {
	wc := c.storageClient.Bucket(c.gcsBucket).Object(fileName).NewWriter(ctx)
	defer wc.Close()

	encoder := json.NewEncoder(wc)

	safetySettings := []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockThresholdBlockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockThresholdBlockNone},
	}

	// Use batch-specific tool struct to ensure JSON output is not empty object {}
	tools := []*batchTool{
		{
			GoogleSearch: &batchGoogleSearch{
				ExcludeDomains: []string{},
			},
		},
	}

	for _, prompt := range prompts {
		body := &batchRequestBody{
			Contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: prompt},
					},
				},
			},
			SafetySettings: safetySettings,
			Tools:          tools,
		}

		req := batchRequest{
			Request: body,
		}

		if err := encoder.Encode(req); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) submitBatchJob(ctx context.Context, inputURI, outputPrefix, uniqueID string) (string, error) {
	// NOTE: Model resource name for publisher models must be fully qualified:
	// projects/{project}/locations/{location}/publishers/google/models/{model}
	// Ensure modelID is correct for gemini-3-pro, if passed explicitly. 
	// If c.modelID is just "gemini-3-pro", we use it.
	// We might want to update config to "gemini-3-pro" but for now, we rely on c.modelID being set correctly.
	modelName := fmt.Sprintf("publishers/google/models/%s", c.modelID)

	req := &aiplatformpb.CreateBatchPredictionJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", c.projectID, c.location),
		BatchPredictionJob: &aiplatformpb.BatchPredictionJob{
			DisplayName: fmt.Sprintf("ecfr-batch-%d-%s", time.Now().Unix(), uniqueID),
			Model:       modelName,
			InputConfig: &aiplatformpb.BatchPredictionJob_InputConfig{
				InstancesFormat: "jsonl",
				Source: &aiplatformpb.BatchPredictionJob_InputConfig_GcsSource{
					GcsSource: &aiplatformpb.GcsSource{Uris: []string{inputURI}},
				},
			},
			OutputConfig: &aiplatformpb.BatchPredictionJob_OutputConfig{
				PredictionsFormat: "jsonl",
				Destination: &aiplatformpb.BatchPredictionJob_OutputConfig_GcsDestination{
					GcsDestination: &aiplatformpb.GcsDestination{
						OutputUriPrefix: outputPrefix,
					},
				},
			},
		},
	}

	job, err := c.jobClient.CreateBatchPredictionJob(ctx, req)
	if err != nil {
		return "", err
	}
	return job.GetName(), nil
}

func (c *Client) pollBatchJob(ctx context.Context, jobName string) (*aiplatformpb.BatchPredictionJob, error) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := c.jobClient.GetBatchPredictionJob(ctx, &aiplatformpb.GetBatchPredictionJobRequest{
				Name: jobName,
			})
			if err != nil {
				return nil, err
			}

			switch job.GetState() {
			case aiplatformpb.JobState_JOB_STATE_SUCCEEDED:
				return job, nil
			case aiplatformpb.JobState_JOB_STATE_FAILED, aiplatformpb.JobState_JOB_STATE_CANCELLED:
				return nil, fmt.Errorf("job failed with state: %s, error: %v", job.GetState(), job.GetError())
			}
			// Continue polling
		}
	}
}

type batchOutputLine struct {
	Response *genai.GenerateContentResponse `json:"response"`
	Status   json.RawMessage                `json:"status,omitempty"`
}

func (c *Client) downloadBatchResults(ctx context.Context, outputURI string, expectedCount int) ([]string, error) {
	// outputURI is gs://bucket/path/to/prediction-.../
	bucketName, prefix, err := parseGCSURI(outputURI)
	if err != nil {
		return nil, fmt.Errorf("invalid output URI %q: %w", outputURI, err)
	}

	// fmt.Printf("Listing objects in bucket: %s, prefix: %s\n", bucketName, prefix)

	it := c.storageClient.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: prefix})
	
	var results []string

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// fmt.Printf("Found object: %s\n", attrs.Name)

		if !strings.HasSuffix(attrs.Name, ".jsonl") {
			continue
		}

		rc, err := c.storageClient.Bucket(bucketName).Object(attrs.Name).NewReader(ctx)
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			var line batchOutputLine
			if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
				fmt.Printf("Error unmarshaling batch output line: %v\n", err)
				continue
			}
			
			if len(line.Status) > 0 && string(line.Status) != "{}" && string(line.Status) != "null" {
				// Failed row
				results = append(results, "") 
				continue
			}
			
			if line.Response != nil {
				text := line.Response.Text()
				results = append(results, text)
			} else {
				results = append(results, "")
			}
		}
	}

	return results, nil
}

func parseGCSURI(uri string) (bucket, prefix string, err error) {
	const scheme = "gs://"
	if !strings.HasPrefix(uri, scheme) {
		return "", "", fmt.Errorf("GCS URI %q must start with %q", uri, scheme)
	}
	remain := strings.TrimPrefix(uri, scheme)
	parts := strings.SplitN(remain, "/", 2)
	bucket = parts[0]
	if bucket == "" {
		return "", "", fmt.Errorf("GCS URI %q missing bucket", uri)
	}
	if len(parts) == 2 {
		prefix = parts[1]
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
	}
	return bucket, prefix, nil
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

func (c *Client) CallWithTools(ctx context.Context, prompt string, tools []Tool) (string, error) {
	// Placeholder for tool use implementation with Vertex AI
	return "", nil
}
