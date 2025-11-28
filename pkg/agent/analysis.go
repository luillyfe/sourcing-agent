package agent

import (
	"fmt"
	"strings"

	"github.com/luillyfe/sourcing-agent/pkg/github"
)

// analyzeRepositoryRelevance analyzes a repository's relevance to job requirements
func analyzeRepositoryRelevance(repo github.Repository, requiredSkills []string, keywords []string) RelevanceAnalysis {
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

	// Cap score at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return RelevanceAnalysis{
		Score:   score,
		Reasons: reasons,
	}
}
