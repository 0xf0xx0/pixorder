package intervals

import (
	"math"
	mathRand "math/rand"
	"slices"

	"pixelsort_go/comparators"
	"pixelsort_go/shared"
	"pixelsort_go/types"
)

var SortingFunctionMappings = map[string]func([]types.PixelWithMask){
	"row":         Row,
	"random":      Random,
	"shuffle":     Shuffle,
	"smear":       Smear,
	"wave":        Wave,
}

func Sort(section []types.PixelWithMask) {
	sorter := SortingFunctionMappings[shared.Config.Interval]
	stretches := getUnmaskedStretches(section)
	for i := 0; i < len(stretches); i++ {
		stretch := stretches[i]
		sorter(section[stretch.Start:stretch.End])
	}
}

// sorters
func Shuffle(interval []types.PixelWithMask) {
	/// we want shuffling to respect thresholds/masks too, so
	/// use the result to determine whether to skip or not
	comparator := comparators.ComparatorFunctionMappings[shared.Config.Comparator]
	mathRand.Shuffle(len(interval), func(i, j int) {
		skip := comparator(interval[i], interval[j])
		if skip != 0 {
			interval[i], interval[j] = interval[j], interval[i]
		}
	})
}

// TODO
// WHY doesnt rotation work??? :<
// it DOES work, but only if the smear doesnt touch the edge
func Smear(interval []types.PixelWithMask) {
	intervalLength := len(interval)
	if intervalLength == 0 {
		return
	}
	smearedPixel := interval[0] /// TODO: why is this empty with an angle?

	for idx := range interval {
		// println(idx, pix.R, pix.G, pix.B, pix.A)
		interval[idx] = smearedPixel
	}
}

// randomly skips rows
func Row(interval []types.PixelWithMask) {
	if mathRand.Float32() > shared.Config.Randomness {
		return
	}
	commonSort([]types.PixelStretch{{Start: 0, End: len(interval)}}, interval)
}

// takes a randomly-sized chunk of the remaining pixels and sorts them
func Random(interval []types.PixelWithMask) {
	stretches := make([]types.PixelStretch, 0)
	intervalLength := len(interval)

	j := 0
	for {
		if j >= intervalLength {
			break
		}
		randLength := randBetween((intervalLength - j), 1)
		if mathRand.Float32() < shared.Config.Randomness {
			endIdx := min(j+randLength, intervalLength)
			stretches = append(stretches, types.PixelStretch{Start: j, End: endIdx})
		}
		j += randLength
	}

	commonSort(stretches, interval)
}

// sorts in "waves" across the interval
// not very useful with complex masks
func Wave(interval []types.PixelWithMask) {
	stretches := make([]types.PixelStretch, 0)
	intervalLength := len(interval)
	baseLength := shared.Config.SectionLength

	j := 0
	for {
		if j >= intervalLength {
			break
		}
		/// how far out waves will reach past their base length
		/// clamp to no further than baseLen
		waveOffsetMin := math.Floor(float64(float32(baseLength) * shared.Config.Randomness))

		/// waves can reach forward or hang back
		waveLength := baseLength + randBetween(int(waveOffsetMin), int(-waveOffsetMin))

		/// now add to stretches
		endIdx := min(j+waveLength, intervalLength)
		stretches = append(stretches, types.PixelStretch{Start: j, End: endIdx})
		j += waveLength
	}
	commonSort(stretches, interval)
}

///

///

// util
func randBetween(max int, min_opt ...int) int {
	min := 0
	if len(min_opt) > 0 {
		min = min_opt[0]
	}
	randNum := mathRand.Float64()
	if min != 0 {
		return int(math.Floor(randNum*float64(((+max)+1)-(+min)))) + (+min)
	}
	return int(math.Floor(randNum * float64((+max)+1)))
}

func commonSort(stretches []types.PixelStretch, interval []types.PixelWithMask) {
	for stretchIdx := 0; stretchIdx < len(stretches); stretchIdx++ {
		stretch := stretches[stretchIdx]
		/// grab the pixels we want
		pixels := interval[stretch.Start:stretch.End]
		jInit := len(pixels) - 1
		comparator := comparators.ComparatorFunctionMappings[shared.Config.Comparator]
		slices.SortStableFunc(pixels, comparator)

		if shared.Config.Reverse {
			/// do a flip!
			for i, j := 0, jInit; i < j; i, j = i+1, j-1 {
				pixels[i], pixels[j] = pixels[j], pixels[i]
			}
		}
	}
}

// select all pixels not masked off
func getUnmaskedStretches(interval []types.PixelWithMask) []types.PixelStretch {
	stretches := make([]types.PixelStretch, 0)
	baseIdx := 0
	intervalLen := len(interval)

	for j := 0; j < intervalLen; j++ {
		pixel := interval[j]
		/// if masked off, or nil
		if pixel.Mask == 255 || (pixel.R == 0 && pixel.G == 0 && pixel.B == 0 && pixel.A == 0) {
			/// look ahead for the end of the mask
			endMaskIdx := j
			for {
				if endMaskIdx >= intervalLen {
					break
				}
				if interval[endMaskIdx].Mask != 255 || !(pixel.R == 0 && pixel.G == 0 && pixel.B == 0 && pixel.A == 0) {
					break
				}
				endMaskIdx++
			}

			stretch := types.PixelStretch{Start: baseIdx, End: j}
			stretches = append(stretches, stretch)

			/// jump past the mask and continue
			baseIdx = endMaskIdx
			j = baseIdx
		}
	}
	/// and then add any remaning unmasked pixels
	if baseIdx < intervalLen {
		stretches = append(stretches, types.PixelStretch{Start: baseIdx, End: intervalLen})
	}
	return stretches
}
