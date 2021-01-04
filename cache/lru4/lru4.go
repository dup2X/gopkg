// Package lru4 ...
package lru4

import (
	"container/list"
	"sync"
)

const defaultStep = 4

// Key is search_key
type Key interface{}

// Cache struct manage mem
type Cache struct {
	// MaxEntries is max
	MaxEntries int

	pool  sync.Pool
	ll    *list.List
	cache map[Key]*list.Element
}

type entry struct {
	key   Key
	value interface{}
}

// New return lru4 cache obj
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[Key]*list.Element),
		pool: sync.Pool{
			New: func() interface{} {
				return &entry{}
			},
		},
	}
}

// Get search k,v
func (c *Cache) Get(key Key) (val interface{}, hit bool) {
	if c.cache == nil {
		return
	}
	if ele, ok := c.cache[key]; ok {
		c.promote(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Add set k,v
func (c *Cache) Add(key Key, val interface{}) {
	if c.cache == nil {
		c.cache = make(map[Key]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.promote(ee)
		ee.Value.(*entry).value = val
		return
	}

	et := c.pool.Get().(*entry)
	et.key, et.value = key, val
	ele := c.ll.PushFront(et)
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.pool.Put(e.Value.(*entry))
}

func (c *Cache) promote(e *list.Element) {
	if e == c.ll.Front() {
		return
	}
	p := e.Prev()
	for i := 0; i < defaultStep && p != c.ll.Front(); i++ {
		p = p.Prev()
	}
}

// RemoveOldest clear back ele
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// Remove clear k
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, ok := c.cache[key]; ok {
		c.removeElement(ele)
	}
}

// FlushAll remove all keys
func (c *Cache) FlushAll() {
	for k := range c.cache {
		c.Remove(k)
	}
}
