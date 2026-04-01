package sessionphotos

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/google/uuid"
)

const presignedPutTTL = 30 * time.Minute

type PhotoPutURL struct {
	PhotoID   string
	PutURL    *url.URL
	ObjectKey string
	Error     bool
}

type FileInput struct {
	Filename  string
	MimeType  string
	SizeBytes int64
}

type Service struct {
	storage *storage.Client
}

func New(storageClient *storage.Client) *Service {
	return &Service{
		storage: storageClient,
	}
}

func (s *Service) PresignedUploadURLs(ctx context.Context, sessionID string, files []FileInput) ([]PhotoPutURL, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session_id cannot be empty")
	}

	output := make([]PhotoPutURL, len(files))

	for i, file := range files {
		if strings.TrimSpace(file.Filename) == "" || file.SizeBytes <= 0 {
			output[i] = PhotoPutURL{Error: true}
			continue
		}

		ext, ok := sourceExtFromMIME(file.MimeType)
		if !ok {
			output[i] = PhotoPutURL{Error: true}
			continue
		}

		photoID := uuid.NewString()
		objectKey := fmt.Sprintf("sessions/%s/source/%s%s", sessionID, photoID, ext)

		putURL, err := s.storage.PresignedPutObject(ctx, objectKey, presignedPutTTL)
		if err != nil {
			return nil, fmt.Errorf("presign put object for %q: %w", file.Filename, err)
		}

		output[i] = PhotoPutURL{
			PhotoID:   photoID,
			PutURL:    putURL,
			ObjectKey: objectKey,
			Error:     false,
		}
	}

	return output, nil
}

func sourceExtFromMIME(mimeType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
