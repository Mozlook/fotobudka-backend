package finalphotos

import "github.com/google/uuid"

type FinalPhotoForDelivery struct {
	ID               uuid.UUID
	PhotoID          uuid.UUID
	FinalKey         string
	OriginalFilename string
}
