package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/rs/zerolog"
)

func New(cfg config.Config) (zerolog.Logger, io.Closer, error) {
	logDir := filepath.Join(cfg.SIEM.LogDir, cfg.App.Name)
	err := os.MkdirAll(logDir, 0o755)
	if err != nil {
		return zerolog.Logger{}, nil, err
	}

	filePath := filepath.Join(logDir, "app.jsonl")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return zerolog.Logger{}, nil, err
	}

	host, err := os.Hostname()
	if err != nil {
		file.Close()
		return zerolog.Logger{}, nil, err
	}

	zerolog.TimestampFieldName = "ts"
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	var writer io.Writer = file
	if cfg.App.Env == "dev" {
		writer = zerolog.MultiLevelWriter(file, os.Stdout)
	}

	logger := zerolog.New(writer).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Str("app", cfg.App.Name).
		Str("host", host).
		Logger()

	return logger, file, nil
}
