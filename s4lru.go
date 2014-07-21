// Package s4lru implements the 4-segmented LRU caching algorithm
/*

From http://www.cs.cornell.edu/~qhuang/papers/sosp_fbanalysis.pdf

    Four queues are maintained at levels 0 to 3. On a cache miss, the item is
    inserted at the head of queue 0. On a cache hit, the item is moved to the
    head of the next higher queue (items in queue 3 move to the head of queue
    3).  Each queue is allocated 1/4 of the total cache size and items are
    evicted from the tail of a queue to the head of the next lower queue to
    maintain the size invariants. Items evicted from queue 0 are evicted from
    the cache

*/
package s4lru

import (
	"container/list"
	"sync"
)

type cacheItem struct {
	lidx  int
	key   string
	value interface{}
}

// Cache is an LRU cache
type Cache struct {
	mu       sync.Mutex
	capacity int
	data     map[string]*list.Element
	lists    []*list.List
}

// New returns a new S4LRU cache that with the given capacity.  Each of the lists will have 1/4 of the capacity.
func New(capacity int) *Cache {
	return &Cache{
		capacity: capacity / 4,
		data:     make(map[string]*list.Element),
		lists:    []*list.List{list.New(), list.New(), list.New(), list.New()},
	}
}

// Get returns a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.data[key]

	if !ok {
		return nil, false
	}

	item := v.Value.(*cacheItem)

	// already on final list?
	if item.lidx == len(c.lists)-1 {
		c.lists[item.lidx].MoveToFront(v)
	} else {
		// move to head of next list
		c.lists[item.lidx].Remove(v)
		delete(c.data, key)
		item.lidx++

		// next list is full, so move the last element of that one to the front of this list
		if c.lists[item.lidx].Len() == c.capacity {
			back := c.lists[item.lidx].Back()
			old := c.lists[item.lidx].Remove(back).(*cacheItem)
			old.lidx--
			c.data[old.key] = c.lists[old.lidx].PushFront(old)
		}

		c.data[key] = c.lists[item.lidx].PushFront(item)
	}

	return item.value, true
}

// Set sets a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lists[0].Len() == c.capacity {
		delete(c.data, c.lists[0].Back().Value.(*cacheItem).key)
		c.lists[0].Remove(c.lists[0].Back())
	}
	c.data[key] = c.lists[0].PushFront(&cacheItem{0, key, value})
}

// Len returns the total number of items in the cache
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data)
}

// Remove removes an item from the cache, returning the item and a boolean indicating if it was found
func (c *Cache) Remove(key string) (interface{}, bool) {

	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.data[key]

	if !ok {
		return nil, false
	}

	item := v.Value.(*cacheItem)

	c.lists[item.lidx].Remove(v)

	delete(c.data, key)

	return item.value, true
}
