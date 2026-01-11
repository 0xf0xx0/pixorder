package shared

/// global var to hold config instead of passing it everywhere
import (
	"git.0xf0xx0.eth.limo/0xf0xx0/pixorder/types"
)

var Config struct {
	// loading pattern
	Pattern string
	// function used to split seams
	Interval string
	// pixel comparison function
	Comparator string
	// used by some interval functions
	SectionLength int
	// used by some interval functions
	Randomness float32
	// flip sorted seam
	Reverse bool
	// pixels outside of these arent sorted
	Thresholds types.ThresholdConfig
	// rotate image
	Angle float64
}
