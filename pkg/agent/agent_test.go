package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/github"
)

func TestExecuteTool(t *testing.T) {
	// Create a mock server for the tool execution
	mux := http.NewServeMux()
	mux.HandleFunc("/search/users", func(w http.ResponseWriter, r *http.Request) {
		response := github.SearchResponse{
			TotalCount: 0,
			Items:      []github.User{},
		}
		json.NewEncoder(w).Encode(response)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &github.Client{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: &http.Client{},
	}

	t.Run("UnknownTool", func(t *testing.T) {
		_, err := executeTool(client, "unknown_tool", map[string]interface{}{})
		if err == nil {
			t.Error("Expected error for unknown tool")
		}
		if err.Error() != "unknown tool: unknown_tool" {
			t.Errorf("Expected error message 'unknown tool: unknown_tool', got '%s'", err.Error())
		}
	})

	t.Run("ValidToolName", func(t *testing.T) {
		// We need to provide valid input for the tool
		input := map[string]interface{}{
			"language": "go",
		}

		_, err := executeTool(client, "search_github_developers", input)
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	})
}
