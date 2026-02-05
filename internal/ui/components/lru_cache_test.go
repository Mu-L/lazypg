package components

import (
	"testing"
)

func TestLRUCache_GetSet(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Set(0, []string{"row0"})
	cache.Set(1, []string{"row1"})
	cache.Set(2, []string{"row2"})

	row, ok := cache.Get(1)
	if !ok {
		t.Fatal("expected to find row 1")
	}
	if row[0] != "row1" {
		t.Fatalf("expected row1, got %s", row[0])
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set(0, []string{"row0"})
	cache.Set(1, []string{"row1"})
	cache.Set(2, []string{"row2"}) // Should evict row 0

	_, ok := cache.Get(0)
	if ok {
		t.Fatal("expected row 0 to be evicted")
	}

	_, ok = cache.Get(1)
	if !ok {
		t.Fatal("expected row 1 to exist")
	}
}

func TestLRUCache_AccessUpdatesOrder(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set(0, []string{"row0"})
	cache.Set(1, []string{"row1"})

	// Access row 0 to make it recently used
	cache.Get(0)

	// Add row 2, should evict row 1 (least recently used)
	cache.Set(2, []string{"row2"})

	_, ok := cache.Get(1)
	if ok {
		t.Fatal("expected row 1 to be evicted")
	}

	_, ok = cache.Get(0)
	if !ok {
		t.Fatal("expected row 0 to exist")
	}
}
