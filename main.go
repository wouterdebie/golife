package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const windowWidth = 1920
const windowHeight = 1080

const cellSize = 3
const gridWidth = windowWidth / cellSize
const gridHeight = windowHeight / cellSize
const nrOfCells = gridWidth * gridHeight
const auto = true
const colorIn = true
const seed = 0.1

func main() {
	fmt.Println(fmt.Sprintf("gw: %d, gh: %d", gridWidth, gridHeight))
	rand.Seed(time.Now().UnixNano())
	pixelgl.Run(run)
}

func run() {

	cfg := pixelgl.WindowConfig{
		Title:  "Go Game of Life",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	canvas := pixelgl.NewCanvas(win.Bounds())

	var cells [nrOfCells]byte
	var nextGen [nrOfCells]byte

	var cellsPtr = &cells
	var nextGenPtr = &nextGen

	fmt.Println(fmt.Sprintf("cellsPtr %p, nextGenPtr %p", cellsPtr, nextGenPtr))

	// Precompute neighbors indices for fast lookup
	var neighborIdxs [nrOfCells][]int
	for row := 0; row < gridHeight; row++ {
		for col := 0; col < gridWidth; col++ {
			idx := row*gridWidth + col

			var n []int
			for nRow := row - 1; nRow <= row+1; nRow++ {
				for nCol := col - 1; nCol <= col+1; nCol++ {
					nIdx := nRow*gridWidth + nCol
					if (nRow >= 0 && nRow < gridHeight && nCol >= 0 && nCol < gridWidth) && (idx != nIdx) {
						n = append(n, nIdx)
					}
				}
			}
			neighborIdxs[idx] = n
		}
	}

	for row := 0; row < gridHeight; row++ {
		for col := 0; col < gridWidth; col++ {
			r := rand.Float64()
			if r < seed {
				idx := row*gridWidth + col
				*cellFromGrid(cellsPtr, idx) ^= 0x80
				updateNeighborCount(cellsPtr, &neighborIdxs, idx, 1)
			}
		}
	}

	buffer := image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))

	for !win.Closed() {
		win.Clear(color.RGBA{0, 0, 0, 255})
		canvas.SetPixels(buffer.Pix)
		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()

		if win.JustPressed(pixelgl.MouseButtonLeft) || auto {
			nextGenPtr, cellsPtr = runRules(cellsPtr, nextGenPtr, &neighborIdxs)
		}
		drawBuffer(cellsPtr, buffer)
	}
}

func drawBuffer(cellsPtr *[nrOfCells]byte, buffer *image.RGBA) {
	for row := 0; row < gridHeight; row++ {
		for col := 0; col < gridWidth; col++ {
			idx := row*gridWidth + col
			cell := *cellFromGrid(cellsPtr, idx)

			n := cell & 0x7F
			var c color.Color

			if cell&0x80 == 0x80 {
				if colorIn {
					if n == 0 {
						c = colornames.Yellow
					} else if n == 1 {
						c = colornames.Orange
					} else if n == 2 {
						c = colornames.Lightgreen
					} else if n == 3 {
						c = colornames.Green
					} else {
						c = colornames.Red
					}
				} else {
					c = colornames.White
				}

				drawCell(buffer, col, row, c)
			} else {
				drawCell(buffer, col, row, colornames.Black)
			}
		}
	}
}

func runRules(cellsPtr *[nrOfCells]byte, nextGenPtr *[nrOfCells]byte,
	neighborIdxs *[nrOfCells][]int) (*[nrOfCells]byte, *[nrOfCells]byte) {

	// Copy everything over to new array. This can be optimized by recursively update neighbors
	for row := 0; row < gridHeight; row++ {
		for col := 0; col < gridWidth; col++ {
			idx := row*gridWidth + col
			cell := cellFromGrid(cellsPtr, idx)
			nextCell := cellFromGrid(nextGenPtr, idx)
			*nextCell = *cell
		}
	}

	for row := 0; row < gridHeight; row++ {
		for col := 0; col < gridWidth; col++ {
			idx := row*gridWidth + col
			runRulesForCell(cellsPtr, nextGenPtr, neighborIdxs, idx)
		}
	}
	return cellsPtr, nextGenPtr
}

func runRulesForCell(cellsPtr *[nrOfCells]byte, nextGenPtr *[nrOfCells]byte,
	neighborIdxs *[nrOfCells][]int, idx int) {
	cell := cellFromGrid(cellsPtr, idx)
	nextCell := cellFromGrid(nextGenPtr, idx)

	active := (*cell&0x80 == 0x80)
	neighbors := *cell & 0x7F

	if !active && neighbors == 3 {
		*nextCell |= (1 << 7)
		updateNeighborCount(nextGenPtr, neighborIdxs, idx, 1)
	} else if active && (neighbors == 3 || neighbors == 2) {
		// cell stays active
	} else if active {
		*nextCell &^= (1 << 7)
		updateNeighborCount(nextGenPtr, neighborIdxs, idx, -1)
	}
}

func updateNeighborCount(c *[nrOfCells]byte, n *[nrOfCells][]int, idx, amount int) {
	for _, i := range n[idx] {
		if amount > 0 {
			c[i]++
		} else {
			c[i]--
		}
	}
}

func cellFromGrid(c *[nrOfCells]byte, idx int) *byte {
	return &c[idx]
}

func drawCell(buffer *image.RGBA, col, row int, color color.Color) {
	startX := col * cellSize
	startY := row * cellSize

	for x := startX; x < startX+cellSize; x++ {
		for y := startY; y < startY+cellSize; y++ {
			buffer.Set(x, y, color)
		}
	}

}
