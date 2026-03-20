package me

import "github.com/Mozlook/fotobudka-backend/internal/repository/profiles"

type Handler struct{ profiles *profiles.Repository }

func NewHandler(profiles *profiles.Repository) *Handler {
	return &Handler{profiles: profiles}
}
