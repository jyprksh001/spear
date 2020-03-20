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

//RGBToYUV converts image.RGBA to image.YCbCr
func RGBToYUV(rgbaImage *image.RGBA) *image.YCbCr {
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

//YUVToRGB converts image.YCbCr to image.RGBA
func YUVToRGB(yuvImage *image.YCbCr) *image.RGBA {
	bounds := yuvImage.Bounds()
	output := image.NewRGBA(bounds)

	width := C.uint(bounds.Dx())
	height := C.uint(bounds.Dy())
	rgba := (*C.uchar)(C.CBytes(output.Pix))
	stride := C.uint(output.Stride)

	y := (*C.uchar)(C.CBytes(yuvImage.Y))
	u := (*C.uchar)(C.CBytes(yuvImage.Cb))
	v := (*C.uchar)(C.CBytes(yuvImage.Cr))

	yStride := C.uint(yuvImage.YStride)
	uvStride := C.uint(yuvImage.CStride)
	C.yuv420_rgb24_sseu(width, height, y, u, v, yStride, uvStride, rgba, stride, C.YCBCR_709)

	rgb := C.GoBytes((unsafe.Pointer)(rgba), C.int(len(output.Pix)))
	copy(output.Pix, rgb)

	return output
}
