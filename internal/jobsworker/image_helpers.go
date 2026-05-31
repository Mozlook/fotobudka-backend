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
	b := src.Bounds()

	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	if b.Empty() {
		return dst, nil
	}

	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	shortSide := minInt(b.Dx(), b.Dy())

	// Skala watermarka zależna od rozmiaru proofa.
	// Dla proofów około 1600–2500 px daje duży, ale nadal czytelny pattern.
	fontSize := clampFloat64(float64(shortSide)/9.5, 48, 160)

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, err
	}

	text := "FotoBudka"
	textW := font.MeasureString(face, text).Ceil()

	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	textH := ascent + descent

	// Odstępy między watermarkami. Im mniejsze wartości, tym gęstszy pattern.
	stepX := textW + maxInt(80, int(fontSize*1.05))
	stepY := textH + maxInt(90, int(fontSize*1.65))

	r := rand.New(rand.NewSource(int64(seed)))

	baseX := b.Min.X - textW - stepX + r.Intn(stepX)
	baseY := b.Min.Y - textH - stepY + r.Intn(stepY)

	maxJitter := maxInt(8, int(fontSize/10))
	shadowOffset := maxInt(2, int(fontSize/24))

	// Przy powtarzalnym watermarku alfa nie może być za duża,
	// bo inaczej proof będzie za ciężki wizualnie.
	ink := image.NewUniform(color.NRGBA{R: 255, G: 255, B: 255, A: 82})
	inkShadow := image.NewUniform(color.NRGBA{R: 0, G: 0, B: 0, A: 70})

	row := 0

	for baseline := baseY + ascent; baseline < b.Max.Y+textH+stepY; baseline += stepY {
		rowShift := 0
		if row%2 == 1 {
			rowShift = stepX / 2
		}

		// Delikatny dryf tworzy mniej regularną siatkę.
		drift := (row * stepX / 6) % stepX
		jitterX := r.Intn(2*maxJitter+1) - maxJitter

		x := baseX + rowShift - drift + jitterX

		for x < b.Max.X+stepX {
			drawWatermarkText(
				dst,
				face,
				text,
				x,
				baseline,
				shadowOffset,
				ink,
				inkShadow,
			)

			x += stepX
		}

		row++
	}

	return dst, nil
}

func drawWatermarkText(
	dst draw.Image,
	face font.Face,
	text string,
	x int,
	y int,
	shadowOffset int,
	ink image.Image,
	inkShadow image.Image,
) {
	shadow := &font.Drawer{
		Dst:  dst,
		Src:  inkShadow,
		Face: face,
		Dot:  fixed.P(x+shadowOffset, y+shadowOffset),
	}
	shadow.DrawString(text)

	main := &font.Drawer{
		Dst:  dst,
		Src:  ink,
		Face: face,
		Dot:  fixed.P(x, y),
	}
	main.DrawString(text)
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func clampFloat64(value float64, minValue float64, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}

	if value > maxValue {
		return maxValue
	}

	return value
}
