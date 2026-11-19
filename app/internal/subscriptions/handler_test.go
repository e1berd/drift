package subscriptions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestCacheUsageDropsTransientErrors(t *testing.T) {
	h := Handler{usage: make(map[string]cachedUsage)}
	h.cacheUsage("s1", Usage{})
	if _, ok := h.usage["s1"]; !ok {
		t.Fatal("successful usage must be cached")
	}
	h.cacheUsage("s1", Usage{Unavailable: "error"})
	if _, ok := h.usage["s1"]; ok {
		t.Fatal("transient usage errors must not be cached")
	}
}

func TestCustomMethodRouteKeepsTheVerb(t *testing.T) {
	e := echo.New()
	var got string
	e.POST("/subscriptions/:subscription", func(c *echo.Context) error {
		got = c.Param("subscription")
		return c.NoContent(http.StatusNoContent)
	})
	e.DELETE("/subscriptions/:subscription", func(c *echo.Context) error { return nil })
	e.GET("/subscriptions/:subscription/usage", func(c *echo.Context) error { return nil })

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60:activate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || got != "2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60:activate" {
		t.Fatalf("status %d, param %q", rec.Code, got)
	}
}
