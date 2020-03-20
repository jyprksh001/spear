package video

/*
#include <stdio.h>
#include "yuv_rgb.h"
*/
import "C"
import (
	"image"
	"unsafe"
)

func rgbToIYUV(rgbaImage *image.RGBA) *image.YCbCr {
	bounds := rgbaImage.Bounds()
	output := image.NewYCbCr(bounds, image.YCbCrSubsampleRatio420)

	width := C.uint(bounds.Dx())
	height := C.uint(bounds.Dy())
	rgba := (*C.uchar)(C.CBytes(rgbaImage.Pix))
	stride := C.uint(rgbaImage.Stride)

	y := (*C.uchar)(C.CBytes(output.Y))
	u := (*C.uchar)(C.CBytes(output.Cb))
	v := (*C.uchar)(C.CBytes(output.Cr))

	yStride := C.uint(output.YStride)
	uvStride := C.uint(output.CStride)
	C.rgb32_yuv420_sseu(width, height, rgba, stride, y, u, v, yStride, uvStride, C.YCBCR_709)

	copy(output.Y, C.GoBytes((unsafe.Pointer)(y), C.int(len(output.Y))))
	copy(output.Cb, C.GoBytes((unsafe.Pointer)(u), C.int(len(output.Cb))))
	copy(output.Cr, C.GoBytes((unsafe.Pointer)(v), C.int(len(output.Cr))))
	return output
}
