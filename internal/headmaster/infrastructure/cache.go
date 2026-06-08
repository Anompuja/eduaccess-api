package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// HeadmasterCache wraps go-cache to provide specific caching for headmaster lists.
type HeadmasterCache struct {
	store *cache.Cache
}

// NewHeadmasterCache creates a new HeadmasterCache with a default expiration and cleanup interval.
func NewHeadmasterCache(defaultExpiration, cleanupInterval time.Duration) *HeadmasterCache {
	return &HeadmasterCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *HeadmasterCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *HeadmasterCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *HeadmasterCache) InvalidatePrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
