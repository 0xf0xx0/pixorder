package patterns

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"pixelsort_go/types"
)

var Loader = map[string]func(img *image.RGBA, mask *image.Gray) (*[][]types.PixelWithMask, any){
	/// theres a better way, right? please tell me im dumb
	"rowload":    LoadRow,
	"spiralload": LoadSpiral,
	"seamload":   LoadSeamCarving,
}
var Saver = map[string]func(outputImg *image.RGBA, rows *[][]types.PixelWithMask, dims image.Rectangle, data ...any) *image.RGBA{
	"rowsave":    SaveRow,
	"spiralsave": SaveSpiral,
	"seamsave":   SaveSeamCarving,
}

func LoadRow(img *image.RGBA, mask *image.Gray) (*[][]types.PixelWithMask, any) {
	dims := img.Bounds().Max
	/// split image into rows
	rows := make([][]types.PixelWithMask, dims.Y)
	for y := 0; y < dims.Y; y++ {
		row := make([]types.PixelWithMask, dims.X)

		for x := 0; x < dims.X; x++ {
			pixel := img.RGBAAt(x, y)
			masked := mask.GrayAt(x, y).Y
			wrapped := types.PixelWithMask{R: pixel.R, G: pixel.G, B: pixel.B, A: pixel.A, Mask: masked}
			row[x] = wrapped
		}
		rows[y] = row
	}
	return &rows, nil
}
func SaveRow(outputImg *image.RGBA, rows *[][]types.PixelWithMask, dims image.Rectangle, _ ...any) *image.RGBA {
	for i := 0; i < len(*rows); i++ {
		row := (*rows)[i]
		for j := 0; j < len(row); j++ {
			currPixWithMask := row[j]
			pixel := color.RGBA{currPixWithMask.R, currPixWithMask.G, currPixWithMask.B, currPixWithMask.A}
			outputImg.SetRGBA(j, i, pixel)
		}
	}
	return outputImg
}

// / based on https://github.com/jeffThompson/PixelSorting/blob/master/SpiralSortPixels/SpiralSortPixels.pde
// / prayge, i'm not a mathy fomx
// / lots of help from fren fixing it
func LoadSpiral(img *image.RGBA, mask *image.Gray) (*[][]types.PixelWithMask, any) {
	dims := img.Bounds().Max
	width := dims.X
	height := dims.Y

	seams := make([][]types.PixelWithMask, 0)

	max := int(math.Min(float64(height), float64(width))) / 2
	for offset := 0; offset <= max; offset++ {
		seam := make([]types.PixelWithMask, 0)

		top := offset
		bottom := height - offset - 1
		left := offset
		right := width - offset - 1

		/// right
		for x := left; x <= right; x++ {
			pixel := img.RGBAAt(x, top)
			maskVal := mask.GrayAt(x, top).Y
			seam = append(seam, types.PixelWithMaskFromColor(pixel, maskVal))
		}
		/// down
		for y := top + 1; y <= bottom; y++ {
			pixel := img.RGBAAt(right, y)
			maskVal := mask.GrayAt(right, y).Y
			seam = append(seam, types.PixelWithMaskFromColor(pixel, maskVal))
		}
		/// left
		for x := right - 1; x > left; x-- {
			pixel := img.RGBAAt(x, bottom)
			maskVal := mask.GrayAt(x, bottom).Y
			seam = append(seam, types.PixelWithMaskFromColor(pixel, maskVal))
		}
		/// up
		for y := bottom; y > top; y-- {
			pixel := img.RGBAAt(left, y)
			maskVal := mask.GrayAt(left, y).Y
			seam = append(seam, types.PixelWithMaskFromColor(pixel, maskVal))
		}

		seams = append(seams, seam)
	}

	return &seams, nil
}
func SaveSpiral(outputImg *image.RGBA, seams *[][]types.PixelWithMask, dims image.Rectangle, _ ...any) *image.RGBA {

	width := dims.Max.X
	height := dims.Max.Y

	for offset, seam := range *seams {
		top := offset
		bottom := height - offset - 1
		left := offset
		right := width - offset - 1
		currPixIdx := 0

		/// right
		for x := left; x <= right; x++ {
			outputImg.Set(x, top, seam[currPixIdx].ToColor())
			currPixIdx++
		}
		/// down
		for y := top + 1; y <= bottom; y++ {
			outputImg.Set(right, y, seam[currPixIdx].ToColor())
			currPixIdx++
		}
		/// left
		for x := right - 1; x > left; x-- {
			outputImg.Set(x, bottom, seam[currPixIdx].ToColor())
			currPixIdx++
		}
		/// up
		for y := bottom; y > top; y-- {
			outputImg.Set(left, y, seam[currPixIdx].ToColor())
			currPixIdx++
		}
	}

	return outputImg
}
// https://github.com/jeffThompson/PixelSorting/tree/master/SortThroughSeamCarving/SortThroughSeamCarving
func LoadSeamCarving(img *image.RGBA, mask *image.Gray) (*[][]types.PixelWithMask, any) {
	dims := img.Bounds()

	/// grayscale
	grayed := image.NewGray(dims)
	draw.Draw(grayed, grayed.Bounds(), img.SubImage(dims), dims.Min, draw.Src)

	runKernels(*grayed)
	sums := getSums(*grayed, grayed.Rect.Max)

	width := img.Rect.Dx()
	height := img.Rect.Dy()
	byteCount := len(img.Pix)
	bottomIndex := width / 2

	path := make([]int, height)
	path = findPath(bottomIndex, sums, path, grayed.Rect.Max)

	seams := make([][]types.PixelWithMask, width+1)
	for bi := 0; bi <= width; bi++ {
		pathLen := len(path)
		seam := make([]types.PixelWithMask, pathLen)
		/// populate path with original pixels
		for i := 0; i < pathLen; i++ {
			index := (i*width + path[i] + bi) * 4
			if index >= byteCount {
				/// :C
				continue
			}
			rawPix := img.Pix[index : index+4]
			seam[i] = types.PixelWithMask{
				R:    rawPix[0],
				G:    rawPix[1],
				B:    rawPix[2],
				A:    rawPix[3],
				Mask: mask.Pix[index/4],
			}
		}
		seams[bi] = seam
	}
	return &seams, path
}
func SaveSeamCarving(outputImg *image.RGBA, seams *[][]types.PixelWithMask, dims image.Rectangle, data ...any) *image.RGBA {
	path := data[0].([]int)
	width := dims.Max.X
	byteCount := len(outputImg.Pix)

	for bi, seam := range (*seams) {
		seamLen := len(seam)
		/// write out
		for i := 0; i < seamLen; i++ {
			index := (i*width + path[i] + bi) * 4
			if index >= byteCount {
				/// :C
				break
			}
			sortedPix := seam[i]
			outputImg.Pix[index] = sortedPix.R
			outputImg.Pix[index+1] = sortedPix.G
			outputImg.Pix[index+2] = sortedPix.B
			outputImg.Pix[index+3] = sortedPix.A
		}
	}
	return outputImg
}
func unrollImage(img image.Image) []color.Gray {
	dims := img.Bounds().Max
	pixels := make([]color.Gray, dims.X*dims.Y)
	for y := 0; y < dims.Y; y++ {
		for x := 0; x < dims.X; x++ {
			pixel := img.At(x, y)
			pixels[y*dims.X+x] = pixel.(color.Gray)
		}
	}
	return pixels
}
func runKernels(img image.Gray) {
	/// kernels are black magic
	vertKernel := [][]int8{
		{-1, 0, 1},
		{-1, 0, 1},
		{-1, 0, 1},
	}
	horizKernel := [][]int8{
		{1, 1, 1},
		{0, 0, 0},
		{-1, -1, -1},
	}

	/// split image
	vImg := unrollImage(&img)
	hImg := unrollImage(&img)

	/// edge detect
	dims := img.Bounds()
	width := dims.Max.X
	height := dims.Max.Y
	totalLen := width * height
	/// horiz
	for y := 1; y < height; y++ {
		for x := 1; x < width; x++ {
			sum := 0
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pos := min((y+ky)*width+(x+kx), totalLen-1)
					val := img.Pix[pos]
					sum += int(horizKernel[ky+1][kx+1]) * int(val)
				}
			}
			hImg[y*width+x] = color.Gray{Y: uint8(sum)}
		}
	}
	/// then vert
	for y := 1; y < height; y++ {
		for x := 1; x < width; x++ {
			sum := 0
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pos := min((y+ky)*width+(x+kx), totalLen-1)
					val := img.Pix[pos]
					sum += int(vertKernel[ky+1][kx+1]) * int(val)
				}
			}
			vImg[y*width+x] = color.Gray{Y: uint8(sum)}
		}
	}
	/// merge
	for y := 1; y < height; y++ {
		for x := 1; x < width; x++ {
			index := y*width + x
			hPixel := hImg[index]
			vPixel := vImg[index]
			img.Set(x, y, color.Gray{Y: hPixel.Y + vPixel.Y})
		}
	}
}
func getSums(img image.Gray, dims image.Point) [][]float32 {
	width := dims.X
	height := dims.Y
	sums := make([][]float32, height)
	sumRows := make([]float32, width*height)
	for i := 0; i < dims.Y; i++ {
		sums[i] = sumRows[i*width : (i+1)*width]
	}

	// read furst row
	for x := 0; x < width; x++ {
		sums[0][x] = float32(img.Pix[x])
	}

	for y := 1; y < height; y++ {
		for x := 1; x < width-1; x++ {

			currentPx := float32(img.Pix[y*width+x])

			// test above L,C, and R sums
			sumL := sums[y-1][x-1] + currentPx
			sumC := sums[y-1][x] + currentPx
			sumR := sums[y-1][x+1] + currentPx
			if sumL < sumC && sumL < sumR {
				sums[y][x] = sumL
			} else if sumC < sumL && sumC < sumR {
				sums[y][x] = sumC
			} else {
				sums[y][x] = sumR
			}
		}
	}
	return sums
}
func findPath(bottomIndex int, sums [][]float32, path []int, dims image.Point) []int {
	currIndex := bottomIndex
	width := dims.X
	height := dims.Y
	for i := height - 1; i > 0; i -= 1 {
		if currIndex-1 <= 0 {
			path[i] = 0
		} else if currIndex+1 >= width {
			path[i] = width
		} else {
			upL := sums[i-1][currIndex-1]
			upC := sums[i-1][currIndex]
			upR := sums[i-1][currIndex+1]

			if upL < upC && upL < upR {
				currIndex += -1
			} else if upR < upC && upR < upL {
				currIndex += 1
			}

			path[i] = currIndex
		}
	}
	return path
}
