package s4lru

import (
	"fmt"
	"testing"
)

func TestCache(t *testing.T) {

	c := New(4)

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("got a valid from an empty cache")
	}

	c.Set("foo1", "bar1")

	if v, ok := c.Get("foo1"); !ok || v.(string) != "bar1" {
		t.Errorf("failed to get key from cache")
	}

	c.Set("foo2", "bar2")
	c.Get("foo2")
	c.Get("foo1")
	c.Get("foo2")

	if v, ok := c.Get("foo2"); !ok || v.(string) != "bar2" {
		t.Errorf("failed to get key foo2 from cache after promotions")
	}

	c.Remove("foo1")

	if _, ok := c.Get("foo1"); ok {
		t.Errorf("failed to delete key from cache")
	}

	// flush out the cache
	for i := 0; i < 4; i++ {
		key := fmt.Sprintf("extra%d", i)
		c.Set(key, i)
		for j := 0; j < 4; j++ {
			c.Get(key)
		}

	}

	if _, ok := c.Get("foo2"); ok {
		t.Errorf("failed to flush key foo2 from cache")
	}

}
