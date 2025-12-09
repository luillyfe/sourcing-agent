package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Repository represents a GitHub repository
type Repository struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Stars       int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Topics      []string `json:"topics"`
	URL         string   `json:"html_url"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// Client handles interactions with the GitHub API
type Client struct {
	BaseURL string
	Token   string
}

// NewClient creates a new GitHubClient
func NewClient(token string) *Client {
	return &Client{
		BaseURL: "https://api.github.com",
		Token:   token,
	}
}

// SearchDevelopers searches GitHub for developers matching criteria
func (c *Client) SearchDevelopers(input ToolInput) (*SearchResult, error) {
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
	// Request up to 100 results per page to allow for filtering attrition
	url := fmt.Sprintf("%s/search/users?q=%s&per_page=100", c.BaseURL, query)
	fmt.Println("SearchDevelopers: ", url)

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

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	fmt.Println("SearchResponse: ", searchResponse)
	// Enrich each user with detailed information
	candidates := []Candidate{}
	for _, user := range searchResponse.Items {
		// Stop if we have collected enough candidates
		if len(candidates) >= input.MaxResults {
			break
		}

		detail, err := c.GetUserDetail(user.Login)
		if err != nil {
			// Log error but continue with other users
			fmt.Fprintf(os.Stderr, "Warning: failed to get details for user %s: %v\n", user.Login, err)
			continue
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
func (c *Client) GetUserDetail(username string) (*UserDetail, error) {
	// Note: We use the BaseURL from the client
	url := fmt.Sprintf("%s/users/%s", c.BaseURL, username)
	fmt.Println("GetUserDetail: ", url)

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

	var userDetail UserDetail
	if err := json.Unmarshal(body, &userDetail); err != nil {
		return nil, fmt.Errorf("failed to parse user detail response: %w", err)
	}

	return &userDetail, nil
}

// GetDeveloperRepositories retrieves repositories for a developer
func (c *Client) GetDeveloperRepositories(username string, maxRepos int) ([]Repository, error) {
	url := fmt.Sprintf("%s/users/%s/repos?sort=stars&per_page=%d", c.BaseURL, username, maxRepos)

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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to parse repositories: %w", err)
	}

	return repos, nil
}
