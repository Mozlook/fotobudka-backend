package me

import "github.com/Mozlook/fotobudka-backend/internal/repository/profiles"

// Handler serves authenticcated /api/me endpoints/
type Handler struct{ profiles *profiles.Repository }

// NewHandler creates a me handler backed by the photogrtapher profiles repository
func NewHandler(profiles *profiles.Repository) *Handler {
	return &Handler{profiles: profiles}
}
