package worker

import (
	"context"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/Mozlook/fotobudka-backend/internal/platform/db"
	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	applog "github.com/Mozlook/fotobudka-backend/internal/platform/logger"
	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStartup()

	pool, err := db.NewPool(startupCtx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	query := dbgen.New(pool)

	jobsRepo := jobs.New(query)

	log, closer, err := applog.New(cfg)
	if err != nil {
		return err
	}
	defer closer.Close()
	log.Info().Str("event_type", "app_started").Msg("worker starting")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Debug().Msg("worker heartbeat")
	}

	return nil
}
