package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiURL    = "https://api.anthropic.com/v1/messages"
	modelName = "claude-sonnet-4-20250514"
	maxTokens = 4096
)

// Client handles interactions with the Anthropic API
type Client struct {
	APIKey string
}

// NewClient creates a new Anthropic Client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
	}
}

// CallAPI calls the Anthropic API with messages and tools
func (c *Client) CallAPI(messages []Message, tools []Tool) (*Response, error) {
	requestBody := Request{
		Model:     modelName,
		MaxTokens: maxTokens,
		Messages:  messages,
		Tools:     tools,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse Response
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResponse, nil
}
