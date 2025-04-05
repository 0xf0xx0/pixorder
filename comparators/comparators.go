package comparators

import (
	"pixorder/shared"
	"pixorder/types"
)

var ComparatorFunctionMappings = map[string]types.ComparatorFunc{
	"red":        Red,
	"green":      Green,
	"blue":       Blue,
	"hue":        Hue,
	"saturation": Saturation,
	"lightness":  Lightness,
	"max":        Max,
	"min":        Min,
}

func Red(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	return int(a.R) - int(b.R)
}

func Green(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	return int(a.G) - int(b.G)
}

func Blue(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	return int(a.B) - int(b.B)
}

func Hue(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	aHue := calculateHue(a)
	bHue := calculateHue(b)

	return int(aHue - bHue)
}

func Saturation(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	aSat := calculateSaturation(a)
	bSat := calculateSaturation(b)
	return int(aSat - bSat)
}

func Lightness(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	aLightness := calculateLightness(a)
	bLightness := calculateLightness(b)
	return int(aLightness - bLightness)
}

func Max(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	aMax := int(max(a.R, a.G, a.B))
	bMax := int(max(b.R, b.G, b.B))
	return aMax - bMax
}
func Min(a, b types.PixelWithMask) int {
	if skipPixel(a) || skipPixel(b) {
		return 0
	}

	aMin := int(min(a.R, a.G, a.B))
	bMin := int(min(b.R, b.G, b.B))
	return aMin - bMin
}

func skipPixel(pixel types.PixelWithMask) bool {
	/// skip if masked
	if pixel.Mask == 255 {
		return true
	}
	/// and if null
	if pixel.R == 0 && pixel.G == 0 && pixel.B == 0 && pixel.A == 0 {
		return true
	}
	/// and if beyond thresholds
	/// FIXME: figure out why thresholds with spiral results in holes in the image
	lightness := calculateLightness(pixel)
	if lightness < shared.Config.Thresholds.Lower*255 || lightness > shared.Config.Thresholds.Upper*255 {
		return true
	}
	return false
}

func calculateLightness(pixel types.PixelWithMask) float32 {
	// 299, 587, 114
	return float32(pixel.R)*0.29 + float32(pixel.G)*0.59 + float32(pixel.B)*0.11
}
func calculateHue(pixel types.PixelWithMask) float32 {
	hue := float32(0)
	maxV := max(pixel.R, pixel.G, pixel.B)
	minV := min(pixel.R, pixel.G, pixel.B)
	switch maxV {
	case pixel.R:
		{
			hue = max(1, float32(pixel.G)-float32(pixel.B))
			break
		}
	case pixel.G:
		{
			hue = 2 + (float32(pixel.B) - float32(pixel.R))
			break
		}
	case pixel.B:
		{
			hue = 4 + (float32(pixel.R) - float32(pixel.G))
		}
	}
	/// finish formula and convert to degrees
	/// and avoid divide-by-zero
	hue = (hue / max(1, float32(maxV)-float32(minV))) * 60
	if hue < 0 {
		hue += 360
	}
	return hue
}
func calculateSaturation(pixel types.PixelWithMask) float32 {
	saturation := float32(0)
	/// pixels are RGBA so skip the A
	minc := float32(min(pixel.R, pixel.G, pixel.B))
	maxc := float32(max(pixel.R, pixel.G, pixel.B))

	if minc == maxc {
		return saturation
	}

	sum := maxc + minc
	diff := maxc - minc
	lightness := (sum / 2) / 255
	if lightness < 0.5 {
		saturation = diff / sum
	} else {
		saturation = diff / (2 - diff)
	}
	return saturation * 1000
}
