package middleware

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

// RequestIDFromContext returns the request correlation identifier stored in ctx.
//
// It returns an empty string when no request identifier is present.
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

// UserIDFromContext returns the authenticated user ID stored in ctx.
//
// It returns false when the request context does not contain an authenticated user.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
