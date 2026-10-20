package main

import (
	"image"
	"math"
	"slices"

	"github.com/disintegration/imaging"
	img "github.com/disintegration/imaging"
)

var rsfMap = map[string]img.ResampleFilter{
	"box":                img.Box,
	"linear":             img.Linear,
	"lanczos":            img.Lanczos,
	"catmull-rom":        img.CatmullRom,
	"nearest-neighbor":   img.NearestNeighbor,
	"mitchell-netravali": img.MitchellNetravali,
}

type FitType int

const (
	Fit FitType = iota
	Resize
	Fill
)

var ftMap = map[string]FitType{
	"fit":    Fit,
	"fill":   Fill,
	"resize": Resize,
}

func GenImage(path string, fitType FitType, resize img.ResampleFilter) (string, error) {
	f, err := img.Open(path, img.AutoOrientation(true))
	if err != nil {
		return "", err
	}

	var scaleX, scaleY int

	bounds := f.Bounds()

	if bounds.Size().X >= optLineWidth-1 {
		scaleX = optLineWidth - 1
	} else if bounds.Size().Y >= windowSize.Height {
		scaleY = windowSize.Height
	} else {
		return asciiImage(f), nil
	}

	f = scaleImage(f, scaleX, scaleY, fitType, resize)
	return asciiImage(f), nil
}

func scaleImage(file image.Image, scaleX, scaleY int, fitType FitType, resizeType img.ResampleFilter) image.Image {
	if scaleX < 0 || scaleY < 0 {
		panic("Invalid scale")
	}

	switch fitType {
	case Fit:
		return imaging.Fit(file, scaleX, scaleY, resizeType)
	case Fill:
		return imaging.Fill(file, scaleX, scaleY, imaging.Center, resizeType)
	case Resize:
		return imaging.Resize(file, scaleX, scaleY, resizeType)
	}

	panic("Unreachable")
}

type lvlShift struct {
	low  float64
	char rune
}

func genCharScale(chars string) []lvlShift {
	step := float64(255.0 / len(chars))

	var charVec = []rune(chars)
	var out []lvlShift
	for fl, idx := 0., 0; idx < len(chars); fl += step {
		out = append(out, lvlShift{
			low:  fl,
			char: charVec[idx],
		})

		idx++
	}

	slices.Reverse(out)
	return out
}

func asciiImage(f image.Image) string {
	imgBounds := f.Bounds().Size()

	f = imaging.Grayscale(f)
	charScale := genCharScale(optImageChars)

	var out []rune
	for x := range imgBounds.X {
		out = slices.Concat(out, []rune("<img>"))

		for y := range imgBounds.Y {
			// In grayscale, RGB are the same
			intensity, _, _, _ := f.At(x, y).RGBA()

			for _, elem := range charScale {
				if intensity > uint32(math.Trunc(elem.low)) {
					out = append(out, elem.char)
					break
				}
			}
		}

		out = append(out, '\n')
	}

	return string(out)
}
