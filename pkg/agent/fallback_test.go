package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/github"
)

// MockLLMClient is already defined in run_test.go

func TestFindAndEnrichCandidates_FallbackStrategies(t *testing.T) {
	// Setup Mock GitHub Server
	mockGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock Search API
		if strings.Contains(r.URL.Path, "/search/users") {
			query := r.URL.Query().Get("q")

			// Scenario:
			// 1. Primary: language:go location:lima -> 0 results
			// 2. Fallback 1: language:go location:peru -> 0 results
			// 3. Fallback 2: language:go -> 1 result

			emptyResponse := `{"total_count": 0, "items": []}`
			successResponse := `{
				"total_count": 1, 
				"items": [
					{
						"login": "success_user",
						"html_url": "https://github.com/success_user",
						"type": "User"
					}
				]
			}`

			if strings.Contains(query, "location:lima") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(emptyResponse))
				return
			}
			if strings.Contains(query, "location:peru") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(emptyResponse))
				return
			}
			// General fallback (no location)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(successResponse))
			return
		}

		// Mock User Detail API
		if strings.Contains(r.URL.Path, "/users/success_user") {
			if !strings.Contains(r.URL.Path, "/repos") {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"login": "success_user",
					"name": "Success User",
					"location": "Remote",
					"bio": "Go developer",
					"public_repos": 10,
					"followers": 5,
					"html_url": "https://github.com/success_user"
				}`))
				return
			}
		}

		// Mock Repos API
		if strings.Contains(r.URL.Path, "/repos") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`)) // Return empty repos for simplicity
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHub.Close()

	// Create Clients
	ghClient := &github.Client{BaseURL: mockGitHub.URL, Token: "mock-token"}
	llmClient := &MockLLMClient{}

	// Create Strategy with 2 fallbacks
	strategy := &SearchStrategy{
		PrimarySearch: SearchQuery{
			Language: "go",
			Location: "lima",
		},
		FallbackSearches: []SearchQuery{
			{
				Language: "go",
				Location: "peru", // Should fail
			},
			{
				Language: "go", // Should succeed
				Location: "",
			},
		},
		RepositorySearch: RepositorySearch{
			Keywords: []string{"backend"},
		},
		PostFilters: PostFilters{
			MinRepos: 1,
		},
	}

	reqs := &Requirements{RequiredSkills: []string{"Go"}}

	// Execute
	results, err := findAndEnrichCandidates(llmClient, ghClient, strategy, reqs)
	if err != nil {
		t.Fatalf("findAndEnrichCandidates failed: %v", err)
	}

	// Verify
	if results.SearchMetadata.TotalProfilesFound != 1 {
		t.Errorf("Expected 1 profile found, got %d", results.SearchMetadata.TotalProfilesFound)
	}
	if len(results.Candidates) != 1 {
		t.Fatalf("Expected 1 candidate, got %d", len(results.Candidates))
	}
	if results.Candidates[0].Username != "success_user" {
		t.Errorf("Expected candidate 'success_user', got '%s'", results.Candidates[0].Username)
	}
}
