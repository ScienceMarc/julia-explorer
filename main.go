package main

import (
	"fmt"
	"image/color"
	"math"
	"runtime"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	width  = 1440
	height = 720
)

func main() {
	runtime.LockOSThread()
	/*
		f, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	*/

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(width), int32(height), "Julia Explorer")
	rl.SetTargetFPS(60)

	screenBuffer := rl.LoadTextureFromImage(rl.GenImageColor(width, height, rl.Yellow))
	renderer := NewRenderer(width, height)
	var wg sync.WaitGroup
	jobs := make(chan Job, width)

	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(&wg, &renderer, jobs)
	}

	batches := []int{1, 2, 3, 4, 5, 6, 8, 9, 10, 12, 15, 16, 18, 20, 24, 30, 32, 36, 40, 45, 48, 60, 72, 80, 90, 96, 120, 144, 160, 180, 240, 288, 360, 480, 720, 1440}
	batchIdx := 0
	maxIters := 64
	samples := 4.0
	zoom := 1.0
	offset := complex(0, 0)
	mandelbrotMode := false

	frameTimes := make([]float64, 180)

	update := true
	oldMouse := complex(0, 0)
	updateFrame := 0
	frameCount := 0

	rendering := false
	needToAA := false
	AAbeginTime := time.Now()
	timeTaken := time.Since(AAbeginTime)

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			width = rl.GetScreenWidth()
			height = rl.GetScreenHeight()
			screenBuffer = rl.LoadTextureFromImage(rl.GenImageColor(width, height, rl.Yellow))
			renderer = NewRenderer(width, height)
			batches = []int{}
			batchIdx = 0
			for i := 1; i <= width; i++ {
				if width%i == 0 {
					batches = append(batches, i)
				}
			}
			update = true
		}

		mouseX := float64(rl.GetMouseX())
		mouseY := float64(rl.GetMouseY())
		mouse := complex(mapRange(mouseX, 0, float64(width), -4, 4), mapRange(mouseY, 0, float64(height), 4*float64(height)/float64(width), -2*float64(height)/float64(width)))

		if mouse != oldMouse && !mandelbrotMode {
			update = true
		}
		oldMouse = mouse

		key := rl.GetKeyPressed()

		switch key {
		case 'X':
			samples *= 4
			needToAA = true
		case 'Z':
			samples = math.Max(samples/4, 1)
			needToAA = true
		case '=':
			maxIters *= 2
			update = true
		case '-':
			maxIters = int(math.Max(float64(maxIters/2), 1))
			update = true
		case 'M':
			mandelbrotMode = !mandelbrotMode
			update = true
		case 'R':
			//batchIdx = 0
			maxIters = 64
			zoom = 1.0
			offset = complex(0, 0)
			update = true
		case 'O':
			batchIdx = (batchIdx + 1) % len(batches)
			fmt.Println(batchIdx, batches[batchIdx])
		case 'P':
			batchIdx--
			if batchIdx < 0 {
				batchIdx = len(batches) - 1
			}
			fmt.Println(batchIdx, batches[batchIdx])
		case 0:

		default:
			update = true
		}

		if rl.IsKeyDown(rl.KeySpace) {
			zoom *= 1.1
			update = true
		}
		if rl.IsKeyDown(rl.KeyLeftShift) {
			zoom /= 1.1
			update = true
		}
		if rl.IsKeyDown(rl.KeyD) {
			offset += complex(0.1/zoom, 0)
			update = true
		}
		if rl.IsKeyDown(rl.KeyA) {
			offset -= complex(0.1/zoom, 0)
			update = true
		}
		if rl.IsKeyDown(rl.KeyW) {
			offset += complex(0, 0.1/zoom)
			update = true
		}
		if rl.IsKeyDown(rl.KeyS) {
			offset -= complex(0, 0.1/zoom)
			update = true
		}

		frameTimes = append(frameTimes[1:], 1000*float64(rl.GetFrameTime()))

		rl.BeginDrawing()
		if update {
			//Dump jobs
		L:
			for {
				select {
				case <-jobs:
					wg.Done()
				default:
					break L
				}
			}
			render(&renderer, mouse, batches[batchIdx], maxIters, 1.0, zoom, offset, mandelbrotMode, &wg, jobs)
			updateFrame = frameCount
			needToAA = true
		} else if needToAA && !rendering && updateFrame+60 <= frameCount {
			rendering = true
			needToAA = false
			AAbeginTime = time.Now()
			go func() {
				render(&renderer, mouse, batches[batchIdx], maxIters, samples, zoom, offset, mandelbrotMode, &wg, jobs)
				rendering = false
			}()
		}
		/*
			renderer.FillRect(width-180, 40, 180, 100, rl.DarkGray)
			renderer.PlotRect(width-180, 40, 180, 100, rl.White)
			for i, ft := range frameTimes {
				x := int(mapRange(float64(i), 0, float64(len(frameTimes)), float64(width)-180, float64(width)))
				y := int(math.Max(40, mapRange(1000.0/ft, 0, 60, 40+100, 40)))
				renderer.Plot(x, y, getColor(int(1000.0/ft), 60))
			}*/

		rl.UpdateTexture(screenBuffer, renderer.pixels)
		rl.DrawTexture(screenBuffer, 0, 0, rl.White)

		fps := fmt.Sprintf("%.0ffps", 1.0/rl.GetFrameTime())
		rl.DrawText(fps, int32(rl.GetScreenWidth()-int(rl.MeasureText(fps, 20))), 0, 20, rl.LightGray)

		ft := fmt.Sprint(time.Duration(time.Millisecond * time.Duration(frameTimes[len(frameTimes)-1])))
		rl.DrawText(ft, int32(rl.GetScreenWidth()-int(rl.MeasureText(ft, 20))), 20, 20, rl.LightGray)

		rText := fmt.Sprintf("(%v)", timeTaken)
		if rendering {
			timeTaken = time.Since(AAbeginTime).Round(time.Millisecond)
			rText = "!RENDERING! " + rText
		}
		rl.DrawText(fmt.Sprintf("C=%v \ncenter: %v, zoom: %fx\n%d iterations (%.0fxSSAA) [%d jobs] %s", mouse, offset, zoom, maxIters, samples, width/batches[batchIdx], rText), 0, int32(height)-80, 20, rl.LightGray)

		rl.EndDrawing()
		update = false
		frameCount++
	}
}

func worker(wg *sync.WaitGroup, r *Renderer, jobs <-chan Job) {
	for {
		job := <-jobs
		for i := 0; i < job.BatchSize; i++ {
			x := job.X + i
			for y := 0; y < height; y++ {
				iterAvg := 0.0
				for s := 0.0; s < job.Samples; s++ {
					xs := float64(int(s) % int(job.Sqrt))
					ys := float64(int(s) / int(job.Sqrt))
					z := complex(float64(x)+xs/job.Sqrt, float64(y)+ys/job.Sqrt)
					//z = mapComplex(z, complex(0, 0), screenRange, 3-3i, -3+3i)
					z = complex(mapRange(real(z), 0, float64(width), -job.Scale, job.Scale), mapRange(imag(z), 0, float64(height), job.Scale*float64(height)/float64(width), -job.Scale*float64(height)/float64(width)))
					z += job.Offset

					//mouse = -0.4 + 0.6i
					iters := 0
					if !job.MandelbrotMode {
						_, iters = inJulia(z, job.Mouse, job.MaxIters)
					} else {
						_, iters = inJulia(0+0i, z, job.MaxIters)
					}
					iterAvg += float64(iters)
				}
				iterAvg /= float64(job.Samples)
				r.Plot(x, y, getColor(int(iterAvg), job.MaxIters))
			}
		}
		wg.Done()
	}
}

type Job struct {
	Mouse          complex128
	Offset         complex128
	BatchSize      int
	MaxIters       int
	Samples        float64
	MandelbrotMode bool

	X     int
	Sqrt  float64
	Scale float64
}

// TODO: Potentially re-evaluate/re-order parameters
func render(r *Renderer, mouse complex128, batchSize, maxIters int, samples, zoom float64, offset complex128, mandelbrotMode bool, wg *sync.WaitGroup, jobs chan<- Job) {
	wg.Add(width / batchSize)
	//fmt.Println(mouse, struct{ X, Y float64 }{float64(rl.GetMouseX()), float64(rl.GetMouseY())})
	sqrt := math.Sqrt(samples)

	var scale = 4.0 / zoom
	for x := 0; x < width; x += batchSize {
		jobs <- Job{
			Mouse:          mouse,
			Offset:         offset,
			BatchSize:      batchSize,
			MaxIters:       maxIters,
			Samples:        samples,
			MandelbrotMode: mandelbrotMode,

			X:     x,
			Sqrt:  sqrt,
			Scale: scale,
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
