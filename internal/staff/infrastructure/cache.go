package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// StaffCache wraps go-cache to provide specific caching for staff lists.
type StaffCache struct {
	store *cache.Cache
}

// NewStaffCache creates a new StaffCache with a default expiration and cleanup interval.
func NewStaffCache(defaultExpiration, cleanupInterval time.Duration) *StaffCache {
	return &StaffCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *StaffCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *StaffCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *StaffCache) InvalidatePrefix(prefix string) {
	// go-cache doesn't have an out-of-the-box prefix invalidation that doesn't require iteration.
	// We iterate over all items and delete those that match the prefix.
	// This is acceptable for our use case.
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
