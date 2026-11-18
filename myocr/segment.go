package myocr

// CharSegment holds the cropped binary matrix of a single character
// and whether a space precedes it.
type CharSegment struct {
	Grid         Matrix
	SpaceBefore  bool
	NewLineAfter bool
}

// SegmentText slices a full-page binary Matrix into individual character matrices.
func SegmentText(page Matrix) []CharSegment {
	var segments []CharSegment
	lines := splitLines(page)

	for lineIdx, line := range lines {
		charBoxes := splitCharsInLine(line)
		if len(charBoxes) == 0 {
			continue
		}

		// Calculate average character width to detect word spaces
		totalWidth := 0
		for _, box := range charBoxes {
			totalWidth += (box.endX - box.startX)
		}

		avgWidth := float64(totalWidth) / float64(len(charBoxes))
		spaceThreshold := int(avgWidth * 0.45)
		if spaceThreshold < 2 {
			spaceThreshold = 2
		}

		for i, box := range charBoxes {
			cropped := cropTight(line, box.startX, box.endX)
			if len(cropped) == 0 {
				continue
			}

			spaceBefore := false
			if i > 0 {
				gap := box.startX - charBoxes[i-1].endX
				if gap > spaceThreshold {
					spaceBefore = true
				}
			}

			isLastInLine := (i == len(charBoxes)-1) && (lineIdx < len(line)-1)
			segments = append(segments, CharSegment{
				Grid:         cropped,
				SpaceBefore:  spaceBefore,
				NewLineAfter: isLastInLine,
			})
		}
	}
	return segments
}

type span struct {
	startX, endX int
}

// cropTight extracts columns [startX:endX] and trims empty rows from top and bottom.
func cropTight(line Matrix, startX, endX int) Matrix {
	topY, bottomY := -1, -1
	for y := 0; y < len(line); y++ {
		for x := startX; x < endX; x++ {
			if line[y][x] {
				if topY == -1 {
					topY = y
				}
				bottomY = y
			}
		}
	}
	if topY == -1 {
		return nil
	}

	height := bottomY - topY + 1
	width := endX - startX
	cropped := make(Matrix, height)
	for y := 0; y < height; y++ {
		cropped[y] = make([]bool, width)
		copy(cropped[y], line[topY+y][startX:endX])
	}
	return cropped
}

// splitCharsInLine finds the left and right X boundaries of each character in a line.
func splitCharsInLine(line Matrix) []span {
	if len(line) == 0 {
		return nil
	}

	width := len(line[0])
	var spans []span
	inChar := false
	startX := 0

	for x := 0; x < width; x++ {
		hasInk := colHasInk(line, x)
		if hasInk && !inChar {
			inChar = true
			startX = x
		} else if !hasInk && inChar {
			inChar = false
			spans = append(spans, span{startX: startX, endX: x})
		}
	}

	if inChar {
		spans = append(spans, span{startX: startX, endX: width})
	}

	return spans
}

func splitLines(page Matrix) []Matrix {
	var lines []Matrix
	inLine := false
	startY := 0

	for y := 0; y < len(page); y++ {
		hasInk := rowHasInk(page[y])
		if hasInk && !inLine {
			inLine = true
			startY = y
		} else if !hasInk && inLine {
			inLine = false
			lines = append(lines, page[startY:y])
		}
	}

	if inLine {
		lines = append(lines, page[startY:])
	}

	return lines
}

func rowHasInk(row []bool) bool {
	for _, pixel := range row {
		if pixel {
			return true
		}
	}
	return false
}

func colHasInk(matrix Matrix, x int) bool {
	for y := 0; y < len(matrix); y++ {
		if matrix[y][x] {
			return true
		}
	}
	return false
}
