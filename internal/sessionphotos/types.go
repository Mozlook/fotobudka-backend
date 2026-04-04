package sessionphotos

import (
	"net/url"

	"github.com/google/uuid"
)

type PhotoPutURL struct {
	PhotoID   uuid.UUID
	PutURL    *url.URL
	ObjectKey string
	Error     bool
}

type FileInput struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

const JobTypeGenerateSessionPhotoVariants = "generate_session_photo_variants"

type GenerateSessionPhotoVariantsPayload struct {
	SessionID     uuid.UUID `json:"session_id"`
	PhotoID       uuid.UUID `json:"photo_id"`
	SourceKey     string    `json:"source_key"`
	WatermarkSeed int32     `json:"watermark_seed"`
}
