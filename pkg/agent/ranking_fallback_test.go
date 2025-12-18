package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/github"
	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

// MockLLMClientForFallback is a custom mock that fails on the ranking prompt
type MockLLMClientForFallback struct {
	CallCount int
}

func (m *MockLLMClientForFallback) CallAPI(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	m.CallCount++

	// Prompt 1: Requirements Analysis
	if m.CallCount == 1 {
		return &llm.Response{
			Content: []llm.ContentBlock{
				{Type: "text", Text: `{"required_skills": ["Go"], "experience_level": "mid", "locations": ["Remote"]}`},
			},
		}, nil
	}

	// Prompt 2: Search Strategy
	if m.CallCount == 2 {
		return &llm.Response{
			Content: []llm.ContentBlock{
				{Type: "text", Text: `
				{
					"primary_search": {"language": "go", "location": "remote"},
					"fallback_searches": [],
					"repository_search": {"keywords": ["backend"], "language": "go"},
					"post_filters": {"min_repos": 1, "bio_keywords": []},
					"strategy_notes": "test"
				}`},
			},
		}, nil
	}

	// Prompt 3 (Ranking): Simulate Failure
	// rankAndPresent is the 3rd LLM call in RunStage2
	if m.CallCount == 3 {
		return nil, fmt.Errorf("simulated LLM failure during ranking")
	}

	return nil, fmt.Errorf("unexpected call count: %d", m.CallCount)
}

func TestRunStage2_RankingFallback(t *testing.T) {
	// Setup Mock GitHub Server
	mockGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock Search API
		if strings.Contains(r.URL.Path, "/search/users") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"total_count": 1, 
				"items": [
					{
						"login": "fallback_user",
						"html_url": "https://github.com/fallback_user",
						"type": "User"
					}
				]
			}`))
			return
		}
		// Mock User Details
		if strings.Contains(r.URL.Path, "/users/fallback_user") {
			if !strings.Contains(r.URL.Path, "/repos") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"login": "fallback_user",
					"name": "Fallback User",
					"location": "Remote",
					"html_url": "https://github.com/fallback_user",
					"public_repos": 5
				}`))
				return
			}
		}
		// Mock Repos
		if strings.Contains(r.URL.Path, "/repos") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"name": "go-repo", "language": "Go", "stars": 10}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHub.Close()

	ghClient := &github.Client{
		BaseURL:    mockGitHub.URL,
		Token:      "mock-token",
		HTTPClient: &http.Client{},
	}
	llmClient := &MockLLMClientForFallback{}

	// Execute RunStage2
	result, err := RunStage2(llmClient, ghClient, "find go developers")

	// We expect NO error, because fallback should handle it
	if err != nil {
		t.Fatalf("RunStage2 returned error despite fallback logic: %v", err)
	}

	// Verify Fallback Result
	if result == nil {
		t.Fatal("Result is nil")
	}
	if result.Summary.SearchQuality != "Fallback (Ranking Unavailable)" {
		t.Errorf("Expected SearchQuality 'Fallback (Ranking Unavailable)', got '%s'", result.Summary.SearchQuality)
	}
	if len(result.TopCandidates) != 1 {
		t.Fatalf("Expected 1 candidate, got %d", len(result.TopCandidates))
	}
	cand := result.TopCandidates[0]
	if cand.Username != "fallback_user" {
		t.Errorf("Expected username 'fallback_user', got '%s'", cand.Username)
	}
	if cand.MatchReasoning != "Ranking step unavailable; score is based on initial keyword match." {
		t.Errorf("Unexpected match reasoning: %s", cand.MatchReasoning)
	}
}
