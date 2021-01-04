// Package cache defined local cache
package cache

import (
	"sync"

	"github.com/dup2X/gopkg/cache/lru4"
)

type localCache struct {
	rwl   *sync.RWMutex
	store *lru4.Cache
}

var _ Cacher = new(localCache)

func newLocalCache(max int) Cacher {
	return &localCache{
		rwl:   &sync.RWMutex{},
		store: lru4.New(max),
	}
}

func (l *localCache) Set(key, val []byte) {
	l.rwl.Lock()
	l.store.Add(string(key), val)
	l.rwl.Unlock()
}

func (l *localCache) Get(key []byte) []byte {
	l.rwl.RLock()
	if val, ok := l.store.Get(string(key)); ok {
		l.rwl.RUnlock()
		return val.([]byte)
	}
	l.rwl.RUnlock()
	return nil
}

func (l *localCache) Del(key []byte) {
	l.rwl.Lock()
	l.store.Remove(string(key))
	l.rwl.Unlock()
}

func (l *localCache) FlushAll() {
	l.rwl.Lock()
	l.store.FlushAll()
	l.rwl.Unlock()
}
