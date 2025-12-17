// internal/ui/components/tree_filter.go
package components

import (
	"strings"
)

// SearchQuery represents a parsed search query
type SearchQuery struct {
	Pattern    string // The search pattern (after removing prefix/type)
	Negate     bool   // True if query starts with !
	TypeFilter string // Normalized type filter (e.g., "table", "function")
}

// Type prefix mappings
var typePrefixes = map[string]string{
	// Short prefixes
	"t:":   "table",
	"v:":   "view",
	"f:":   "function",
	"s:":   "schema",
	"seq:": "sequence",
	"ext:": "extension",
	"col:": "column",
	"idx:": "index",
	// Long prefixes
	"table:":     "table",
	"view:":      "view",
	"func:":      "function",
	"function:":  "function",
	"schema:":    "schema",
	"sequence:":  "sequence",
	"extension:": "extension",
	"column:":    "column",
	"index:":     "index",
}

// ParseSearchQuery parses a search query string into structured form
// Examples:
//   - "plan" → {Pattern: "plan", Negate: false, TypeFilter: ""}
//   - "!test" → {Pattern: "test", Negate: true, TypeFilter: ""}
//   - "t:plan" → {Pattern: "plan", Negate: false, TypeFilter: "table"}
//   - "!f:get" → {Pattern: "get", Negate: true, TypeFilter: "function"}
func ParseSearchQuery(query string) SearchQuery {
	q := SearchQuery{}

	// Check for negation prefix
	if strings.HasPrefix(query, "!") {
		q.Negate = true
		query = query[1:]
	}

	// Check for type prefix
	queryLower := strings.ToLower(query)
	for prefix, typeName := range typePrefixes {
		if strings.HasPrefix(queryLower, prefix) {
			q.TypeFilter = typeName
			query = query[len(prefix):]
			break
		}
	}

	q.Pattern = query
	return q
}
