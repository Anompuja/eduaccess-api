package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// AdminCache wraps go-cache to provide specific caching for admin lists.
type AdminCache struct {
	store *cache.Cache
}

// NewAdminCache creates a new AdminCache with a default expiration and cleanup interval.
func NewAdminCache(defaultExpiration, cleanupInterval time.Duration) *AdminCache {
	return &AdminCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *AdminCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *AdminCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *AdminCache) InvalidatePrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
