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
	siteweb "github.com/authfly/website/web"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	// Vercel (and most PaaS) inject PORT; prefer it over APP_BIND.
	if port := os.Getenv("PORT"); port != "" {
		cfg.AppBind = "0.0.0.0:" + port
	} else if os.Getenv("APP_BIND") == "" {
		cfg.AppBind = "127.0.0.1:3011"
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
	var store *fixture.Store
	if fixtureDir != "" {
		store, err = fixture.Load(fixtureDir)
	} else {
		// Embedded packs — required on Vercel (binary-only runtime).
		store, err = fixture.LoadLocale(defaultLocale)
		fixtureDir = "embed:" + defaultLocale
	}
	if err != nil {
		slog.Error("fixtures", "err", err)
		os.Exit(1)
	}
	var content ports.Content = store

	// On Vercel the working tree is not shipped with the binary.
	if os.Getenv("VERCEL") != "" || os.Getenv("VERCEL_ENV") != "" {
		staticDir, matErr := siteweb.MaterializeStatic()
		if matErr != nil {
			slog.Error("static", "err", matErr)
			os.Exit(1)
		}
		cfg.StaticDir = staticDir
		defer os.RemoveAll(staticDir)
	} else if os.Getenv("APP_STATIC_DIR") == "" {
		cfg.StaticDir = "web/static"
	}
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
