// Campaign-02 materials generator: deterministic images without
// external dependencies (stdlib only). Output goes to
// campaigns/campaign-02/materials.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
)

func noiseFill(img *image.RGBA, r *rand.Rand, x0, y0, w, h int) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			g := uint8(r.Intn(256))
			// structural stripes over the noise — living content
			if (x/7+y/11)%2 == 0 {
				g = uint8(int(g)/2 + 120)
			}
			img.Set(x, y, color.RGBA{g, g, g, 255})
		}
	}
}

func saveJPEG(img image.Image, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 92}); err != nil {
		panic(err)
	}
}

func main() {
	base := os.Args[1]
	r := rand.New(rand.NewSource(42))

	// bordered.jpg — 250×250, an even 4 px black frame, a living
	// middle.
	bordered := image.NewRGBA(image.Rect(0, 0, 250, 250))
	for y := 0; y < 250; y++ {
		for x := 0; x < 250; x++ {
			bordered.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	noiseFill(bordered, r, 4, 4, 242, 242)
	saveJPEG(bordered, filepath.Join(base, "bordered.jpg"))

	// clear.jpg — 510×350, no border: living content throughout.
	clear := image.NewRGBA(image.Rect(0, 0, 510, 350))
	noiseFill(clear, r, 0, 0, 510, 350)
	saveJPEG(clear, filepath.Join(base, "clear.jpg"))

	// pic1.png — a small image for the pressmark record.
	pic := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			pic.Set(x, y, color.RGBA{uint8(30 * x), uint8(40 * y), 200, 255})
		}
	}
	pf, err := os.Create(filepath.Join(base, "pic1.png"))
	if err != nil {
		panic(err)
	}
	defer pf.Close()
	if err := png.Encode(pf, pic); err != nil {
		panic(err)
	}
	fmt.Println("materials written to", base)
}
