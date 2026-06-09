package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// AcademicCache wraps go-cache to provide specific caching for academic entities.
type AcademicCache struct {
	store *cache.Cache
}

// NewAcademicCache creates a new AcademicCache with a default expiration and cleanup interval.
func NewAcademicCache(defaultExpiration, cleanupInterval time.Duration) *AcademicCache {
	return &AcademicCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *AcademicCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *AcademicCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *AcademicCache) InvalidatePrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
