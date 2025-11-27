package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/luillyfe/sourcing-agent/pkg/anthropic"
	"github.com/luillyfe/sourcing-agent/pkg/github"
)

// Run executes the sourcing agent with a user query
func Run(anthropicClient *anthropic.Client, githubClient *github.Client, query string) (string, error) {
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

	// Call Claude API
	response, err := anthropicClient.CallAPI(messages, tools)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	// Check if Claude wants to use a tool
	if response.StopReason == "tool_use" {
		// Find tool use blocks
		var toolResults []anthropic.ContentBlock

		for _, block := range response.Content {
			if block.Type == "tool_use" {
				// Execute the tool
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

		// Add assistant response to messages
		messages = append(messages, anthropic.Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Add tool results to messages
		messages = append(messages, anthropic.Message{
			Role:    "user",
			Content: toolResults,
		})

		// Call Claude again with tool results
		finalResponse, err := anthropicClient.CallAPI(messages, tools)
		if err != nil {
			return "", fmt.Errorf("failed to call Anthropic API with tool results: %w", err)
		}

		// Extract final text response
		var textParts []string
		for _, block := range finalResponse.Content {
			if block.Type == "text" {
				textParts = append(textParts, block.Text)
			}
		}

		return strings.Join(textParts, "\n"), nil
	}

	// If no tool use, just return the text response
	var textParts []string
	for _, block := range response.Content {
		if block.Type == "text" {
			textParts = append(textParts, block.Text)
		}
	}

	return strings.Join(textParts, "\n"), nil
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
