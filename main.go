package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"image"
	"math"
	"math/rand"
	"time"
)

type World struct {
	Window *pixelgl.Window
	RNG    *rand.Rand
	Size   int
	Speed  float64 // Lower number = faster to win.
	Step   float64
	Sum_1  float64

	Colour1 pixel.RGBA
	Colour2 pixel.RGBA

	Old    [][]float64
	New    [][]float64
	Neigh  [][]float64
	Ratio1 [][]float64
}

func (w *World) Setup() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Fight",
		Bounds: pixel.R(0, 0, 500, 500),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	w.Window = win

	// Setup Variables
	w.init()

	// Begin World Loop

	clock := time.Now()

	var (
		frames = 0
		second = time.Tick(time.Second)
	)

	win.SetTitle(fmt.Sprintf("%s | %v/%v | FPS %d", cfg.Title, 50, 50, frames))

	for !win.Closed() {
		dt := time.Since(clock).Seconds()
		clock = time.Now()

		w.ratio(dt)
		w.draw(dt)
		w.calc(dt)
		frames++

		left := math.Round((1 - w.Sum_1) * 100)
		right := math.Round(w.Sum_1 * 100)

		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | %v/%v | FPS %d", cfg.Title, left, right, frames))
			frames = 0
		default:
		}

		if w.Sum_1 == 1 || w.Sum_1 == 0 {
			w.init()
		}

		win.Update()
	}
}

func (w *World) init() {
	w.RNG = rand.New(rand.NewSource(time.Now().UnixNano()))

	w.Colour1 = pixel.RGB(w.randFloats(0, 1), w.randFloats(0, 1), w.randFloats(0, 1))
	w.Colour2 = pixel.RGB(w.randFloats(0, 1), w.randFloats(0, 1), w.randFloats(0, 1))

	w.Step = 500 / float64(w.Size)
	w.Old = make([][]float64, w.Size)
	w.New = make([][]float64, w.Size)
	w.Neigh = make([][]float64, w.Size)
	w.Ratio1 = make([][]float64, w.Size)

	for i := range w.Old {
		w.Old[i] = make([]float64, w.Size)
	}
	for i := range w.Old {
		w.New[i] = make([]float64, w.Size)
	}
	for i := range w.Old {
		w.Neigh[i] = make([]float64, w.Size)
	}
	for i := range w.Old {
		w.Ratio1[i] = make([]float64, w.Size)
	}

	for i := 0; i < w.Size; i++ {
		for j := 0; j < w.Size; j++ {
			w.Ratio1[i][j] = 0
			w.Neigh[i][j] = 8

			if i == 0 || i == w.Size-1 {
				w.Neigh[i][j] = 5
				if j == 0 || j == w.Size-1 {
					w.Neigh[i][j] = 3
				}
			}

			if j == 0 || j == w.Size-1 {
				w.Neigh[i][j] = 5
				if i == 0 || i == w.Size-1 {
					w.Neigh[i][j] = 3
				}
			}

			if i < w.Size/2 {
				w.Old[i][j] = 1
				w.Sum_1 += 1
			} else {
				w.Old[i][j] = 0
			}
			w.New[i][j] = w.Old[i][j]
		}
	}
	w.Sum_1 = w.Sum_1 / float64(w.Size*w.Size)
}

func (w *World) ratio(dt float64) {
	for i := 0; i < w.Size; i++ {
		for j := 0; j < w.Size; j++ {
			w.Ratio1[i][j] = 0
			if i > 0 {
				if j > 0 {
					w.Ratio1[i][j] += w.Old[i-1][j-1]
				}
				w.Ratio1[i][j] += w.Old[i-1][j]
				if j < w.Size-1 {
					w.Ratio1[i][j] += w.Old[i-1][j+1]
				}
			}

			if j > 0 {
				w.Ratio1[i][j] += w.Old[i][j-1]
			}
			if j < w.Size-1 {
				w.Ratio1[i][j] += w.Old[i][j+1]
			}

			if i < w.Size-1 {
				if j > 0 {
					w.Ratio1[i][j] += w.Old[i+1][j-1]
				}
				w.Ratio1[i][j] += w.Old[i+1][j]
				if j < w.Size-1 {
					w.Ratio1[i][j] += w.Old[i+1][j+1]
				}
			}

			w.Ratio1[i][j] = w.Ratio1[i][j] / w.Neigh[i][j]
		}
	}
}

func (w *World) draw(dt float64) {
	b := image.NewRGBA(image.Rect(0, 0, w.Size, w.Size))
	f := image.NewRGBA(image.Rect(0, 0, w.Size, w.Size))

	var (
		cR, cG, cB float64
		count      int
	)

	count = 0

	for x := 0; x < w.Size; x++ {
		for y := 0; y < w.Size; y++ {
			c := w.Colour1
			if w.Old[x][y] == 1 {
				c = w.Colour2
			}
			f.Set(x, y, c)

			cR += c.R
			cG += c.G
			cB += c.B
			count++
		}
	}

	cR = cR / float64(count)
	cG = cG / float64(count)
	cB = cB / float64(count)

	if cR > 255 {
		cR = 255
	}

	if cG > 255 {
		cG = 255
	}

	if cB > 255 {
		cB = 255
	}

	for x := 0; x < w.Size; x++ {
		for y := 0; y < w.Size; y++ {
			b.Set(x, y, pixel.RGBA{cR, cG, cB, 1})
		}
	}

	pB := pixel.PictureDataFromImage(b)
	pixel.NewSprite(pB, pB.Bounds()).Draw(w.Window, pixel.IM.ScaledXY(pixel.ZV, pixel.V(-4.0, 4.0)).Moved(w.Window.Bounds().Center()))

	pF := pixel.PictureDataFromImage(f)
	pixel.NewSprite(pF, pF.Bounds()).Draw(w.Window, pixel.IM.ScaledXY(pixel.ZV, pixel.V(-3.8, 3.8)).Moved(w.Window.Bounds().Center())) // 4.0 is 500 / 125
}

func (w *World) calc(dt float64) {
	for i := 0; i < w.Size; i++ {
		for j := 0; j < w.Size; j++ {
			help := w.randFloats(0, 1) - ((w.Sum_1 - 0.5) / w.Speed)

			if (w.Ratio1[i][j]) > help {
				w.Old[i][j] = 1
			} else {
				w.Old[i][j] = 0
			}
		}
	}

	w.Sum_1 = 0
	for i := 0; i < w.Size; i++ {
		for j := 0; j < w.Size; j++ {
			if w.Old[i][j] == 1 {
				w.Sum_1 += 1
			}
		}
	}
	w.Sum_1 = w.Sum_1 / float64(w.Size*w.Size)
}

func (w *World) randFloats(min, max float64) float64 {
	return min + w.RNG.Float64()*(max-min)
}

func main() {
	world := World{Size: 125, Speed: 100}
	pixelgl.Run(world.Setup)
}
