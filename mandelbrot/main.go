package main // import "ale-re.net/bors/mandelbrot"

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"math/rand"
	"net/http"
	"net/http/cgi"
	"os"
	"time"
)

/* from wikipedia:
The Mandelbrot set is the set of complex numbers c for which the function
	f_c(z)=z^2+c
does not diverge when iterated from z=0, i.e., for which the sequence
	f_c(0), f_c(f_c(0)), etc.,
remains bounded in absolute value.


*/

type Funge func(z, c complex128) complex128
type Pixel struct {
	x, y int
	col  byte
}

type Job struct {
	x, y int
}

func sqabs(c complex128) float64 {
	return real(c)*real(c) + imag(c)*imag(c)
}

func absgt(c complex128, x float64) bool {
	return sqabs(c) > x*x
}

func fc(z, c complex128) complex128 {
	return z*z + c
}

type ZoomPoint struct {
	cx, cy, r float64
	depth     uint
}

var interestingPoints = []ZoomPoint{
	ZoomPoint{0, 0, 2, 50},
	// Seahorse valley
	ZoomPoint{-0.747, 0.1, 0.001, 800},
	ZoomPoint{-0.747, 0.1, 0.0001, 1000},
	// Triple spiral
	ZoomPoint{-0.088, 0.654, 0.005, 500},
	ZoomPoint{-0.088, 0.656, 0.001, 800},
	ZoomPoint{-0.088, 0.656, 0.0001, 1000},
	// Quad spiral
	ZoomPoint{0.274, 0.482, 0.005, 500},
	// Double scepter
	ZoomPoint{-0.1002, 0.836, 0.0001, 1000},
	ZoomPoint{-0.100, 0.836, 0.001, 750},
	ZoomPoint{-0.1002, 0.8383, 0.1, 500},
	// Scepter valley
	ZoomPoint{-1.36, 0.005, 0.1, 500},
	ZoomPoint{-1.3683, 0.005, 0.0001, 2000},
	ZoomPoint{-1.3685, 0.005, 0.0005, 2000},
}

var zp ZoomPoint

const (
	win_w, win_h = 1024, 768

	ratio  = float64(win_w) / float64(win_h)
	iratio = float64(win_h) / float64(win_w)
)

func normalize(min, v, max int) float64 {
	delta := float64(max - min)
	return float64(v) / delta
}

func setPixel(x, y int, img image.Image, col color.Color) {
	img.(*image.RGBA).Set(x, y, col)
}

func diverges(f Funge, c complex128, iters uint) uint {
	for i, z := uint(1), 0i; i <= iters; i++ {
		if absgt(z, 2) {
			return i // It is diverging, tell how much it took
		}
		z = f(z, c) // Apply again
	}
	return 0 // We did it
}

func divergez(f Funge, c complex128, iters uint) float64 {
	for i, z := uint(1), 0i; i <= iters; i++ {
		if absgt(z, 2) {
			z = f(z, c) // extra iteration to reduce error
			return float64(i) - math.Log(math.Log(cmplx.Abs(z)))/math.Ln2
		}
		z = f(z, c) // Apply again
	}
	return 0 // We did it
}

func orbitColor(i, j int, min_x, min_y, max_x, max_y float64) *Pixel {
	// Normalize coordinates
	nx := float64(j) / float64(win_w)
	ny := float64(i) / float64(win_h)
	// Fix ranges for non-1 ratios
	if ratio < 1 {
		ny = ny*iratio - (iratio-1)*0.5
	} else if ratio > 1 {
		nx = nx*ratio - (ratio-1)*0.5
	}

	// Coordinate in space
	x := min_x + nx*(max_x-min_x)
	y := min_y + ny*(max_y-min_y)
	// Check divergence
	if c := divergez(fc, complex(x, y), zp.depth); c > 0 {
		k := 1.0 - c/float64(zp.depth)
		b := uint8(int(k*255.0) & 0xff)
		return &Pixel{j, i, b}
	} else {
		return &Pixel{j, i, 0x00}
	}
}

func renderMandelbrot() image.Image {
	// Surface where to write
	image := image.NewRGBA(image.Rect(0, 0, win_w, win_h))
	// Extents for our search
	min_x, max_x := zp.cx-zp.r, zp.cx+zp.r
	min_y, max_y := zp.cy-zp.r, zp.cy+zp.r
	// Mandelbrot set
	for i := 0; i < win_h; i++ {
		for j := 0; j < win_w; j++ {
			p := orbitColor(i, j, min_x, min_y, max_x, max_y)
			setPixel(p.x, p.y, image, color.RGBA{p.col, p.col, p.col, 0xff})
		}
	}
	return image
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if false {
		if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Set("Content-Type", "image/png")
			header.Set("Cache-Control:", "no-store, no-cache, must-revalidate, max-age=0")
			header.Set("Cache-Control:", "post-check=0, pre-check=0")
			header.Set("Pragma:", "no-cache")

			zp = interestingPoints[rand.Intn(len(interestingPoints))]
			png.Encode(w, renderMandelbrot())
		})); err != nil {
			fmt.Println(err)
		}
	} else {
		// Save image
		file, err := os.Create("mande.png")
		if err != nil {
			panic(file)
		}
		defer file.Close()
		zp = interestingPoints[rand.Intn(len(interestingPoints))]
		png.Encode(file, renderMandelbrot())
	}
}
