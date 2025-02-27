# pixelsort_go
pixelsorter cli written in golang

rewritten from my nodejs implementation, which used [satyarth/pixelsort](https://github.com/satyarth/pixelsort/) as reference

furst time using go, might be cursed

## features
- row, spiral, and seam carving patterns
- shuffle pixels, sort in waves, random lengths, or smear instead
- sort by lightness, hue, saturation, and r/g/b
- sort with a mask
- sort images in parallel
- sort in reverse
- rotation

## wanted features
- even more patterns
- completion? im lazy

## install
### from source
```sh
git clone https://github.com/0xf0xx0/pixelsort-go.git
cd pixelsort-go
go build
go install
```

## examples/usage
**Original**
![webb picture-of-the-month 2023/10](https://cdn.esawebb.org/archives/images/screen/potm2310a.jpg) - [src: ESA/Webb, NASA & CSA, A. Adamo (Stockholm University) and the FEAST JWST team](https://esawebb.org/images/potm2310a/)

**Mask**
![b&w mask](./examples/webb-mask.jpg)

**Row sort w/ default options**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --output ./examples/webb-row.jpg`
![row sort](./examples/webb-row.jpg)

**Row sort w/ threshold**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --lower_threshold 0.3 --upper_threshold 0.6 --output ./examples/webb-row-thresh.jpg`
![row sort](./examples/webb-row-thresh.jpg)

**Row sort w/ default options & mask**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --mask ./examples/webb-mask.jpg --output ./examples/webb-row-masked.jpg`
![masked row sort](./examples/webb-row-masked.jpg)

**Spiral sort w/ default options**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --pattern spiral --output ./examples/webb-spiral.jpg`
![spiral sort](./examples/webb-spiral.jpg)

**Spiral sort using most options & mask**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --mask ./examples/webb-mask.jpg --pattern spiral --interval wave --randomness 0.4 --comparator saturation --section_length 400 --output ./examples/webb-spiral-masked.jpg`
![masked spiral sort](./examples/webb-spiral-masked.jpg)

**Seam sort w/ default options**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --pattern seam --output ./examples/webb-seam.jpg`
![seam sort](./examples/webb-seam.jpg)

**Seam sort w/ default options & mask**
`pixelsort_go --input ~/Downloads/potm2310a.jpg --mask ./examples/webb-mask.jpg --pattern seam --output ./examples/webb-seam-masked.jpg`
![masked seam sort](./examples/webb-seam-masked.jpg)

did you know webb and hubble pics are cc4?

## "benchmark"
"elapsed" is time spent actually sorting  \
src img is the "large jpg" dl from [esawebb.org](https://esawebb.org/images/potm2310a/)
```
=> time pixelsort_go -i ~/Downloads/potm2310a.jpg -o ./benchmark
Sorting 1 images with a config of {Pattern:row Interval:random Comparator:lightness SectionLength:69 Randomness:1 Reverse:false Thresholds:{Lower:0 Upper:1} Angle:0}.
Loading image 1 (/home/________/Downloads/potm2310a.jpg -> ./benchmark)...
Sorting /home/________/Downloads/potm2310a.jpg...
./benchmark elapsed: 1.903s
Writing ./benchmark...
0:02.62 real 2.54 usermode 99% 89720 Kb mem

=> time pixelsort_go -i ~/Downloads/potm2310a.jpg -m ./examples/webb-mask.jpg -o ./benchmark
Sorting 1 images with a config of {Pattern:row Interval:random Comparator:lightness SectionLength:69 Randomness:1 Reverse:false Thresholds:{Lower:0 Upper:1} Angle:0}.
Loading image 1 (/home/________/Downloads/potm2310a.jpg -> ./benchmark)...
Sorting /home/________/Downloads/potm2310a.jpg...
./benchmark elapsed: 1.068s
Writing ./benchmark...
0:02.14 real 1.99 usermode 99% 224736 Kb mem
```
