package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	anthropicAPIURL  = "https://api.anthropic.com/v1/messages"
	githubAPIURL     = "https://api.github.com/search/users"
	githubUserAPIURL = "https://api.github.com/users"
	modelName        = "claude-sonnet-4-20250514"
	maxTokens        = 4096
)

// ============================================================================
// ANTHROPIC API STRUCTURES
// ============================================================================

// AnthropicRequest represents the request payload for Anthropic API
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentBlock
}

// ContentBlock represents a content block (text or tool_use or tool_result)
type ContentBlock struct {
	Type      string      `json:"type"`
	Text      string      `json:"text,omitempty"`
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	Input     interface{} `json:"input,omitempty"`
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   string      `json:"content,omitempty"`
}

// Tool represents a tool definition for Claude
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the input parameters for a tool
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property defines a property in the tool input schema
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     int    `json:"default,omitempty"`
}

// AnthropicResponse represents the response from Anthropic API
type AnthropicResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model"`
	StopReason string         `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ============================================================================
// GITHUB API STRUCTURES
// ============================================================================

// GitHubSearchResponse represents the response from GitHub search API
type GitHubSearchResponse struct {
	TotalCount        int          `json:"total_count"`
	IncompleteResults bool         `json:"incomplete_results"`
	Items             []GitHubUser `json:"items"`
}

// GitHubUser represents a GitHub user from search results
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubUserDetail represents detailed user information
type GitHubUserDetail struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Company     string `json:"company"`
	Blog        string `json:"blog"`
	Location    string `json:"location"`
	Email       string `json:"email"`
	Bio         string `json:"bio"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	HTMLURL     string `json:"html_url"`
	AvatarURL   string `json:"avatar_url"`
}

// ============================================================================
// TOOL RESULT STRUCTURES
// ============================================================================

// Candidate represents a developer candidate
type Candidate struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Bio         string `json:"bio"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	GitHubURL   string `json:"github_url"`
	AvatarURL   string `json:"avatar_url"`
}

// SearchResult represents the complete search result
type SearchResult struct {
	Candidates     []Candidate            `json:"candidates"`
	TotalFound     int                    `json:"total_found"`
	SearchCriteria map[string]interface{} `json:"search_criteria"`
}

// ToolInput represents the input for the search_github_developers tool
type ToolInput struct {
	Language   string `json:"language"`
	Location   string `json:"location,omitempty"`
	Keywords   string `json:"keywords,omitempty"`
	MinRepos   int    `json:"min_repos"`
	MaxResults int    `json:"max_results"`
}

// ============================================================================
// GITHUB API CLIENT
// ============================================================================

// GitHubClient handles interactions with the GitHub API
type GitHubClient struct {
	BaseURL string
	Token   string
}

// NewGitHubClient creates a new GitHubClient
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		BaseURL: "https://api.github.com",
		Token:   token,
	}
}

// SearchDevelopers searches GitHub for developers matching criteria
func (c *GitHubClient) SearchDevelopers(input ToolInput) (*SearchResult, error) {
	// Set defaults
	if input.MinRepos == 0 {
		input.MinRepos = 5
	}
	if input.MaxResults == 0 {
		input.MaxResults = 10
	}

	// Build search query
	queryParts := []string{
		fmt.Sprintf("language:%s", input.Language),
		fmt.Sprintf("repos:>%d", input.MinRepos),
	}

	if input.Location != "" {
		queryParts = append(queryParts, fmt.Sprintf("location:%s", input.Location))
	}

	query := strings.Join(queryParts, "+")

	// Call GitHub Search API
	// Note: We use the BaseURL from the client
	url := fmt.Sprintf("%s/search/users?q=%s&per_page=%d", c.BaseURL, query, input.MaxResults)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
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
		return nil, fmt.Errorf("GitHub API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var searchResponse GitHubSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Enrich each user with detailed information
	candidates := []Candidate{}
	for i, user := range searchResponse.Items {
		if i >= input.MaxResults {
			break
		}

		detail, err := c.GetUserDetail(user.Login)
		if err != nil {
			// Log error but continue with other users
			fmt.Fprintf(os.Stderr, "Warning: failed to get details for user %s: %v\n", user.Login, err)
			continue
		}

		// Filter by keywords if specified
		if input.Keywords != "" {
			keywords := strings.ToLower(input.Keywords)
			bio := strings.ToLower(detail.Bio)
			if !strings.Contains(bio, keywords) {
				continue
			}
		}

		candidate := Candidate{
			Username:    detail.Login,
			Name:        detail.Name,
			Location:    detail.Location,
			Bio:         detail.Bio,
			PublicRepos: detail.PublicRepos,
			Followers:   detail.Followers,
			GitHubURL:   detail.HTMLURL,
			AvatarURL:   detail.AvatarURL,
		}

		candidates = append(candidates, candidate)
	}

	result := &SearchResult{
		Candidates: candidates,
		TotalFound: len(candidates),
		SearchCriteria: map[string]interface{}{
			"language":    input.Language,
			"location":    input.Location,
			"keywords":    input.Keywords,
			"min_repos":   input.MinRepos,
			"max_results": input.MaxResults,
		},
	}

	return result, nil
}

// GetUserDetail retrieves detailed information for a GitHub user
func (c *GitHubClient) GetUserDetail(username string) (*GitHubUserDetail, error) {
	// Note: We use the BaseURL from the client
	url := fmt.Sprintf("%s/users/%s", c.BaseURL, username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
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
		return nil, fmt.Errorf("GitHub API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userDetail GitHubUserDetail
	if err := json.Unmarshal(body, &userDetail); err != nil {
		return nil, fmt.Errorf("failed to parse user detail response: %w", err)
	}

	return &userDetail, nil
}

// ============================================================================
// TOOL EXECUTION
// ============================================================================

// executeTool executes a tool call and returns the result
func executeTool(githubClient *GitHubClient, toolName string, toolInput interface{}) (string, error) {
	if toolName != "search_github_developers" {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	// Parse the input
	inputJSON, err := json.Marshal(toolInput)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool input: %w", err)
	}

	var input ToolInput
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

// ============================================================================
// ANTHROPIC API FUNCTIONS
// ============================================================================

// getToolDefinition returns the tool definition for search_github_developers
func getToolDefinition() Tool {
	return Tool{
		Name:        "search_github_developers",
		Description: "Search GitHub for developers matching specific criteria. Returns ready-to-present candidate profiles with their GitHub information.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
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

// callAnthropicAPI calls the Anthropic API with messages and tools
func callAnthropicAPI(apiKey string, messages []Message, tools []Tool) (*AnthropicResponse, error) {
	requestBody := AnthropicRequest{
		Model:     modelName,
		MaxTokens: maxTokens,
		Messages:  messages,
		Tools:     tools,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
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

	var apiResponse AnthropicResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResponse, nil
}

// ============================================================================
// SOURCING AGENT
// ============================================================================

// runSourcingAgent executes the sourcing agent with a user query
func runSourcingAgent(anthropicKey string, githubClient *GitHubClient, query string) (string, error) {
	// System prompt
	systemPrompt := `You are a developer sourcing assistant. Your job is to search GitHub for developers matching hiring requirements.

You have ONE tool: search_github_developers

Process:
1. Extract: programming language, location, and relevant keywords from the query
2. Call search_github_developers with appropriate parameters
3. Present the results in a clear, readable format

Keep it simple. One search, one response.`

	// Initial messages
	messages := []Message{
		{
			Role:    "user",
			Content: fmt.Sprintf("%s\n\nUser query: %s", systemPrompt, query),
		},
	}

	// Tools
	tools := []Tool{getToolDefinition()}

	// Call Claude API
	response, err := callAnthropicAPI(anthropicKey, messages, tools)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}

	// Check if Claude wants to use a tool
	if response.StopReason == "tool_use" {
		// Find tool use blocks
		var toolResults []ContentBlock

		for _, block := range response.Content {
			if block.Type == "tool_use" {
				// Execute the tool
				result, err := executeTool(githubClient, block.Name, block.Input)
				if err != nil {
					return "", fmt.Errorf("failed to execute tool %s: %w", block.Name, err)
				}

				// Add tool result
				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				})
			}
		}

		// Add assistant response to messages
		messages = append(messages, Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Add tool results to messages
		messages = append(messages, Message{
			Role:    "user",
			Content: toolResults,
		})

		// Call Claude again with tool results
		finalResponse, err := callAnthropicAPI(anthropicKey, messages, tools)
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

// ============================================================================
// MAIN
// ============================================================================

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Get API keys from environment
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is not set")
		fmt.Println("Please create a .env file with your API key or set it as an environment variable")
		os.Exit(1)
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable is not set")
		fmt.Println("Please create a .env file with your GitHub token or set it as an environment variable")
		os.Exit(1)
	}

	// Check for command line arguments
	if len(os.Args) < 2 {
		fmt.Println("=== GitHub Developer Sourcing Agent ===")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run main.go \"<your query>\"")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go \"Find Go developers in Lima\"")
		fmt.Println("  go run main.go \"Looking for Python engineers in Peru\"")
		fmt.Println("  go run main.go \"Need React developers with TypeScript experience\"")
		fmt.Println()
		os.Exit(0)
	}

	// Get query from command line
	query := strings.Join(os.Args[1:], " ")

	fmt.Println("=== GitHub Developer Sourcing Agent ===")
	fmt.Printf("Query: %s\n\n", query)
	fmt.Println("Searching...")
	fmt.Println()

	// Initialize GitHub client
	githubClient := NewGitHubClient(githubToken)

	// Run the sourcing agent
	result, err := runSourcingAgent(anthropicKey, githubClient, query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display result
	fmt.Println(result)
}
