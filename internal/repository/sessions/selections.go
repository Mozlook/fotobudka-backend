package sessions

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PhotographerSelectionOverview struct {
	SessionID          uuid.UUID
	SessionStatus      string
	SelectedCount      int32
	PaymentStatus      string
	PaymentAmountCents int32
	PaymentPaidAt      *time.Time
	Photos             []PhotographerSelectedPhoto
}

type PhotographerSelectedPhoto struct {
	PhotoID          uuid.UUID
	OriginalFilename string
	ThumbKey         string
	Note             string
	SelectedAt       time.Time
	FinalID          *uuid.UUID
	FinalUploaded    bool
}

func (r *Repository) GetPhotographerSelectionOverview(
	ctx context.Context,
	sessionID uuid.UUID,
) (PhotographerSelectionOverview, error) {
	overviewRow, err := r.q.GetPhotographerSelectionOverview(ctx, sessionID)
	if err != nil {
		return PhotographerSelectionOverview{}, err
	}

	photoRows, err := r.q.ListPhotographerSelectedPhotos(ctx, sessionID)
	if err != nil {
		return PhotographerSelectionOverview{}, err
	}

	photos := make([]PhotographerSelectedPhoto, 0, len(photoRows))

	for _, row := range photoRows {
		photos = append(photos, PhotographerSelectedPhoto{
			PhotoID:          row.PhotoID,
			OriginalFilename: row.OriginalFilename,
			ThumbKey:         *row.ThumbKey,
			Note:             row.Note,
			SelectedAt:       row.SelectedAt,
			FinalID:          &row.FinalID,
			FinalUploaded:    row.FinalUploaded,
		})
	}

	return PhotographerSelectionOverview{
		SessionID:          overviewRow.SessionID,
		SessionStatus:      overviewRow.SessionStatus,
		SelectedCount:      overviewRow.SelectedCount,
		PaymentStatus:      overviewRow.PaymentStatus,
		PaymentAmountCents: overviewRow.PaymentAmountCents,
		PaymentPaidAt:      overviewRow.PaymentPaidAt,
		Photos:             photos,
	}, nil
}
