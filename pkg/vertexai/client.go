package vertexai

import (
	"context"
	"fmt"

	"github.com/luillyfe/sourcing-agent/pkg/anthropic"
	"google.golang.org/genai"
)

const (
	modelName = "gemini-3-pro-preview"
)

// Client handles interactions with the Gemini API on Vertex AI
type Client struct {
	ProjectID string
	Region    string
	client    *genai.Client
}

// NewClient creates a new Vertex AI Gemini Client
func NewClient(ctx context.Context, projectID, region string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: region,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create vertexai client: %w", err)
	}

	return &Client{
		ProjectID: projectID,
		Region:    region,
		client:    client,
	}, nil
}

// Close closes the underlying client connection
// The new SDK Client doesn't have a Close method exposed in the interface shown by go doc?
// Wait, go doc didn't show Close.
// But usually clients have Close.
// If not, we can remove it or check if it's needed.
// The http client inside might need closing if custom, but default one is shared.
// Let's assume no Close for now or check if it compiles.
// Actually, `genai.Client` struct has unexported fields.
// If it doesn't have Close, we can't call it.
func (c *Client) Close() error {
	// genai.Client does not appear to have a Close method in the doc output.
	return nil
}

// CallAPI calls the Gemini API and adapts the response to Anthropic format
func (c *Client) CallAPI(messages []anthropic.Message, tools []anthropic.Tool) (*anthropic.Response, error) {
	// 1. Configure Tools
	var toolConfig *genai.Tool
	if len(tools) > 0 {
		var functionDecls []*genai.FunctionDeclaration
		for _, tool := range tools {
			functionDecls = append(functionDecls, convertTool(tool))
		}
		toolConfig = &genai.Tool{
			FunctionDeclarations: functionDecls,
		}
	}

	// 2. Convert Messages to Gemini Contents
	var contents []*genai.Content
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		parts := convertMessageContent(msg.Content)
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: parts,
		})
	}

	// 3. Generate Content
	// We pass the full history as contents.
	// We need to set Tools in the config.

	config := &genai.GenerateContentConfig{
		Temperature: float32Ptr(0),
	}
	if toolConfig != nil {
		config.Tools = []*genai.Tool{toolConfig}
	}

	resp, err := c.client.Models.GenerateContent(context.Background(), modelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// 4. Convert Response to Anthropic format
	return convertResponse(resp), nil
}

func float32Ptr(v float32) *float32 {
	return &v
}

// --- Adapter Helpers ---

func convertTool(tool anthropic.Tool) *genai.FunctionDeclaration {
	// Convert InputSchema to OpenAPI Schema
	// Anthropic InputSchema is already very similar to JSON Schema
	// We need to map it to genai.Schema

	// Simplified conversion for the specific tool we have
	// For a generic solution, we would need a recursive converter

	properties := make(map[string]*genai.Schema)
	for name, prop := range tool.InputSchema.Properties {
		propType := genai.TypeString
		if prop.Type == "integer" {
			propType = genai.TypeInteger
		}

		properties[name] = &genai.Schema{
			Type:        propType,
			Description: prop.Description,
		}
	}

	return &genai.FunctionDeclaration{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters: &genai.Schema{
			Type:       genai.TypeObject,
			Properties: properties,
			Required:   tool.InputSchema.Required,
		},
	}
}

func convertMessageContent(content interface{}) []*genai.Part {
	var parts []*genai.Part

	switch v := content.(type) {
	case string:
		parts = append(parts, &genai.Part{Text: v})
	case []anthropic.ContentBlock:
		for _, block := range v {
			switch block.Type {
			case "text":
				parts = append(parts, &genai.Part{Text: block.Text})
			case "tool_use":
				// block.Input is interface{}, likely map[string]interface{}
				var args map[string]interface{}
				if inputMap, ok := block.Input.(map[string]interface{}); ok {
					args = inputMap
				}

				part := &genai.Part{
					FunctionCall: &genai.FunctionCall{
						Name: block.Name,
						Args: args,
					},
				}

				// Restore ThoughtSignature if present
				if block.ThoughtSignature != "" {
					part.ThoughtSignature = []byte(block.ThoughtSignature)
				}

				parts = append(parts, part)

			case "tool_result":
				// FunctionResponse
				response := map[string]interface{}{
					"content": block.Content,
				}

				parts = append(parts, &genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						Name:     "search_github_developers", // Hardcoded as before
						Response: response,
					},
				})
			}
		}
	}
	return parts
}

func convertResponse(resp *genai.GenerateContentResponse) *anthropic.Response {
	anthropicResp := &anthropic.Response{
		Role: "assistant",
		Type: "message",
	}

	var content []anthropic.ContentBlock

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if part.Text != "" {
					content = append(content, anthropic.ContentBlock{
						Type: "text",
						Text: part.Text,
					})
				}

				if part.FunctionCall != nil {
					toolID := fmt.Sprintf("call_%s", part.FunctionCall.Name)

					// Capture ThoughtSignature
					var thoughtSig string
					if len(part.ThoughtSignature) > 0 {
						thoughtSig = string(part.ThoughtSignature)
					}

					content = append(content, anthropic.ContentBlock{
						Type:             "tool_use",
						Name:             part.FunctionCall.Name,
						ID:               toolID,
						Input:            part.FunctionCall.Args,
						ThoughtSignature: thoughtSig,
					})
					anthropicResp.StopReason = "tool_use"
				}
			}
		}
	}

	anthropicResp.Content = content
	if anthropicResp.StopReason == "" {
		anthropicResp.StopReason = "end_turn"
	}

	return anthropicResp
}
