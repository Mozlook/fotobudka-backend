package sessionphotos

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/google/uuid"
)

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
	minio *storage.Client
}

func New(minio *storage.Client) *Service {
	return &Service{
		minio: minio,
	}
}

func (s *Service) PresignedUploadURLs(ctx context.Context, sessionID string, files []FileInput) ([]PhotoPutURL, error) {
	outputList := make([]PhotoPutURL, 0, len(files))
	expires := time.Duration(time.Minute * 60)

	for _, file := range files {
		photoID := uuid.NewString()

		ext := strings.ToLower(path.Ext(file.Filename))
		if ext == "" {
			outputList = append(outputList, PhotoPutURL{PhotoID: photoID, PutURL: nil, ObjectKey: "", Error: true})
			continue
		}

		objectKey := fmt.Sprintf("sessions/%s/source/%s%s", sessionID, photoID, ext)

		putURL, err := s.minio.PresignedPutObject(ctx, objectKey, expires)
		if err != nil {
			return nil, err
		}

		outputList = append(outputList, PhotoPutURL{PhotoID: photoID, PutURL: putURL, ObjectKey: objectKey, Error: false})

	}

	return outputList, nil
}
