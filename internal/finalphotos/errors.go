package finalphotos

import "errors"

var (
	ErrInvalidSessionID      = errors.New("invalid session id")
	ErrEmptyFiles            = errors.New("files cannot be empty")
	ErrInvalidPhotoID        = errors.New("invalid photo id")
	ErrDuplicatePhotoInBatch = errors.New("duplicate photo id in request")
	ErrInvalidFilename       = errors.New("invalid filename")
	ErrInvalidMimeType       = errors.New("invalid mime type")
	ErrInvalidFileSize       = errors.New("invalid file size")
	ErrSessionNotFound       = errors.New("session not found")
	ErrFinalUploadLocked     = errors.New("final upload is not allowed in current session state")
	ErrPhotoNotSelected      = errors.New("photo is not selected for this session")
	ErrFinalAlreadyExists    = errors.New("final photo already exists")
)
