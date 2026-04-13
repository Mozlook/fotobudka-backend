package sessionphotos

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ClientSessionPhotoListItem struct {
	PhotoID  uuid.UUID `json:"photo_id"`
	ThumbURL string    `json:"thumb_url"`
	Selected bool      `json:"selected"`
	Note     *string   `json:"note,omitempty"`
}

const presignedGetTTL = 30 * time.Minute

func (s *Service) ListReadyClientSessionThumbs(ctx context.Context, sessionID uuid.UUID, offsetCount int32) ([]ClientSessionPhotoListItem, error) {
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id cannot be nil")
	}
	if offsetCount < 0 {
		return nil, fmt.Errorf("offset_count cannot be negative")
	}

	photos, err := s.photosRepo.ListReadyClientSessionPhotos(ctx, sessionID, offsetCount)
	if err != nil {
		return nil, fmt.Errorf("list ready client session photos: %w", err)
	}

	items := make([]ClientSessionPhotoListItem, 0, len(photos))

	for _, photo := range photos {
		if photo.ThumbKey == nil || *photo.ThumbKey == "" {
			return nil, fmt.Errorf("photo %s has empty thumb key", photo.PhotoID)
		}

		thumbURL, err := s.storage.PresignedGetObject(ctx, *photo.ThumbKey, presignedGetTTL)
		if err != nil {
			return nil, fmt.Errorf("presign thumb for photo %s: %w", photo.PhotoID, err)
		}

		items = append(items, ClientSessionPhotoListItem{
			PhotoID:  photo.PhotoID,
			ThumbURL: thumbURL,
			Selected: photo.Selected,
			Note:     photo.Note,
		})
	}

	return items, nil
}
