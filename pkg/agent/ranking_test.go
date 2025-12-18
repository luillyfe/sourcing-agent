package agent

import (
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

func TestRankAndPresent(t *testing.T) {
	// Mock response
	mockResp := &llm.Response{
		Content: []llm.ContentBlock{
			{
				Type: "text",
				Text: `
` + "```json" + `
{
  "top_candidates": [
    {
      "rank": 1,
      "username": "testuser",
      "name": "Test User",
      "location": "Lima",
      "github_url": "https://github.com/testuser",
      "final_match_score": 0.95,
      "match_breakdown": {
        "required_skills_score": 0.4,
        "repository_relevance_score": 0.3,
        "experience_score": 0.2,
        "profile_quality_score": 0.05
      },
      "key_qualifications": ["Go", "Kubernetes"],
      "top_relevant_projects": [],
      "match_reasoning": "Strong match"
    }
  ],
  "summary": {
    "total_candidates_found": 10,
    "candidates_presented": 1,
    "average_match_score": 0.8,
    "search_quality": "High"
  }
}
` + "```",
			},
		},
	}

	client := &MockLLMClient{
		CallAPIFunc: func(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
			return mockResp, nil
		},
	}

	candidates := &EnrichedCandidates{}
	requirements := &Requirements{}

	result, _, err := rankAndPresent(client, candidates, requirements)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.TopCandidates) != 1 {
		t.Errorf("Expected 1 top candidate, got %d", len(result.TopCandidates))
	}
	if result.TopCandidates[0].Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", result.TopCandidates[0].Username)
	}
	if result.Summary.TotalCandidatesFound != 10 {
		t.Errorf("Expected 10 candidates found, got %d", result.Summary.TotalCandidatesFound)
	}
}
