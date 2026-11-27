package scrape

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CookieCache holds the NYT cookie in memory so fetches do not read the
// database. It loads once, and after Invalidate (call it when a fetch fails) it
// reloads at most once per MinReload, so a broken cookie cannot turn every
// fetch into a database query. To rotate the cookie, update the cookies table;
// the next reload after a failed fetch picks it up.
type CookieCache struct {
	Load      func(ctx context.Context) (string, error)
	MinReload time.Duration

	mu        sync.Mutex
	value     string
	attempted time.Time // last Load attempt
	stale     bool
	loadErr   error
}

// Get returns the cached cookie, loading it first if none is cached or a
// reload is due.
func (c *CookieCache) Get(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.attempted.IsZero() || (c.stale && time.Since(c.attempted) >= c.MinReload) {
		c.attempted = time.Now()
		v, err := c.Load(ctx)
		switch {
		case err != nil:
			c.loadErr = err
			c.stale = true // retry after MinReload
		case v == "":
			c.loadErr = errors.New("stored cookie is empty")
			c.stale = true
		default:
			c.value, c.loadErr, c.stale = v, nil, false
		}
	}

	if c.value == "" {
		return "", fmt.Errorf("no NYT cookie: %w", c.loadErr)
	}
	return c.value, nil
}

// Invalidate marks the cookie as possibly expired so the next Get may reload it.
func (c *CookieCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stale = true
}
