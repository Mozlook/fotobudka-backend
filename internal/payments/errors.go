package payments

import "errors"

var (
	ErrInvalidSessionID               = errors.New("invalid session id")
	ErrSessionNotFound                = errors.New("session not found")
	ErrMarkPaidLocked                 = errors.New("mark paid is not allowed in current session state")
	ErrUnpaidPaymentNotFound          = errors.New("unpaid payment not found")
	ErrPaymentMarkPaidConflict        = errors.New("payment could not be marked as paid")
	ErrSessionEditingTransitionFailed = errors.New("session could not transition to editing")
)
