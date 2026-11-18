package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/NunoRifa/go-pure-ocr/myocr"
)

func main() {
	genDemo := flag.Bool("gen-demo", false, "Generate sample train.png and test.png images")
	trainImg := flag.String("train", "", "Path to training image")
	labels := flag.String("labels", "", "Characters present in the training image")
	modelPath := flag.String("model", "model.json", "Path to save/load trained JSON model")
	inputImg := flag.String("i", "", "Path to input image to recognize")
	outputTxt := flag.String("o", "", "Path to save output .txt file (optional)")
	flag.Parse()

	// Mode 1: Generate sample images for instant testing
	if *genDemo {
		if err := createSampleImages(); err != nil {
			log.Fatalf("Failed to generate demo images: %v", err)
		}
		fmt.Println("Generated 'train.png' (alphabet/numbers) and 'test.png' (sample text)!")
		return
	}

	engine := myocr.NewEngine()

	// Mode 2: Train the OCR engine and save model.json
	if *trainImg != "" {
		if *labels == "" {
			log.Fatal("Error: Please provide -labels matching the characters in -train image")
		}

		// Load existing model if present so you can train incrementally
		_ = engine.LoadModel(*modelPath)

		if err := engine.TrainFromImage(*trainImg, *labels); err != nil {
			log.Fatalf("Training failed: %v", err)
		}
		if err := engine.SaveModel(*modelPath); err != nil {
			log.Fatalf("Failed to save model: %v", err)
		}
		fmt.Printf("Successfully trained %d characters and saved to %s\n", len(engine.Templates), *modelPath)
		return
	}

	// Mode 3: Run OCR Recognition on an input image
	if *inputImg != "" {
		if err := engine.LoadModel(*modelPath); err != nil {
			log.Fatalf("Failed to load model '%s': %v (run -train first)", *modelPath, err)
		}

		text, err := engine.Recognize(*inputImg)
		if err != nil {
			log.Fatalf("OCR Recognition failed: %v", err)
		}

		if *outputTxt != "" {
			if err := os.WriteFile(*outputTxt, []byte(text+"\n"), 0644); err != nil {
				log.Fatalf("Failed to write output file: %v", err)
			}
			fmt.Printf("Saved recognized text to %s\n", *outputTxt)
		} else {
			fmt.Println("--- RECOGNIZED TEXT ---")
			fmt.Println(text)
			fmt.Println("-----------------------")
		}

		return
	}

	fmt.Println("Usage:")
	fmt.Println("  1. Generate demo images:  go run ./cmd/ocr-cli -gen-demo")
	fmt.Println("  2. Train OCR model:       go run ./cmd/ocr-cli -train train.png -labels \"GOLANG0123456789\"")
	fmt.Println("  3. Recognize image text:  go run ./cmd/ocr-cli -i test.png -o result.txt")
}

func createSampleImages() error {
	glyphs := map[rune][]string{
		'G': {"011110", "100001", "100000", "100111", "100001", "100001", "011110"},
		'O': {"011110", "100001", "100001", "100001", "100001", "100001", "011110"},
		'L': {"100000", "100000", "100000", "100000", "100000", "100000", "111111"},
		'A': {"001100", "010010", "100001", "111111", "100001", "100001", "100001"},
		'N': {"100001", "110001", "101001", "100101", "100011", "100001", "100001"},
		'2': {"011110", "100001", "000001", "001110", "010000", "100000", "111111"},
		'0': {"011110", "100001", "100011", "10101", "110001", "100001", "011110"},
		'6': {"011110", "100000", "100000", "111110", "100001", "100001", "011110"},
	}

	// Fix '0' row width to 6 columns
	glyphs['0'] = []string{"011110", "100001", "100101", "101001", "100001", "100001", "011110"}

	if err := renderStringImage("train.png", "GOLANG206", glyphs, 4); err != nil {
		return err
	}
	return renderStringImage("test.png", "GO LANG 2026", glyphs, 3)
}

func renderStringImage(filename, text string, glyphs map[rune][]string, scale int) error {
	charW, charH := 6*scale, 7*scale
	spacing := 2 * scale
	margin := 10
	width := margin*2 + len(text)*(charW+spacing)
	height := margin*2 + charH

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill white background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	cursorX := margin
	for _, ch := range text {
		if ch == ' ' {
			cursorX += charW
			continue
		}

		pattern := glyphs[ch]
		for rowIdx, row := range pattern {
			for colIdx, bit := range row {
				if bit == '1' {
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							img.Set(cursorX+colIdx*scale+dx, margin+rowIdx*scale+dy, color.Black)
						}
					}
				}
			}
		}
		cursorX += charW + spacing
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
