package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/config"
	auth "github.com/Mozlook/fotobudka-backend/internal/http/handler/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/me"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/sessions"
	hrouter "github.com/Mozlook/fotobudka-backend/internal/http/router"
	"github.com/Mozlook/fotobudka-backend/internal/oauth"
	"github.com/Mozlook/fotobudka-backend/internal/platform/db"
	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	applog "github.com/Mozlook/fotobudka-backend/internal/platform/logger"
	"github.com/Mozlook/fotobudka-backend/internal/repository/profiles"
	sessionsrep "github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/repository/users"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return err
	}

	log, closer, err := applog.New(cfg)
	if err != nil {
		return err
	}
	defer closer.Close()

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStartup()

	pool, err := db.NewPool(startupCtx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := dbgen.New(pool)
	usersRepo := users.New(queries)
	profilesRepo := profiles.New(queries)
	sessionsRepo := sessionsrep.New(queries)

	manager := appauth.New(cfg)
	provider := oauth.New(cfg)
	authHandler := auth.NewAuthHandler(cfg, provider, usersRepo, manager)
	meHandler := me.NewHandler(profilesRepo)
	sessionsHandler := sessions.NewHandler(sessionsRepo)

	srv := &http.Server{
		Addr:              cfg.HTTP.APIAddr,
		Handler:           hrouter.New(log, authHandler, meHandler, sessionsHandler, manager),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Info().Str("addr", cfg.HTTP.APIAddr).Str("event_type", "app_started").Msg("api starting")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
