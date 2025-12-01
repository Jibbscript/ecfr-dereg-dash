package anthropic

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Jibbscript/ecfr-dereg-dashboard/internal/domain"
)

type Client struct {
	apiKey  string
	version string
	modelID string
	client  *http.Client
}

func NewClient(apiKey, version, modelID string) *Client {
	return &Client{
		apiKey:  apiKey,
		version: version,
		modelID: modelID,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type MessageRequest struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	MaxTokens  int       `json:"max_tokens"`
	System     string    `json:"system,omitempty"`
	Tools      []Tool    `json:"tools,omitempty"`
	ToolChoice string    `json:"tool_choice,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

func (c *Client) GenerateSummary(prompt string) (string, error) {
	if c.apiKey == "" {
		return "", domain.ErrAPI
	}
	reqBody := MessageRequest{
		Model:     c.modelID,
		Messages:  []Message{{Role: "user", Content: prompt}},
		MaxTokens: 1024,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", c.version)
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var res struct {
		Content []struct{ Text string } `json:"content"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if len(res.Content) > 0 {
		return res.Content[0].Text, nil
	}
	return "", domain.ErrInvalidData
}

func (c *Client) CallWithTools(prompt string, tools []Tool) (string, error) {
	return "", nil // Placeholder
}
