package selections

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SelectionItem struct {
	PhotoID  uuid.UUID
	Selected bool
	Note     *string
}

type SubmitSelectionResult struct {
	Status        string
	SelectedCount int64
	AmountCents   int32
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

func (s *Service) SubmitSelection(ctx context.Context, sessionID uuid.UUID) (SubmitSelectionResult, error) {
	if sessionID == uuid.Nil {
		return SubmitSelectionResult{}, ErrInvalidSessionID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SubmitSelectionResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	session, err := qtx.GetSessionSubmitDataForUpdate(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SubmitSelectionResult{}, ErrSessionNotFound
		}
		return SubmitSelectionResult{}, fmt.Errorf("get session submit data for update: %w", err)
	}

	if session.Status != "selecting" {
		return SubmitSelectionResult{}, ErrSubmitLocked
	}

	selectedCount, err := qtx.CountSelectionsBySessionID(ctx, sessionID)
	if err != nil {
		return SubmitSelectionResult{}, fmt.Errorf("count selections by session_id: %w", err)
	}

	if selectedCount < int64(session.MinSelectCount) {
		return SubmitSelectionResult{}, ErrMinimumSelectionNotMet
	}

	extraCount := max(0, selectedCount-int64(session.IncludedCount))
	amount := int64(session.BasePriceCents) + extraCount*int64(session.ExtraPriceCents)

	var method string
	switch session.PaymentMode {
	case "manual":
		method = "manual"
	case "platform_future":
		method = "online_future"
	default:
		return SubmitSelectionResult{}, fmt.Errorf("unsupported payment mode: %s", session.PaymentMode)
	}

	err = qtx.InsertSessionPayment(ctx, dbgen.InsertSessionPaymentParams{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Method:      method,
		Status:      "unpaid",
		AmountCents: int32(amount),
	})
	if err != nil {
		return SubmitSelectionResult{}, fmt.Errorf("insert session payment: %w", err)
	}

	rows, err := qtx.MarkSessionWaitingForPayment(ctx, sessionID)
	if err != nil {
		return SubmitSelectionResult{}, fmt.Errorf("mark session waiting for payment: %w", err)
	}
	if rows != 1 {
		return SubmitSelectionResult{}, ErrSubmitLocked
	}

	if err := tx.Commit(ctx); err != nil {
		return SubmitSelectionResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return SubmitSelectionResult{
		Status:        "waiting_for_payment",
		SelectedCount: selectedCount,
		AmountCents:   int32(amount),
	}, nil
}
