package video

import (
	"image"
	"math"

	"github.com/nfnt/resize"

	"github.com/kbinani/screenshot"
)

//A list of supported aspect ratio
var (
	Ratio4to3  = &Resolution{960, 720}
	Ratio16to9 = &Resolution{1280, 480}
	Ratios     = []*Resolution{Ratio4to3, Ratio16to9}
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

		img, _ := fitNearestRatio(screenshot)

		images = append(images, rgbToIYUV(img))
	}
	return images, nil
}

func fitNearestRatio(img *image.RGBA) (*image.RGBA, *Resolution) {
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
	casted = resize.Resize(uint(fittestRatio.Width), uint(fittestRatio.Height), casted, resize.Bilinear)
	return casted.(*image.RGBA), fittestRatio
}
