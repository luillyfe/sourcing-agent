package agent

// Requirements structure (output of Prompt 1)
type Requirements struct {
	RequiredSkills        []string `json:"required_skills"`
	ExperienceLevel       string   `json:"experience_level"`
	Locations             []string `json:"locations"`
	Keywords              []string `json:"keywords"`
	NiceToHave            []string `json:"nice_to_have"`
	UnclearRequest        bool     `json:"unclear_request,omitempty"`
	ClarificationQuestion string   `json:"clarification_question,omitempty"`
}

// Search Strategy structure (output of Prompt 2)
type SearchStrategy struct {
	PrimarySearch    SearchQuery      `json:"primary_search"`
	FallbackSearches []SearchQuery    `json:"fallback_searches"`
	RepositorySearch RepositorySearch `json:"repository_search"`
	PostFilters      PostFilters      `json:"post_filters"`
	StrategyNotes    string           `json:"strategy_notes"`
}

type SearchQuery struct {
	Language  string  `json:"language"`
	Location  string  `json:"location"`
	Followers *string `json:"followers,omitempty"`
	Rationale string  `json:"rationale,omitempty"`
}

type RepositorySearch struct {
	Keywords []string `json:"keywords"`
	MinStars *int     `json:"min_stars,omitempty"`
	Language string   `json:"language"`
}

type PostFilters struct {
	MinRepos           int      `json:"min_repos"`
	BioKeywords        []string `json:"bio_keywords"`
	RecentActivityDays *int     `json:"recent_activity_days,omitempty"`
}

// Enriched Candidates structure (output of Prompt 3)
type EnrichedCandidates struct {
	Candidates     []EnrichedCandidate `json:"candidates"`
	SearchMetadata SearchMetadata      `json:"search_metadata"`
}

type EnrichedCandidate struct {
	Username             string               `json:"username"`
	Name                 string               `json:"name"`
	Location             string               `json:"location"`
	Bio                  string               `json:"bio"`
	PublicRepos          int                  `json:"public_repos"`
	Followers            int                  `json:"followers"`
	GitHubURL            string               `json:"github_url"`
	RelevantRepositories []RelevantRepository `json:"relevant_repositories"`
	SkillsFound          []string             `json:"skills_found"`
	ExperienceIndicators ExperienceIndicators `json:"experience_indicators"`
	InitialMatchScore    float64              `json:"initial_match_score"`
}

type RelevantRepository struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Language        string   `json:"language"`
	Stars           int      `json:"stars"`
	Topics          []string `json:"topics"`
	RelevanceScore  float64  `json:"relevance_score"`
	RelevanceReason string   `json:"relevance_reason"`
}

type ExperienceIndicators struct {
	AccountAgeYears    float64 `json:"account_age_years"`
	TotalStars         int     `json:"total_stars"`
	HasPopularProjects bool    `json:"has_popular_projects"`
}

type SearchMetadata struct {
	SearchesExecuted   int `json:"searches_executed"`
	TotalProfilesFound int `json:"total_profiles_found"`
	ProfilesAnalyzed   int `json:"profiles_analyzed"`
}

// Final Result structure (output of Prompt 4)
type FinalResult struct {
	TopCandidates []RankedCandidate `json:"top_candidates"`
	Summary       ResultSummary     `json:"summary"`
}

type RankedCandidate struct {
	Rank                int               `json:"rank"`
	Username            string            `json:"username"`
	Name                string            `json:"name"`
	Location            string            `json:"location"`
	GitHubURL           string            `json:"github_url"`
	FinalMatchScore     float64           `json:"final_match_score"`
	MatchBreakdown      MatchBreakdown    `json:"match_breakdown"`
	KeyQualifications   []string          `json:"key_qualifications"`
	TopRelevantProjects []RelevantProject `json:"top_relevant_projects"`
	MatchReasoning      string            `json:"match_reasoning"`
	PotentialConcerns   string            `json:"potential_concerns,omitempty"`
}

type MatchBreakdown struct {
	RequiredSkillsScore      float64 `json:"required_skills_score"`
	RepositoryRelevanceScore float64 `json:"repository_relevance_score"`
	ExperienceScore          float64 `json:"experience_score"`
	ProfileQualityScore      float64 `json:"profile_quality_score"`
}

type RelevantProject struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	WhyRelevant string `json:"why_relevant"`
}

type ResultSummary struct {
	TotalCandidatesFound int     `json:"total_candidates_found"`
	CandidatesPresented  int     `json:"candidates_presented"`
	AverageMatchScore    float64 `json:"average_match_score"`
	SearchQuality        string  `json:"search_quality"`
}

// RelevanceAnalysis result
type RelevanceAnalysis struct {
	Score   float64
	Reasons []string
}
