package sessionphotos

import (
	"hash/crc32"
	"strings"

	"github.com/google/uuid"
)

func SourceExtFromMIME(mimeType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func watermarkSeedFromPhotoID(id uuid.UUID) int32 {
	sum := crc32.ChecksumIEEE(id[:])
	return int32(sum & 0x7fffffff)
}
