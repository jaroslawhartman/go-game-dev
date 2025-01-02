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
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

const (
	RUNNING = iota
	CRASHED
	CRASHING
)

type Point struct {
	x int
	y int
}

type Game struct {
	snake     []*Point
	food      *Point
	offscreen *ebiten.Image
	direction *Point
	color     uint8
	score     int
	state     int
}

var (
	mplusFaceSource *text.GoTextFaceSource
	mplusNormalFace *text.GoTextFace
	mplusBigFace    *text.GoTextFace
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s

	mplusNormalFace = &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   24,
	}
	mplusBigFace = &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   32,
	}
}

func (g *Game) handleKeyboard() {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		g.direction.x = 0
		g.direction.y = -1
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		g.direction.x = 1
		g.direction.y = 0
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		g.direction.x = 0
		g.direction.y = 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		g.direction.x = -1
		g.direction.y = 0
	}
}

func (p *Point) String() string {
	return fmt.Sprintf("[%d,%d]", p.x, p.y)
}

func (g *Game) detectCollision(h *Point) bool {
	for i := 1; i < len(g.snake); i++ {
		p := g.snake[i]
		if p.x == h.x && p.y == h.y {
			return true
		}
	}

	return false
}

func (g *Game) Update() error {
	g.handleKeyboard()

	// new segment will be the last snake's tail
	tail := &Point{
		g.snake[len(g.snake)-1].x,
		g.snake[len(g.snake)-1].y,
	}

	if uint16(g.color)+speed >= math.MaxUint8 {
		switch g.state {
		case RUNNING:
			// update color (= sync)
			// Snake
			//
			// Iterate backward (i.e. tail -> head) as the new segment
			// position should be in the point where the predecesor (still) is
			for i, v := range slices.Backward(g.snake) {
				// head update
				if i == 0 {
					v.x += g.direction.x
					v.y += g.direction.y

					// Grabbing the food? If so:
					// - set a new peiece
					// - append the new segment where the last tail was
					if v.x == g.food.x && v.y == g.food.y {
						g.setFood()
						g.snake = append(g.snake, tail)
						g.score += 1
					}
				} else {
					v.x = g.snake[i-1].x
					v.y = g.snake[i-1].y
				}
			}

			// check for collision and reinit if needed
			if g.detectCollision(g.snake[0]) {
				g.state = CRASHED
			}
		case CRASHED:
			g.score = 0
			g.state = CRASHING

		case CRASHING:
			if len(g.snake) > 1 {
				g.snake = g.snake[0 : len(g.snake)-1]
			} else {
				g.state = RUNNING
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

		if i == 0 {
			// head update
			c = color.Gray{g.color}
		} else if i == len(g.snake)-1 && len(g.snake) > 1 {
			// last tail section
			c = color.Gray{math.MaxUint8 - g.color}
		} else {
			// middle sections
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

	// score

	op := &text.DrawOptions{}
	op.GeoM.Translate(5, 3)

	text.Draw(g.offscreen, fmt.Sprintf("Score: %d", g.score),
		&text.GoTextFace{Source: mplusFaceSource, Size: 16},
		op,
	)

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
	g := &Game{
		offscreen: ebiten.NewImage(screenWidth, screenHeight),
		snake: []*Point{
			{boardHeight / 2, boardWidth / 2},
		},
		direction: &Point{1, 0},
		food:      &Point{},
		state:     RUNNING,
	}

	g.setFood()

	return g
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Snake game")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
