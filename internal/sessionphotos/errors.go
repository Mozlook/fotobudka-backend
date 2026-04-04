package sessionphotos

import "errors"

var (
	ErrInvalidPhotoStatus     = errors.New("invalid photo status")
	ErrSessionPhotoNotFound   = errors.New("session photo not found")
	ErrUploadedObjectNotFound = errors.New("photo object not found")
)
