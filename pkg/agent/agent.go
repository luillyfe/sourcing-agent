package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/luillyfe/sourcing-agent/pkg/github"
	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

// Run executes the sourcing agent with a user query
func Run(client llm.Client, githubClient *github.Client, query string) (string, error) {
	// System prompt
	systemPrompt := `You are a developer sourcing assistant. Your job is to search GitHub for developers matching hiring requirements.

You have ONE tool: search_github_developers

Process:
1. Extract: programming language, location, and relevant keywords from the query
2. Call search_github_developers with appropriate parameters
3. Present the results in a clear, readable format

Keep it simple. One search, one response.`

	// Initial messages
	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("User query: %s", query),
		},
	}

	// Tools
	tools := []llm.Tool{getToolDefinition()}

	// Initial search
	fmt.Println("Analyzing query and searching GitHub...")
	resp, err := client.CallAPI(messages, tools)
	if err != nil {
		return "", fmt.Errorf("failed to call LLM API: %w", err)
	}

	// Check if LLM wants to use a tool
	if resp.StopReason == "tool_use" {
		// Find tool use blocks
		var toolResults []llm.ContentBlock

		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				fmt.Printf("Agent wants to use tool: %s\n", block.Name)

				// Execute tool
				result, err := executeTool(githubClient, block.Name, block.Input)
				if err != nil {
					return "", fmt.Errorf("failed to execute tool %s: %w", block.Name, err)
				}

				// Add tool result
				toolResults = append(toolResults, llm.ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				})
			}
		}

		// Append assistant's tool use to messages
		messages = append(messages, llm.Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		// Append tool results to messages
		messages = append(messages, llm.Message{
			Role:    "user",
			Content: toolResults,
		})

		// Call LLM again with tool results
		fmt.Println("Processing search results...")
		resp, err = client.CallAPI(messages, tools)
		if err != nil {
			return "", fmt.Errorf("failed to call LLM API with tool results: %w", err)
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
func getToolDefinition() llm.Tool {
	return llm.Tool{
		Name:        "search_github_developers",
		Description: "Search GitHub for developers matching specific criteria. Returns ready-to-present candidate profiles with their GitHub information.",
		InputSchema: llm.InputSchema{
			Type: "object",
			Properties: map[string]llm.Property{
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

// RunStage2 executes the multi-prompt sourcing agent (Stage 2)
func RunStage2(client llm.Client, githubClient *github.Client, query string) (*FinalResult, error) {
	fmt.Println("Step 1: Analyzing requirements...")
	// Step 1: Analyze Requirements
	requirements, err := analyzeRequirements(client, query)
	if err != nil {
		return nil, fmt.Errorf("requirements analysis failed: %w", err)
	}
	fmt.Printf("Requirements: %+v\n", requirements)

	// Check for unclear requirements (Fail Fast)
	if requirements.UnclearRequest {
		return nil, fmt.Errorf("request unclear: %s", requirements.ClarificationQuestion)
	}

	fmt.Println("Step 2: Generating search strategy...")
	// Step 2: Generate Search Strategy
	strategy, err := generateSearchStrategy(client, requirements)
	if err != nil {
		return nil, fmt.Errorf("strategy generation failed: %w", err)
	}
	strategyJSON, _ := json.MarshalIndent(strategy, "", "  ")
	fmt.Printf("Strategy: %s\n", string(strategyJSON))

	fmt.Println("Step 3: Finding and enriching candidates...")
	// Step 3: Find and Enrich Candidates
	enrichedCandidates, err := findAndEnrichCandidates(client, githubClient, strategy, requirements)
	if err != nil {
		return nil, fmt.Errorf("candidate search failed: %w", err)
	}
	fmt.Printf("Found %d candidates, analyzed %d\n", enrichedCandidates.SearchMetadata.TotalProfilesFound, enrichedCandidates.SearchMetadata.ProfilesAnalyzed)

	fmt.Println("Step 4: Ranking and presenting...")
	// Step 4: Rank and Present
	finalResult, err := rankAndPresent(client, enrichedCandidates, requirements)
	if err != nil {
		fmt.Printf("Ranking step failed (%v), falling back to unranked results.\n", err)
		finalResult = createFallbackResult(enrichedCandidates)
	}

	return finalResult, nil
}
