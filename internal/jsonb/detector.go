package jsonb

import (
	"encoding/json"
	"strings"
)

// IsJSONB checks if a string value looks like JSONB
func IsJSONB(value string) bool {
	if value == "" {
		return false
	}

	// Quick check for JSON-like start
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return false
	}

	first := value[0]
	if first != '{' && first != '[' && first != '"' {
		// Could be null, true, false, or number
		if value == "null" || value == "true" || value == "false" {
			return true
		}
		// Try parsing as number
		var f float64
		err := json.Unmarshal([]byte(value), &f)
		return err == nil
	}

	// Try to parse as JSON
	var parsed interface{}
	err := json.Unmarshal([]byte(value), &parsed)
	return err == nil
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
