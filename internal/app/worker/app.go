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

	log := applog.New(cfg.AppName)
	log.Info().Msg("worker starting")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Debug().Msg("worker heartbeat")
	}

	return nil
}
