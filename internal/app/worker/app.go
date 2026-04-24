package worker

import (
	"context"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/Mozlook/fotobudka-backend/internal/jobsworker"
	"github.com/Mozlook/fotobudka-backend/internal/platform/db"
	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	applog "github.com/Mozlook/fotobudka-backend/internal/platform/logger"
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/Mozlook/fotobudka-backend/internal/repository/deliveries"
	finalphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/finalphotos"
	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStartup()

	pool, err := db.NewPool(startupCtx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	q := dbgen.New(pool)

	log, closer, err := applog.New(cfg)
	if err != nil {
		return err
	}
	defer closer.Close()

	storageClient, err := storage.New(cfg.S3)
	if err != nil {
		return err
	}

	jobsRepo := jobs.New(q)
	sessionPhotosRepo := sessionphotosrepo.New(q, pool)
	sessionsRepo := sessions.New(q)
	deliveriesRepo := deliveries.New(q, pool)
	finalphotosRepo := finalphotosrepo.New(q)
	worker := jobsworker.New(
		jobsRepo,
		sessionPhotosRepo,
		sessionsRepo,
		deliveriesRepo,
		finalphotosRepo,
		storageClient,
		10, // limit
	)

	log.Info().Str("event_type", "app_started").Msg("worker starting")

	runCtx, cancelRun := context.WithTimeout(context.Background(), 30*time.Second)
	if err := worker.RunOnce(runCtx); err != nil {
		log.Error().Err(err).Msg("worker run once failed")
	}
	cancelRun()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		runCtx, cancelRun := context.WithTimeout(context.Background(), 30*time.Second)

		if err := worker.RunOnce(runCtx); err != nil {
			log.Error().Err(err).Msg("worker run once failed")
		}

		cancelRun()
	}

	return nil
}
