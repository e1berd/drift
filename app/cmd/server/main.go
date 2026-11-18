package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	schema "github.com/e1berd/drift/app/db"
	"github.com/e1berd/drift/app/internal/agents"
	"github.com/e1berd/drift/app/internal/askuser"
	"github.com/e1berd/drift/app/internal/auth"
	"github.com/e1berd/drift/app/internal/config"
	"github.com/e1berd/drift/app/internal/database"
	"github.com/e1berd/drift/app/internal/delivery"
	"github.com/e1berd/drift/app/internal/events"
	"github.com/e1berd/drift/app/internal/frontend"
	"github.com/e1berd/drift/app/internal/gitrepo"
	"github.com/e1berd/drift/app/internal/health"
	"github.com/e1berd/drift/app/internal/i18n"
	"github.com/e1berd/drift/app/internal/identity"
	"github.com/e1berd/drift/app/internal/kratos"
	"github.com/e1berd/drift/app/internal/migrate"
	"github.com/e1berd/drift/app/internal/runner"
	"github.com/e1berd/drift/app/internal/subscriptions"
	"github.com/e1berd/drift/app/internal/terminal"
	"github.com/e1berd/drift/app/internal/workspace"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == askuser.Command {
		if err := askuser.Serve(os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(delivery.Version())
		return
	}
	relaunchPath, err := run()
	if relaunchPath != "" {
		err = delivery.Relaunch(relaunchPath)
	}
	if err != nil {
		slog.Error("drift stopped", "err", err)
		os.Exit(1)
	}
}

func run() (string, error) {
	signals, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	ctx, shutdown := context.WithCancel(signals)
	defer shutdown()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cfg.WorkspaceRoot, 0o755); err != nil {
		return "", fmt.Errorf("create workspace root: %w", err)
	}
	if err := os.MkdirAll(cfg.AgentsRoot, 0o700); err != nil {
		return "", fmt.Errorf("create agents folder: %w", err)
	}

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return "", err
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, schema.Migrations, "migrations", log); err != nil {
		return "", err
	}

	fe, err := frontend.New(cfg.ViteDevURL)
	if err != nil {
		return "", err
	}

	translations, err := i18n.Load()
	if err != nil {
		return "", err
	}

	kc := kratos.New(cfg.KratosPublicURL, cfg.KratosAdminURL)
	authHandler, err := auth.New(kc, fe, translations, cfg.PublicURL, log)
	if err != nil {
		return "", err
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
	cli := agents.CLI{Root: cfg.AgentsRoot, Claude: cfg.ClaudeBin, Codex: cfg.CodexBin, Antigravity: cfg.AntigravityBin}
	folders := workspace.Folders{Workspace: cfg.WorkspaceRoot, Chats: cfg.ChatsRoot}
	agentRunner, err := runner.New(ctx, pool, hub, cli, folders, log)
	if err != nil {
		return "", err
	}
	workspaceHandler := workspace.New(pool, hub, folders, log).WithRunner(agentRunner)
	terminals := terminal.New(terminal.DetectShell(cfg.Shell), workspaceHandler.SessionFolder, log)
	workspaceHandler.WithTerminals(terminals).Register(api)
	terminals.Register(api)
	gitrepo.New(workspaceHandler.SessionFolder).Register(api)
	subs, err := subscriptions.New(ctx, pool, hub, cli, log)
	if err != nil {
		return "", err
	}
	subs.Register(api)
	updates, err := delivery.New(delivery.Options{Repository: cfg.DeliveryRepo, APIURL: cfg.DeliveryAPIURL, Shutdown: shutdown}, hub, log)
	if err != nil {
		return "", err
	}
	updates.Register(api)
	go updates.Watch(ctx)

	e.Any("/*", echo.WrapHandler(fe), authHandler.RequireSession)

	log.Info("listening", "addr", cfg.Addr, "version", delivery.Version(), "public_url", cfg.PublicURL.String(), "workspace_root", cfg.WorkspaceRoot)
	serveErr := echo.StartConfig{Address: cfg.Addr, HideBanner: true}.Start(ctx, e)
	return updates.RelaunchPath(), serveErr
}
