package middleware

import (
	"context"

	"github.com/google/uuid"
)

type (
	clientSessionIDKey       struct{}
	clientSessionAccessIDKey struct{}
)

func clientContextWithSessionID(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, clientSessionIDKey{}, sessionID)
}

func clientContextWithSessionAccessID(ctx context.Context, sessionAccessID uuid.UUID) context.Context {
	return context.WithValue(ctx, clientSessionAccessIDKey{}, sessionAccessID)
}

func ClientSessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(clientSessionIDKey{}).(uuid.UUID)
	return sessionID, ok
}

func ClientSessionAccessIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	sessionAccessID, ok := ctx.Value(clientSessionAccessIDKey{}).(uuid.UUID)
	return sessionAccessID, ok
}
