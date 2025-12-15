package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

const (
	apiURL    = "https://api.anthropic.com/v1/messages"
	modelName = "claude-sonnet-4-20250514"
	maxTokens = 4096
)

// Client handles interactions with the Anthropic API
type Client struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Anthropic Client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// CallAPI calls the Anthropic API with messages and tools
func (c *Client) CallAPI(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	// Convert llm.Message to anthropic.Message
	var anthropicMessages []Message
	for _, msg := range messages {
		anthropicMessages = append(anthropicMessages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Convert llm.Tool to anthropic.Tool
	var anthropicTools []Tool
	for _, tool := range tools {
		// Convert llm.InputSchema to anthropic.InputSchema
		props := make(map[string]Property)
		for k, v := range tool.InputSchema.Properties {
			props[k] = Property{
				Type:        v.Type,
				Description: v.Description,
				Default:     v.Default,
			}
		}

		anthropicTools = append(anthropicTools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: InputSchema{
				Type:       tool.InputSchema.Type,
				Properties: props,
				Required:   tool.InputSchema.Required,
			},
		})
	}

	requestBody := Request{
		Model:     modelName,
		MaxTokens: maxTokens,
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
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

	client := c.HTTPClient
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

	// Convert anthropic.Response to llm.Response
	var content []llm.ContentBlock
	for _, block := range apiResponse.Content {
		content = append(content, llm.ContentBlock{
			Type:             block.Type,
			Text:             block.Text,
			ID:               block.ID,
			Name:             block.Name,
			Input:            block.Input,
			ToolUseID:        block.ToolUseID,
			Content:          block.Content,
			ThoughtSignature: block.ThoughtSignature,
		})
	}

	return &llm.Response{
		ID:         apiResponse.ID,
		Type:       apiResponse.Type,
		Role:       apiResponse.Role,
		Content:    content,
		Model:      apiResponse.Model,
		StopReason: apiResponse.StopReason,
		Usage: llm.Usage{
			InputTokens:  apiResponse.Usage.InputTokens,
			OutputTokens: apiResponse.Usage.OutputTokens,
		},
	}, nil
}
