package repositories

import (
	"testing"
)

func TestCache_BasicOperations(t *testing.T) {
	cache := NewCache[string, int](3)

	// Test Set and Get
	cache.Set("key1", 1)
	val, err := cache.Get("key1")
	if err != nil || val != 1 {
		t.Errorf("Get after Set failed: got %v, %v, want 1, nil", val, err)
	}

	// Test Get non-existent key
	val, err = cache.Get("nonexistent")
	if err != ErrNotFound || val != 0 {
		t.Errorf("Get non-existent key failed: got %v, %v, want 0, ErrNotFound", val, err)
	}

	// Test Update existing key
	cache.Set("key1", 2)
	val, err = cache.Get("key1")
	if err != nil || val != 2 {
		t.Errorf("Update existing key failed: got %v, %v, want 2, nil", val, err)
	}
}

func TestCache_Capacity(t *testing.T) {
	cache := NewCache[string, int](2)

	// Fill cache to capacity
	cache.Set("key1", 1)
	cache.Set("key2", 2)

	// Add one more item, should evict oldest
	cache.Set("key3", 3)

	// key1 should be evicted
	if _, err := cache.Get("key1"); err != ErrNotFound {
		t.Error("Oldest item was not evicted")
	}

	// key2 and key3 should still be present
	val, err := cache.Get("key2")
	if err != nil || val != 2 {
		t.Errorf("key2 not found or wrong value: got %v, %v, want 2, nil", val, err)
	}
	val, err = cache.Get("key3")
	if err != nil || val != 3 {
		t.Errorf("key3 not found or wrong value: got %v, %v, want 3, nil", val, err)
	}
}

func TestCache_LRUBehavior(t *testing.T) {
	cache := NewCache[string, int](2)

	// Fill cache
	cache.Set("key1", 1)
	cache.Set("key2", 2)

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add new item, should evict key2 (least recently used)
	cache.Set("key3", 3)

	// key1 should still be present
	val, err := cache.Get("key1")
	if err != nil || val != 1 {
		t.Errorf("key1 not found or wrong value: got %v, %v, want 1, nil", val, err)
	}

	// key2 should be evicted
	if _, err := cache.Get("key2"); err != ErrNotFound {
		t.Error("Least recently used item was not evicted")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache[string, int](3)

	// Add items
	cache.Set("key1", 1)
	cache.Set("key2", 2)

	// Delete existing item
	cache.Delete("key1")
	if _, err := cache.Get("key1"); err != ErrNotFound {
		t.Error("Deleted item still exists")
	}

	// Delete non-existent item (should not panic)
	cache.Delete("nonexistent")

	// Other items should still be present
	val, err := cache.Get("key2")
	if err != nil || val != 2 {
		t.Errorf("key2 not found or wrong value: got %v, %v, want 2, nil", val, err)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache[string, int](3)

	// Add items
	cache.Set("key1", 1)
	cache.Set("key2", 2)

	// Clear cache
	cache.Clear()

	// Check length
	if cache.Len() != 0 {
		t.Errorf("Cache not empty after Clear: got %d, want 0", cache.Len())
	}

	// Check items are gone
	if _, err := cache.Get("key1"); err != ErrNotFound {
		t.Error("Item still exists after Clear")
	}
	if _, err := cache.Get("key2"); err != ErrNotFound {
		t.Error("Item still exists after Clear")
	}
}

func TestCache_GetAll(t *testing.T) {
	cache := NewCache[string, int](3)

	// Add items
	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Test GetAll
	entries := cache.GetAll()
	if len(entries) != 3 {
		t.Errorf("GetAll returned wrong number of entries: got %d, want 3", len(entries))
	}

	// Check all items are present with correct values
	expected := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}
	for _, entry := range entries {
		if val, ok := expected[entry.Key]; !ok || val != entry.Value {
			t.Errorf("GetAll entry mismatch: got %v=%v, want %v", entry.Key, entry.Value, val)
		}
	}
}

func TestCache_Range(t *testing.T) {
	cache := NewCache[string, int](3)

	// Add items
	cache.Set("key1", 1)
	cache.Set("key2", 2)
	cache.Set("key3", 3)

	// Test Range
	visited := make(map[string]bool)
	cache.Range(func(key string, value int) bool {
		visited[key] = true
		return true
	})

	// Check all items were visited
	if len(visited) != 3 {
		t.Errorf("Range did not visit all items: got %d, want 3", len(visited))
	}
	for _, key := range []string{"key1", "key2", "key3"} {
		if !visited[key] {
			t.Errorf("Range did not visit %s", key)
		}
	}

	// Test Range with early exit
	count := 0
	cache.Range(func(key string, value int) bool {
		count++
		return count < 2 // Stop after visiting 2 items
	})

	if count != 2 {
		t.Errorf("Range did not stop after visiting 2 items: got %d, want 2", count)
	}
}

func TestCache_ComplexTypes(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}

	cache := NewCache[string, TestStruct](2)

	// Test with struct values
	value1 := TestStruct{ID: 1, Name: "test1"}
	value2 := TestStruct{ID: 2, Name: "test2"}

	cache.Set("key1", value1)
	cache.Set("key2", value2)

	// Test retrieval
	val, err := cache.Get("key1")
	if err != nil || val != value1 {
		t.Errorf("Get struct value failed: got %v, %v, want %v, nil", val, err, value1)
	}

	// Test update
	value1.Name = "updated"
	cache.Set("key1", value1)
	val, err = cache.Get("key1")
	if err != nil || val.Name != "updated" {
		t.Errorf("Update struct value failed: got %v, %v, want %v, nil", val, err, value1)
	}
}
