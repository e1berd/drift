package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	schema "github.com/e1berd/drift/app/db"
	"github.com/e1berd/drift/app/internal/auth"
	"github.com/e1berd/drift/app/internal/config"
	"github.com/e1berd/drift/app/internal/database"
	"github.com/e1berd/drift/app/internal/events"
	"github.com/e1berd/drift/app/internal/frontend"
	"github.com/e1berd/drift/app/internal/health"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/e1berd/drift/app/internal/kratos"
	"github.com/e1berd/drift/app/internal/migrate"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	if err := run(); err != nil {
		slog.Error("drift stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, schema.Migrations, "migrations", log); err != nil {
		return err
	}

	fe, err := frontend.New(cfg.ViteDevURL)
	if err != nil {
		return err
	}

	kc := kratos.New(cfg.KratosPublicURL, cfg.KratosAdminURL)
	authHandler, err := auth.New(kc, fe, cfg.PublicURL, log)
	if err != nil {
		return err
	}

	e := echo.New()
	e.Logger = log
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "0",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		ReferrerPolicy:     "same-origin",
	}))

	health.New(pool).Register(e)
	authHandler.Register(e)

	hub := events.NewHub()
	api := e.Group("/api", identity.New(kc).Require)
	api.GET("/events", hub.Stream)
	workspace.New(pool, hub, cfg.WorkspaceRoot, log).Register(api)

	e.Any("/*", echo.WrapHandler(fe), authHandler.RequireSession)

	log.Info("listening", "addr", cfg.Addr, "public_url", cfg.PublicURL.String(), "workspace_root", cfg.WorkspaceRoot)
	return echo.StartConfig{Address: cfg.Addr, HideBanner: true}.Start(ctx, e)
}
