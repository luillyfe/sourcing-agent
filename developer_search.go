package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// GitHub API structures
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
	PublicRepos int  `json:"public_repos"`
}

type GitHubSearchResponse struct {
	TotalCount int          `json:"total_count"`
	Items      []GitHubUser `json:"items"`
}

type GitHubRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	StargazersCount int `json:"stargazers_count"`
	ForksCount  int    `json:"forks_count"`
	Topics      []string `json:"topics"`
	HTMLURL     string `json:"html_url"`
}

// Claude API structures for function calling
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type ToolUse struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type ContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
}

type ClaudeMessage struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ClaudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Tools     []ToolDefinition `json:"tools,omitempty"`
	Messages  []ClaudeMessage  `json:"messages"`
}

type ClaudeResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Candidate represents a ranked developer candidate
type Candidate struct {
	Username           string
	Name               string
	Location           string
	Bio                string
	ProfileURL         string
	Score              float64
	MicroservicesRepos []string
	RelevantRepos      []GitHubRepo
}

// GitHub API client functions

func searchGitHubDevelopers(language, location string, limit int) ([]GitHubUser, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	// Build search query
	query := "type:user"
	if language != "" {
		query += fmt.Sprintf(" language:%s", language)
	}
	if location != "" {
		query += fmt.Sprintf(" location:%s", location)
	}

	url := fmt.Sprintf("https://api.github.com/search/users?q=%s&per_page=%d", query, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if githubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	}

	client := &http.Client{}
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

	var searchResp GitHubSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return searchResp.Items, nil
}

func getUserRepos(username string, limit int) ([]GitHubRepo, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=%d&sort=stars", username, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if githubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	}

	client := &http.Client{}
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

	var repos []GitHubRepo
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return repos, nil
}

// Tool definitions for Claude
func getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "search_github_developers",
			Description: "Search for GitHub developers based on programming language and location. Returns a list of matching users.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Programming language (e.g., 'go', 'python', 'javascript')",
					},
					"location": map[string]interface{}{
						"type":        "string",
						"description": "Location to search for developers (e.g., 'lima', 'san francisco', 'berlin')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of developers to return (default: 10)",
					},
				},
				"required": []string{"language", "location"},
			},
		},
		{
			Name:        "get_user_repos",
			Description: "Get the repositories for a specific GitHub user. Returns detailed information about their projects.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type":        "string",
						"description": "GitHub username to get repositories for",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of repositories to return (default: 20)",
					},
				},
				"required": []string{"username"},
			},
		},
	}
}

// Execute tool calls
func executeTool(toolName string, input map[string]interface{}) (string, error) {
	switch toolName {
	case "search_github_developers":
		language, _ := input["language"].(string)
		location, _ := input["location"].(string)
		limit := 10
		if l, ok := input["limit"].(float64); ok {
			limit = int(l)
		}

		users, err := searchGitHubDevelopers(language, location, limit)
		if err != nil {
			return "", err
		}

		result, err := json.Marshal(users)
		if err != nil {
			return "", err
		}
		return string(result), nil

	case "get_user_repos":
		username, _ := input["username"].(string)
		limit := 20
		if l, ok := input["limit"].(float64); ok {
			limit = int(l)
		}

		repos, err := getUserRepos(username, limit)
		if err != nil {
			return "", err
		}

		result, err := json.Marshal(repos)
		if err != nil {
			return "", err
		}
		return string(result), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// Call Claude API with function calling support
func callClaudeWithTools(apiKey string, messages []ClaudeMessage, tools []ToolDefinition) (*ClaudeResponse, error) {
	requestBody := ClaudeRequest{
		Model:     modelName,
		MaxTokens: 4096,
		Tools:     tools,
		Messages:  messages,
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

	client := &http.Client{}
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

	var apiResponse ClaudeResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResponse, nil
}

// Rank candidates based on microservices experience
func rankCandidates(users []GitHubUser, reposMap map[string][]GitHubRepo, keywords []string) []Candidate {
	var candidates []Candidate

	for _, user := range users {
		repos, ok := reposMap[user.Login]
		if !ok {
			continue
		}

		candidate := Candidate{
			Username:      user.Login,
			Name:          user.Name,
			Location:      user.Location,
			Bio:           user.Bio,
			ProfileURL:    user.HTMLURL,
			Score:         0,
			RelevantRepos: []GitHubRepo{},
		}

		// Score based on repository analysis
		for _, repo := range repos {
			repoScore := 0.0
			isRelevant := false

			// Check description and topics for keywords
			repoText := strings.ToLower(repo.Description + " " + strings.Join(repo.Topics, " "))

			for _, keyword := range keywords {
				if strings.Contains(repoText, strings.ToLower(keyword)) {
					repoScore += 10
					isRelevant = true
					candidate.MicroservicesRepos = append(candidate.MicroservicesRepos, repo.Name)
				}
			}

			// Add bonus for stars and forks
			repoScore += float64(repo.StargazersCount) * 0.5
			repoScore += float64(repo.ForksCount) * 0.3

			if isRelevant {
				candidate.RelevantRepos = append(candidate.RelevantRepos, repo)
			}

			candidate.Score += repoScore
		}

		if candidate.Score > 0 {
			candidates = append(candidates, candidate)
		}
	}

	// Sort candidates by score (bubble sort for simplicity)
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Score > candidates[i].Score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	return candidates
}

// Main workflow orchestration
func runDeveloperSearchWorkflow(apiKey, userQuery string) error {
	fmt.Println("\n🔍 Analyzing your search query...")

	// Initial message to Claude
	messages := []ClaudeMessage{
		{
			Role: "user",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: fmt.Sprintf(`You are a developer sourcing assistant. The user wants to find developers with specific skills.

User query: "%s"

Your task:
1. First, use search_github_developers to find developers matching the language and location
2. Then, for each developer found, use get_user_repos to analyze their repositories
3. Look for evidence of the requested skills in their repositories (e.g., microservices, docker, kubernetes, etc.)
4. Provide a summary of the best candidates

Start by calling search_github_developers with the appropriate parameters.`, userQuery),
				},
			},
		},
	}

	tools := getToolDefinitions()
	maxIterations := 10
	iteration := 0

	// Agentic loop: keep calling Claude until it stops requesting tools
	for iteration < maxIterations {
		iteration++

		response, err := callClaudeWithTools(apiKey, messages, tools)
		if err != nil {
			return fmt.Errorf("failed to call Claude: %w", err)
		}

		// Add Claude's response to message history
		messages = append(messages, ClaudeMessage{
			Role:    "assistant",
			Content: response.Content,
		})

		// Check stop reason
		if response.StopReason == "end_turn" {
			// Claude is done, extract final response
			for _, block := range response.Content {
				if block.Type == "text" {
					fmt.Println("\n" + strings.Repeat("=", 60))
					fmt.Println("🎯 RESULTS")
					fmt.Println(strings.Repeat("=", 60))
					fmt.Println(block.Text)
					fmt.Println(strings.Repeat("=", 60))
				}
			}
			break
		}

		if response.StopReason == "tool_use" {
			// Execute tools and add results
			var toolResults []ContentBlock

			for _, block := range response.Content {
				if block.Type == "tool_use" {
					fmt.Printf("\n📞 Calling tool: %s\n", block.Name)

					result, err := executeTool(block.Name, block.Input)
					if err != nil {
						result = fmt.Sprintf("Error: %v", err)
					}

					toolResults = append(toolResults, ContentBlock{
						Type:      "tool_result",
						ToolUseID: block.ID,
						Content:   result,
					})

					fmt.Printf("✓ Tool executed successfully\n")
				}
			}

			// Add tool results as user message
			messages = append(messages, ClaudeMessage{
				Role:    "user",
				Content: toolResults,
			})
		}
	}

	if iteration >= maxIterations {
		fmt.Println("\n⚠ Maximum iterations reached")
	}

	return nil
}
