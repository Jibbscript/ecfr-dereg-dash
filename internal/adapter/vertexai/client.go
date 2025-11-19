package vertexai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xai/ecfr-dereg-dashboard/internal/domain"
	"google.golang.org/genai"
)

type Client struct {
	projectID string
	location  string
	modelID   string
	client    *genai.Client
}

func NewClient(ctx context.Context, projectID, location, modelID string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &Client{
		projectID: projectID,
		location:  location,
		modelID:   modelID,
		client:    client,
	}, nil
}

func NewMockClient() *Client {
	return &Client{
		projectID: "mock",
		location:  "mock",
		modelID:   "mock",
		client:    nil,
	}
}

func (c *Client) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	if c.client == nil {
		return "Mock summary for testing", nil
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.modelID, genai.Text(prompt), nil)
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", domain.ErrInvalidData
	}

	// Assuming text response
	if part := resp.Candidates[0].Content.Parts[0]; part.Text != "" {
		return part.Text, nil
	}
	return "", domain.ErrInvalidData
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

func (c *Client) CallWithTools(ctx context.Context, prompt string, tools []Tool) (string, error) {
	// Placeholder for tool use implementation with Vertex AI
	// This would involve defining FunctionDeclarations in the model config
	return "", nil
}
