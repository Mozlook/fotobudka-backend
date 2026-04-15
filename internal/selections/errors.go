package selections

import "errors"

var (
	ErrInvalidSessionID       = errors.New("invalid session id")
	ErrEmptySelectionItems    = errors.New("selection items cannot be empty")
	ErrInvalidPhotoID         = errors.New("invalid photo id")
	ErrDuplicatePhotoInBatch  = errors.New("duplicate photo id in request")
	ErrSelectionLocked        = errors.New("selection is locked")
	ErrPhotoNotSelectable     = errors.New("photo is not selectable")
	ErrSubmitLocked           = errors.New("submit is not allowed in current session status")
	ErrMinimumSelectionNotMet = errors.New("minimum selection count not met")
	ErrSessionNotFound        = errors.New("session not found")
)
