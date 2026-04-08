package jobsworker

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"

	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

func decodeImage(r io.Reader) (image.Image, string, error) {
	img, formatName, err := image.Decode(r)
	if err != nil {
		return nil, "", err
	}

	return img, formatName, nil
}

func resieToFit(src image.Image, maxW, maxH uint) image.Image {
	return resize.Thumbnail(maxW, maxH, src, resize.Lanczos3)
}

func encodeJPEG(w io.Writer, img image.Image, quality int) error {
	if quality < 1 || quality > 100 {
		return fmt.Errorf("jpeg quality must be between 1 and 100")
	}

	return jpeg.Encode(w, img, &jpeg.Options{
		Quality: quality,
	})
}

func thumbObjectKey(sessionID, photoID uuid.UUID) string {
	return fmt.Sprintf("sessions/%s/thumb/%s.jpg", sessionID, photoID)
}

func proofObjectKey(sessionID, photoID uuid.UUID) string {
	return fmt.Sprintf("sessions/%s/proof/%s.jpg", sessionID, photoID)
}
