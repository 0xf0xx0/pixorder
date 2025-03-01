package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"slices"
	"strings"
	"time"

	"pixelsort_go/comparators"
	"pixelsort_go/intervals"
	"pixelsort_go/patterns"
	"pixelsort_go/shared"

	"github.com/kovidgoyal/imaging"
	"github.com/remeh/sizedwaitgroup"
	"github.com/urfave/cli/v2"
)

func main() {
	validPatterns := make([]string, 0, len(patterns.Loader))
	validIntervals := make([]string, 0, len(intervals.IntervalFunctionMappings))
	validComparators := make([]string, 0, len(comparators.ComparatorFunctionMappings))

	for k := range patterns.Loader {
		/// why -4, you ask? cause patterns always come in pairs, xLoad and xSave
		/// "load" is 4 chars, so trim them off and you get the pattern name uwu
		validPatterns = append(validPatterns, k[:len(k)-4])
	}
	for k := range intervals.IntervalFunctionMappings {
		validIntervals = append(validIntervals, k)
	}
	for k := range comparators.ComparatorFunctionMappings {
		validComparators = append(validComparators, k)
	}

	app := &cli.App{
		Name:                   "pixelsort_go",
		Usage:                  "Organize pixels.",
		Version:                "0.8.0",
		UseShortOptionHandling: true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "`image`(s) to sort, or a dir full of images (supported: png, jpg)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "`file` to output to",
			},
			&cli.StringFlag{
				Name:    "pattern",
				Aliases: []string{"p"},
				Usage:   fmt.Sprintf("`pattern` loader to use [%s]", strings.Join(validPatterns, ", ")),
				Value:   "row",
				Action: func(ctx *cli.Context, v string) error {
					if !slices.Contains(validPatterns, v) {
						return fmt.Errorf("invalid interval \"%s\" [%s]", v, strings.Join(validPatterns, ", "))
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "interval",
				Value:   "none",
				Aliases: []string{"I"},
				Usage:   fmt.Sprintf("interval `func`tion to use [%s]", strings.Join(validIntervals, ", ")),
				Action: func(ctx *cli.Context, v string) error {
					if !slices.Contains(validIntervals, v) {
						return fmt.Errorf("invalid interval \"%s\" [%s]", v, strings.Join(validIntervals, ", "))
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "comparator",
				Value:   "lightness",
				Aliases: []string{"c"},
				Usage:   fmt.Sprintf("pixel comparison `func`tion to use [%s]", strings.Join(validComparators, ", ")),
				Action: func(ctx *cli.Context, v string) error {
					if !slices.Contains(validComparators, v) {
						return fmt.Errorf("invalid comparator \"%s\" [%s]", v, strings.Join(validComparators, ", "))
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "mask",
				Aliases: []string{"m"},
				Usage:   "b&w `mask` to determine which pixels to touch; white is skipped",
			},
			&cli.Float64Flag{
				Name:    "lower_threshold",
				Value:   0.0,
				Aliases: []string{"l"},
				Usage:   "pixels below this `thresh`old won't be sorted",
				Action: func(ctx *cli.Context, v float64) error {
					if v < 0.0 || v > 1.0 {
						return fmt.Errorf("lower_threshold is outside of range [0.0-1.0]")
					}
					return nil
				},
			},
			&cli.Float64Flag{
				Name:    "upper_threshold",
				Value:   1.0,
				Aliases: []string{"u"},
				Usage:   "pixels above this `thresh`old won't be sorted",
				Action: func(ctx *cli.Context, v float64) error {
					if v < 0.0 || v > 1.0 {
						return fmt.Errorf("upper_threshold is outside of range [0.0-1.0]")
					}
					return nil
				},
			},
			&cli.Float64Flag{
				Name:    "angle",
				Value:   0.0,
				Aliases: []string{"a"},
				Usage:   "rotate the image by `deg`rees, pos or neg",
			},
			&cli.IntFlag{
				Name:    "section_length",
				Value:   69,
				Aliases: []string{"L"},
				Usage:   "The base `len`gth of each slice",
			},
			&cli.BoolFlag{
				Name:    "reverse",
				Value:   false,
				Aliases: []string{"r"},
				Usage:   "reverse the sort direction",
			},
			&cli.Float64Flag{
				Name:    "randomness",
				Value:   1,
				Aliases: []string{"R"},
				Usage:   "used to determine the percentage of [row]s to skip and how wild [wave] edges should be, among other things",
				Action: func(ctx *cli.Context, v float64) error {
					if v < 0.0 || v > 1.0 {
						return fmt.Errorf("randomness is outside of range [0.0-1.0]")
					}
					return nil
				},
			},
			&cli.IntFlag{
				Name:    "threads",
				Value:   1,
				Aliases: []string{"t"},
				Usage:   "Sort images in parallel across `N` threads",
			},
			&cli.BoolFlag{
				Name:  "profile",
				Value: false,
				Usage: "use pprof to spit out a cpu profile",
			},
		},
		Action: func(ctx *cli.Context) error {
			inputs := ctx.StringSlice("input")
			output := ctx.String("output")
			mask := ctx.String("mask")
			masks := make([]string, 0)
			shared.Config.Pattern = ctx.String("pattern")
			shared.Config.Interval = ctx.String("interval")
			shared.Config.Comparator = ctx.String("comparator")
			shared.Config.Thresholds.Lower = float32(ctx.Float64("lower_threshold"))
			shared.Config.Thresholds.Upper = float32(ctx.Float64("upper_threshold"))
			shared.Config.SectionLength = ctx.Int("section_length")
			shared.Config.Reverse = ctx.Bool("reverse")
			shared.Config.Randomness = float32(ctx.Float64("randomness"))
			shared.Config.Angle = ctx.Float64("angle")
			threadCount := ctx.Int("threads")

			/// profiling
			if ctx.Bool("profile") {
				masked := "unmasked"
				if mask != "" {
					masked = "masked"
				}
				profileFile, err := os.Create(fmt.Sprintf("cpuprofile-%s-%s-%s-%s.prof", shared.Config.Pattern, shared.Config.Interval, shared.Config.Comparator, masked))
				if err != nil {
					log.Fatal(err)
				}
				pprof.StartCPUProfile(profileFile)
				defer pprof.StopCPUProfile()
			}

			/// this can be done better but im lazy and braindead
			/// MAYBE: accept multiple dirs? pop them and append contents?
			inputLen := len(inputs)
			if inputLen == 1 {
				input := inputs[0]

				inputfile, err := os.Open(input)
				if err != nil {
					return cli.Exit(fmt.Sprintf("%s could not be opened", input), 1)
				}
				defer inputfile.Close()

				inputStat, err := inputfile.Stat()
				if err != nil {
					return cli.Exit(fmt.Sprintf("Error getting %s file stats: %s", input, err), 1)
				}
				inputfile = nil

				if inputStat.IsDir() {
					res, err := readdirForImages(input)
					if err != nil {
						return err
					}
					inputs = res
					inputLen = len(inputs)
				}
			}

			/// masking
			if mask != "" {
				maskfile, err := os.Open(mask)
				if err != nil {
					return cli.Exit(fmt.Sprintf("Mask %s could not be opened", mask), 1)
				}
				defer maskfile.Close()
				maskStat, err := maskfile.Stat()
				if err != nil {
					return cli.Exit(fmt.Sprintf("Error getting mask %s file stats: %s", mask, err), 1)
				}
				maskfile = nil
				if maskStat.IsDir() {
					res, err := readdirForImages(mask)
					if err != nil {
						return err
					}
					masks = res
				} else {
					masks = append(masks, mask)
				}
			}

			maskLen := len(masks)
			if maskLen == 0 {
				/// empty string, will be ignored by sorting
				masks = append(masks, "")
				maskLen = len(masks)
			}
			fmt.Println(fmt.Sprintf("Sorting %d images with a config of %+v.", inputLen, shared.Config))

			/// multiple imgs
			/// sort em first so frames dont get jumbled
			slices.SortFunc(inputs, func(a, b string) int {
				return strings.Compare(a, b)
			})
			slices.SortFunc(masks, func(a, b string) int {
				return strings.Compare(a, b)
			})

			/// create workgroup
			wg := sizedwaitgroup.New(+threadCount)

			for i := 0; i < inputLen; i++ {
				wg.Add()
				go func(i int) {
					defer wg.Done()

					in := inputs[i]
					out := output
					splitFileName := strings.Split(inputs[0], ".")
					fileSuffix := splitFileName[len(splitFileName)-1]

					if inputLen > 1 {
						out = fmt.Sprintf("frame%04d.%s", i, fileSuffix)
					} else if out == "" {
						out = fmt.Sprintf("%s.%s", "sorted", fileSuffix)
					}

					fmt.Println(fmt.Sprintf("Loading image %d (%s -> %s)...", i+1, in, out))
					maskIdx := min(i, maskLen-1)
					err := sortingTime(in, out, masks[maskIdx])
					if err != nil {
						cli.Exit(fmt.Sprintf("Error occured during sort of image %d (%q): %q", i+1, in, err), 1)
					}
				}(i)
			}
			wg.Wait()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func readdirForImages(input string) ([]string, error) {
	files, err := os.ReadDir(input)
	if err != nil {
		return nil, cli.Exit("Couldn't read input dir", 1)
	}
	inputs := make([]string, len(files)) /// allocate enough space
	for idx, file := range files {
		if !file.IsDir() && file.Type().IsRegular() {
			name := file.Name()
			if strings.HasSuffix(name, "jpg") || strings.HasSuffix(name, "jpeg") || strings.HasSuffix(name, "png") {
				inputs[idx] = fmt.Sprintf("%s/%s", input, name)
			}
		}
	}
	/// remove empty elms
	inputs = slices.DeleteFunc(inputs, func(s string) bool {
		return len(strings.TrimSpace(s)) == 0
	})
	return inputs, nil
}

func sortingTime(input, output, maskpath string) error {
	file, err := os.Open(input)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Input %q could not be opened", input), 1)
	}
	defer file.Close()

	rawImg, format, err := image.Decode(file)
	if err != nil {
		println(err.Error())
		// for some reason this error specficially doesnt display?
		return cli.Exit(fmt.Sprintf("Input %q could not be decoded", input), 1)
	}

	/// RO TA TE
	/// god why do i have to do thissssssswddenwfiosbduglzx er agdxbv
	/// this is used in the writing step cause `imaging` doesnt have a option to
	/// auto-crop transparency
	originalDims := rawImg.Bounds()
	if math.Mod(shared.Config.Angle, 360) != 0 {
		rawImg = (*image.RGBA)(imaging.Rotate(rawImg, shared.Config.Angle, color.Transparent))
	}

	/// convert to rgba (gray for the mask)
	b := rawImg.Bounds()
	sortingDims := image.Rect(0, 0, b.Dx(), b.Dy())
	img := image.NewRGBA(sortingDims)
	mask := image.NewGray(sortingDims)
	draw.Draw(img, img.Bounds(), rawImg, b.Min, draw.Src)
	rawImg = nil

	/// MAYBE: figure out how to load mask once if used multiple times
	if maskpath != "" {
		maskFile, err := os.Open(maskpath)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Mask %q could not be opened", maskpath), 1)
		}
		defer maskFile.Close()

		rawMask, _, err := image.Decode(maskFile)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Mask %q could not be decoded", maskpath), 1)
		}

		/// RO TA TE (again)
		if math.Mod(shared.Config.Angle, 360) != 0 {
			rawMask = imaging.Rotate(rawMask, float64(shared.Config.Angle), color.Transparent)
		}

		draw.Draw(mask, mask.Bounds(), rawMask, b.Min, draw.Src)
		rawMask = nil
	}

	/// load seams
	loader := patterns.Loader[fmt.Sprintf("%sload", shared.Config.Pattern)]
	if loader == nil {
		fmt.Println("invalid pattern")
		return cli.Exit("invalid pattern", 2)
	}
	seams, data := loader(img, mask)
	/// more whitespace
	/// im not gonna rant again
	/// just
	/// *sigh*

	println(fmt.Sprintf("Sorting %s...", input))
	/// pass the rows to the sorter
	start := time.Now()
	for _, seam := range *seams  {
		intervals.Sort(seam)
	}
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Println(output, "elapsed:", elapsed.Truncate(time.Millisecond).String())

	/// cant believe i have to do this so the stupid extension doesnt fucking trim my lines
	/// like fuck dude i just want some fucking whitespace, its not that big of a deal

	/// now write
	outputImg := patterns.Saver[fmt.Sprintf("%ssave", shared.Config.Pattern)](img, seams, img.Bounds(), data)
	// outputImg := patterns.Saver[fmt.Sprintf("%ssave", shared.Config.Pattern)](image.NewRGBA(sortingDims), stretches, img.Bounds(), data)

	/// ET AT OR
	if math.Mod(shared.Config.Angle, 360) != 0 {
		outputImg = (*image.RGBA)(imaging.Rotate(outputImg, -shared.Config.Angle, color.Transparent))
		/// gotta crop the invisible pixels
		if math.Mod(shared.Config.Angle, 90) != 0 {
			outputImg = (*image.RGBA)(imaging.CropCenter(outputImg, originalDims.Dx(), originalDims.Dy()))
		}
	}

	fmt.Println(fmt.Sprintf("Writing %s...", output))
	f, err := os.Create(output)
	if err != nil {
		return cli.Exit("Could not create output file", 1)
	}

	/// spit the result out
	if format == "jpeg" {
		jpeg.Encode(f, outputImg.SubImage(outputImg.Rect), &jpeg.Options{
			Quality: 100,
		})
	} else {
		pngcoder := png.Encoder{
			CompressionLevel: png.NoCompression,
		}
		pngcoder.Encode(f, outputImg)
	}
	return nil
}
