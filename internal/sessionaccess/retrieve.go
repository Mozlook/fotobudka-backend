package sessionaccess

import (
	"context"
	"errors"

	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/jackc/pgx/v5"
)

var ErrSessionAccessNotFound = errors.New("access not found")

func (s *Service) GetClientSessionByToken(ctx context.Context, token string) (sessions.ClientSession, error) {
	tokenHMAC := hmacHex(s.secret, token)

	clientSession, err := s.repo.GetClientSessionByTokenHMAC(ctx, tokenHMAC)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sessions.ClientSession{}, ErrSessionAccessNotFound
		}

		return sessions.ClientSession{}, err
	}
	return clientSession, nil
}

func (s *Service) GetClientSessionByCode(ctx context.Context, code string) (sessions.ClientSession, error) {
	codeHMAC := hmacHex(s.secret, code)

	clientSession, err := s.repo.GetClientSessionByCodeHMAC(ctx, codeHMAC)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sessions.ClientSession{}, ErrSessionAccessNotFound
		}

		return sessions.ClientSession{}, err
	}
	return clientSession, nil
}
