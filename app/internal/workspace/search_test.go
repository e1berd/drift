package workspace

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

func TestLikeLiteralEscapesWildcards(t *testing.T) {
	cases := map[string]string{
		"plan":       "plan",
		"100%":       `100\%`,
		"snake_case": `snake\_case`,
		`C:\path`:    `C:\\path`,
	}
	for in, want := range cases {
		if got := likeLiteral(in); got != want {
			t.Errorf("likeLiteral(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSearchRouteIsNotASessionID(t *testing.T) {
	e := echo.New()
	h := &Handler{}
	h.Register(e.Group(""))

	cases := map[string]int{
		"/sessions:search":                   http.StatusBadRequest,
		"/sessions:search?query=" + "%20%20": http.StatusBadRequest,
		"/sessions:search?query=" + strings.Repeat("x", maxSearchQueryRunes+1): http.StatusBadRequest,
	}
	for target, want := range cases {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != want {
			t.Errorf("GET %s = %d, want %d: %s", target, rec.Code, want, rec.Body)
		}
	}
}

func TestSessionPageCutsAtSizeAndPointsPastTheLastRow(t *testing.T) {
	rows := make([]db.Session, 3)
	for i := range rows {
		if err := rows[i].ID.Scan("2f1c2a4e-9d0b-4c8e-8a7e-1b2c3d4e5f6" + string(rune('0'+i))); err != nil {
			t.Fatal(err)
		}
		rows[i].UpdatedAt = pgtype.Timestamptz{Time: time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC), Valid: true}
	}

	page := sessionPage(rows, 2)
	if len(page.Sessions) != 2 || page.NextPageToken != encodeCursor(rows[1].UpdatedAt, rows[1].ID) {
		t.Fatalf("page = %+v", page)
	}
	if last := sessionPage(rows, 3); len(last.Sessions) != 3 || last.NextPageToken != "" {
		t.Fatalf("last page = %+v", last)
	}
}
