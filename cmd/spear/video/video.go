package video

import (
	"image"
	"image/color"
	"math"

	"github.com/nfnt/resize"

	"github.com/kbinani/screenshot"
)

//A list of supported aspect ratio
var (
	Ratio4to3   = &Resolution{1024, 768}
	Ratio5to4   = &Resolution{1280, 1024}
	Ratio3to2   = &Resolution{2160, 1440}
	Ratio16to10 = &Resolution{1280, 800}
	Ratio16to9  = &Resolution{1366, 768}
	Ratio21to9  = &Resolution{2560, 1080}
	Ratio32to9  = &Resolution{3840, 1080}
	Ratios      = []*Resolution{Ratio4to3, Ratio5to4, Ratio3to2, Ratio16to10, Ratio16to9, Ratio21to9, Ratio32to9}
)

//ScreencastFPS refers to screensharing fps
const ScreencastFPS = 30

//Resolution refers to the resolution of one casted display
type Resolution struct {
	Width, Height int
}

//Screenshot takes a screenshot from all display
func Screenshot() ([]*image.YCbCr, error) {
	images := []*image.YCbCr{}
	n := screenshot.NumActiveDisplays()
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		screenshot, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return nil, err
		}

		img, size := fitNearestRatio(screenshot)

		images = append(images, rgbToYCbCr420(&img, size))
	}
	return images, nil
}

func fitNearestRatio(img *image.RGBA) (image.Image, *Resolution) {
	ratioDiff := 10.0
	var fittestRatio *Resolution
	for _, ratio := range Ratios {
		whRato := float64(ratio.Width) / float64(ratio.Height)
		imgWhRatio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
		if diff := math.Abs(whRato - imgWhRatio); diff < ratioDiff {
			ratioDiff = diff
			fittestRatio = ratio
		}
	}

	var casted image.Image = img
	casted = resize.Resize(uint(fittestRatio.Width), uint(fittestRatio.Height), casted, resize.Lanczos3)
	return casted, fittestRatio
}

func rgbToYCbCr420(rgba *image.Image, size *Resolution) *image.YCbCr {
	img := image.NewYCbCr(image.Rect(0, 0, size.Width, size.Height), image.YCbCrSubsampleRatio420)

	//Init Y
	i := 0
	for x := 0; x < size.Width; x++ {
		for y := 0; y < size.Height; y++ {
			r, g, b, _ := (*rgba).At(x, y).RGBA()
			Y, _, _ := color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))
			img.Y[i] = Y
			i++
		}
	}

	//Init Cr, Cb
	resized := resize.Resize(uint(size.Width/2), uint(size.Height/2), *rgba, resize.Lanczos3)
	i = 0
	for x := 0; x < size.Width/2; x++ {
		for y := 0; y < size.Height/2; y++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			_, Cb, Cr := color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))
			img.Cb[i] = Cb
			img.Cr[i] = Cr
			i++
		}
	}

	return img
}
