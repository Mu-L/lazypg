package jsonb

import (
	"strings"
)

// IsJSONB checks if a string value looks like JSONB
// This is a fast heuristic check - it does NOT validate JSON syntax.
// Used in render hot path, so must be O(1) without full JSON parsing.
func IsJSONB(value string) bool {
	if value == "" {
		return false
	}

	// Quick check for JSON-like start - O(1) operation
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return false
	}

	first := value[0]
	last := value[len(value)-1]

	// Check for object: starts with { ends with }
	if first == '{' && last == '}' {
		return true
	}

	// Check for array: starts with [ ends with ]
	if first == '[' && last == ']' {
		return true
	}

	return false
}

// Truncate truncates a JSON string for table display
func Truncate(jsonStr string, maxLen int) string {
	if len(jsonStr) <= maxLen {
		return jsonStr
	}

	// Try to truncate at a reasonable boundary
	truncated := jsonStr[:maxLen-3]

	// Find last space, comma, or bracket
	lastGood := strings.LastIndexAny(truncated, " ,{}[]")
	if lastGood > maxLen/2 {
		truncated = truncated[:lastGood]
	}

	return truncated + "..."
}
