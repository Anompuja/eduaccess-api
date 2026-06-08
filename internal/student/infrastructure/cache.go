package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// StudentCache wraps go-cache to provide specific caching for student lists.
type StudentCache struct {
	store *cache.Cache
}

// NewStudentCache creates a new StudentCache with a default expiration and cleanup interval.
func NewStudentCache(defaultExpiration, cleanupInterval time.Duration) *StudentCache {
	return &StudentCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *StudentCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *StudentCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *StudentCache) InvalidatePrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
