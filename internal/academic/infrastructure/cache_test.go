package infrastructure_test

import (
	"testing"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/academic/infrastructure"
)

func newTestCache() *infrastructure.AcademicCache {
	// Very short TTL so we can also test natural expiry if needed.
	return infrastructure.NewAcademicCache(1*time.Minute, 2*time.Minute)
}

// TestAcademicCache_SetAndGet verifies that an item stored with Set
// can be retrieved with Get before it expires.
func TestAcademicCache_SetAndGet(t *testing.T) {
	c := newTestCache()

	c.Set("academic:levels:admin:school-1", "payload-A")

	got, found := c.Get("academic:levels:admin:school-1")
	if !found {
		t.Fatal("expected cache hit, got miss")
	}
	if got.(string) != "payload-A" {
		t.Fatalf("expected %q, got %q", "payload-A", got)
	}
}

// TestAcademicCache_Miss verifies that a key that was never set returns
// (nil, false).
func TestAcademicCache_Miss(t *testing.T) {
	c := newTestCache()

	_, found := c.Get("academic:levels:admin:unknown")
	if found {
		t.Fatal("expected cache miss, got hit")
	}
}

// TestAcademicCache_InvalidatePrefix verifies that InvalidatePrefix removes
// all keys sharing a given prefix while leaving unrelated keys intact.
func TestAcademicCache_InvalidatePrefix(t *testing.T) {
	c := newTestCache()

	c.Set("academic:levels:admin:school-1", "A")
	c.Set("academic:levels:superadmin:all", "B")
	c.Set("academic:classes:admin:school-1", "C") // different prefix

	c.InvalidatePrefix("academic:levels:")

	if _, found := c.Get("academic:levels:admin:school-1"); found {
		t.Error("expected 'academic:levels:admin:school-1' to be evicted")
	}
	if _, found := c.Get("academic:levels:superadmin:all"); found {
		t.Error("expected 'academic:levels:superadmin:all' to be evicted")
	}
	if _, found := c.Get("academic:classes:admin:school-1"); !found {
		t.Error("expected 'academic:classes:admin:school-1' to survive invalidation of 'academic:levels:'")
	}
}

// TestAcademicCache_InvalidatePrefix_Empty ensures no panic when the cache is
// empty.
func TestAcademicCache_InvalidatePrefix_Empty(t *testing.T) {
	c := newTestCache()
	c.InvalidatePrefix("academic:levels:") // must not panic
}

// TestAcademicCache_Overwrite verifies that Set replaces an existing value.
func TestAcademicCache_Overwrite(t *testing.T) {
	c := newTestCache()

	c.Set("key", "first")
	c.Set("key", "second")

	got, found := c.Get("key")
	if !found {
		t.Fatal("expected cache hit after overwrite")
	}
	if got.(string) != "second" {
		t.Fatalf("expected %q after overwrite, got %q", "second", got)
	}
}
