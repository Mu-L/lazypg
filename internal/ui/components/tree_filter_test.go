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
