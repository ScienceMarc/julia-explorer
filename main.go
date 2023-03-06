package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	width  = 1440
	height = 720
)

func main() {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	rl.InitWindow(int32(width), int32(height), "Julia Explorer")
	rl.SetTargetFPS(60)

	screenBuffer := rl.LoadTextureFromImage(rl.GenImageColor(rl.GetScreenWidth(), rl.GetScreenHeight(), rl.Yellow))
	renderer := NewRenderer(width, height)

	batches := []int{1, 2, 3, 4, 5, 6, 8, 9, 10, 12, 15, 16, 18, 20, 24, 30, 36, 40, 45, 48, 60, 72, 80, 90, 120, 144, 180, 240, 360, 720}
	batchSize := 29
	maxIters := 64
	samples := 4.0
	zoom := 1.0
	offset := complex(0, 0)
	mandelbrotMode := false

	frameTimes := make([]float64, 180)

	for !rl.WindowShouldClose() {
		mouseX := float64(rl.GetMouseX())
		mouseY := float64(rl.GetMouseY())
		mouse := complex(mapRange(mouseX, 0, width, -4, 4), mapRange(mouseY, 0, height, 2, -2))

		switch rl.GetKeyPressed() {
		case rl.KeyLeftAlt:
			samples *= 4
		case rl.KeyLeftControl:
			samples = math.Max(samples/4, 1)
		case '=':
			maxIters *= 2
		case '-':
			maxIters = int(math.Max(float64(maxIters/2), 1))
		case 'M':
			mandelbrotMode = !mandelbrotMode
		case 'R':
			batchSize = 29
			maxIters = 64
			if mandelbrotMode {
				samples = 1.0
			} else {
				samples = 4.0
			}
			zoom = 1.0
			offset = complex(0, 0)
		}

		if rl.IsKeyDown(rl.KeySpace) {
			zoom *= 1.1
		}
		if rl.IsKeyDown(rl.KeyLeftShift) {
			zoom /= 1.1
		}
		if rl.IsKeyDown(rl.KeyD) {
			offset += complex(0.1/zoom, 0)
		}
		if rl.IsKeyDown(rl.KeyA) {
			offset -= complex(0.1/zoom, 0)
		}
		if rl.IsKeyDown(rl.KeyW) {
			offset += complex(0, 0.1/zoom)
		}
		if rl.IsKeyDown(rl.KeyS) {
			offset -= complex(0, 0.1/zoom)
		}

		frameTimes = append(frameTimes[1:], 1000*float64(rl.GetFrameTime()))

		rl.BeginDrawing()
		render(&renderer, mouse, batchSize, maxIters, batches, samples, zoom, offset, mandelbrotMode)
		renderer.PlotRect(width-180, 40, 180, 100, rl.White)
		for i, ft := range frameTimes {
			x := int(mapRange(float64(i), 0, float64(len(frameTimes)), width-180, width))
			y := int(math.Max(40, mapRange(1000.0/ft, 0, 60, 40+100, 40)))
			renderer.Plot(x, y, getColor(int(1000.0/ft), 60))
		}

		rl.UpdateTexture(screenBuffer, renderer.pixels)
		rl.DrawTexture(screenBuffer, 0, 0, rl.White)

		fps := fmt.Sprintf("%.0ffps", 1.0/rl.GetFrameTime())
		rl.DrawText(fps, int32(rl.GetScreenWidth()-int(rl.MeasureText(fps, 20))), 0, 20, rl.White)

		ft := fmt.Sprint(time.Duration(time.Millisecond * time.Duration(frameTimes[len(frameTimes)-1])))
		rl.DrawText(ft, int32(rl.GetScreenWidth()-int(rl.MeasureText(ft, 20))), 20, 20, rl.White)

		rl.DrawText(fmt.Sprintf("C=%v \ncenter: %v, zoom: %fx\n%d iterations (%.0fxSSAA)", mouse, offset, zoom, maxIters, samples), 0, height-80, 20, rl.LightGray)

		rl.EndDrawing()
	}
}

func render(r *Renderer, mouse complex128, batchSize, maxIters int, batches []int, samples, zoom float64, offset complex128, mandelbrotMode bool) {
	var wg sync.WaitGroup

	wg.Add(width * height / batches[batchSize])

	//fmt.Println(mouse, struct{ X, Y float64 }{float64(rl.GetMouseX()), float64(rl.GetMouseY())})
	sqrt := math.Sqrt(samples)

	var scale = 2.0 / zoom
	for x := 0; x < width; x++ {
		for y := 0; y < height; y += batches[batchSize] {
			//TODO: Don't spawn 1440 threads
			go func(x, y int) {
				defer wg.Done()
				for i := 0; i < batches[batchSize]; i++ {
					y := y + i
					iterAvg := 0.0
					for s := 0.0; s < samples; s++ {
						xs := float64(int(s) % int(sqrt))
						ys := float64(int(s) / int(sqrt))
						z := complex(float64(x)+xs/sqrt, float64(y)+ys/sqrt)
						//z = mapComplex(z, complex(0, 0), screenRange, 3-3i, -3+3i)
						z = complex(mapRange(real(z), 0, width, -scale*2, scale*2), mapRange(imag(z), 0, height, scale, -scale))
						z += offset

						//mouse = -0.4 + 0.6i
						iters := 0
						if !mandelbrotMode {
							_, iters = inJulia(z, mouse, maxIters)
						} else {
							_, iters = inJulia(0+0i, z, maxIters)
						}
						iterAvg += float64(iters)
					}
					iterAvg /= float64(samples)
					r.Plot(x, y, getColor(int(iterAvg), maxIters))
				}
			}(x, y)
		}
	}
	wg.Wait()
}

func getColor(iters, maxIters int) rl.Color {
	if iters == maxIters-1 {
		return color.RGBA{255, 255, 180, 255}
	}
	t := float64(iters) / float64(maxIters)

	if t <= 0.5 {
		return color.RGBA{uint8(255 * t * 2), 0, 0, 255}
	}
	return color.RGBA{255, uint8(255 * (t - 0.5) * 2), 0, 255}
}

func inJulia(z, c complex128, maxIter int) (bool, int) {
	iters := 0
	for i := 0; i < maxIter; i++ {
		z = z*z + c
		if real(z)*real(z)+imag(z)*imag(z) > 4.0 {
			return false, i
		}
		iters = i
	}
	return false, iters
}

func mapRange(x, inMin, inMax, outMin, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
