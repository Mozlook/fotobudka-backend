package sessionaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Result struct {
	Code      string
	Token     string
	CreatedAt time.Time
}

type Service struct {
	pool   *pgxpool.Pool
	repo   *sessions.Repository
	secret []byte
}

func New(pool *pgxpool.Pool, repo *sessions.Repository, secret []byte) *Service {
	return &Service{pool: pool, repo: repo, secret: secret}
}

func (s *Service) RegenerateSessionAccess(ctx context.Context, sessionID uuid.UUID) (Result, error) {
	if len(s.secret) == 0 {
		return Result{}, fmt.Errorf("secret cannot be empty")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.repo.WithTx(tx)
	_, err = txRepo.RevokeSessionAccess(ctx, sessionID)
	if err != nil {
		return Result{}, err
	}
	sessionAccessInput, plainTextResponse, err := GenerateAccess(sessionID, s.secret)
	if err != nil {
		return Result{}, err
	}
	sessionAccess, err := txRepo.InsertSessionAccess(ctx, sessionAccessInput)
	if err != nil {
		return Result{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Code:      plainTextResponse.PlainTextCode,
		Token:     plainTextResponse.PlainTextToken,
		CreatedAt: sessionAccess.CreatedAt,
	}, nil
}
