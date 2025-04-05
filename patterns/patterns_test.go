package patterns_test
/// i need to expand this to the whole repo
/// ...lazyyyyyy
import (
	//"fmt"
	"crypto/rand"
	"testing"

	"image"
	"image/color"
	"pixorder/patterns"
	// "pixorder/types"
)

func TestArdenLoadSpiral(t *testing.T) {
	// Constants are not to be modified; test expects 3 and 3
	WIDTH := 3
	HEIGHT := 3
	input := image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT))
	mask := image.NewGray(image.Rect(0, 0, WIDTH, HEIGHT))
	// Fill input with random pixels
	n, err := rand.Read(input.Pix)
	if n != WIDTH*HEIGHT*4 {
		t.Errorf("Read %d bytes, expected %d", n, WIDTH*HEIGHT*4)
	} else if err != nil {
		t.Errorf("Error reading: %s", err)
	}
	// Algorithm reads top then bottom from left to right, then right then left from top to bottom
	// Each spiral is a separate slice
	expected := [][]color.RGBA{
		{
			input.RGBAAt(0, 0), input.RGBAAt(1, 0), input.RGBAAt(2, 0),
			input.RGBAAt(2, 1),
			input.RGBAAt(2, 2), input.RGBAAt(1, 2), input.RGBAAt(0, 2),
			input.RGBAAt(0, 1),
		},
		{
			input.RGBAAt(1, 1),
		},
	}
	actual, _ := patterns.LoadSpiral(input, mask)
	t.Logf("input: %v", input.Pix)
	t.Logf("expected: %v", expected)
	t.Logf("actual: %v", *actual)
	// Compare equality of each element in each slice
	for slice := 0; slice < len(expected); slice++ {
		for pixel := 0; pixel < len(expected[slice]); pixel++ {
			colorExpected := expected[slice][pixel]
			colorActual := (*actual)[slice][pixel].ToColor()
			if colorActual != colorExpected {
				t.Errorf("Pixel %d of slice %d is inequal. Expected %v, got %v", pixel, slice, colorExpected, colorActual)
			}
		}
	}

	/// compare input and output (should be equal)
	res := patterns.SaveSpiral(image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT)), actual, input.Rect)
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			inPix := input.At(x, y)
			outPix := res.At(x, y)
			if inPix != outPix {
				t.Errorf("pixel (%d,%d) differs from input:\nexpected: %v\nactual:  %v", x,y,inPix,outPix)
			}
		}
	}
}
