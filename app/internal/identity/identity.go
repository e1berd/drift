// Package identity resolves the Kratos identity behind an API request and caches it briefly,
// so a burst of requests from one browser costs a single whoami call.
package identity

import (
	"crypto/sha256"
	"net/http"
	"sync"
	"time"

	"github.com/e1berd/drift/app/internal/kratos"
	"github.com/labstack/echo/v5"
)

const (
	contextKey     = "drift.identity"
	roleContextKey = "drift.identity.role"
	superuserRole  = "superuser"
	cacheTTL       = 30 * time.Second
	cacheMax       = 512
)

type entry struct {
	id      string
	role    string
	expires time.Time
}

// Resolver turns the Kratos session cookie into an identity id.
type Resolver struct {
	kratos *kratos.Client
	mu     sync.Mutex
	cache  map[[32]byte]entry
}

// New returns a resolver backed by the Kratos public API.
func New(k *kratos.Client) *Resolver {
	return &Resolver{kratos: k, cache: make(map[[32]byte]entry)}
}

// Require rejects requests without an active Kratos session with 401 and stores the identity id.
func (r *Resolver) Require(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie := c.Request().Header.Get("Cookie")
		if cookie == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "not signed in")
		}
		key := sha256.Sum256([]byte(cookie))
		now := time.Now()

		r.mu.Lock()
		e, ok := r.cache[key]
		r.mu.Unlock()

		if !ok || now.After(e.expires) {
			session, err := r.kratos.Session(c.Request().Context(), cookie)
			if err != nil {
				return err
			}
			if session == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "not signed in")
			}
			e = entry{id: session.Identity.ID, role: session.Identity.Role(), expires: now.Add(cacheTTL)}
			r.store(key, e, now)
		}

		c.Set(contextKey, e.id)
		c.Set(roleContextKey, e.role)
		return next(c)
	}
}

func (r *Resolver) store(key [32]byte, e entry, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.cache) >= cacheMax {
		for k, v := range r.cache {
			if now.After(v.expires) {
				delete(r.cache, k)
			}
		}
		if len(r.cache) >= cacheMax {
			clear(r.cache)
		}
	}
	r.cache[key] = e
}

// From returns the identity id stored by Require.
func From(c *echo.Context) string {
	id, _ := c.Get(contextKey).(string)
	return id
}

// IsSuperuser reports whether the identity stored by Require is the instance owner.
func IsSuperuser(c *echo.Context) bool {
	role, _ := c.Get(roleContextKey).(string)
	return role == superuserRole
}

// RequireSuperuser rejects requests from identities other than the instance owner.
func RequireSuperuser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if !IsSuperuser(c) {
			return echo.NewHTTPError(http.StatusForbidden, "only the instance owner can do this")
		}
		return next(c)
	}
}
