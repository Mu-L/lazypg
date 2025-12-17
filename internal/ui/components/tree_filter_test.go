// internal/ui/components/tree_filter_test.go
package components

import (
	"testing"
)

func TestParseSearchQuery_Simple(t *testing.T) {
	q := ParseSearchQuery("plan")

	if q.Pattern != "plan" {
		t.Errorf("expected pattern 'plan', got '%s'", q.Pattern)
	}
	if q.Negate {
		t.Error("expected Negate=false")
	}
	if q.TypeFilter != "" {
		t.Errorf("expected empty TypeFilter, got '%s'", q.TypeFilter)
	}
}

func TestParseSearchQuery_Negate(t *testing.T) {
	q := ParseSearchQuery("!test")

	if q.Pattern != "test" {
		t.Errorf("expected pattern 'test', got '%s'", q.Pattern)
	}
	if !q.Negate {
		t.Error("expected Negate=true")
	}
}

func TestParseSearchQuery_TypeShort(t *testing.T) {
	q := ParseSearchQuery("t:plan")

	if q.Pattern != "plan" {
		t.Errorf("expected pattern 'plan', got '%s'", q.Pattern)
	}
	if q.TypeFilter != "table" {
		t.Errorf("expected TypeFilter 'table', got '%s'", q.TypeFilter)
	}
}

func TestParseSearchQuery_TypeLong(t *testing.T) {
	q := ParseSearchQuery("table:plan")

	if q.Pattern != "plan" {
		t.Errorf("expected pattern 'plan', got '%s'", q.Pattern)
	}
	if q.TypeFilter != "table" {
		t.Errorf("expected TypeFilter 'table', got '%s'", q.TypeFilter)
	}
}

func TestParseSearchQuery_NegateWithType(t *testing.T) {
	q := ParseSearchQuery("!f:get")

	if q.Pattern != "get" {
		t.Errorf("expected pattern 'get', got '%s'", q.Pattern)
	}
	if !q.Negate {
		t.Error("expected Negate=true")
	}
	if q.TypeFilter != "function" {
		t.Errorf("expected TypeFilter 'function', got '%s'", q.TypeFilter)
	}
}

func TestParseSearchQuery_TypeOnlyNoPattern(t *testing.T) {
	q := ParseSearchQuery("!t:")

	if q.Pattern != "" {
		t.Errorf("expected empty pattern, got '%s'", q.Pattern)
	}
	if !q.Negate {
		t.Error("expected Negate=true")
	}
	if q.TypeFilter != "table" {
		t.Errorf("expected TypeFilter 'table', got '%s'", q.TypeFilter)
	}
}

func TestFuzzyMatch_ExactPrefix(t *testing.T) {
	match, positions := FuzzyMatch("plan", "plan_check_run")

	if !match {
		t.Error("expected match")
	}
	if len(positions) != 4 || positions[0] != 0 || positions[1] != 1 || positions[2] != 2 || positions[3] != 3 {
		t.Errorf("expected positions [0,1,2,3], got %v", positions)
	}
}

func TestFuzzyMatch_Subsequence(t *testing.T) {
	match, positions := FuzzyMatch("pcr", "plan_check_run")

	if !match {
		t.Error("expected match")
	}
	// p=0, c=5, r=11
	if len(positions) != 3 {
		t.Errorf("expected 3 positions, got %d", len(positions))
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	match, _ := FuzzyMatch("xyz", "plan_check_run")

	if match {
		t.Error("expected no match")
	}
}

func TestFuzzyMatch_CaseInsensitive(t *testing.T) {
	match, _ := FuzzyMatch("PLAN", "plan_check_run")

	if !match {
		t.Error("expected case-insensitive match")
	}
}

func TestFuzzyMatch_EmptyPattern(t *testing.T) {
	match, positions := FuzzyMatch("", "anything")

	if !match {
		t.Error("empty pattern should match everything")
	}
	if len(positions) != 0 {
		t.Error("empty pattern should have no positions")
	}
}
