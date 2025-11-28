package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/luillyfe/sourcing-agent/pkg/anthropic"
	"github.com/luillyfe/sourcing-agent/pkg/github"
)

// LLMClient defines the interface for interacting with an LLM
type LLMClient interface {
	CallAPI(messages []anthropic.Message, tools []anthropic.Tool) (*anthropic.Response, error)
}

// Run executes the sourcing agent with a user query
func Run(client LLMClient, githubClient *github.Client, query string) (string, error) {
	// System prompt
	systemPrompt := `You are a developer sourcing assistant. Your job is to search GitHub for developers matching hiring requirements.

You have ONE tool: search_github_developers

Process:
1. Extract: programming language, location, and relevant keywords from the query
2. Call search_github_developers with appropriate parameters
3. Present the results in a clear, readable format

Keep it simple. One search, one response.`

	// Initial messages
	messages := []anthropic.Message{
		{
			Role:    "user",
			Content: fmt.Sprintf("%s\n\nUser query: %s", systemPrompt, query),
		},
	}

	// Tools
	tools := []anthropic.Tool{getToolDefinition()}

	// Initial search
	fmt.Println("Analyzing query and searching GitHub...")
	resp, err := client.CallAPI(messages, tools)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	// Check if Claude wants to use a tool
	if resp.StopReason == "tool_use" {
		// Find tool use blocks
		var toolResults []anthropic.ContentBlock

		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				fmt.Printf("Agent wants to use tool: %s\n", block.Name)

				// Execute tool
				result, err := executeTool(githubClient, block.Name, block.Input)
				if err != nil {
					return "", fmt.Errorf("failed to execute tool %s: %w", block.Name, err)
				}

				// Add tool result
				toolResults = append(toolResults, anthropic.ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				})
			}
		}

		// Append assistant's tool use to messages
		messages = append(messages, anthropic.Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		// Append tool results to messages
		messages = append(messages, anthropic.Message{
			Role:    "user",
			Content: toolResults,
		})

		// Call Anthropic again with tool results
		fmt.Println("Processing search results...")
		resp, err = client.CallAPI(messages, tools)
		if err != nil {
			return "", fmt.Errorf("failed to call Anthropic API with tool results: %w", err)
		}
	}

	// Extract text content from final response
	var finalContent strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			finalContent.WriteString(block.Text)
		}
	}

	return finalContent.String(), nil
}

// executeTool executes a tool call and returns the result
func executeTool(githubClient *github.Client, toolName string, toolInput interface{}) (string, error) {
	if toolName != "search_github_developers" {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	// Parse the input
	inputJSON, err := json.Marshal(toolInput)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool input: %w", err)
	}

	var input github.ToolInput
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return "", fmt.Errorf("failed to parse tool input: %w", err)
	}

	// Execute the search
	result, err := githubClient.SearchDevelopers(input)
	if err != nil {
		return "", fmt.Errorf("failed to search GitHub developers: %w", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// getToolDefinition returns the tool definition for search_github_developers
func getToolDefinition() anthropic.Tool {
	return anthropic.Tool{
		Name:        "search_github_developers",
		Description: "Search GitHub for developers matching specific criteria. Returns ready-to-present candidate profiles with their GitHub information.",
		InputSchema: anthropic.InputSchema{
			Type: "object",
			Properties: map[string]anthropic.Property{
				"language": {
					Type:        "string",
					Description: "Programming language (required) - e.g., 'python', 'go', 'javascript'",
				},
				"location": {
					Type:        "string",
					Description: "Geographic location (optional) - e.g., 'lima', 'peru', 'san francisco'",
				},
				"keywords": {
					Type:        "string",
					Description: "Keywords to search in user bio (optional) - e.g., 'microservices', 'mongodb', 'react'",
				},
				"min_repos": {
					Type:        "integer",
					Description: "Minimum number of public repositories (default: 5)",
					Default:     5,
				},
				"max_results": {
					Type:        "integer",
					Description: "Maximum number of candidates to return (default: 10)",
					Default:     10,
				},
			},
			Required: []string{"language"},
		},
	}
}
