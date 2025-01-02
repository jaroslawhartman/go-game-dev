// Copyright 2018 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 320
	screenHeight = 240
	boardWidth   = screenWidth / boxSize
	boardHeight  = screenHeight / boxSize
	boxSize      = 8
	speed        = math.MaxUint8 / 10
)

type Point struct {
	x int
	y int
}

type Game struct {
	snake     []*Point
	food      *Point
	offscreen *ebiten.Image
	dirX      int
	dirY      int
	color     uint8
}

func (g *Game) handleKeyboard() {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		g.dirX = 0
		g.dirY = -1
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		g.dirX = 1
		g.dirY = 0
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		g.dirX = 0
		g.dirY = 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		g.dirX = -1
		g.dirY = 0
	}
}

func (p *Point) String() string {
	return fmt.Sprintf("[%d,%d]", p.x, p.y)
}

func (g *Game) Update() error {
	g.handleKeyboard()

	// new segment will be the last snake's tail
	tail := &Point{
		g.snake[len(g.snake)-1].x,
		g.snake[len(g.snake)-1].y,
	}

	// update color (= sync)
	if uint16(g.color)+speed >= math.MaxUint8 {
		// snake
		for i, v := range slices.Backward(g.snake) {
			// head update
			if i == 0 {
				if uint16(g.color)+speed >= math.MaxUint8 {
					v.x += g.dirX
					v.y += g.dirY
				}

				if v.x == g.food.x && v.y == g.food.y {
					g.setFood()
					g.snake = append(g.snake, tail)
				}
			} else {
				v.x = g.snake[i-1].x
				v.y = g.snake[i-1].y
			}
		}
	}
	g.color += speed

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.offscreen.Clear()

	// board
	vector.StrokeRect(g.offscreen, 2, 2, screenWidth-4, screenHeight-4, 2, color.Gray{200}, true)

	// snake
	for i, v := range slices.Backward(g.snake) {
		var c color.Color

		// head update
		if i == 0 {
			c = color.Gray{g.color}
		} else {
			c = color.White
		}

		vector.DrawFilledRect(g.offscreen,
			float32(v.x*boxSize),
			float32(v.y*boxSize),
			float32(boxSize-1),
			float32(boxSize-1),
			c,
			true)
	}

	// food
	vector.DrawFilledRect(g.offscreen,
		float32(g.food.x*boxSize),
		float32(g.food.y*boxSize),
		float32(boxSize-1),
		float32(boxSize-1),
		color.RGBA{255, 0, 0, 0},
		true)

	screen.DrawImage(g.offscreen, nil)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) setFood() {
	g.food.x = rand.IntN(boardWidth)
	g.food.y = rand.IntN(boardHeight)
}

func NewGame() ebiten.Game {
	snake := Point{boardHeight / 2, boardWidth / 2}

	g := &Game{
		offscreen: ebiten.NewImage(screenWidth, screenHeight),
		snake: []*Point{
			&snake,
		},
		dirX: 1,
		dirY: 0,
		food: &Point{},
	}

	g.setFood()

	return g
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Animation (Ebitengine Demo)")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
