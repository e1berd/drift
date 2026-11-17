package workspace

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

func TestCleanRel(t *testing.T) {
	cases := map[string]string{
		"":                 ".",
		"/":                ".",
		"src/../main.go":   "main.go",
		"../../etc/passwd": "etc/passwd",
		"a\\b":             "a/b",
	}
	for in, want := range cases {
		got, err := cleanRel(in)
		if err != nil || got != want {
			t.Errorf("cleanRel(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

func TestCursorRoundTrip(t *testing.T) {
	var id pgtype.UUID
	if err := id.Scan("2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60"); err != nil {
		t.Fatal(err)
	}
	ts := pgtype.Timestamptz{Time: time.Date(2026, 11, 17, 12, 0, 0, 123456000, time.UTC), Valid: true}
	token := encodeCursor(ts, id)

	e := echo.New()
	req := httptest.NewRequest("GET", "/?page_token="+url.QueryEscape(token), nil)
	c := e.NewContext(req, httptest.NewRecorder())
	gotTS, gotID, err := decodeCursor(c)
	if err != nil {
		t.Fatal(err)
	}
	if !gotTS.Time.Equal(ts.Time) || uuidString(gotID) != uuidString(id) {
		t.Fatalf("round trip mismatch: %v %v", gotTS.Time, uuidString(gotID))
	}
}

func TestDraftKey(t *testing.T) {
	ok := []string{"session:2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60", "new:2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60"}
	bad := []string{"session:x", "other:2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60", "new:2F1C2A4E-9D0B-4C8E-8A7E-1B2C3D4E5F60"}
	for _, k := range ok {
		if !draftKey.MatchString(k) {
			t.Errorf("%q should be valid", k)
		}
	}
	for _, k := range bad {
		if draftKey.MatchString(k) {
			t.Errorf("%q should be invalid", k)
		}
	}
}

func TestTitleFrom(t *testing.T) {
	if got := titleFrom("  first line\nsecond"); got != "first line" {
		t.Fatalf("got %q", got)
	}
	long := ""
	for i := 0; i < 70; i++ {
		long += "я"
	}
	if got := []rune(titleFrom(long)); len(got) != 61 {
		t.Fatalf("want 60 runes + ellipsis, got %d", len(got))
	}
}

func TestPathKeyUnescapes(t *testing.T) {
	e := echo.New()
	e.PUT("/drafts/:key", func(c *echo.Context) error {
		key, err := pathKey(c)
		if err != nil {
			return err
		}
		return c.String(200, key)
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest("PUT", "/drafts/new%3A2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60", nil))
	if rec.Code != 200 || rec.Body.String() != "new:2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f60" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}
