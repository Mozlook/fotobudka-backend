package galleries

import "github.com/google/uuid"

type PresignGalleryPhotoFile struct {
	Filename  string
	MimeType  string
	SizeBytes int64
}

type PresignGalleryPhotosInput struct {
	GalleryID      uuid.UUID
	PhotographerID uuid.UUID
	Files          []PresignGalleryPhotoFile
}

type GalleryPhotoUploadTarget struct {
	PhotoID   uuid.UUID
	PutURL    string
	ObjectKey string
}

type CompleteGalleryPhotoInput struct {
	GalleryID      uuid.UUID
	PhotoID        uuid.UUID
	PhotographerID uuid.UUID
}
