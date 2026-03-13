package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(app string) zerolog.Logger {
	zerolog.TimestampFieldName = "ts"
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	return zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Str("app", app).
		Logger()
}
