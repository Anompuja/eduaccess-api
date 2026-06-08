package infrastructure

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// TeacherCache wraps go-cache to provide specific caching for teacher lists.
type TeacherCache struct {
	store *cache.Cache
}

// NewTeacherCache creates a new TeacherCache with a default expiration and cleanup interval.
func NewTeacherCache(defaultExpiration, cleanupInterval time.Duration) *TeacherCache {
	return &TeacherCache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves an item from the cache. Returns the item and a boolean indicating if it was found.
func (c *TeacherCache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Set adds an item to the cache with the default expiration.
func (c *TeacherCache) Set(key string, value interface{}) {
	c.store.SetDefault(key, value)
}

// InvalidatePrefix removes all items from the cache whose keys start with the given prefix.
func (c *TeacherCache) InvalidatePrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if strings.HasPrefix(k, prefix) {
			c.store.Delete(k)
		}
	}
}
