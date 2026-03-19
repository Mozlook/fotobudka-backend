package profiles

import (
	"time"

	"github.com/google/uuid"
)

type SocialLinks struct {
	Instagram string `json:"instagram"`
	Tiktok    string `json:"tiktok"`
	Website   string `json:"website"`
	Facebook  string `json:"facebook"`
	Behance   string `json:"behance"`
}

type UpsertInput struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	Bio         string
	SocialLinks SocialLinks
}

type Profile struct {
	UserID      uuid.UUID   `json:"user_id"`
	Username    string      `json:"username"`
	DisplayName string      `json:"display_name"`
	Bio         string      `json:"bio"`
	SocialLinks SocialLinks `json:"social_links"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
