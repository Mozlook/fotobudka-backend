package payments

import (
	"context"
	"errors"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MarkPaidResult struct {
	SessionStatus string
	PaymentStatus string
	AmountCents   int32
}

func (s *Service) MarkPaid(ctx context.Context, sessionID uuid.UUID) (MarkPaidResult, error) {
	if sessionID == uuid.Nil {
		return MarkPaidResult{}, ErrInvalidSessionID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MarkPaidResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	session, err := qtx.GetSessionStatusForUpdate(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MarkPaidResult{}, ErrSessionNotFound
		}
		return MarkPaidResult{}, fmt.Errorf("get session status for update: %w", err)
	}

	if session.Status != "waiting_for_payment" {
		return MarkPaidResult{}, ErrMarkPaidLocked
	}

	payment, err := qtx.GetUnpaidPaymentBySessionIDForUpdate(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MarkPaidResult{}, ErrUnpaidPaymentNotFound
		}
		return MarkPaidResult{}, fmt.Errorf("get unpaid payment by session id for update: %w", err)
	}

	rows, err := qtx.MarkPaymentPaid(ctx, payment.ID)
	if err != nil {
		return MarkPaidResult{}, fmt.Errorf("mark payment paid: %w", err)
	}
	if rows != 1 {
		return MarkPaidResult{}, ErrPaymentMarkPaidConflict
	}

	row, err := qtx.MarkSessionEditing(ctx, sessionID)
	if err != nil {
		return MarkPaidResult{}, fmt.Errorf("mark session editing: %w", err)
	}
	if row != 1 {
		return MarkPaidResult{}, ErrSessionEditingTransitionFailed
	}

	if err := tx.Commit(ctx); err != nil {
		return MarkPaidResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return MarkPaidResult{
		SessionStatus: "editing",
		PaymentStatus: "paid",
		AmountCents:   payment.AmountCents,
	}, nil
}
