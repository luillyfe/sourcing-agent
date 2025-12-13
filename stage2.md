# STAGE 2: MULTI-PROMPT SOURCING AGENT - DESIGN SPECIFICATION

## 1. OVERVIEW

*   **Pattern:** Prompt Chaining
*   **Purpose:** Break down sourcing into specialized sequential steps, each handled by a separate LLM prompt
*   **Complexity:** Medium - multiple sequential prompts, each with specific responsibilities
*   **Evolution from Stage 1:** Single shot → Specialized pipeline

## 2. SYSTEM ARCHITECTURE

```text
User Query
    ↓
Prompt 1: Requirements Analysis
    ↓
Prompt 2: Search Strategy Generation
    ↓
Prompt 3: Candidate Search & Enrichment
    ↓
Prompt 4: Ranking & Presentation
    ↓
Final Response
```

**Key Characteristic:** Each prompt is specialized for one task. Output from one prompt becomes input to the next.

## 3. CORE COMPONENTS

### 3.1 Prompt Chain
*   4 specialized prompts in sequence
*   Each prompt has single responsibility
*   Deterministic flow (no branching)
*   Each step builds on previous output

### 3.2 Tool Layer
*   GitHub Search Tools (same as Stage 1)
*   Tools can be called from any prompt in the chain
*   Same rate limiting and error handling

### 3.3 State Management
*   Intermediate results passed between prompts
*   Structured data format for inter-prompt communication
*   No persistent state (stateless between user queries)

## 4. PROMPT CHAIN DESIGN

### PROMPT 1: REQUIREMENTS ANALYZER

**Responsibility:** Parse user query into structured requirements

**Input Example:**
User query: "Find senior Go developers in Lima with microservices and MongoDB experience"

**System Prompt:**
You are a requirements analyzer for technical recruiting.

Your task: Parse the user's hiring request into structured requirements.

Extract:
1. Required skills (programming languages, frameworks, technologies)
2. Experience level (junior, mid, senior, lead)
3. Location requirements (city, country, region, remote)
4. Keywords for relevance matching
5. Nice-to-have skills (optional qualifications)

**Output Format (JSON):**
```json
{
  "required_skills": ["skill1", "skill2"],
  "experience_level": "senior|mid|junior|lead",
  "locations": ["location1", "location2"],
  "keywords": ["keyword1", "keyword2"],
  "nice_to_have": ["skill3", "skill4"]
}
```

Be specific and extract all relevant information from the query.

**Output Example:**
```json
{
  "required_skills": ["Go", "microservices", "MongoDB"],
  "experience_level": "senior",
  "locations": ["Lima", "Peru"],
  "keywords": ["distributed systems", "backend", "API"],
  "nice_to_have": ["Docker", "Kubernetes", "gRPC"]
}
```

### PROMPT 2: SEARCH STRATEGY GENERATOR
 
 **Responsibility:** Generate optimal GitHub search strategies based on requirements
 
 **Input:** Output from Prompt 1 (Requirements structure)
 
 **System Prompt:**
 You are a search strategy expert for GitHub developer sourcing.
 
 ## Available Search Capabilities
 
 The system can search GitHub using these parameters:
 
 **User Search (primary)**
 - language: programming language (inferred from user's repos)
 - location: matches user's profile location field (freeform text, inconsistent)
 - followers: minimum follower count (e.g., ">10", ">100")
 
 **Repository Search (secondary)**
 - keywords: searches repo names, descriptions, and READMEs
 - stars: minimum star count
 - language: exact match on repo primary language
 
 **Post-Search Filtering (applied locally after fetching results)**
 - min_repos: minimum public repository count
 - bio_keywords: substring match against user bio
 - recent_activity_days: only users with commits within N days
 
 ## Limitations
 
 - Cannot search by years of experience directly
 - Location is unreliable (~40% of users have it filled, format varies)
 - Language filter only works if user has public repos in that language
 - GitHub API rate limits: prefer precise queries over broad ones
 
 ## Your Task
 
 Given structured job requirements, generate an optimal search strategy:
 
 1. Create a primary search (most specific, highest signal)
 2. Create fallback searches (progressively broader for when primary yields few results)
 3. Configure repository search to find users via their project work
 4. Set post-filters to refine results locally
 5. Plan for low/no results scenario in your fallbacks
 
 ## Output Format (JSON)
 
 ```json
 {
   "primary_search": {
     "language": "string",
     "location": "string", 
     "followers": "string (e.g., '>10') or null"
   },
   "fallback_searches": [
     {
       "language": "string",
       "location": "string or null (broader)",
       "followers": "string or null",
       "rationale": "string (why this fallback)"
     }
   ],
   "repository_search": {
     "keywords": ["keyword1", "keyword2"],
     "min_stars": "number or null",
     "language": "string"
   },
   "post_filters": {
     "min_repos": "number",
     "bio_keywords": ["keyword1", "keyword2"],
     "recent_activity_days": "number or null"
   },
   "strategy_notes": "string (brief explanation of your approach)"
 }
 ```
 
 **Output Example:**
 ```json
 {
   "primary_search": {
     "language": "go",
     "location": "lima",
     "followers": ">10"
   },
   "fallback_searches": [
     {
       "language": "go",
       "location": "peru",
       "followers": ">5",
       "rationale": "Broadening location to country level"
     },
     {
       "language": "go",
       "location": "lima",
       "followers": null,
       "rationale": "Removing follower constraint"
     }
   ],
   "repository_search": {
     "keywords": ["microservices", "mongodb", "grpc"],
     "min_stars": 5,
     "language": "go"
   },
   "post_filters": {
     "min_repos": 15,
     "bio_keywords": ["senior", "lead", "backend"],
     "recent_activity_days": 30
   },
   "strategy_notes": "Starting with strict location and follower count in Lima. Falling back to all of Peru, then removing follower constraints if needed. Using repo search to ensure microservices experience."
 }
 ```

### PROMPT 3: CANDIDATE FINDER & ENRICHER

**Responsibility:** Execute searches and enrich candidate profiles with detailed analysis

**Input:** Output from Prompt 2 (Search Strategy structure)

**Tools Available:**
- `search_github_developers` (from Stage 1)
- `get_developer_repositories` (NEW for Stage 2)
- `analyze_repository_relevance` (NEW for Stage 2)

**System Prompt:**
You are a candidate sourcing specialist.

Given search strategies, execute searches and enrich candidate data.

Your task:
1. Execute primary search using search_github_developers
2. If insufficient results, try fallback searches
3. For each candidate, call get_developer_repositories
4. Analyze repositories for relevance to requirements
5. Calculate initial match scores

Process:
- Start with primary search
- Aim for 15-20 candidates
- Get top 5-10 repositories per candidate
- Look for repository_keywords in repo names, descriptions, topics
- Note experience indicators (repo age, stars, contribution patterns)

**Output Format (JSON):**
```json
{
  "candidates": [
    {
      "username": "string",
      "name": "string",
      "location": "string",
      "bio": "string",
      "public_repos": number,
      "followers": number,
      "github_url": "string",
      "relevant_repositories": [
        {
          "name": "string",
          "description": "string",
          "language": "string",
          "stars": number,
          "topics": ["topic1", "topic2"],
          "relevance_score": number,
          "relevance_reason": "string"
        }
      ],
      "skills_found": ["skill1", "skill2"],
      "experience_indicators": {
        "account_age_years": number,
        "total_stars": number,
        "has_popular_projects": boolean
      },
      "initial_match_score": number
    }
  ],
  "search_metadata": {
    "searches_executed": number,
    "total_profiles_found": number,
    "profiles_analyzed": number
  }
}
```

### PROMPT 4: RANKER & PRESENTER

**Responsibility:** Final ranking and formatting for user presentation

**Input:** Output from Prompt 3 (Enriched Candidates structure)

**System Prompt:**
You are a candidate ranking and presentation specialist.

Given enriched candidate data, produce final rankings and presentation.

Your task:
1. Calculate final match scores based on:
   - Required skills coverage
   - Repository relevance
   - Experience indicators
   - Location match
   - Profile quality (bio, followers, activity)
2. Rank candidates by match score
3. Format top 10 for presentation
4. Provide reasoning for each candidate

**Scoring weights:**
- Required skills match: 40%
- Repository relevance: 30%
- Experience indicators: 20%
- Profile quality: 10%

**Output Format (JSON):**
```json
{
  "top_candidates": [
    {
      "rank": number,
      "username": "string",
      "name": "string",
      "location": "string",
      "github_url": "string",
      "final_match_score": number,
      "match_breakdown": {
        "required_skills_score": number,
        "repository_relevance_score": number,
        "experience_score": number,
        "profile_quality_score": number
      },
      "key_qualifications": ["qual1", "qual2", "qual3"],
      "top_relevant_projects": [
        {
          "name": "string",
          "url": "string",
          "why_relevant": "string"
        }
      ],
      "match_reasoning": "string - why this candidate is a good match",
      "potential_concerns": "string - any gaps or considerations"
    }
  ],
  "summary": {
    "total_candidates_found": number,
    "candidates_presented": number,
    "average_match_score": number,
    "search_quality": "excellent|good|fair|limited"
  }
}
```

## 5. DATA FLOW BETWEEN PROMPTS

### Complete Flow Example:

**User Query:**
"Find senior Go developers in Lima with microservices and MongoDB experience"

↓ **[Prompt 1: Requirements Analyzer]**

**Structured Requirements:**
```json
{
  "required_skills": ["Go", "microservices", "MongoDB"],
  "experience_level": "senior",
  "locations": ["Lima", "Peru"],
  "keywords": ["distributed systems", "backend", "API"],
  "nice_to_have": ["Docker", "Kubernetes", "gRPC"]
}
```

↓ **[Prompt 2: Search Strategy Generator]**

**Search Strategy:**
```json
{
  "primary_search": {
    "language": "go",
    "location": "lima",
    "followers": ">20"
  },
  "fallback_searches": [
    {
      "language": "go",
      "location": "peru",
      "followers": ">10",
      "rationale": "Broadening location to country level"
    }
  ],
  "repository_search": {
    "keywords": ["microservices", "mongodb", "grpc"],
    "min_stars": 5,
    "language": "go"
  },
  "post_filters": {
    "min_repos": 15,
    "bio_keywords": ["senior", "lead", "backend"],
    "recent_activity_days": 30
  },
  "strategy_notes": "Starting with strict location and follower count in Lima. Falling back to all of Peru. Using repo search to ensure microservices experience."
}
```

↓ **[Prompt 3: Candidate Finder & Enricher]**
   [Uses: `search_github_developers`, `get_developer_repositories`]

**Enriched Candidates:**
```json
{
  "candidates": [
    {
      "username": "dev1",
      "name": "Developer One",
      "relevant_repositories": [
        {
          "name": "microservice-api",
          "relevance_score": 0.9,
          "relevance_reason": "Go microservices with MongoDB"
        }
      ],
      "skills_found": ["Go", "MongoDB", "Docker"],
      "initial_match_score": 0.85
    }
  ],
  "search_metadata": {
    "searches_executed": 1,
    "total_profiles_found": 12,
    "profiles_analyzed": 12
  }
}
```

↓ **[Prompt 4: Ranker & Presenter]**

**Final Output:**
```json
{
  "top_candidates": [
    {
      "rank": 1,
      "username": "dev1",
      "final_match_score": 0.92,
      "match_breakdown": {
        "required_skills_score": 0.95,
        "repository_relevance_score": 0.90,
        "experience_score": 0.88,
        "profile_quality_score": 0.85
      },
      "key_qualifications": [
        "8 years Go experience",
        "Proven microservices expertise",
        "Active MongoDB user"
      ],
      "match_reasoning": "Strong match: Active Go developer in Lima with proven microservices and MongoDB experience..."
    }
  ],
  "summary": {
    "total_candidates_found": 12,
    "candidates_presented": 10,
    "average_match_score": 0.78,
    "search_quality": "excellent"
  }
}
```

## 6. IMPLEMENTATION STRUCTURE (GO)

### 6.1 Main Orchestrator

```go
func runSourcingAgentStage2(userQuery string) (string, error) {
    // Step 1: Analyze Requirements
    requirements, err := analyzeRequirements(userQuery)
    if err != nil {
        return "", fmt.Errorf("requirements analysis failed: %w", err)
    }
    
    // Step 2: Generate Search Strategy
    strategy, err := generateSearchStrategy(requirements)
    if err != nil {
        return "", fmt.Errorf("strategy generation failed: %w", err)
    }
    
    // Step 3: Find and Enrich Candidates
    enrichedCandidates, err := findAndEnrichCandidates(strategy)
    if err != nil {
        return "", fmt.Errorf("candidate search failed: %w", err)
    }
    
    // Step 4: Rank and Present
    finalResult, err := rankAndPresent(enrichedCandidates, requirements)
    if err != nil {
        return "", fmt.Errorf("ranking failed: %w", err)
    }
    
    return finalResult, nil
}
```

### 6.2 Individual Prompt Functions

```go
// Prompt 1: Requirements Analyzer
func analyzeRequirements(client llm.Client, userQuery string) (*Requirements, error) {
    systemPrompt := `You are a requirements analyzer for technical 
                     recruiting...`
    
    messages := []llm.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: fmt.Sprintf("User query: %s", userQuery)},
    }
    
    resp, err := client.CallAPI(messages, nil)
    if err != nil {
        return nil, err
    }
    
    var content string
    for _, block := range resp.Content {
        if block.Type == "text" {
            content += block.Text
        }
    }
    
    // Parse JSON response into Requirements struct
    var requirements Requirements
    err = json.Unmarshal([]byte(extractJSON(content)), &requirements)
    
    return &requirements, err
}

// Prompt 2: Search Strategy Generator
func generateSearchStrategy(
    client llm.Client,
    requirements *Requirements,
) (*SearchStrategy, error) {
    systemPrompt := `You are a search strategy expert for GitHub developer 
                     sourcing...`
    
    requirementsJSON, _ := json.Marshal(requirements)
    
    messages := []llm.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: string(requirementsJSON)},
    }
    
    resp, err := client.CallAPI(messages, nil)
    if err != nil {
        return nil, err
    }
    
    var content string
    for _, block := range resp.Content {
        if block.Type == "text" {
            content += block.Text
        }
    }
    
    // Parse JSON response
    var strategy SearchStrategy
    err = json.Unmarshal([]byte(extractJSON(content)), &strategy)
    
    return &strategy, err
}

// Prompt 3: Candidate Finder & Enricher
func findAndEnrichCandidates(
    client llm.Client,
    strategy *SearchStrategy,
) (*EnrichedCandidates, error) {
    // This function orchestrates the search and enrichment
    // It may use programmatic tools or LLM calls depending on implementation
    
    // ... Implementation logic ...
    
    return enrichedCandidates, nil
}

// Prompt 4: Ranker & Presenter
func rankAndPresent(
    client llm.Client,
    candidates *EnrichedCandidates,
    requirements *Requirements,
) (string, error) {
    systemPrompt := `You are a candidate ranking and presentation 
                     specialist...`
    
    input := map[string]interface{}{
        "candidates": candidates,
        "requirements": requirements,
    }
    inputJSON, _ := json.Marshal(input)
    
    messages := []llm.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: string(inputJSON)},
    }
    
    resp, err := client.CallAPI(messages, nil)
    if err != nil {
        return "", err
    }
    
    var content string
    for _, block := range resp.Content {
        if block.Type == "text" {
            content += block.Text
        }
    }
    
    return extractJSON(content), nil
}
```

### 6.3 Data Structures

```go
// Requirements structure (output of Prompt 1)
type Requirements struct {
    RequiredSkills   []string `json:"required_skills"`
    ExperienceLevel  string   `json:"experience_level"`
    Locations        []string `json:"locations"`
    Keywords         []string `json:"keywords"`
    NiceToHave       []string `json:"nice_to_have"`
}

// Search Strategy structure (output of Prompt 2)
type SearchStrategy struct {
    PrimarySearch     SearchQuery      `json:"primary_search"`
    FallbackSearches  []SearchQuery    `json:"fallback_searches"`
    RepositorySearch  RepositorySearch `json:"repository_search"`
    PostFilters       PostFilters      `json:"post_filters"`
    StrategyNotes     string           `json:"strategy_notes"`
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
    Username              string                `json:"username"`
    Name                  string                `json:"name"`
    Location              string                `json:"location"`
    Bio                   string                `json:"bio"`
    PublicRepos           int                   `json:"public_repos"`
    Followers             int                   `json:"followers"`
    GitHubURL             string                `json:"github_url"`
    RelevantRepositories  []RelevantRepository  `json:"relevant_repositories"`
    SkillsFound           []string              `json:"skills_found"`
    ExperienceIndicators  ExperienceIndicators  `json:"experience_indicators"`
    InitialMatchScore     float64               `json:"initial_match_score"`
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
    SearchesExecuted    int `json:"searches_executed"`
    TotalProfilesFound  int `json:"total_profiles_found"`
    ProfilesAnalyzed    int `json:"profiles_analyzed"`
}

// Final Result structure (output of Prompt 4)
type FinalResult struct {
    TopCandidates []RankedCandidate `json:"top_candidates"`
    Summary       ResultSummary     `json:"summary"`
}

type RankedCandidate struct {
    Rank                 int                  `json:"rank"`
    Username             string               `json:"username"`
    Name                 string               `json:"name"`
    Location             string               `json:"location"`
    GitHubURL            string               `json:"github_url"`
    FinalMatchScore      float64              `json:"final_match_score"`
    MatchBreakdown       MatchBreakdown       `json:"match_breakdown"`
    KeyQualifications    []string             `json:"key_qualifications"`
    TopRelevantProjects  []RelevantProject    `json:"top_relevant_projects"`
    MatchReasoning       string               `json:"match_reasoning"`
    PotentialConcerns    string               `json:"potential_concerns,omitempty"`
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
```

## 7. NEW TOOLS FOR STAGE 2

### Tool 1: get_developer_repositories

**Purpose:** Retrieve a developer's repositories for analysis

**Parameters:**
- `username`: string (required) - GitHub username
- `maxRepos`: int (default: 10) - Maximum repos to return
- `sortBy`: string (default: "stars") - Sort order: "stars", "updated", "created"

**Implementation:**
```go
func getDeveloperRepositories(
    username string,
    maxRepos int,
    sortBy string,
) ([]Repository, error) {
    githubToken := os.Getenv("GITHUB_TOKEN")
    
    // GET /users/{username}/repos?sort={sortBy}&per_page={maxRepos}
    url := fmt.Sprintf(
        "https://api.github.com/users/%s/repos?sort=%s&per_page=%d",
        username,
        sortBy,
        maxRepos,
    )
    
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var repos []Repository
    json.NewDecoder(resp.Body).Decode(&repos)
    
    return repos, nil
}

Repository Structure:
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
```

**Tool Definition:**
```json
{
    "name": "get_developer_repositories",
    "description": "Get a developer's repositories sorted by stars, update date, or creation date. Returns repository details including name, description, language, stars, and topics.",
    "input_schema": {
        "type": "object",
        "properties": {
            "username": {
                "type": "string",
                "description": "GitHub username"
            },
            "max_repos": {
                "type": "integer",
                "description": "Maximum number of repositories to return (default: 10)"
            },
            "sort_by": {
                "type": "string",
                "enum": ["stars", "updated", "created"],
                "description": "Sort order for repositories (default: stars)"
            }
        },
        "required": ["username"]
    }
}
```

### Tool 2: analyze_repository_relevance

**Purpose:** Analyze a repository's relevance to job requirements

**Note:** This can be implemented as either:
  A) A programmatic helper function (no LLM call)
  B) Another LLM call for deeper analysis
  
**Recommended:** Start with programmatic approach (simpler, faster)

**Parameters:**
- `repo`: Repository object
- `requiredSkills`: []string - List of required skills
- `keywords`: []string - List of relevant keywords

**Implementation (Programmatic):**
```go
func analyzeRepositoryRelevance(
    repo Repository,
    requiredSkills []string,
    keywords []string,
) RelevanceAnalysis {
    score := 0.0
    reasons := []string{}
    
    // Check language match
    for _, skill := range requiredSkills {
        if strings.EqualFold(repo.Language, skill) {
            score += 0.3
            reasons = append(reasons, fmt.Sprintf("Uses %s", skill))
        }
    }
    
    // Check keywords in name/description
    repoText := strings.ToLower(repo.Name + " " + repo.Description)
    for _, keyword := range keywords {
        if strings.Contains(repoText, strings.ToLower(keyword)) {
            score += 0.2
            reasons = append(
                reasons,
                fmt.Sprintf("Contains '%s'", keyword),
            )
        }
    }
    
    // Check topics
    for _, topic := range repo.Topics {
        for _, keyword := range keywords {
            topicLower := strings.ToLower(topic)
            keywordLower := strings.ToLower(keyword)
            if strings.Contains(topicLower, keywordLower) {
                score += 0.15
                reasons = append(reasons, fmt.Sprintf("Topic: %s", topic))
            }
        }
    }
    
    // Check popularity (stars)
    if repo.Stars > 50 {
        score += 0.1
        reasons = append(reasons, "Popular project")
    }
    
    return RelevanceAnalysis{
        Score:   min(score, 1.0),
        Reasons: reasons,
    }
}

type RelevanceAnalysis struct {
    Score   float64
    Reasons []string
}
```

## 8. EXECUTION FLOW EXAMPLE

### Complete Example Flow:

**User Query:**
"Find senior Go developers in Lima with microservices and MongoDB experience"

### PROMPT 1: Requirements Analyzer

**Input:** Raw user query

**LLM Response:**
```json
{
  "required_skills": ["Go", "microservices", "MongoDB"],
  "experience_level": "senior",
  "locations": ["Lima", "Peru"],
  "keywords": ["distributed systems", "backend", "API"],
  "nice_to_have": ["Docker", "Kubernetes", "gRPC"]
}
```

### PROMPT 2: Search Strategy Generator

**Input:** Requirements from Prompt 1

**LLM Response:**
```json
{
  "primary_search": {
    "language": "go",
    "location": "lima",
    "followers": ">20"
  },
  "fallback_searches": [
    {
      "language": "go",
      "location": "peru",
      "followers": ">10",
      "rationale": "Broader location match"
    },
    {
      "language": "go",
      "location": "lima",
      "followers": ">10",
      "rationale": "Lower follower count"
    }
  ],
  "repository_search": {
    "keywords": ["microservices", "mongodb", "grpc", "api", "docker"],
    "min_stars": 5,
    "language": "go"
  },
  "post_filters": {
    "min_repos": 15,
    "bio_keywords": ["senior", "lead", "backend"],
    "recent_activity_days": 30
  },
  "strategy_notes": "Prioritizing Lima-based candidates with significant following."
}
```

### PROMPT 3: Candidate Finder & Enricher

**Input:** Search strategy from Prompt 2

**LLM Actions:**
1. Calls `search_github_developers(language="go", location="lima", min_repos=15)`
   → Returns 12 candidates

2. For each candidate, calls `get_developer_repositories(username, max=10)`
   → Returns top 10 repos per candidate

3. Analyzes repositories for keywords
   → Calculates relevance scores

**LLM Response:**
```json
{
  "candidates": [
    {
      "username": "fermin_dev",
      "name": "Fermin Rodriguez",
      "location": "Lima, Peru",
      "bio": "Cloud Engineer | Go microservices | MongoDB",
      "public_repos": 25,
      "followers": 89,
      "github_url": "https://github.com/fermin_dev",
      "relevant_repositories": [
        {
          "name": "product-catalog-service",
          "description": "Multi-tenant catalog with MongoDB",
          "language": "Go",
          "stars": 45,
          "topics": ["microservices", "mongodb", "grpc"],
          "relevance_score": 0.95,
          "relevance_reason": "Matches Go, microservices, MongoDB"
        },
        {
          "name": "event-streaming-platform",
          "description": "Distributed event streaming with Kafka",
          "language": "Go",
          "stars": 32,
          "topics": ["distributed-systems", "kafka"],
          "relevance_score": 0.75,
          "relevance_reason": "Distributed systems experience"
        }
      ],
      "skills_found": ["Go", "MongoDB", "gRPC", "Docker"],
      "experience_indicators": {
        "account_age_years": 8,
        "total_stars": 234,
        "has_popular_projects": true
      },
      "initial_match_score": 0.88
    },
    // ... 11 more candidates
  ],
  "search_metadata": {
    "searches_executed": 1,
    "total_profiles_found": 12,
    "profiles_analyzed": 12
  }
}
```

### PROMPT 4: Ranker & Presenter

**Input:** Enriched candidates from Prompt 3 + Original requirements

**LLM Actions:**
1. Calculates final scores with weighted components
2. Ranks all candidates
3. Selects top 10
4. Formats for presentation

**LLM Response (formatted for user):**

I found 12 Go developers in Lima and ranked them by match quality. 
Here are the top 10:

1. ★★★★★ fermin_dev (92% match)
   Fermin Rodriguez | Lima, Peru
   https://github.com/fermin_dev
   
   Why this candidate:
   • Strong Go developer with 8 years experience
   • Proven microservices expertise (product-catalog-service)
   • Active MongoDB user with production projects
   • Has gRPC and Docker experience (nice-to-have)
   
   Top relevant projects:
   • product-catalog-service - Multi-tenant catalog with MongoDB, gRPC
     https://github.com/fermin_dev/product-catalog-service
   • event-streaming-platform - Distributed system with Go
     https://github.com/fermin_dev/event-streaming-platform
   
   Match breakdown:
   • Required skills: 95%
   • Repository relevance: 90%
   • Experience level: 88%
   • Profile quality: 85%
   
   Potential concerns: None identified

2. ★★★★☆ carlos_backend (87% match)
   Carlos Mendoza | Lima, Peru
   https://github.com/carlos_backend
   
   Why this candidate:
   • Senior Go developer (10 years on GitHub)
   • Built multiple microservices architectures
   • MongoDB integration in 3+ projects
   
   Top relevant projects:
   • order-management-api - Microservices with MongoDB
   • payment-processor - Distributed Go system
   
   Match breakdown:
   • Required skills: 90%
   • Repository relevance: 85%
   • Experience level: 92%
   • Profile quality: 78%
   
   Potential concerns: Limited gRPC experience

3. ★★★★☆ ana_gomez (85% match)
   Ana Gomez | Lima, Peru
   https://github.com/ana_gomez
   
   [... details ...]

[... 7 more candidates ...]

Summary:
• Found 12 qualified candidates total
• Presenting top 10 ranked by match quality
• Average match score: 78%
• Search quality: Excellent
• All candidates have strong Go and microservices background
• 8 out of 10 have proven MongoDB experience

## 9. KEY DIFFERENCES FROM STAGE 1

### Comparison Table:

| Aspect | Stage 1 | Stage 2 |
|---|---|---|
| Prompts | 1 prompt with tools | 4 specialized prompts |
| Complexity | Simple search | Multi-step pipeline |
| Analysis | Basic filtering | Deep repository analysis |
| Ranking | No ranking | Weighted scoring system |
| Strategy | Fixed approach | Adaptive search strategy |
| Enrichment | Profile data only | Profile + repos + analysis |
| Output | Simple list | Detailed rankings w/ reasoning |
| Execution Time | <30 seconds | 1-2 minutes |
| Requirements Parsing | Within single prompt | Dedicated prompt |
| Search Planning | Fixed query | Primary + fallback strategies |
| Repository Analysis | None | Detailed relevance scoring |
| Score Transparency | None | Multi-factor breakdown |
| Quality Assessment | None | Search quality rating |

### Benefits of Stage 2:
✓ Better understanding of complex requirements\
✓ More intelligent search strategies\
✓ Deeper candidate analysis\
✓ Transparent scoring and reasoning\
✓ Higher quality matches\
✓ More actionable information for recruiters

### Trade-offs:
✗ Longer execution time\
✗ More API calls (higher cost)\
✗ More complex to debug\
✗ Requires more sophisticated error handling

## 10. SUCCESS CRITERIA

### 10.1 Functional Requirements
✓ Successfully parse complex queries into structured requirements\
✓ Generate intelligent search strategies with fallbacks\
✓ Execute searches and enrich with repository data\
✓ Rank candidates with transparent scoring\
✓ Complete pipeline in < 2 minutes

### 10.2 Quality Metrics
- Relevance: Top 5 candidates match >= 80% of requirements (verified by manual review)
- Coverage: Analyze 10-20 candidate repositories
- Accuracy: Scoring aligns with actual candidate fit (verified by manual review)
- Transparency: Clear reasoning for each ranking
- Consistency: Similar queries produce similar results

### 10.3 User Experience
- Detailed candidate profiles with project evidence
- Clear match scores and reasoning
- Transparent about potential gaps or concerns
- Actionable information for recruiter follow-up
- Easy to understand rankings

### 10.4 Performance
- Total execution time: 60-120 seconds
- API calls: 15-30 per query
- Memory usage: Reasonable for web service
- Error rate: < 5% of queries

## 11. LIMITATIONS

### Still Missing (Reserved for Stage 3+):
✗ No self-evaluation or refinement loops
✗ No quality assessment of results
✗ No dynamic strategy adjustment
✗ No learning from feedback
✗ Pipeline is fixed (can't skip or repeat steps)
✗ No conversation/iteration with user
✗ No handling of ambiguous requirements
✗ No cross-platform search (only GitHub)

### By Design (Appropriate for Stage 2):
- Linear pipeline (no branching)
- Fixed scoring weights
- No user interaction during execution
- Stateless (no memory between queries)

## 12. WHAT MAKES THIS STAGE 2?

✓ **Prompt Chaining Pattern**
  - Multiple specialized prompts
  - Sequential execution
  - Each prompt has single responsibility
  - Deterministic flow

✓ **Increased Sophistication**
  - Breaks down complex task
  - Each step adds value
  - Intermediate structured outputs
  - Better results than Stage 1

✓ **Foundation for Stage 3**
  - Stage 3 will add reflection/evaluation
  - Can insert evaluation between steps
  - Can add loops for refinement

### Anthropic's Building Blocks Mapping:
*   Stage 1: Single-Shot Tool = Augmented LLM
*   Stage 2: Multi-Prompt Generator = Prompt Chaining ← **WE ARE HERE**
*   Stage 3: Reflective Agent = Evaluator-Optimizer
*   Stage 4: Integrated Engine = Parallelization + Orchestrator-Workers
*   Stage 5: Full Autonomous Agent = Beyond patterns (Agentic System)

## 13. ERROR HANDLING

### 13.1 Prompt-Level Errors
- Requirements parsing fails → Return error, ask for clarification
- Strategy generation fails → Use default search strategy
- Search execution fails → Try fallback searches
- Ranking fails → Return unranked enriched candidates

### 13.2 Tool-Level Errors
- GitHub API rate limit → Wait and retry or return partial results
- Repository fetch fails → Skip that candidate or use basic profile only
- Network timeout → Retry with exponential backoff

### 13.3 Pipeline Recovery
If any step fails:
1. Log the error with context
2. Attempt recovery if possible
3. Proceed with partial results if acceptable
4. Return informative error message to user

**Example Recovery Strategy:**
Prompt 3 finds only 5 candidates (target was 15)
→ Try fallback search
→ If still insufficient, proceed with what we have
→ Note in final output: "Limited results, consider broader criteria"

### 13.4 Data Validation
Between each prompt:
- Validate JSON structure
- Check required fields present
- Verify data types
- Log warnings for unexpected values

## 14. TESTING STRATEGY

### 14.1 Unit Tests
- Test each prompt function in isolation
- Mock LLM responses with sample data
- Verify JSON parsing/serialization
- Test error handling for each function

### 14.2 Integration Tests
- Test full pipeline with real APIs
- Verify data flows correctly between prompts
- Check tool execution in Prompt 3
- Validate final output format

### 14.3 End-to-End Tests
**Test Cases:**
1. Simple query: "Find Python developers in Lima"
2. Complex query: "Senior Go developers with microservices and MongoDB"
3. Query with no results: "COBOL developers in Antarctica"
4. Query with typos: "Goa develoopers in Limma"
5. Ambiguous query: "Find good developers"

### 14.4 Performance Tests
- Measure execution time per prompt
- Track total pipeline duration
- Monitor API call counts
- Profile memory usage

### 14.5 Quality Tests
- Manual review of top 10 candidates
- Verify relevance scores make sense
- Check reasoning quality
- Validate match breakdown accuracy

## 15. IMPLEMENTATION CHECKLIST

### Phase 1: Core Prompts
- [ ] Implement Requirements Analyzer (Prompt 1)
- [ ] Implement Search Strategy Generator (Prompt 2)
- [ ] Implement Candidate Finder (Prompt 3)
- [ ] Implement Ranker & Presenter (Prompt 4)
- [ ] Create data structures for all intermediate formats

### Phase 2: Tools & Integration
- [ ] Implement get_developer_repositories tool
- [ ] Implement repository relevance analysis
- [ ] Connect prompts in pipeline orchestrator
- [ ] Add error handling between steps
- [ ] Add logging for debugging

### Phase 3: Testing
- [ ] Write unit tests for each prompt
- [ ] Test each prompt individually
- [ ] Test full pipeline end-to-end
- [ ] Test with various query types
- [ ] Validate scoring accuracy

### Phase 4: Optimization
- [ ] Tune prompt instructions
- [ ] Optimize API calls (minimize requests)
- [ ] Improve scoring weights
- [ ] Add better error messages
- [ ] Add execution time monitoring

### Phase 5: Documentation
- [ ] Document each prompt's purpose
- [ ] Create examples for each prompt
- [ ] Document data structures
- [ ] Add troubleshooting guide
- [ ] Create usage examples

**Estimated Implementation Time:** 3-5 days for MVP

## 16. DEPLOYMENT CONSIDERATIONS

### 16.1 Environment Setup
**Required:**
- Go 1.21 or higher
- GitHub Personal Access Token
- Vertex AI Project ID

**Configuration:**
- Set appropriate timeouts for each prompt
- Configure retry logic for API failures
- Set up logging infrastructure

### 16.2 Monitoring
**Track:**
- Execution time per prompt
- Total pipeline duration
- API call counts (GitHub + LLM)
- Success/failure rates
- Quality metrics (match scores)

### 16.3 Cost Management
**LLM API Costs:**
- Prompt 1: ~500 tokens
- Prompt 2: ~800 tokens
- Prompt 3: ~2000 tokens (with tool calls)
- Prompt 4: ~1500 tokens
- Total per query: ~4800 tokens + tool responses

**GitHub API:**
- Stay within rate limits
- Cache results when possible
- Batch requests efficiently

## 17. FUTURE EXTENSIONS (STAGE 3+)

### Planned for Stage 3 (Reflective Agent):
- Add evaluation loop after Prompt 4
- Assess search quality and decide if refinement needed
- Iteratively improve results
- Handle ambiguous requirements through clarification

### Planned for Stage 4 (Integrated Engine):
- Add LinkedIn API integration
- Add internal ATS integration
- Parallel search across multiple platforms
- Orchestrator coordinates multiple specialized agents

### Planned for Stage 5 (Full Autonomous):
- Goal-driven behavior
- Adaptive strategy selection
- Learning from feedback
- Multi-session campaigns

### Not in Scope for Any Stage:
- Automated outreach/contact
- Interview scheduling
- Candidate tracking over time
- Email/message generation

## 18. EXAMPLE QUERIES AND EXPECTED BEHAVIOR

### Example 1: Simple Query
**Input:** "Find Python developers in Peru"

**Expected:**
- Prompt 1: Extract language=Python, location=Peru
- Prompt 2: Generate straightforward search strategy
- Prompt 3: Find 15-20 candidates, analyze repos
- Prompt 4: Rank by Python proficiency and activity

### Example 2: Complex Query
**Input:** "Senior Go developers in Lima with microservices, MongoDB, and preferably gRPC experience"

**Expected:**
- Prompt 1: Separate required vs nice-to-have
- Prompt 2: Create primary + fallback strategies
- Prompt 3: Deep dive into repos for all keywords
- Prompt 4: Weight required skills higher than nice-to-have

### Example 3: Ambiguous Query
**Input:** "Find developers"

**Expected:**
- Prompt 1: Extract minimal requirements
- Prompt 2: Request clarification OR use defaults
- Prompt 3: Broad search
- Prompt 4: Return results with note about ambiguity

### Example 4: No Results Query
**Input:** "Find COBOL developers in Antarctica"

**Expected:**
- Prompt 1: Extract requirements normally
- Prompt 2: Create search strategies with fallbacks
- Prompt 3: Execute searches, return empty
- Prompt 4: Inform user, suggest alternatives

## 19. API CALL PATTERNS

### Typical API Call Sequence:

1. **Prompt 1 (Requirements Analyzer)**
   - 1 LLM API call
   - Input: ~100 tokens
   - Output: ~300 tokens

2. **Prompt 2 (Search Strategy Generator)**
   - 1 LLM API call
   - Input: ~300 tokens
   - Output: ~500 tokens

3. **Prompt 3 (Candidate Finder & Enricher)**
   - 1 LLM API call (with tools)
   - LLM calls `search_github_developers`: 1-2 times
   - LLM calls `get_developer_repositories`: 10-15 times
   - Total GitHub API calls: 12-17
   - Input: ~500 tokens
   - Output: ~1500 tokens

4. **Prompt 4 (Ranker & Presenter)**
   - 1 LLM API call
   - Input: ~1500 tokens
   - Output: ~1000 tokens

**Total:**
- LLM API calls: 4
- GitHub API calls: 12-17
- Total tokens: ~4300
- Execution time: 60-120 seconds

## 20. LLM BUILDING BLOCKS MAPPING

This Stage 2 implements the "Prompt Chaining" pattern:

### Characteristics:
✓ Multiple prompts in sequence
✓ Each prompt has single responsibility
✓ Output of one prompt feeds into next
✓ Deterministic flow (no branching)
✓ Stateless between executions

### Benefits:
✓ Easier to debug (inspect each step)
✓ Easier to optimize (tune individual prompts)
✓ More transparent (see intermediate results)
✓ More modular (swap prompts independently)

### When to Use Prompt Chaining:
✓ Task can be broken into clear sequential steps
✓ Each step benefits from specialized instructions
✓ Intermediate outputs are useful for debugging
✓ No need for loops or branching

### When NOT to Use:
✗ Task is simple enough for single prompt
✗ Steps need to be executed in parallel
✗ Need dynamic branching based on results
✗ Need evaluation/refinement loops

### Next Pattern (Stage 3):
Stage 3 will add "Evaluator-Optimizer" pattern on top of chaining:
- Chain remains the same
- Add evaluation step after Prompt 4
- Loop back if quality insufficient
- This transforms Stage 2 into Stage 3

---

**Document Version:** 1.0\
**Last Updated:** 2025-12-10\
**Target Implementation:** Go 1.21+\
**Pattern:** Prompt Chaining (Stage 2 of 5)\
**Previous Stage:** Stage 1 (Augmented LLM)\
**Next Stage:** Stage 3 (Reflective Agent)
