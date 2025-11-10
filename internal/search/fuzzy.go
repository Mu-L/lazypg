package search

import "strings"

// Match represents a fuzzy match result
type Match struct {
	Score   int
	Indices []int
	Matched bool
}

// FuzzyMatch performs fuzzy matching between query and target
func FuzzyMatch(query, target string) Match {
	// Convert to lowercase for case-insensitive matching
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	// Empty query matches everything
	if query == "" {
		return Match{Score: 100, Matched: true}
	}

	// Exact substring match gets highest score
	if idx := strings.Index(target, query); idx >= 0 {
		// Earlier matches score higher
		score := 100 - idx
		if score < 50 {
			score = 50
		}
		return Match{Score: score, Matched: true}
	}

	// Character-by-character fuzzy match
	queryIdx := 0
	score := 0
	indices := []int{}
	lastMatchIdx := -1

	for i, ch := range target {
		if queryIdx < len(query) && rune(query[queryIdx]) == ch {
			indices = append(indices, i)

			// Score based on:
			// 1. Position (earlier is better)
			// 2. Consecutive matches (bonus)
			positionScore := 10
			if i < 10 {
				positionScore = 15
			}

			// Consecutive match bonus
			if lastMatchIdx >= 0 && i == lastMatchIdx+1 {
				positionScore += 5
			}

			score += positionScore
			lastMatchIdx = i
			queryIdx++
		}
	}

	// All query characters must be matched
	if queryIdx == len(query) {
		return Match{Score: score, Indices: indices, Matched: true}
	}

	return Match{Matched: false}
}

// RankMatches ranks a list of targets by their fuzzy match score
func RankMatches(query string, targets []string) []int {
	scores := make([]int, len(targets))

	for i, target := range targets {
		match := FuzzyMatch(query, target)
		if match.Matched {
			scores[i] = match.Score
		} else {
			scores[i] = -1 // Not matched
		}
	}

	return scores
}
