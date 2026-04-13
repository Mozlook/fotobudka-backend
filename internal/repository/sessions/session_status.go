package sessions

import (
	"context"

	"github.com/google/uuid"
)

func (r *Repository) TryMarkSessionProcessing(ctx context.Context, SessionID uuid.UUID) (bool, error) {
}
