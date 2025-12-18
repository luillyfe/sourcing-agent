package agent

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/luillyfe/sourcing-agent/pkg/github"
	"github.com/luillyfe/sourcing-agent/pkg/vertexai"
)

// TestConsistency checks if semantically equivalent queries produce overlapping results.
// It is an integration test that requires real API keys.
func TestConsistency(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TEST not set to 1")
	}

	// Load .env from project root if possible (optional, but helpful if running locally)
	_ = godotenv.Load("../../.env")

	// Setup clients
	projectID := os.Getenv("VERTEX_PROJECT_ID")
	region := os.Getenv("VERTEX_REGION")
	githubToken := os.Getenv("GITHUB_TOKEN")

	if projectID == "" || region == "" || githubToken == "" {
		t.Fatal("Missing required environment variables for integration test (VERTEX_PROJECT_ID, VERTEX_REGION, GITHUB_TOKEN)")
	}

	ctx := context.Background()
	vertexClient, err := vertexai.NewClient(ctx, projectID, region)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer vertexClient.Close()

	githubClient := github.NewClient(githubToken)

	// Test Cases: Pairs of queries that should yield similar results
	testCases := []struct {
		name   string
		queryA string
		queryB string
	}{
		{
			name:   "Go Developers in Lima",
			queryA: "Find Senior Go developers in Lima",
			queryB: "Looking for expert Golang engineers based in Lima, Peru",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run Query A
			t.Logf("Running Query A: %s", tc.queryA)
			resultA, err := RunStage2(vertexClient, githubClient, tc.queryA)
			if err != nil {
				t.Fatalf("Query A failed: %v", err)
			}

			// Run Query B
			t.Logf("Running Query B: %s", tc.queryB)
			resultB, err := RunStage2(vertexClient, githubClient, tc.queryB)
			if err != nil {
				t.Fatalf("Query B failed: %v", err)
			}

			// Calculate Consistency (Jaccard Similarity)
			// J = (A ∩ B) / (A ∪ B)
			// We use usernames as unique identifiers

			usersA := make(map[string]bool)
			for _, c := range resultA.TopCandidates {
				usersA[c.Username] = true
			}

			usersB := make(map[string]bool)
			for _, c := range resultB.TopCandidates {
				usersB[c.Username] = true
			}

			intersection := 0
			union := len(usersA)

			for user := range usersB {
				if usersA[user] {
					intersection++
				} else {
					union++
				}
			}

			if union == 0 {
				t.Fatal("No candidates found for either query, cannot calculate consistency")
			}

			consistency := float64(intersection) / float64(union)
			t.Logf("Query A Candidates: %d", len(usersA))
			t.Logf("Query B Candidates: %d", len(usersB))
			t.Logf("Intersection: %d", intersection)
			t.Logf("Consistency Score: %.2f", consistency)

			// Threshold: We expect at least some overlap for semantically identical queries.
			// 30% is a conservative starting point given search API volatility.
			if consistency < 0.3 {
				t.Errorf("Consistency score %.2f is below threshold 0.3", consistency)
			}
		})
	}
}
