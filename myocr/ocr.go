package myocr

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

// Template stores a learned character label and its 16x16 normalized bitmap string.
type Template struct {
	Char   string `json:"char"`
	Bitmap string `json:"bitmap"` // 256-character string of '0' and '1'
}

// Engine is our custom pure-Go OCR engine.
type Engine struct {
	Threshold uint8      `json:"threshold"`
	Templates []Template `json:"templates"`
}

// NewEngine creates a new OCR engine with a default luminance threshold (128).
func NewEngine() *Engine {
	return &Engine{
		Threshold: 128,
		Templates: []Template{},
	}
}

// TrainFromImage learns character templates from an image containing known characters.
func (e *Engine) TrainFromImage(imagePath string, expectedChars string) error {
	matrix, err := LoadAndBinarize(imagePath, e.Threshold)
	if err != nil {
		return err
	}

	segments := SegmentText(matrix)

	// Remove spaces/newlines from expectedChars so it matches segmented glyphs 1-to-1
	var runes []rune
	for _, r := range expectedChars {
		if r != ' ' && r != '\n' && r != '\r' {
			runes = append(runes, r)
		}
	}

	if len(segments) != len(runes) {
		return fmt.Errorf(
			"training mismatch: image contains %d segmented characters, but label string has %d non-space characters",
			len(segments), len(runes),
		)
	}

	for i, seg := range segments {
		norm := Normalize(seg.Grid)
		e.Templates = append(e.Templates, Template{
			Char:   string(runes[i]),
			Bitmap: norm,
		})
	}

	return nil
}

// Recognize extracts text from an image file using the trained templates.
func (e *Engine) Recognize(imagePath string) (string, error) {
	if len(e.Templates) == 0 {
		return "", fmt.Errorf("OCR engine has no trained templates; train or load a model first")
	}

	matrix, err := LoadAndBinarize(imagePath, e.Threshold)
	if err != nil {
		return "", err
	}

	segments := SegmentText(matrix)
	var result strings.Builder

	for _, seg := range segments {
		if seg.SpaceBefore {
			result.WriteByte(' ')
		}

		norm := Normalize(seg.Grid)
		bestChar := e.matchBest(norm)
		result.WriteString(bestChar)

		if seg.NewLineAfter {
			result.WriteByte('\n')
		}
	}

	return result.String(), nil
}

// matchBest finds the template with the smallest Hamming distance (k-NN with k=1).
func (e *Engine) matchBest(bitmap string) string {
	bestDist := math.MaxInt32
	bestChar := "?"

	for _, tmpl := range e.Templates {
		dist := HammingDistance(bitmap, tmpl.Bitmap)
		if dist < bestDist {
			bestDist = dist
			bestChar = tmpl.Char
		}
	}

	return bestChar
}

// SaveModel exports the learned character templates to a JSON file.
func (e *Engine) SaveModel(filePath string) error {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// LoadModel loads learned character templates from a JSON file.
func (e *Engine) LoadModel(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, e)
}
