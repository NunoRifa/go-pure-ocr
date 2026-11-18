package myocr

import (
	"fmt"
	"image"
	"os"
)

// Matrix represents a 2D binary image where true = black ink, false = white background.
type Matrix [][]bool

// LoadAndBinarize opens an image file and converts it into a 2D binary matrix.
func LoadAndBinarize(filepath string, threshold uint8) (Matrix, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return BinarizeImage(img, threshold), nil
}

// BinarizeImage converts any Go image.Image into a binary Matrix.
func BinarizeImage(img image.Image, threshold uint8) Matrix {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	matrix := make(Matrix, height)
	for y := 0; y < height; y++ {
		matrix[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()

			// Transparent pixels are treated as white background
			if a < 32768 {
				matrix[y][x] = false
				continue
			}

			// Convert 16-bit color channels (0-65535) to 8-bit (0-255)
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			// Calculate grayscale luminance
			luminance := uint8(0.299*r8 + 0.587*g8 + 0.114*b8)

			// If pixel is darker than threshold, mark as ink (true)
			matrix[y][x] = luminance < threshold
		}
	}
	return matrix
}
