package galleries

import (
	"fmt"
	"regexp"
	"strings"
)

var gallerySlugRegex = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,60}[a-z0-9])?$`)

func normalizeGallerySlug(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func validateGalleryInput(title string, slug string) error {
	title = strings.TrimSpace(title)
	slug = normalizeGallerySlug(slug)

	if title == "" {
		return fmt.Errorf("title is required")
	}

	if len([]rune(title)) > 120 {
		return fmt.Errorf("title must be at most 120 characters")
	}

	if !gallerySlugRegex.MatchString(slug) {
		return fmt.Errorf("invalid slug")
	}

	if isReservedGallerySlug(slug) {
		return fmt.Errorf("reserved slug")
	}

	return nil
}

func isReservedGallerySlug(slug string) bool {
	switch slug {
	case "api",
		"auth",
		"public",
		"me",
		"login",
		"logout",
		"s",
		"client",
		"app":
		return true
	default:
		return false
	}
}
