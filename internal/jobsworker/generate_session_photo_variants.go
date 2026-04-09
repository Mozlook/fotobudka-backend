package jobsworker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
)

func (w *Worker) handleGenerateSessionPhotoVariants(ctx context.Context, job jobs.Job) error {
	var payload sessionphotos.GenerateSessionPhotoVariantsPayload
	err := json.Unmarshal(job.Payload, &payload)
	if err != nil {
		return fmt.Errorf("unmarshal generate_session_photo_variants payload: %w", err)
	}

	if payload.SessionID == uuid.Nil {
		return fmt.Errorf("sessionID cannot be empty")
	}
}
