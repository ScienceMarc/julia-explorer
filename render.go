//Optimized rendering with raylib
//TODO: Add support for textured surfaces

package main

import (
	"math"
	"runtime"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Renderer struct {
	W      int
	H      int
	pixels []rl.Color
	images []rl.Image
}

func (r *Renderer) Plot(x, y int, color rl.Color) {
	if x < 0 || y < 0 || x >= r.W || y >= r.H {
		return
	}
	idx := x + y*r.W
	if idx < len(r.pixels) && idx >= 0 {
		r.pixels[idx] = color
	}
}

func (r *Renderer) PlotSlice(x, y, length int, color rl.Color) {
	if x < 0 || y < 0 || x >= r.W || y >= r.H {
		return
	}
	idx := x + y*r.W
	sub := r.pixels[idx : idx+length]
	for i := range sub {
		sub[i] = color
	}
}

// Bresenham's line algorithm
func (r *Renderer) PlotLine(x1, y1, x2, y2 int, color rl.Color) {
	lineHigh := func(x1, y1, x2, y2 int, color rl.Color, r *Renderer) {
		dx := x2 - x1
		dy := y2 - y1
		xi := 1
		if dx < 0 {
			xi = -1
			dx = -dx
		}
		D := (2 * dx) - dy
		x := x1

		for y := y1; y <= y2; y++ {
			r.Plot(x, y, color)
			if D > 0 {
				x = x + xi
				D = D + (2 * (dx - dy))
			} else {
				D = D + 2*dx
			}
		}
	}
	lineLow := func(x1, y1, x2, y2 int, color rl.Color, r *Renderer) {
		dx := x2 - x1
		dy := y2 - y1
		yi := 1
		if dy < 0 {
			yi = -1
			dy = -dy
		}
		D := (2 * dy) - dx
		y := y1

		for x := x1; x < x2; x++ {
			r.Plot(x, y, color)
			if D > 0 {
				y += yi
				D += 2 * (dy - dx)
			} else {
				D += 2 * dy
			}
		}
	}
	if x1 == x2 { //Vertical line
		if y1 > y2 {
			temp := y2
			y2 = y1
			y1 = temp
		}
		for i := y1; i < y2; i++ {
			r.Plot(x1, i, color)
		}
	} else if y1 == y2 { //Horizontal line
		if x1 > x2 {
			temp := x2
			x2 = x1
			x1 = temp
		}
		for i := x1; i < x2; i++ {
			r.Plot(i, y1, color)
		}
	} else {
		dx := x2 - x1
		if dx < 0 {
			dx = -dx
		}
		dy := y2 - y1
		if dy < 0 {
			dy = -dy
		}

		if dy < dx {
			if x1 > x2 {
				lineLow(x2, y2, x1, y1, color, r)
			} else {
				lineLow(x1, y1, x2, y2, color, r)
			}
		} else {
			if y1 > y2 {
				lineHigh(x2, y2, x1, y1, color, r)
			} else {
				lineHigh(x1, y1, x2, y2, color, r)
			}
		}
	}
}

func (r *Renderer) PlotRect(x, y, width, height int, color rl.Color) {
	//Have to subtract 1 because of how line drawing works
	width--
	height--
	r.PlotLine(x, y, x+width, y, color)
	r.PlotLine(x, y, x, y+height, color)
	r.PlotLine(x, y+height, x+width, y+height, color)
	r.PlotLine(x+width, y, x+width, y+height, color)
}

func (r *Renderer) FillRect(x, y, width, height int, color rl.Color) {
	for i := y; i < height+y; i++ {
		r.PlotSlice(x, i, width, color)
	}
}

func (r *Renderer) PlotCircle(xc, yc, radius int, color rl.Color) {
	drawCircle := func(xc, yc, x, y int) {
		r.Plot(xc+x, yc+y, color)
		r.Plot(xc-x, yc+y, color)
		r.Plot(xc+x, yc-y, color)
		r.Plot(xc-x, yc-y, color)
		r.Plot(xc+y, yc+x, color)
		r.Plot(xc-y, yc+x, color)
		r.Plot(xc+y, yc-x, color)
		r.Plot(xc-y, yc-x, color)
	}

	x, y := 0, radius
	d := 3 - 2*radius
	drawCircle(xc, yc, x, y)
	for y >= x {
		x++
		if d > 0 {
			y--
			d += 4*(x-y) + 10
		} else {
			d += 4*x + 6
		}
		drawCircle(xc, yc, x, y)
	}
}

func (r *Renderer) FillCircle(xc, yc, radius int, color rl.Color) {
	sideLength := int(float64(radius)*math.Sqrt2) / 2
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		for x := -radius; x <= -sideLength; x++ {
			for y := -sideLength; y <= sideLength; y++ {
				if x*x+y*y <= radius*radius {
					r.Plot(x+xc, y+yc, color)
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for x := sideLength; x <= radius; x++ {
			for y := -sideLength; y <= sideLength; y++ {
				if x*x+y*y <= radius*radius {
					r.Plot(x+xc, y+yc, color)
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for x := -sideLength; x <= sideLength; x++ {
			for y := -radius; y <= -sideLength; y++ {
				if x*x+y*y <= radius*radius {
					r.Plot(x+xc, y+yc, color)
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for x := -sideLength; x <= sideLength; x++ {
			for y := sideLength; y <= radius; y++ {
				if x*x+y*y <= radius*radius {
					r.Plot(x+xc, y+yc, color)
				}
			}
		}
	}()
	r.FillRect(xc-sideLength, yc-sideLength, sideLength*2, sideLength*2, color)
	wg.Wait()
}

func (r *Renderer) PlotArrow(x, y int, heading float32, color rl.Color) {
	center := rl.Vector2{X: float32(x), Y: float32(y)}
	tip := rl.Vector2{X: 0, Y: -9}
	corner1 := rl.Vector2{X: -5, Y: 6}
	corner2 := rl.Vector2{X: 5, Y: 6}
	divot := rl.Vector2{X: 0, Y: 3}

	rad := float64(heading * 0.0174533) //Convert to radians
	//Rotation matrices are magic: https://en.wikipedia.org/wiki/Rotation_matrix
	rotMat := rl.Mat2{M00: float32(math.Cos(rad)), M01: float32(-math.Sin(rad)), M10: float32(math.Sin(rad)), M11: float32(math.Cos(rad))}
	tip = rl.Mat2MultiplyVector2(rotMat, tip)
	corner1 = rl.Mat2MultiplyVector2(rotMat, corner1)
	corner2 = rl.Mat2MultiplyVector2(rotMat, corner2)
	divot = rl.Mat2MultiplyVector2(rotMat, divot)

	//Offset arrow to be centered on given coordinates
	tip = rl.Vector2Add(center, tip)
	corner1 = rl.Vector2Add(center, corner1)
	corner2 = rl.Vector2Add(center, corner2)
	divot = rl.Vector2Add(center, divot)

	r.PlotLine(int(tip.X), int(tip.Y), int(corner1.X), int(corner1.Y), color)
	r.PlotLine(int(tip.X), int(tip.Y), int(corner2.X), int(corner2.Y), color)
	r.PlotLine(int(corner1.X), int(corner1.Y), int(divot.X), int(divot.Y), color)
	r.PlotLine(int(divot.X), int(divot.Y), int(corner2.X), int(corner2.Y), color)
	r.Plot(int(center.X), int(center.Y), color)
}

func (r *Renderer) FillArrow(x, y int, heading float32, color rl.Color) {
	center := rl.Vector2{X: float32(x), Y: float32(y)}
	tip := rl.Vector2{X: 0, Y: -9}
	corner1 := rl.Vector2{X: -5, Y: 6}
	corner2 := rl.Vector2{X: 5, Y: 6}
	divot := rl.Vector2{X: 0, Y: 3}

	rad := float64(heading * 0.0174533) //Convert to radians
	//Rotation matrices are magic: https://en.wikipedia.org/wiki/Rotation_matrix
	rotMat := rl.Mat2{M00: float32(math.Cos(rad)), M01: float32(-math.Sin(rad)), M10: float32(math.Sin(rad)), M11: float32(math.Cos(rad))}
	tip = rl.Mat2MultiplyVector2(rotMat, tip)
	corner1 = rl.Mat2MultiplyVector2(rotMat, corner1)
	corner2 = rl.Mat2MultiplyVector2(rotMat, corner2)
	divot = rl.Mat2MultiplyVector2(rotMat, divot)

	//Offset arrow to be centered on given coordinates
	tip = rl.Vector2Add(center, tip)
	corner1 = rl.Vector2Add(center, corner1)
	corner2 = rl.Vector2Add(center, corner2)
	divot = rl.Vector2Add(center, divot)

	r.FillTrig(int(tip.X), int(tip.Y), int(corner1.X), int(corner1.Y), int(divot.X), int(divot.Y), color)
	r.FillTrig(int(tip.X), int(tip.Y), int(corner2.X), int(corner2.Y), int(divot.X), int(divot.Y), color)
	r.PlotLine(int(tip.X), int(tip.Y), int(divot.X), int(divot.Y), color)
}

func (r *Renderer) PlotTrig(x1, y1, x2, y2, x3, y3 int, color rl.Color) {
	r.PlotLine(x1, y1, x2, y2, color)
	r.PlotLine(x1, y1, x3, y3, color)
	r.PlotLine(x2, y2, x3, y3, color)
}

func (r *Renderer) FillTrig(x1, y1, x2, y2, x3, y3 int, color rl.Color) {
	//Most efficient way to find the bounding box I could think of
	xMin, xMax, yMin, yMax := x1, x1, y1, y1
	if x2 < xMin {
		xMin = x2
	}
	if x3 < xMin {
		xMin = x3
	}

	if x2 > xMax {
		xMax = x2
	}
	if x3 > xMax {
		xMax = x3
	}

	if y2 < yMin {
		yMin = y2
	}
	if y3 < yMin {
		yMin = y3
	}

	if y2 > yMax {
		yMax = y2
	}
	if y3 > yMax {
		yMax = y3
	}

	//Check if winding is correct
	isClockwise := ((y2-y1)*(x3-x2) - (x2-x1)*(y3-y2)) > 0
	if isClockwise {
		temp1, temp2 := x1, y1
		x1, y1 = x2, y2
		x2, y2 = temp1, temp2
	}
	lineTest := func(v0x, v0y, v1x, v1y, px, py int) bool {
		return (v1x-v0x)*(py-v0y)-(v1y-v0y)*(px-v0x) > 0
	}
	inTrig := func(px, py int) bool {
		return lineTest(x1, y1, x2, y2, px, py) && lineTest(x2, y2, x3, y3, px, py) && lineTest(x3, y3, x1, y1, px, py)
	}
	//* Maybe this could be optimized through a worker pool?
	var wg sync.WaitGroup
	coreCount := runtime.NumCPU() * 5 //TODO: Tune this multiplier to better utilize CPU
	jobSize := int(math.Ceil(float64(xMax-xMin) / float64(coreCount)))
	for i := 0; i < coreCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for px := xMin + (jobSize * (i)); px < xMin+(jobSize*(i+1)); px++ {
				for py := yMin; py < yMax; py++ {
					if inTrig(px, py) {
						r.Plot(px, py, color)
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

func (r *Renderer) PlotSprite(x, y, index int) {
	img := r.images[index]
	colors := rl.LoadImageColors(&img)
	for i := range colors {
		r.Plot(x+i%int(img.Width), y+i/int(img.Width), colors[i])
	}
}

// TODO: Perhaps make image dimensions variable
func (r *Renderer) PlotSpriteStretched(x, y, s, index int) {
	img := r.images[index]
	colors := rl.LoadImageColors(&img)
	sf := float64(256) / float64(s)
	if sf > 1 {
		for i := 0; int(float64(i)*sf) < 256; i++ {
			row := colors[int(float64(i)*sf)*256 : (int(float64(i)*sf)+1)*256]
			for j := 0; int(float64(j)*sf) < 256; j++ {
				r.Plot(x+j, y+i, row[int(float64(j)*sf)])
			}
		}
	} else if sf < 1 {
		for i := 0.0; i < 256; i++ {
			for j := 0.0; j < 256; j++ {
				r.FillRect(x+int(i/sf), y+int(j/sf), int(math.Ceil(1.0/sf)), int(math.Ceil(1.0/sf)), colors[int(i)+int(j)*256])
			}
		}
	} else {
		r.PlotSprite(x, y, index)
	}
}

func (r *Renderer) PlotSpriteSlice(x, y, w, offset, h, index int) {
	img := r.images[index]
	colors := rl.LoadImageColors(&img)
	_ = colors
	//TODO: Finish sprite slice rendering
}

func NewRenderer(w, h int) Renderer {
	r := Renderer{W: w, H: h, pixels: make([]rl.Color, w*h), images: make([]rl.Image, 0)}

	//TODO: Dynamically load needed images
	//* Maybe move this to level.go?
	////r.images = append(r.images, *rl.LoadImage("test.jpg"))
	r.FillRect(0, 0, r.W, r.H, rl.White)
	return r
}
