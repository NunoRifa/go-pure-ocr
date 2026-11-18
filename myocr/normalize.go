package myocr

import "strings"

const NormSize = 16 // Normalize every character to 16x16 pixels

// Normalize scales any cropped character Matrix to a fixed 16x16 bitmap string of '0' and '1'.
func Normalize(charGrid Matrix) string {
	srcH := len(charGrid)
	if srcH == 0 {
		return strings.Repeat("0", NormSize*NormSize)
	}

	srcW := len(charGrid[0])
	var sb strings.Builder
	sb.Grow(NormSize * NormSize)

	for y := 0; y < NormSize; y++ {
		// Nearest-neighbor vertical mapping
		srcY := (y * srcH) / NormSize
		if srcY >= srcH {
			srcY = srcH - 1
		}

		for x := 0; x < NormSize; x++ {
			// Nearest-neighbor horizontal mapping
			srcX := (x * srcW) / NormSize
			if srcX >= srcW {
				srcX = srcW - 1
			}

			if charGrid[srcY][srcX] {
				sb.WriteByte('1')
			} else {
				sb.WriteByte('0')
			}
		}
	}
	return sb.String()
}

// HammingDistance counts how many pixels differ between two 256-bit normalized strings.
// Lower distance = higher similarity.
func HammingDistance(a, b string) int {
	if len(a) != len(b) {
		return NormSize * NormSize
	}

	diff := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			diff++
		}
	}
	return diff
}
