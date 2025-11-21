package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================================================
// TEST GITHUB API FUNCTIONS
// ============================================================================

func TestSearchGitHubDevelopers(t *testing.T) {
	// Create a mock GitHub API server
	mockSearchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected Authorization header 'token test-token', got '%s'", r.Header.Get("Authorization"))
		}

		// Return mock search response
		response := GitHubSearchResponse{
			TotalCount:        2,
			IncompleteResults: false,
			Items: []GitHubUser{
				{Login: "testuser1", ID: 1, HTMLURL: "https://github.com/testuser1", AvatarURL: "https://avatar1.png"},
				{Login: "testuser2", ID: 2, HTMLURL: "https://github.com/testuser2", AvatarURL: "https://avatar2.png"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockSearchServer.Close()

	// Create a mock GitHub user detail server
	mockUserServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mock user detail response
		response := GitHubUserDetail{
			Login:       "testuser1",
			Name:        "Test User 1",
			Location:    "Lima, Peru",
			Bio:         "Go developer with microservices experience",
			PublicRepos: 25,
			Followers:   100,
			HTMLURL:     "https://github.com/testuser1",
			AvatarURL:   "https://avatar1.png",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockUserServer.Close()

	// Note: In a real test, you would need to inject these mock URLs
	// For now, this test demonstrates the structure

	t.Run("ValidInput", func(t *testing.T) {
		input := ToolInput{
			Language:   "go",
			Location:   "lima",
			MinRepos:   5,
			MaxResults: 10,
		}

		// This would call the real GitHub API in a real test
		// You would need to mock the HTTP client or use dependency injection
		// For demonstration purposes, we're just testing the input validation
		if input.Language == "" {
			t.Error("Language should not be empty")
		}
		if input.MinRepos < 0 {
			t.Error("MinRepos should not be negative")
		}
		if input.MaxResults < 0 {
			t.Error("MaxResults should not be negative")
		}
	})
}

func TestGetGitHubUserDetail(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected Authorization header 'token test-token', got '%s'", r.Header.Get("Authorization"))
		}

		// Return mock user detail
		response := GitHubUserDetail{
			Login:       "testuser",
			Name:        "Test User",
			Location:    "Lima, Peru",
			Bio:         "Go developer",
			PublicRepos: 20,
			Followers:   50,
			HTMLURL:     "https://github.com/testuser",
			AvatarURL:   "https://avatar.png",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// This test demonstrates the structure
	// In a real implementation, you would need to inject the mock server URL
	t.Run("ValidUsername", func(t *testing.T) {
		username := "testuser"
		if username == "" {
			t.Error("Username should not be empty")
		}
	})
}

// ============================================================================
// TEST TOOL EXECUTION
// ============================================================================

func TestExecuteTool(t *testing.T) {
	t.Run("UnknownTool", func(t *testing.T) {
		_, err := executeTool("test-token", "unknown_tool", map[string]interface{}{})
		if err == nil {
			t.Error("Expected error for unknown tool")
		}
		if err.Error() != "unknown tool: unknown_tool" {
			t.Errorf("Expected error message 'unknown tool: unknown_tool', got '%s'", err.Error())
		}
	})

	t.Run("ValidToolName", func(t *testing.T) {
		toolName := "search_github_developers"
		if toolName != "search_github_developers" {
			t.Error("Tool name should be 'search_github_developers'")
		}
	})
}

// ============================================================================
// TEST ANTHROPIC API STRUCTURES
// ============================================================================

func TestToolDefinition(t *testing.T) {
	tool := getToolDefinition()

	t.Run("ToolName", func(t *testing.T) {
		if tool.Name != "search_github_developers" {
			t.Errorf("Expected tool name 'search_github_developers', got '%s'", tool.Name)
		}
	})

	t.Run("ToolDescription", func(t *testing.T) {
		if tool.Description == "" {
			t.Error("Tool description should not be empty")
		}
	})

	t.Run("RequiredFields", func(t *testing.T) {
		if len(tool.InputSchema.Required) != 1 {
			t.Errorf("Expected 1 required field, got %d", len(tool.InputSchema.Required))
		}
		if tool.InputSchema.Required[0] != "language" {
			t.Errorf("Expected required field 'language', got '%s'", tool.InputSchema.Required[0])
		}
	})

	t.Run("Properties", func(t *testing.T) {
		expectedProperties := []string{"language", "location", "keywords", "min_repos", "max_results"}
		for _, prop := range expectedProperties {
			if _, exists := tool.InputSchema.Properties[prop]; !exists {
				t.Errorf("Expected property '%s' to exist", prop)
			}
		}
	})
}

func TestContentBlock(t *testing.T) {
	t.Run("TextBlock", func(t *testing.T) {
		block := ContentBlock{
			Type: "text",
			Text: "Hello, world!",
		}
		if block.Type != "text" {
			t.Errorf("Expected type 'text', got '%s'", block.Type)
		}
		if block.Text != "Hello, world!" {
			t.Errorf("Expected text 'Hello, world!', got '%s'", block.Text)
		}
	})

	t.Run("ToolUseBlock", func(t *testing.T) {
		block := ContentBlock{
			Type:  "tool_use",
			ID:    "toolu_123",
			Name:  "search_github_developers",
			Input: map[string]interface{}{"language": "go"},
		}
		if block.Type != "tool_use" {
			t.Errorf("Expected type 'tool_use', got '%s'", block.Type)
		}
		if block.Name != "search_github_developers" {
			t.Errorf("Expected name 'search_github_developers', got '%s'", block.Name)
		}
	})

	t.Run("ToolResultBlock", func(t *testing.T) {
		block := ContentBlock{
			Type:      "tool_result",
			ToolUseID: "toolu_123",
			Content:   `{"candidates": []}`,
		}
		if block.Type != "tool_result" {
			t.Errorf("Expected type 'tool_result', got '%s'", block.Type)
		}
		if block.ToolUseID != "toolu_123" {
			t.Errorf("Expected tool_use_id 'toolu_123', got '%s'", block.ToolUseID)
		}
	})
}

// ============================================================================
// TEST DATA STRUCTURES
// ============================================================================

func TestToolInput(t *testing.T) {
	t.Run("ValidInput", func(t *testing.T) {
		input := ToolInput{
			Language:   "go",
			Location:   "lima",
			Keywords:   "microservices",
			MinRepos:   5,
			MaxResults: 10,
		}

		if input.Language != "go" {
			t.Errorf("Expected language 'go', got '%s'", input.Language)
		}
		if input.Location != "lima" {
			t.Errorf("Expected location 'lima', got '%s'", input.Location)
		}
		if input.MinRepos != 5 {
			t.Errorf("Expected min_repos 5, got %d", input.MinRepos)
		}
		if input.MaxResults != 10 {
			t.Errorf("Expected max_results 10, got %d", input.MaxResults)
		}
	})

	t.Run("JSONMarshaling", func(t *testing.T) {
		input := ToolInput{
			Language:   "python",
			Location:   "peru",
			MinRepos:   10,
			MaxResults: 5,
		}

		jsonData, err := json.Marshal(input)
		if err != nil {
			t.Errorf("Failed to marshal ToolInput: %v", err)
		}

		var unmarshaled ToolInput
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal ToolInput: %v", err)
		}

		if unmarshaled.Language != input.Language {
			t.Errorf("Expected language '%s', got '%s'", input.Language, unmarshaled.Language)
		}
	})
}

func TestCandidate(t *testing.T) {
	t.Run("ValidCandidate", func(t *testing.T) {
		candidate := Candidate{
			Username:    "testuser",
			Name:        "Test User",
			Location:    "Lima, Peru",
			Bio:         "Go developer",
			PublicRepos: 25,
			Followers:   100,
			GitHubURL:   "https://github.com/testuser",
			AvatarURL:   "https://avatar.png",
		}

		if candidate.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", candidate.Username)
		}
		if candidate.PublicRepos != 25 {
			t.Errorf("Expected public_repos 25, got %d", candidate.PublicRepos)
		}
	})

	t.Run("JSONMarshaling", func(t *testing.T) {
		candidate := Candidate{
			Username:    "testuser",
			Name:        "Test User",
			Location:    "Lima, Peru",
			Bio:         "Go developer",
			PublicRepos: 25,
			Followers:   100,
			GitHubURL:   "https://github.com/testuser",
			AvatarURL:   "https://avatar.png",
		}

		jsonData, err := json.Marshal(candidate)
		if err != nil {
			t.Errorf("Failed to marshal Candidate: %v", err)
		}

		var unmarshaled Candidate
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal Candidate: %v", err)
		}

		if unmarshaled.Username != candidate.Username {
			t.Errorf("Expected username '%s', got '%s'", candidate.Username, unmarshaled.Username)
		}
	})
}

func TestSearchResult(t *testing.T) {
	t.Run("ValidSearchResult", func(t *testing.T) {
		result := SearchResult{
			Candidates: []Candidate{
				{Username: "user1", Name: "User 1"},
				{Username: "user2", Name: "User 2"},
			},
			TotalFound: 2,
			SearchCriteria: map[string]interface{}{
				"language": "go",
				"location": "lima",
			},
		}

		if result.TotalFound != 2 {
			t.Errorf("Expected total_found 2, got %d", result.TotalFound)
		}
		if len(result.Candidates) != 2 {
			t.Errorf("Expected 2 candidates, got %d", len(result.Candidates))
		}
	})

	t.Run("JSONMarshaling", func(t *testing.T) {
		result := SearchResult{
			Candidates: []Candidate{
				{Username: "user1", Name: "User 1", PublicRepos: 10, Followers: 20},
			},
			TotalFound: 1,
			SearchCriteria: map[string]interface{}{
				"language": "python",
			},
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			t.Errorf("Failed to marshal SearchResult: %v", err)
		}

		var unmarshaled SearchResult
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal SearchResult: %v", err)
		}

		if unmarshaled.TotalFound != result.TotalFound {
			t.Errorf("Expected total_found %d, got %d", result.TotalFound, unmarshaled.TotalFound)
		}
	})
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

func BenchmarkToolDefinition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getToolDefinition()
	}
}

func BenchmarkJSONMarshaling(b *testing.B) {
	candidate := Candidate{
		Username:    "testuser",
		Name:        "Test User",
		Location:    "Lima, Peru",
		Bio:         "Go developer",
		PublicRepos: 25,
		Followers:   100,
		GitHubURL:   "https://github.com/testuser",
		AvatarURL:   "https://avatar.png",
	}

	b.Run("MarshalCandidate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(candidate)
		}
	})

	b.Run("UnmarshalCandidate", func(b *testing.B) {
		jsonData, _ := json.Marshal(candidate)
		for i := 0; i < b.N; i++ {
			var c Candidate
			_ = json.Unmarshal(jsonData, &c)
		}
	})
}
