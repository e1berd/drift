package uploads

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	schema "github.com/e1berd/drift/app/db"
	"github.com/e1berd/drift/app/internal/database"
	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/migrate"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"net/url"
)

type fixedLimit int64

func (l fixedLimit) MaxUploadBytes(context.Context) (int64, error) { return int64(l), nil }

func testPool(t *testing.T) *pgxpool.Pool {
	raw := os.Getenv("DRIFT_TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("DRIFT_TEST_DATABASE_URL is not set")
	}
	databaseURL, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool, schema.Migrations, "migrations", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatal(err)
	}
	return pool
}

func testServer(t *testing.T, limit int64) (*httptest.Server, *Server) {
	httpServer, server, _ := testServerWithPool(t, limit)
	return httpServer, server
}

func testServerWithPool(t *testing.T, limit int64) (*httptest.Server, *Server, *pgxpool.Pool) {
	pool := testPool(t)
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server, err := New(store, pool, fixedLimit(limit), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	api := e.Group("/api", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("drift.identity", c.Request().Header.Get("X-Test-Owner"))
			return next(c)
		}
	})
	server.Register(api)
	httpServer := httptest.NewServer(e)
	t.Cleanup(httpServer.Close)
	return httpServer, server, pool
}

func tusRequest(t *testing.T, method, target, owner string, body []byte, headers map[string]string) *http.Response {
	request, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Tus-Resumable", "1.0.0")
	request.Header.Set("X-Test-Owner", owner)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	return response
}

func createUpload(t *testing.T, base, owner, name string, size int) *http.Response {
	metadata := "filename " + base64.StdEncoding.EncodeToString([]byte(name)) + ",filetype " + base64.StdEncoding.EncodeToString([]byte("application/zip"))
	return tusRequest(t, http.MethodPost, base+"/api/uploads", owner, nil, map[string]string{
		"Upload-Length": strconv.Itoa(size), "Upload-Metadata": metadata,
	})
}

func TestResumableUploadBecomesAnAttachment(t *testing.T) {
	httpServer, server := testServer(t, 1<<30)
	content := bytes.Repeat([]byte("drift"), 200_000)
	created := createUpload(t, httpServer.URL, "owner-a", "../archive.zip", len(content))
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d", created.StatusCode)
	}
	location := created.Header.Get("Location")
	half := len(content) / 2

	first := tusRequest(t, http.MethodPatch, location, "owner-a", content[:half], map[string]string{
		"Content-Type": "application/offset+octet-stream", "Upload-Offset": "0",
	})
	if first.StatusCode != http.StatusNoContent || first.Header.Get("Attachment-Id") != "" {
		t.Fatalf("first chunk: %d", first.StatusCode)
	}
	if foreign := tusRequest(t, http.MethodHead, location, "owner-b", nil, nil); foreign.StatusCode != http.StatusNotFound {
		t.Fatalf("someone else's upload answered %d", foreign.StatusCode)
	}
	head := tusRequest(t, http.MethodHead, location, "owner-a", nil, nil)
	if head.Header.Get("Upload-Offset") != strconv.Itoa(half) {
		t.Fatalf("offset after reload %q", head.Header.Get("Upload-Offset"))
	}
	last := tusRequest(t, http.MethodPatch, location, "owner-a", content[half:], map[string]string{
		"Content-Type": "application/offset+octet-stream", "Upload-Offset": strconv.Itoa(half),
	})
	attachmentID := last.Header.Get("Attachment-Id")
	if last.StatusCode != http.StatusNoContent || attachmentID == "" {
		t.Fatalf("last chunk: %d, id %q", last.StatusCode, attachmentID)
	}

	var id pgtype.UUID
	if err := id.Scan(attachmentID); err != nil {
		t.Fatal(err)
	}
	row, err := server.q.GetAttachmentContent(context.Background(), db.GetAttachmentContentParams{ID: id, OwnerID: "owner-a"})
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "archive.zip" || row.Path.String != attachmentID+"/archive.zip" {
		t.Fatalf("row %+v", row)
	}
	stored, err := os.ReadFile(server.store.Path(row.Path.String))
	if err != nil || !bytes.Equal(stored, content) {
		t.Fatalf("stored file differs: %v", err)
	}
	if gone := tusRequest(t, http.MethodHead, location, "owner-a", nil, nil); gone.StatusCode != http.StatusNotFound {
		t.Fatalf("finished upload still answers %d", gone.StatusCode)
	}
}

func TestUploadsAreCreatedWithOrWithoutTrailingSlash(t *testing.T) {
	httpServer, _ := testServer(t, 1<<30)
	for _, path := range []string{"/api/uploads", "/api/uploads/"} {
		response := tusRequest(t, http.MethodPost, httpServer.URL+path, "owner-a", nil, map[string]string{"Upload-Length": "10"})
		if response.StatusCode != http.StatusCreated {
			t.Errorf("POST %s: %d", path, response.StatusCode)
		}
	}
}

func TestUploadsOverTheLimitAreRefused(t *testing.T) {
	httpServer, _ := testServer(t, 1<<20)
	if response := createUpload(t, httpServer.URL, "owner-a", "big.bin", 1<<20+1); response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("too large: %d", response.StatusCode)
	}
	if response := createUpload(t, httpServer.URL, "owner-a", "empty.bin", 0); response.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty: %d", response.StatusCode)
	}
}

func TestCancelledUploadIsDeleted(t *testing.T) {
	httpServer, server := testServer(t, 1<<30)
	created := createUpload(t, httpServer.URL, "owner-a", "a.zip", 100)
	location := created.Header.Get("Location")
	if response := tusRequest(t, http.MethodDelete, location, "owner-b", nil, nil); response.StatusCode != http.StatusNotFound {
		t.Fatalf("someone else cancelled: %d", response.StatusCode)
	}
	if response := tusRequest(t, http.MethodDelete, location, "owner-a", nil, nil); response.StatusCode != http.StatusNoContent {
		t.Fatalf("cancel: %d", response.StatusCode)
	}
	entries, _ := os.ReadDir(server.store.partialDir())
	if len(entries) != 0 {
		t.Fatalf("partial files left: %d", len(entries))
	}
}

func TestImagesKeptInTheDatabaseMoveToDisk(t *testing.T) {
	_, server, pool := testServerWithPool(t, 1<<30)
	ctx := context.Background()
	var id pgtype.UUID
	err := pool.QueryRow(ctx,
		`insert into attachments (owner_id, name, media_type, size, data) values ('owner-a', 'shot.png', 'image/png', 3, '\x010203') returning id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.MoveDatabaseAttachments(ctx); err != nil {
		t.Fatal(err)
	}
	row, err := server.q.GetAttachmentContent(ctx, db.GetAttachmentContentParams{ID: id, OwnerID: "owner-a"})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile(server.store.Path(row.Path.String))
	if err != nil || !bytes.Equal(stored, []byte{1, 2, 3}) {
		t.Fatalf("stored %v, err %v", stored, err)
	}
}
