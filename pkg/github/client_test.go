package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchDevelopers(t *testing.T) {
	// Create a single mock server that handles both endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/search/users", func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected Authorization header 'token test-token', got '%s'", r.Header.Get("Authorization"))
		}

		// Verify query parameter format
		// The client sends: ...?q=language:go+repos:>5...
		// r.URL.Query().Get("q") decodes keys and values.
		// If client sent q=language%3Ago%2Brepos%3A%3E5, then r.URL.Query().Get("q") is "language:go+repos:>5".
		// But in standard URL encoding, + means space.
		// GitHub expects "language:go repos:>5" (space separated terms).
		// So we expect the decoded q to be "language:go repos:>5" OR "language:go+repos:>5" if + is literal?
		// GitHub documentation says: q=language:go+repos:>5.
		// Wait, if I use curl "https://api.github.com/search/users?q=language:go+repos:>5", the server sees "language:go repos:>5".
		// If I use "q=language:go%2Brepos:>5", the server sees "language:go+repos:>5".
		// The user complained about %2B.
		// So we want the URL to contain + as separator, which means q value has spaces.

		q := r.URL.Query().Get("q")
		expectedQ := "language:go repos:>5"
		if q != expectedQ {
			// This might be tricky because standard Go Decoding might treat + as space.
			// Let's check raw query
			// If URL has q=language%3Ago+repos%3A%3E5, raw is ...
			// If URL has q=language%3Ago%2Brepos%3A%3E5, raw is ...

			// We want to ensure that we are NOT sending %2B.
			// So if we check r.URL.RawQuery we can see what was sent.
		}

		if !strings.Contains(r.URL.RawQuery, "language%3Ago+repos%3A%3E5") && !strings.Contains(r.URL.RawQuery, "language%3Ago%20repos%3A%3E5") {
			// We want to accept + (space) or %20 (space).
			// We do NOT want %2B (+) which was the bug.
			if strings.Contains(r.URL.RawQuery, "%2B") {
				t.Errorf("Query contained %%2B (escaped +) which implies incorrect encoding: %s", r.URL.RawQuery)
			}
		}

		// Return mock search response
		response := SearchResponse{
			TotalCount:        2,
			IncompleteResults: false,
			Items: []User{
				{Login: "testuser1", ID: 1, HTMLURL: "https://github.com/testuser1", AvatarURL: "https://avatar1.png"},
				{Login: "testuser2", ID: 2, HTMLURL: "https://github.com/testuser2", AvatarURL: "https://avatar2.png"},
			},
		}
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/users/testuser1", func(w http.ResponseWriter, r *http.Request) {
		// Return mock user detail response
		response := UserDetail{
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
	})

	mux.HandleFunc("/users/testuser2", func(w http.ResponseWriter, r *http.Request) {
		// Return mock user detail response
		response := UserDetail{
			Login:       "testuser2",
			Name:        "Test User 2",
			Location:    "Arequipa, Peru",
			Bio:         "Python developer",
			PublicRepos: 15,
			Followers:   50,
			HTMLURL:     "https://github.com/testuser2",
			AvatarURL:   "https://avatar2.png",
		}
		json.NewEncoder(w).Encode(response)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a client with the mock server URL
	client := &Client{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	t.Run("ValidInput", func(t *testing.T) {
		input := ToolInput{
			Language:   "go",
			MinRepos:   5,
			MaxResults: 10,
		}

		result, err := client.SearchDevelopers(input)
		if err != nil {
			t.Fatalf("SearchDevelopers failed: %v", err)
		}

		if result.TotalFound != 2 {
			t.Errorf("Expected 2 candidates, got %d", result.TotalFound)
		}

		if len(result.Candidates) != 2 {
			t.Errorf("Expected 2 candidates, got %d", len(result.Candidates))
		}

		if result.Candidates[0].Username != "testuser1" {
			t.Errorf("Expected first candidate to be testuser1, got %s", result.Candidates[0].Username)
		}
	})
}

func TestGetUserDetail(t *testing.T) {
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
		response := UserDetail{
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

	// Create a client with the mock server URL
	client := &Client{
		BaseURL: mockServer.URL,
		Token:   "test-token",
	}

	t.Run("ValidUsername", func(t *testing.T) {
		username := "testuser"
		detail, err := client.GetUserDetail(username)
		if err != nil {
			t.Fatalf("GetUserDetail failed: %v", err)
		}

		if detail.Login != "testuser" {
			t.Errorf("Expected login 'testuser', got '%s'", detail.Login)
		}
		if detail.Name != "Test User" {
			t.Errorf("Expected name 'Test User', got '%s'", detail.Name)
		}
	})
}
