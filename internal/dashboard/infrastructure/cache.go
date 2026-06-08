package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// DashboardCache wraps go-cache to provide specific caching for dashboard stats.
type DashboardCache struct {
	store *cache.Cache
}

// NewDashboardCache creates a new DashboardCache with a default expiration and cleanup interval.
func NewDashboardCache(defaultExpiration, cleanupInterval time.Duration) *DashboardCache {
	return &DashboardCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *DashboardCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *DashboardCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *DashboardCache) InvalidatePrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
