package repositories

import (
	"container/list"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("cache: key not found")

// Cache is a generic LRU cache implementation that can be used by any repository.
// It is safe for concurrent use: even Get changes the cache (it moves the entry to the front).
type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type cacheEntry[K comparable, V any] struct {
	key   K
	value V
}

// NewCache creates a new cache with the specified capacity
func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves a value from the cache
func (c *Cache[K, V]) Get(key K) (V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, found := c.items[key]; found {
		c.list.MoveToFront(element)
		return element.Value.(*cacheEntry[K, V]).value, nil
	}
	var zero V
	return zero, ErrNotFound
}

// GetAll returns all entries in the cache as a slice of key-value pairs
func (c *Cache[K, V]) GetAll() []struct {
	Key   K
	Value V
} {
	c.mu.Lock()
	defer c.mu.Unlock()
	entries := make([]struct {
		Key   K
		Value V
	}, 0, c.list.Len())
	for element := c.list.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*cacheEntry[K, V])
		entries = append(entries, struct {
			Key   K
			Value V
		}{
			Key:   entry.key,
			Value: entry.value,
		})
	}
	return entries
}

// Set adds or updates a value in the cache
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, found := c.items[key]; found {
		c.list.MoveToFront(element)
		element.Value.(*cacheEntry[K, V]).value = value
		return
	}

	entry := &cacheEntry[K, V]{
		key:   key,
		value: value,
	}
	element := c.list.PushFront(entry)
	c.items[key] = element

	if c.list.Len() > c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.list.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry[K, V]).key)
		}
	}
}

// Delete removes a value from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, found := c.items[key]; found {
		c.list.Remove(element)
		delete(c.items, key)
	}
}

// Clear removes all values from the cache
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*list.Element)
	c.list = list.New()
}

// Len returns the number of items in the cache
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.Len()
}

// Range iterates over all entries in the cache
func (c *Cache[K, V]) Range(f func(key K, value V) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for element := c.list.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*cacheEntry[K, V])
		if !f(entry.key, entry.value) {
			break
		}
	}
}
