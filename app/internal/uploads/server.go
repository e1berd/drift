package uploads

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/e1berd/drift/app/internal/db"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/tus/tusd/v2/pkg/filestore"
	tus "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/memorylocker"
	expslog "golang.org/x/exp/slog"
)

const (
	routePrefix   = "/api/uploads"
	ownerMetadata = "owner"
	nameMetadata  = "filename"
	typeMetadata  = "filetype"
)

// Limits tells how large an upload may be.
type Limits interface {
	MaxUploadBytes(ctx context.Context) (int64, error)
}

// Server is the tus endpoint at /api/uploads; an upload id becomes the attachment id.
type Server struct {
	store  Store
	files  filestore.FileStore
	tus    *tus.Handler
	q      *db.Queries
	limits Limits
	log    *slog.Logger
}

type ownerKey struct{}

// New returns the tus server writing into store.
func New(store Store, pool *pgxpool.Pool, limits Limits, log *slog.Logger) (*Server, error) {
	s := &Server{store: store, files: filestore.New(store.partialDir()), q: db.New(pool), limits: limits, log: log}
	s.files.DirModePerm, s.files.FileModePerm = 0o700, 0o600
	composer := tus.NewStoreComposer()
	s.files.UseIn(composer)
	memorylocker.New().UseIn(composer)
	handler, err := tus.NewHandler(tus.Config{
		BasePath:                  routePrefix + "/",
		StoreComposer:             composer,
		DisableDownload:           true,
		RespectForwardedHeaders:   true,
		Cors:                      &tus.CorsConfig{Disable: true},
		NetworkTimeout:            2 * time.Minute,
		Logger:                    warningsOnly(),
		PreUploadCreateCallback:   s.beforeCreate,
		PreFinishResponseCallback: s.finish,
	})
	if err != nil {
		return nil, fmt.Errorf("tus handler: %w", err)
	}
	s.tus = handler
	return s, nil
}

// Register mounts the tus routes on a group that already requires a signed-in identity.
func (s *Server) Register(g *echo.Group) {
	serve := echo.WrapHandler(http.StripPrefix(routePrefix, s.tus))
	g.Any("/uploads", s.withOwner(serve))
	g.Any("/uploads/", s.withOwner(serve))
	g.Any("/uploads/:upload", s.withOwner(s.requireOwnUpload(serve)))
}

func (s *Server) withOwner(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		request := c.Request()
		c.SetRequest(request.WithContext(context.WithValue(request.Context(), ownerKey{}, identity.From(c))))
		return next(c)
	}
}

func (s *Server) requireOwnUpload(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		id := strings.TrimSpace(c.Param("upload"))
		if !isUUID(id) {
			return echo.NewHTTPError(http.StatusNotFound, "upload not found")
		}
		upload, err := s.files.GetUpload(c.Request().Context(), id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "upload not found")
		}
		info, err := upload.GetInfo(c.Request().Context())
		if err != nil || info.MetaData[ownerMetadata] != identity.From(c) {
			return echo.NewHTTPError(http.StatusNotFound, "upload not found")
		}
		return next(c)
	}
}

func ownerFrom(ctx context.Context) string {
	owner, _ := ctx.Value(ownerKey{}).(string)
	return owner
}

func warningsOnly() *expslog.Logger {
	return expslog.New(expslog.NewTextHandler(os.Stderr, &expslog.HandlerOptions{Level: expslog.LevelWarn}))
}
