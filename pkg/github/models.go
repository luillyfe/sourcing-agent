package github

// GitHubSearchResponse represents the response from GitHub search API
type SearchResponse struct {
	TotalCount        int    `json:"total_count"`
	IncompleteResults bool   `json:"incomplete_results"`
	Items             []User `json:"items"`
}

// GitHubUser represents a GitHub user from search results
type User struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubUserDetail represents detailed user information
type UserDetail struct {
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
