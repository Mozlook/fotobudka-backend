package sessionphotos

import (
	"context"
	"errors"
	"fmt"
	"time"

	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
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

func (s *Service) GetReadyClientPhotoProofURL(ctx context.Context, sessionID, photoID uuid.UUID) (string, error) {
	if sessionID == uuid.Nil {
		return "", fmt.Errorf("session_id cannot be nil")
	}
	if photoID == uuid.Nil {
		return "", fmt.Errorf("photo_id cannot be nil")
	}

	proofKey, err := s.photosRepo.GetReadyClientPhotoProofKey(ctx, sessionID, photoID)
	if err != nil {
		if errors.Is(err, sessionphotosrepo.ErrSessionPhotoNotFound) {
			return "", ErrSessionPhotoNotFound
		}
		return "", fmt.Errorf("get ready client photo proof key: %w", err)
	}

	proofURL, err := s.storage.PresignedGetObject(ctx, proofKey, presignedGetTTL)
	if err != nil {
		return "", fmt.Errorf("presign proof object: %w", err)
	}

	return proofURL, nil
}
