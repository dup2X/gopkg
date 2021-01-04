// Package cache local cache
package cache

const (
	defaultCap = 4096
)

// Cacher interface
type Cacher interface {
	// Set defined cache_set
	Set(key, val []byte)
	// Get defined cache_get
	Get(key []byte) []byte
	// Del defined cache_delete by key
	Del(key []byte)
	// FlushAll defined flush all cache
	FlushAll()
}

// NewLocal return local cache with default_cap=4096
func NewLocal() Cacher {
	return newLocalCache(defaultCap)
}
