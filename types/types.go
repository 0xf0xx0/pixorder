package types

import "image/color"

type PixelWithMask struct {
	R, G, B, A uint8
	Mask       uint8
}

func (pixel PixelWithMask) ToColor() color.RGBA {
	return color.RGBA{
		R: pixel.R,
		G: pixel.G,
		B: pixel.B,
		A: pixel.A,
	}
}
func PixelWithMaskFromColor(color color.RGBA, mask uint8) PixelWithMask {
	return PixelWithMask{
		R:    color.R,
		G:    color.G,
		B:    color.B,
		A:    color.A,
		Mask: mask,
	}
}

type PixelStretch struct {
	Start int
	End   int
}

type ThresholdConfig struct {
	Lower, Upper float32
}

type ComparatorFunc func(a, b PixelWithMask) int

type SorterFunc func(interval []PixelWithMask)
