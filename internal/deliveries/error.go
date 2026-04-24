package deliveries

import "errors"

var (
	ErrInvalidSessionID            = errors.New("invalid session id")
	ErrSessionNotFound             = errors.New("session not found")
	ErrGenerateZIPLocked           = errors.New("generate zip is not allowed in current session state")
	ErrNoSelections                = errors.New("no selected photos to deliver")
	ErrMissingFinalPhotos          = errors.New("missing final photos for selected images")
	ErrLatestReadyDeliveryNotFound = errors.New("latest ready delivery not found")
)
