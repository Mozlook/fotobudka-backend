package galleries

import (
	"time"

	"github.com/google/uuid"
)

type PublicPhotographerProfile struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	Bio         string
	SocialLinks []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Gallery struct {
	ID             uuid.UUID
	PhotographerID uuid.UUID
	Title          string
	Slug           string
	IsPublic       bool
	CreatedAt      time.Time
	PhotoCount     int32
	CoverImageKey  string
}

type GalleryPhoto struct {
	ID        uuid.UUID
	GalleryID uuid.UUID
	ImageKey  string
	Width     int32
	Height    int32
	SortOrder int32
	CreatedAt time.Time
}

type CreateGalleryInput struct {
	ID             uuid.UUID
	PhotographerID uuid.UUID
	Title          string
	Slug           string
	IsPublic       bool
}

type UpdateGalleryInput struct {
	ID             uuid.UUID
	PhotographerID uuid.UUID
	Title          string
	Slug           string
	IsPublic       bool
}

type CreateGalleryPhotoInput struct {
	ID        uuid.UUID
	GalleryID uuid.UUID
	ImageKey  string
	Width     int32
	Height    int32
}

type FeaturedPublicGallery struct {
	ID                      uuid.UUID
	PhotographerID          uuid.UUID
	Title                   string
	Slug                    string
	IsPublic                bool
	CreatedAt               time.Time
	PhotoCount              int32
	CoverImageKey           string
	PhotographerUsername    string
	PhotographerDisplayName string
}
