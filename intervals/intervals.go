package intervals

import (
	"math"
	mathRand "math/rand"
	"slices"

	"pixelsort_go/comparators"
	"pixelsort_go/shared"
	"pixelsort_go/types"
)

/// seam building functions + some alt "sorts"
/// an interval/seam is a full slice of pixels from shared.Config.Pattern
/// a stretch is a section of a seam
/// types.PixelStretch is the "skeleton" stretch
/// i should get this straightened out

// interval sorting algos
var IntervalFunctionMappings = map[string]func([]types.PixelWithMask){
	"none":    None,
	"random":  Random,
	"shuffle": Shuffle,
	"smear":   Smear,
	"wave":    Wave,
}

// sorters

func Sort(seam []types.PixelWithMask) {
	sorter := IntervalFunctionMappings[shared.Config.Interval]
	stretches := getUnmaskedStretches(seam)
	for i := 0; i < len(stretches); i++ {
		stretch := stretches[i]
		sorter(seam[stretch.Start:stretch.End])
	}
}
func Shuffle(seam []types.PixelWithMask) {
	/// we want shuffling to respect thresholds/masks too, so
	/// use the result to determine whether to skip or not
	comparator := comparators.ComparatorFunctionMappings[shared.Config.Comparator]
	mathRand.Shuffle(len(seam), func(i, j int) {
		skip := comparator(seam[i], seam[j])
		if skip != 0 {
			seam[i], seam[j] = seam[j], seam[i]
		}
	})
}
// copy the furst pixel across the rest of the seam
func Smear(seam []types.PixelWithMask) {
	intervalLength := len(seam)
	if intervalLength == 0 {
		return
	}
	smearedPixel := seam[0]

	for idx := range seam {
		seam[idx] = smearedPixel
	}
}

// noop, returns a single stretch containing the full seam
func None(seam []types.PixelWithMask) {
	commonSort([]types.PixelStretch{{Start: 0, End: len(seam)}}, seam)
}

// takes a randomly-sized chunk of the remaining pixels and sorts them
func Random(seam []types.PixelWithMask) {
	stretches := make([]types.PixelStretch, 0)
	intervalLength := len(seam)

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

	commonSort(stretches, seam)
}

// sorts in "waves" across the interval
// not very useful with complex masks
func Wave(seam []types.PixelWithMask) {
	stretches := make([]types.PixelStretch, 0)
	intervalLength := len(seam)
	baseLength := shared.Config.SectionLength

	j := 0
	for {
		if j >= intervalLength {
			break
		}
		/// how far out waves will reach past (or hang behind) their base length
		/// clamp to no further than baseLen
		waveOffsetMin := math.Floor(float64(float32(baseLength) * shared.Config.Randomness))

		/// waves can reach forward or hang back
		waveLength := baseLength + randBetween(int(waveOffsetMin), int(-waveOffsetMin))

		/// now add to stretches
		endIdx := min(j+waveLength, intervalLength)
		stretches = append(stretches, types.PixelStretch{Start: j, End: endIdx})
		j += waveLength
	}
	commonSort(stretches, seam)
}

///

/// util

// inclusive
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

func commonSort(stretches []types.PixelStretch, seam []types.PixelWithMask) {
	for stretchIdx := 0; stretchIdx < len(stretches); stretchIdx++ {
		stretch := stretches[stretchIdx]
		/// grab the pixels we want
		pixels := seam[stretch.Start:stretch.End]
		comparator := comparators.ComparatorFunctionMappings[shared.Config.Comparator]
		slices.SortStableFunc(pixels, comparator)

		if shared.Config.Reverse {
			/// do a flip!
			for i, j := 0, (len(pixels) - 1); i < j; i, j = i+1, j-1 {
				pixels[i], pixels[j] = pixels[j], pixels[i]
			}
		}
	}
}

// select all pixels not masked off
func getUnmaskedStretches(seam []types.PixelWithMask) []types.PixelStretch {
	stretches := make([]types.PixelStretch, 0)
	baseIdx := 0
	intervalLen := len(seam)

	for j := 0; j < intervalLen; j++ {
		pixel := seam[j]
		/// if masked off, or nil
		if pixel.Mask == 255 || (pixel.R == 0 && pixel.G == 0 && pixel.B == 0 && pixel.A == 0) {
			/// look ahead for the end of the mask
			endMaskIdx := j
			for {
				/// increment furst: we wanna look at the next pixel
				endMaskIdx++
				/// off the edge
				if endMaskIdx == intervalLen {
					break
				}
				nextPixel := seam[endMaskIdx]
				/// if its not masked or nil, exit
				/// this is the start index of the next stretch
				if nextPixel.Mask != 255 && !(nextPixel.R == 0 && nextPixel.G == 0 && nextPixel.B == 0 && nextPixel.A == 0) {
					break
				}
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
