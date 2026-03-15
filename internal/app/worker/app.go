package worker

import (
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
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
	log.Info().Str("event_type", "app_started").Msg("worker starting")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Debug().Msg("worker heartbeat")
	}

	return nil
}
