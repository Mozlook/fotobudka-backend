package selections

import (
	"context"
	"fmt"
	"strings"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

type SelectionItem struct {
	PhotoID  uuid.UUID
	Selected bool
	Note     *string
}

func (s *Service) UpdateSelections(ctx context.Context, sessionID uuid.UUID, items []SelectionItem) error {
	if sessionID == uuid.Nil {
		return ErrInvalidSessionID
	}
	if len(items) == 0 {
		return ErrEmptySelectionItems
	}

	session, err := s.sessionsRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session by id: %w", err)
	}
	if session.Status != "selecting" {
		return ErrSelectionLocked
	}

	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item.PhotoID == uuid.Nil {
			return ErrInvalidPhotoID
		}
		if _, exists := seen[item.PhotoID]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicatePhotoInBatch, item.PhotoID)
		}
		seen[item.PhotoID] = struct{}{}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	for _, item := range items {
		isReady, err := qtx.IsReadySessionPhoto(ctx, dbgen.IsReadySessionPhotoParams{
			PhotoID:   item.PhotoID,
			SessionID: sessionID,
		})
		if err != nil {
			return fmt.Errorf("check ready photo %s: %w", item.PhotoID, err)
		}
		if !isReady {
			return fmt.Errorf("%w: %s", ErrPhotoNotSelectable, item.PhotoID)
		}

		if item.Selected {
			if err := qtx.UpsertSelection(ctx, dbgen.UpsertSelectionParams{
				SessionID: sessionID,
				PhotoID:   item.PhotoID,
				Note:      normalizeNote(item.Note),
			}); err != nil {
				return fmt.Errorf("upsert selection for photo %s: %w", item.PhotoID, err)
			}
		} else {
			if _, err := qtx.DeleteSelection(ctx, dbgen.DeleteSelectionParams{
				SessionID: sessionID,
				PhotoID:   item.PhotoID,
			}); err != nil {
				return fmt.Errorf("delete selection for photo %s: %w", item.PhotoID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func normalizeNote(note *string) *string {
	if note == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*note)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
