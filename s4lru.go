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

import "container/list"

type cacheItem struct {
	lidx  int
	key   string
	value interface{}
}

// Cache is an LRU cache.  It is not safe for concurrent access.
type Cache struct {
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
	v, ok := c.data[key]

	if !ok {
		return nil, false
	}

	item := v.Value.(*cacheItem)

	// already on final list?
	if item.lidx == len(c.lists)-1 {
		c.lists[item.lidx].MoveToFront(v)
		return item.value, true
	}

	// is there space on the next list?
	if c.lists[item.lidx+1].Len() < c.capacity {
		// just do the remove/add
		c.lists[item.lidx].Remove(v)
		item.lidx++
		c.data[key] = c.lists[item.lidx].PushFront(item)
		return item.value, true
	}

	// no free space on either list, so we do some in-place swapping to avoid allocations
	// the key/value in bitem need to be moved to the front of c.lists[item.lidx]
	// the key/value in item need to be moved to the front of c.lists[bitem.lidx]
	back := c.lists[item.lidx+1].Back()
	bitem := back.Value.(*cacheItem)

	// swap the key/values
	bitem.key, item.key = item.key, bitem.key
	bitem.value, item.value = item.value, bitem.value

	// update pointers in the map
	c.data[item.key] = v
	c.data[bitem.key] = back

	// move the elements to the front of their lists
	c.lists[item.lidx].MoveToFront(v)
	c.lists[bitem.lidx].MoveToFront(back)

	return bitem.value, true
}

// Set sets a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	if c.lists[0].Len() < c.capacity {
		c.data[key] = c.lists[0].PushFront(&cacheItem{0, key, value})
		return
	}

	// reuse the tail item
	e := c.lists[0].Back()
	item := e.Value.(*cacheItem)

	delete(c.data, item.key)
	item.key = key
	item.value = value
	c.data[key] = e
	c.lists[0].MoveToFront(e)
}

// Len returns the total number of items in the cache
func (c *Cache) Len() int {
	return len(c.data)
}

// Remove removes an item from the cache, returning the item and a boolean indicating if it was found
func (c *Cache) Remove(key string) (interface{}, bool) {
	v, ok := c.data[key]

	if !ok {
		return nil, false
	}

	item := v.Value.(*cacheItem)

	c.lists[item.lidx].Remove(v)

	delete(c.data, key)

	return item.value, true
}
