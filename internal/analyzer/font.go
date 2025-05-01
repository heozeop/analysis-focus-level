package analyzer

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/image/font/opentype"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
)

// InitKoreanFont loads and registers a Korean font with the Gonum plotting library
func InitKoreanFont() error {
	fontFilePath := FindKoreanFont()
	if fontFilePath == "" {
		return fmt.Errorf("no Korean font found")
	}

	// Read the font file
	fontData, err := ioutil.ReadFile(fontFilePath)
	if err != nil {
		return fmt.Errorf("error reading font file: %v", err)
	}

	// Parse the font
	fontFace, err := opentype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("error parsing font: %v", err)
	}

	// Create a font.Collection with the Korean font
	koreanFont := font.Font{Typeface: "KoreanFont"}
	collection := font.Collection{
		{
			Font: koreanFont,
			Face: fontFace,
		},
	}

	// Add the font to the default font cache
	font.DefaultCache.Add(collection)

	// Set it as the default font for plots
	plot.DefaultFont = koreanFont

	fmt.Printf("Successfully registered Korean font from: %s\n", fontFilePath)
	return nil
} 