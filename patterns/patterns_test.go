package patterns_test

/// merge into 1 func?
import (
	//"fmt"
	"crypto/rand"
	"testing"

	"image"
	"image/color"
	"pixorder/patterns"
	"pixorder/types"
	// "pixorder/types"
)

func TestLoadRow(t *testing.T) {
	DIMS := 3
	input := genTestPic(DIMS, DIMS, t)
	mask := image.NewGray(image.Rect(0, 0, DIMS, DIMS))

	expected := [][]color.RGBA{
		{input.RGBAAt(0, 0), input.RGBAAt(1, 0), input.RGBAAt(2, 0)},
		{input.RGBAAt(0, 1), input.RGBAAt(1, 1), input.RGBAAt(2, 1)},
		{input.RGBAAt(0, 2), input.RGBAAt(1, 2), input.RGBAAt(2, 2)},
	}
	actual, _ := patterns.LoadRow(input, mask)
	compareLoadEquality(input, expected, actual, t)
}
func TestLoadSpiral(t *testing.T) {
	// Constants are not to be modified; test expects 3 and 3
	DIMS := 3
	input := genTestPic(DIMS, DIMS, t)
	mask := image.NewGray(image.Rect(0, 0, DIMS, DIMS))

	// Algorithm reads top -> left -> bottom -> right
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
	// Compare equality of each element in each slice
	compareLoadEquality(input, expected, actual, t)
}
func TestLoadSeam(t *testing.T) {

}

func TestSaves(t *testing.T) {
	DIMS := 3
	for key := range patterns.Saver {
		t.Logf("testing %s", key)
		input := genTestPic(DIMS, DIMS, t)
		mask := image.NewGray(image.Rect(0, 0, DIMS, DIMS))

		loaded, extra := patterns.Loader[key[:len(key)-4]+"load"](input, mask)
		res := patterns.Saver[key](image.NewRGBA(image.Rect(0, 0, DIMS, DIMS)), loaded, input.Rect, extra)
		compareSaveEquality(input, res, t)
	}
}

func genTestPic(w, h int, t *testing.T) *image.RGBA {
	input := image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill input with random pixels
	n, err := rand.Read(input.Pix)
	if n != w*h*4 {
		t.Errorf("Read %d bytes, expected %d", n, w*h*4)
	} else if err != nil {
		t.Errorf("Error reading: %s", err)
	}
	return input
}
func compareLoadEquality(input *image.RGBA, expected [][]color.RGBA, actual *[][]types.PixelWithMask, t *testing.T) {
	for slice := 0; slice < len(expected); slice++ {
		for pixel := 0; pixel < len(expected[slice]); pixel++ {
			colorExpected := expected[slice][pixel]
			colorActual := (*actual)[slice][pixel].ToColor()
			if colorActual != colorExpected {
				t.Logf("input: %v", input.Pix)
				t.Logf("expected: %v", expected)
				t.Logf("actual: %v", *actual)
				t.Errorf("Pixel %d of slice %d is inequal. Expected %v, got %v", pixel, slice, colorExpected, colorActual)
			}
		}
	}
}
func compareSaveEquality(input, res *image.RGBA, t *testing.T) {
	for y := 0; y < res.Rect.Dy(); y++ {
		for x := 0; x < res.Rect.Dx(); x++ {
			inPix := input.At(x, y)
			outPix := res.At(x, y)
			if inPix != outPix {
				t.Logf("expected: %v", input.Pix)
				t.Logf("actual: %v", (*res).Pix)
				t.Errorf("pixel (%d,%d) differs from input:\nexpected: %v\nactual:  %v", x, y, inPix, outPix)
			}
		}
	}
}
