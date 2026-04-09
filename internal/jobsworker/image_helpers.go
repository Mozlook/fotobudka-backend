package jobsworker

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"math/rand"

	_ "image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
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

func resizeToFit(src image.Image, maxW, maxH uint) image.Image {
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

func applyProofWatermark(src image.Image, seed int32) (image.Image, error) {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)

	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    72,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}

	text := "FotoBudka"
	textW := font.MeasureString(face, text).Ceil()

	m := face.Metrics()
	ascent := m.Ascent.Ceil()
	descent := m.Descent.Ceil()
	textH := ascent + descent
	b := dst.Bounds()

	centerX := b.Min.X + (b.Dx()-textW)/2
	centerY := b.Min.Y + (b.Dy()-textH)/2 + ascent

	r := rand.New(rand.NewSource(int64(seed)))

	padding := 40

	maxShiftX := max(20, b.Dx()/10)
	maxShiftY := max(20, b.Dy()/10)

	offsetX := r.Intn(2*maxShiftX+1) - maxShiftX
	offsetY := r.Intn(2*maxShiftY+1) - maxShiftY

	x := centerX + offsetX
	y := centerY + offsetY

	minX := b.Min.X + padding
	maxX := b.Max.X - padding - textW

	minY := b.Min.Y + padding + ascent
	maxY := b.Max.Y - padding - descent

	if x < minX {
		x = minX
	}
	if x > maxX {
		x = maxX
	}
	if y < minY {
		y = minY
	}
	if y > maxY {
		y = maxY
	}

	dot := fixed.P(x, y)
	dotShadow := fixed.P(x+3, y+3)

	ink := image.NewUniform(color.NRGBA{R: 255, G: 255, B: 255, A: 110})
	inkShadow := image.NewUniform(color.NRGBA{R: 0, G: 0, B: 0, A: 110})

	d := &font.Drawer{
		Dst:  dst,
		Src:  inkShadow,
		Face: face,
		Dot:  dotShadow,
	}
	d.DrawString(text)

	d = &font.Drawer{
		Dst:  dst,
		Src:  ink,
		Face: face,
		Dot:  dot,
	}
	d.DrawString(text)

	return dst, nil
}
