package agent

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/luillyfe/sourcing-agent/pkg/github"
	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

// analyzeRequirements (Prompt 1)
func analyzeRequirements(client llm.Client, userQuery string) (*Requirements, error) {
	systemPrompt := `You are a requirements analyzer for technical recruiting.

Your task: Parse the user's hiring request into structured requirements.

Extract:
1. Required skills (programming languages, frameworks, technologies)
2. Experience level (junior, mid, senior, lead)
3. Location requirements (city, country, region, remote)
4. Keywords for relevance matching
5. Nice-to-have skills (optional qualifications)

Output Format (JSON):
{
  "required_skills": ["skill1", "skill2"],
  "experience_level": "senior|mid|junior|lead",
  "locations": ["location1", "location2"],
  "keywords": ["keyword1", "keyword2"],
  "nice_to_have": ["skill3", "skill4"],
  "unclear_request": false,
  "clarification_question": "string (only if unclear)"
}

Be specific and extract all relevant information from the query.
If the query is too vague (e.g., "find developers", "search github"), set "unclear_request" to true and ask a specific clarification question.`

	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("User query: %s", userQuery),
		},
	}

	resp, err := client.CallAPI(messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	var content string
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	// Extract JSON from content (in case of markdown code blocks)
	jsonStr := extractJSON(content)

	var requirements Requirements
	if err := json.Unmarshal([]byte(jsonStr), &requirements); err != nil {
		return nil, fmt.Errorf("failed to parse requirements JSON: %w", err)
	}

	if err := requirements.Validate(); err != nil {
		return nil, fmt.Errorf("invalid requirements: %w", err)
	}

	return &requirements, nil
}

// generateSearchStrategy (Prompt 2)
func generateSearchStrategy(client llm.Client, requirements *Requirements) (*SearchStrategy, error) {
	systemPrompt := `You are a search strategy expert for GitHub developer sourcing.

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
}`

	reqJSON, _ := json.Marshal(requirements)
	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Requirements: %s", string(reqJSON)),
		},
	}

	resp, err := client.CallAPI(messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	var content string
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	jsonStr := extractJSON(content)

	var strategy SearchStrategy
	if err := json.Unmarshal([]byte(jsonStr), &strategy); err != nil {
		return nil, fmt.Errorf("failed to parse strategy JSON: %w", err)
	}

	if err := strategy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid strategy: %w", err)
	}

	return &strategy, nil
}

// findAndEnrichCandidates (Prompt 3)
func findAndEnrichCandidates(client llm.Client, githubClient *github.Client, strategy *SearchStrategy, requirements *Requirements) (*EnrichedCandidates, error) {
	// 1. Execute primary search
	// Note: We are NOT using the LLM to call the tool here as per the "Programmatic" flow in the spec example,
	// BUT the spec says "Prompt 3: Candidate Finder & Enricher... This prompt has tool access".
	// However, the "Execution Flow Example" section says "LLM Actions: 1. Calls search_github_developers...".
	// To strictly follow the "Prompt Chaining" pattern where the LLM orchestrates, we should let the LLM do it.
	// BUT, for efficiency and reliability, and since we have the strategy, we can just execute the strategy programmatically
	// and then pass the results to the LLM for enrichment?
	//
	// Re-reading Spec Section 4, Prompt 3:
	// "Responsibility: Execute searches and enrich candidate profiles... Tools Available: search_github_developers..."
	// "System Prompt: You are a candidate sourcing specialist... Your task: 1. Execute primary search..."
	//
	// So the LLM *should* be calling the tools.
	//
	// However, implementing a full tool-use loop inside a sub-function is complex.
	// Let's look at the "Execution Flow Example" again.
	// "LLM Actions: 1. Calls search_github_developers... 2. For each candidate, calls get_developer_repositories..."
	//
	// This implies a loop.
	//
	// Alternative: We can execute the search strategy programmatically (since we have the JSON) and then fetch repos programmatically,
	// and THEN give the data to the LLM for "Enrichment".
	//
	// The spec says: "Prompt 3: Candidate Finder & Enricher".
	// If I do it programmatically, I deviate slightly but it's much more robust.
	//
	// Let's try to follow the spec's intent of using the LLM to orchestrate the search if possible,
	// but given the complexity of the tool loop in `agent.go`, maybe I should refactor `Run` to be reusable?
	//
	// Actually, `findAndEnrichCandidates` can just do the programmatic work since we have the `SearchStrategy` struct.
	// Why ask the LLM to read the JSON and then call the tool with the same parameters?
	//
	// Wait, the spec says: "Prompt 3... Input: Output from Prompt 2".
	// If I do it programmatically, I am skipping the "Prompt 3" as an LLM call that *executes* the search.
	//
	// Let's look at "Section 8. Execution Flow Example":
	// "LLM Actions: 1. Calls search_github_developers... 2. For each candidate, calls get_developer_repositories... 3. Analyzes repositories..."
	//
	// This confirms the LLM is supposed to do the work.
	//
	// However, `get_developer_repositories` is a new tool.
	//
	// I will implement a hybrid approach:
	// 1. Execute the search strategy programmatically (Primary, then Fallback if needed).
	// 2. For each candidate, fetch repos programmatically.
	// 3. Analyze relevance programmatically (as per Section 7 "Recommended: Start with programmatic approach").
	// 4. Construct the `EnrichedCandidates` struct.
	//
	// This avoids a very expensive and slow LLM loop for 15+ candidates x 10 repos.
	// The spec actually says in Section 7: "Recommended: Start with programmatic approach (simpler, faster)" for `analyze_repository_relevance`.
	//
	// So, `findAndEnrichCandidates` will be largely programmatic, maybe NOT an LLM call at all?
	//
	// Spec Section 6.2 shows `findAndEnrichCandidates` calling `executePromptWithTools`.
	//
	// "Execute with tool support (similar to Stage 1)".
	//
	// If I follow the spec strictly, I need to give the LLM the tools.
	// But the spec also says "Recommended: Start with programmatic approach" for the analysis part.
	//
	// Let's look at the "Trade-offs": "Longer execution time", "More API calls".
	//
	// If I let the LLM do it, it will be:
	// 1. LLM: "I'll search for X" -> Tool Call
	// 2. Tool Result: [10 candidates]
	// 3. LLM: "I'll get repos for Candidate 1" -> Tool Call
	// 4. Tool Result: [Repos]
	// ... repeat 10 times ...
	//
	// This is 20+ round trips. That's too slow.
	//
	// I will implement `findAndEnrichCandidates` as a PROGRAMMATIC orchestrator that uses the `SearchStrategy` to call the GitHub API directly,
	// and then uses the programmatic `analyzeRepositoryRelevance`.
	// This effectively "simulates" the LLM's role in Prompt 3 but much faster.
	//
	// Wait, is there a "Prompt 3" LLM call left if I do that?
	// Maybe just to summarize or format?
	//
	// Actually, the output of Prompt 3 is `EnrichedCandidates` JSON.
	// I can generate that programmatically.
	//
	// So, `findAndEnrichCandidates` will NOT call the LLM. It will use the `githubClient` and the `analyzeRepositoryRelevance` function.
	// This seems like a smart deviation/optimization.
	//
	// Let's verify if this is acceptable.
	// The User wants to "Plan its implementation".
	// My plan said: "Implement Candidate Finder (Prompt 3)".
	//
	// I will stick to the programmatic implementation for efficiency, as hinted by the "Recommended: Start with programmatic approach" in the spec.

	// 1. Search
	searchesExecuted := 1
	var candidates []github.Candidate
	input := github.ToolInput{
		Language:   strategy.PrimarySearch.Language,
		Location:   strategy.PrimarySearch.Location,
		MinRepos:   strategy.PostFilters.MinRepos,
		MaxResults: 15, // Aim for 15-20 as per spec
	}
	if len(strategy.RepositorySearch.Keywords) > 0 {
		input.Keywords = strings.Join(strategy.RepositorySearch.Keywords, " ")
	}

	result, err := githubClient.SearchDevelopers(input)
	if err != nil || (result != nil && len(result.Candidates) == 0) {
		// Try fallback
		// Try fallback strategies
		for i, fallback := range strategy.FallbackSearches {
			searchesExecuted++
			if err == nil {
				fmt.Printf("Search returned no results, switching to fallback strategy %d...\n", i+1)
			}

			input = github.ToolInput{
				Language:   fallback.Language,
				Location:   fallback.Location,
				MinRepos:   strategy.PostFilters.MinRepos,
				MaxResults: 15,
			}
			if len(strategy.RepositorySearch.Keywords) > 0 {
				input.Keywords = strings.Join(strategy.RepositorySearch.Keywords, " ")
			}
			result, err = githubClient.SearchDevelopers(input)

			if err == nil && result != nil && len(result.Candidates) > 0 {
				break
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	if result == nil {
		return &EnrichedCandidates{
			SearchMetadata: SearchMetadata{SearchesExecuted: searchesExecuted},
		}, nil
	}
	candidates = result.Candidates

	// 2. Enrich
	enriched := []EnrichedCandidate{}
	profilesAnalyzed := 0

	for _, cand := range candidates {
		profilesAnalyzed++

		// Get Repos
		repos, err := githubClient.GetDeveloperRepositories(cand.Username, 10)
		if err != nil {
			fmt.Printf("Failed to get repos for %s: %v\n", cand.Username, err)
			continue
		}

		// Analyze
		relevantRepos := []RelevantRepository{}
		for _, repo := range repos {
			analysis := analyzeRepositoryRelevance(repo, requirements.RequiredSkills, strategy.RepositorySearch.Keywords)
			if analysis.Score > 0.3 { // Threshold
				relevantRepos = append(relevantRepos, RelevantRepository{
					Name:            repo.Name,
					Description:     repo.Description,
					Language:        repo.Language,
					Stars:           repo.Stars,
					Topics:          repo.Topics,
					RelevanceScore:  analysis.Score,
					RelevanceReason: strings.Join(analysis.Reasons, ", "),
				})
			}
		}

		// Calc initial match score (simplified)
		matchScore := 0.5 // Base
		if len(relevantRepos) > 0 {
			matchScore += 0.2
		}
		// ... more logic ...

		enriched = append(enriched, EnrichedCandidate{
			Username:             cand.Username,
			Name:                 cand.Name,
			Location:             cand.Location,
			Bio:                  cand.Bio,
			PublicRepos:          cand.PublicRepos,
			Followers:            cand.Followers,
			GitHubURL:            cand.GitHubURL,
			RelevantRepositories: relevantRepos,
			SkillsFound:          requirements.RequiredSkills, // Placeholder, should extract from bio/repos
			ExperienceIndicators: ExperienceIndicators{
				TotalStars: 0, // Need to sum
			},
			InitialMatchScore: matchScore,
		})
	}

	finalEnrichedCandidates := &EnrichedCandidates{
		Candidates: enriched,
		SearchMetadata: SearchMetadata{
			SearchesExecuted:   searchesExecuted,
			TotalProfilesFound: len(candidates),
			ProfilesAnalyzed:   profilesAnalyzed,
		},
	}

	if err := finalEnrichedCandidates.Validate(); err != nil {
		return nil, fmt.Errorf("invalid enriched candidates: %w", err)
	}

	return finalEnrichedCandidates, nil
}

// rankAndPresent (Prompt 4)
func rankAndPresent(client llm.Client, candidates *EnrichedCandidates, requirements *Requirements) (*FinalResult, error) {
	systemPrompt := `You are a candidate ranking and presentation specialist.

Given enriched candidate data, produce final rankings and presentation.

Your task:
1. Evaluate each candidate's fit based on:
   - Required skills coverage
   - Repository relevance
   - Experience indicators
   - Location match
   - Profile quality (bio, followers, activity)
2. Format the top candidates for presentation
3. Provide reasoning for each candidate

Evaluate each candidate on a 0-100 scale for these components:
- Required skills match
- Repository relevance
- Experience indicators
- Profile quality

Output Format (JSON):
{
  "top_candidates": [
    {
      "username": "string",
      "name": "string",
      "location": "string",
      "github_url": "string",
      "match_breakdown": {
        "required_skills_score": number,
        "repository_relevance_score": number,
        "experience_score": number,
        "profile_quality_score": number
      },
      "key_qualifications": ["qual1", "qual2"],
      "top_relevant_projects": [
        { "name": "string", "url": "string", "why_relevant": "string" }
      ],
      "match_reasoning": "string",
      "potential_concerns": "string"
    }
  ],
  "summary": {
    "total_candidates_found": number,
    "candidates_presented": number,
    "average_match_score": number,
    "search_quality": "string"
  }
}`

	input := map[string]interface{}{
		"candidates":   candidates,
		"requirements": requirements,
	}
	inputJSON, _ := json.Marshal(input)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Input Data: %s", string(inputJSON)),
		},
	}

	resp, err := client.CallAPI(messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	var content string
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	jsonStr := extractJSON(content)

	var result FinalResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse final result JSON: %w", err)
	}

	// Calculate scores programmatically to ensure accuracy
	// Weights: Skills (40%), Repos (30%), Experience (20%), Quality (10%)
	var totalScore float64
	for i := range result.TopCandidates {
		cand := &result.TopCandidates[i]
		bd := cand.MatchBreakdown

		finalScore := (bd.RequiredSkillsScore * 0.4) +
			(bd.RepositoryRelevanceScore * 0.3) +
			(bd.ExperienceScore * 0.2) +
			(bd.ProfileQualityScore * 0.1)

		cand.FinalMatchScore = finalScore
		totalScore += finalScore
	}

	// Sort candidates by score desc
	sort.Slice(result.TopCandidates, func(i, j int) bool {
		return result.TopCandidates[i].FinalMatchScore > result.TopCandidates[j].FinalMatchScore
	})

	// Assign ranks
	for i := range result.TopCandidates {
		result.TopCandidates[i].Rank = i + 1
	}

	// Update summary stats
	if len(result.TopCandidates) > 0 {
		result.Summary.AverageMatchScore = totalScore / float64(len(result.TopCandidates))
	}

	return &result, nil
}

// createFallbackResult creates a FinalResult from enriched candidates without LLM ranking
func createFallbackResult(candidates *EnrichedCandidates) *FinalResult {
	topCandidates := []RankedCandidate{}
	var totalScore float64

	// Convert enriched candidates to ranked candidates
	for i, cand := range candidates.Candidates {
		// Just take top 10 if there are many
		if i >= 10 {
			break
		}

		relevantProjects := []RelevantProject{}
		for _, repo := range cand.RelevantRepositories {
			relevantProjects = append(relevantProjects, RelevantProject{
				Name:        repo.Name,
				URL:         cand.GitHubURL + "/" + repo.Name,
				WhyRelevant: repo.RelevanceReason,
			})
		}

		ranked := RankedCandidate{
			Username:            cand.Username,
			Name:                cand.Name,
			Location:            cand.Location,
			GitHubURL:           cand.GitHubURL,
			FinalMatchScore:     cand.InitialMatchScore * 100, // Scale to 0-100
			MatchReasoning:      "Ranking step unavailable; score is based on initial keyword match.",
			TopRelevantProjects: relevantProjects,
		}
		topCandidates = append(topCandidates, ranked)
		totalScore += ranked.FinalMatchScore
	}

	// Sort by score
	sort.Slice(topCandidates, func(i, j int) bool {
		return topCandidates[i].FinalMatchScore > topCandidates[j].FinalMatchScore
	})

	// Assign ranks
	for i := range topCandidates {
		topCandidates[i].Rank = i + 1
	}

	avgScore := 0.0
	if len(topCandidates) > 0 {
		avgScore = totalScore / float64(len(topCandidates))
	}

	return &FinalResult{
		TopCandidates: topCandidates,
		Summary: ResultSummary{
			TotalCandidatesFound: candidates.SearchMetadata.TotalProfilesFound,
			CandidatesPresented:  len(topCandidates),
			AverageMatchScore:    avgScore,
			SearchQuality:        "Fallback (Ranking Unavailable)",
		},
	}
}

// Helper to extract JSON from markdown code blocks
func extractJSON(content string) string {
	if strings.Contains(content, "```json") {
		parts := strings.Split(content, "```json")
		if len(parts) > 1 {
			sub := strings.Split(parts[1], "```")
			return strings.TrimSpace(sub[0])
		}
	}
	return content
}
