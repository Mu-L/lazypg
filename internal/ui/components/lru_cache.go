package components

import (
	"container/list"
	"sync"
)

// LRUCache is a thread-safe LRU cache for table rows
type LRUCache struct {
	capacity int
	cache    map[int]*list.Element
	order    *list.List
	mu       sync.RWMutex
}

type cacheEntry struct {
	key   int
	value []string
}

// NewLRUCache creates a new LRU cache with the given capacity
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves a row from the cache
func (c *LRUCache) Get(key int) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}

// Set adds or updates a row in the cache
func (c *LRUCache) Set(key int, value []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	// Evict if at capacity
	if c.order.Len() >= c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.cache, oldest.Value.(*cacheEntry).key)
		}
	}

	entry := &cacheEntry{key: key, value: value}
	elem := c.order.PushFront(entry)
	c.cache[key] = elem
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[int]*list.Element)
	c.order = list.New()
}

// Len returns the number of entries in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}
