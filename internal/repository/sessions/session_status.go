package sessions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) TryMarkSessionSelecting(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	count, err := r.q.MarkSessionSelecting(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("mark session selecting: %w", err)
	}
	if count > 1 {
		return false, fmt.Errorf("mark session selecting: unexpected affected rows: %d", count)
	}
	return count == 1, nil
}

func (r *Repository) TryMarkSessionFailed(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	count, err := r.q.MarkSessionFailed(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("mark session failed: %w", err)
	}
	if count > 1 {
		return false, fmt.Errorf("mark session failed: unexpected affected rows: %d", count)
	}
	return count == 1, nil
}
