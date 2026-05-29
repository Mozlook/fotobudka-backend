package galleries

import (
	"time"

	"github.com/google/uuid"
)

type upsertGalleryRequest struct {
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	IsPublic bool   `json:"is_public"`
}

type galleryResponse struct {
	ID         uuid.UUID `json:"ID"`
	Title      string    `json:"Title"`
	Slug       string    `json:"Slug"`
	IsPublic   bool      `json:"IsPublic"`
	PhotoCount int32     `json:"PhotoCount"`
	CoverURL   string    `json:"CoverURL,omitempty"`
	CreatedAt  time.Time `json:"CreatedAt"`
}

type galleryPhotoResponse struct {
	ID        uuid.UUID `json:"ID"`
	ImageURL  string    `json:"ImageURL,omitempty"`
	Width     int32     `json:"Width"`
	Height    int32     `json:"Height"`
	SortOrder int32     `json:"SortOrder"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type presignPhotosRequest struct {
	Files []presignPhotoFileRequest `json:"files"`
}

type presignPhotoFileRequest struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type presignPhotosResponse struct {
	Uploads []presignPhotoUploadResponse `json:"Uploads"`
}

type presignPhotoUploadResponse struct {
	PhotoID   uuid.UUID `json:"PhotoID"`
	PutURL    string    `json:"PutURL"`
	ObjectKey string    `json:"ObjectKey"`
}
