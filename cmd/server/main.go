package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/framework/pkg/web/locale"
	"github.com/fastygo/framework/pkg/web/security"
	"github.com/authfly/website/internal/adapters/fixture"
	"github.com/authfly/website/internal/app/site"
	"github.com/authfly/website/internal/ports"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	if os.Getenv("APP_BIND") == "" {
		cfg.AppBind = "127.0.0.1:3011"
	}
	if os.Getenv("APP_STATIC_DIR") == "" {
		cfg.StaticDir = "web/static"
	}
	if cfg.HealthLivePath == "" {
		cfg.HealthLivePath = "/healthz"
	}
	if cfg.HealthReadyPath == "" {
		cfg.HealthReadyPath = "/readyz"
	}

	defaultLocale := os.Getenv("APP_DEFAULT_LOCALE")
	if defaultLocale == "" {
		defaultLocale = "ru"
	}
	if os.Getenv("APP_DEFAULT_LOCALE") == "" {
		cfg.DefaultLocale = defaultLocale
	}
	if len(cfg.AvailableLocales) == 0 {
		cfg.AvailableLocales = []string{defaultLocale}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}))
	slog.SetDefault(logger)

	fixtureDir := os.Getenv("FIXTURE_DIR")
	if fixtureDir == "" {
		if defaultLocale == "en" {
			fixtureDir = "internal/adapters/fixture/data-en"
		} else {
			fixtureDir = "internal/adapters/fixture/data"
		}
	}
	store, err := fixture.Load(fixtureDir)
	if err != nil {
		slog.Error("fixtures", "err", err)
		os.Exit(1)
	}
	var content ports.Content = store

	sf := site.New(site.Deps{Content: content})

	application := app.New(cfg).
		WithSecurity(security.LoadConfig()).
		WithLocales(app.LocalesConfig{
			Default:   cfg.DefaultLocale,
			Available: cfg.AvailableLocales,
			Cookie: locale.CookieOptions{
				Enabled: false,
			},
		}).
		WithStaticPrefix("/assets").
		WithHealthEndpoints("/healthz", "/readyz").
		WithFeature(sf).
		Build()

	handler := sf.WithNotFound(application.Handler())
	server := &http.Server{
		Addr:              cfg.AppBind,
		Handler:           handler,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		MaxHeaderBytes:    cfg.HTTPMaxHeaderBytes,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("authfly website starting", "bind", cfg.AppBind, "locale", defaultLocale, "fixtures", fixtureDir)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("run", "err", err)
			os.Exit(1)
		}
	}
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
