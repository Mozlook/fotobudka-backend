package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	hrouter "github.com/Mozlook/fotobudka-backend/internal/http/router"
	applog "github.com/Mozlook/fotobudka-backend/internal/platform/logger"
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

	srv := &http.Server{
		Addr:              cfg.HTTP.APIAddr,
		Handler:           hrouter.New(log),
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
