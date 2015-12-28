package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type direction int
type gameError string

//error types for internal use
const (
	moveError    gameError = "Failed to move"
	fullError    gameError = "No places left to add tiles and no moves to make"
	inputError   gameError = "Illegal input received"
	tooHighError gameError = "We don't have that tile(yet)"
)

type encouragement string

const (
	normal   encouragement = "Good luck getting 2048!"
	past2048 encouragement = "See how far you can go!"
	gameOver encouragement = "Game Over!"
)

const (
	tileCount  = 16
	rowSize    = 4
	columnSize = 4
)

const spawnRate = 0.9

func (g gameError) Error() string {
	return string(g)
}

//movement directions
const (
	goLeft direction = iota
	goDown
	goUp
	goRight
)

func (d direction) String() string {
	switch d {
	case goLeft:
		return "left"
	case goDown:
		return "down"
	case goUp:
		return "up"
	case goRight:
		return "right"
	default:
		return "unknown"
	}
}

//Keycodes used by GTK
const (
	keyLeft  = 65361
	keyDown  = 65364
	keyUp    = 65362
	keyRight = 65363
	keyH     = 104
	keyJ     = 106
	keyK     = 107
	keyL     = 108
)

type image struct {
	*gtk.Image
}

func (i image) Write(p []byte) (n int, err error) {
	i.SetFromFile(string(p))
	return len(p), err
}

//Image resides in the EventBox
type tile struct {
	event *gtk.EventBox
	image
}

type tiles [tileCount]tile

func (t tile) set(n int) (err error) {
	if val, ok := nums[n]; ok {
		_, err = t.Write([]byte(val))
	} else {
		return tooHighError
	}
	return
}

func (t tiles) set(grid grid) {
	for i := 0; i < len(t); i++ {
		t[i].set(grid[i/columnSize][i%rowSize])
	}
}

//Backing state decoupled from GUI
type gameState struct {
	grid   grid
	score  int
	errors map[direction]error
}

//method receiver for the tiles
type grid [columnSize][rowSize]int

//Output for terminal
func (g grid) String() string {
	//Extra newline added on front for log package default logger
	return fmt.Sprintf("\n%4v\n%4v\n%4v\n%4v", g[0], g[1], g[2], g[3])
}

type scoreCounter struct {
	*gtk.Label
}

func (s scoreCounter) set(n int) {
	s.SetLabel(strconv.Itoa(n))
}

type statusBar struct {
	*gtk.Statusbar
	contextID uint
}

func (s statusBar) set(state string) {
	s.Push(s.contextID, state)
}

//All elements from the builder output we will work with
type app struct {
	window        *gtk.Window
	display       *gdk.Display
	about         *gtk.AboutDialog
	headerBar     *gtk.HeaderBar
	resetButton   *gtk.Button
	helpButton    *gtk.Button
	aboutButton   *gtk.Button
	encouragement *gtk.Label
	scoreLabel    *gtk.Label
	tiles         tiles
	statusBar
	scoreCounter
}

//Images of tiles, if a score above 4096 happens we will need new tiles!
var nums = map[int]string{
	0:    "img/empty.png",
	2:    "img/2.png",
	4:    "img/4.png",
	8:    "img/8.png",
	16:   "img/16.png",
	32:   "img/32.png",
	64:   "img/64.png",
	128:  "img/128.png",
	256:  "img/256.png",
	512:  "img/512.png",
	1024: "img/1024.png",
	2048: "img/2048.png",
	4096: "img/4096.png",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	gtk.Init(&os.Args)

	//For autoconnect usage
	signalmap := make(map[string]interface{})

	//Gtk view of the application
	var app app
	//Decoupled view of the application
	var game gameState

	builder, err := gtk.BuilderNew()
	if err != nil {
		panic(err)
	}
	//Read window description
	builder.AddFromFile("ui.glade")

	//Set up all builder elements
	for i := 0; i < tileCount; i++ {
		n := strconv.Itoa(i + 1)
		obj, err := builder.GetObject("eventTile" + n)
		if err != nil {
			panic(err)
		}

		if b, ok := obj.(*gtk.EventBox); ok {
			app.tiles[i].event = b
		}
		obj, err = builder.GetObject("imageTile" + n)
		if err != nil {
			panic(err)
		}
		if b, ok := obj.(*gtk.Image); ok {
			app.tiles[i].image = image{b}
		}
	}
	obj, err := builder.GetObject("resetButton")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		app.resetButton = b
	}

	obj, err = builder.GetObject("helpButton")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		app.helpButton = b
	}

	obj, err = builder.GetObject("aboutButton")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		app.aboutButton = b
	}

	obj, err = builder.GetObject("encouragement")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Label); ok {
		app.encouragement = b
	}

	obj, err = builder.GetObject("scoreLabel")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Label); ok {
		app.scoreLabel = b
	}

	obj, err = builder.GetObject("scoreCounter")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Label); ok {
		app.scoreCounter = scoreCounter{b}
	}

	obj, err = builder.GetObject("statusBar")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.Statusbar); ok {
		app.statusBar = statusBar{w, w.GetContextId("arrow keys")}
	}

	obj, err = builder.GetObject("headerBar")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.HeaderBar); ok {
		app.headerBar = w
	}

	obj, err = builder.GetObject("about")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.AboutDialog); ok {
		app.about = w
	}

	obj, err = builder.GetObject("window")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.Window); ok {
		app.window = w
	}

	//Used for sending audible bells
	app.display, err = gdk.DisplayGetDefault()
	if err != nil {
		panic(err)
	}
	//Done getting elements

	//Signal handlers

	//Placeholder
	signalmap["helpClicked"] = func() {
		randomImages(&game.grid, app.tiles)
	}

	signalmap["aboutClicked"] = func() {
		r := gtk.ResponseType(app.about.Run())
		if r == gtk.RESPONSE_DELETE_EVENT || r == gtk.RESPONSE_CANCEL {
			app.about.Hide()
		}
	}

	//TODO(sjon): Change to storing images in memory instead of loading from disk every single time
	signalmap["resetClicked"] = func() {
		//Reset all tiles in GUI and backing store
		for i := 0; i < tileCount; i++ {
			app.tiles[i].set(0)
			game.grid[i/columnSize][i%rowSize] = 0
		}
		//Reset all score in GUI and backing store
		game.score = 0
		score := strconv.Itoa(game.score)
		app.scoreCounter.SetLabel(score)

		//Reset error count
		game.errors = make(map[direction]error)
		//Reset encouragement
		app.encouragement.SetLabel(string(normal))

		//Perform first tile placement
		err := game.spawn()
		if err != nil {
			panic(err)
		}
		//Display first move in TUI
		//TODO(sjon): display in GUI
		log.Println(game.grid)
	}

	signalmap["removeWindow"] = func() {
		gtk.MainQuit()
	}

	//Currently only for moving the tiles
	signalmap["inputHandler"] = func(win *gtk.Window, ev *gdk.Event) {

		keyEvent := &gdk.EventKey{Event: ev}

		defer func() {
			if len(game.errors) == columnSize {
				for _, v := range game.errors {
					if v != fullError {
						break
					}
				}
				fmt.Println(gameOver)
				app.encouragement.SetLabel(string(gameOver))
			}
		}()

		var moveToMake direction
		switch keyEvent.KeyVal() {
		case keyLeft:
			moveToMake = goLeft
		case keyDown:
			moveToMake = goDown
		case keyUp:
			moveToMake = goUp
		case keyRight:
			moveToMake = goRight
		default:
			app.statusBar.set(fmt.Sprint(inputError))
			return
		}
		app.statusBar.set(moveToMake.String() + " pressed")
		err := game.move(moveToMake)
		if err != nil {
			if err == moveError {
				app.display.Beep()
				return
			} else if err == fullError {
				game.errors[moveToMake] = err
				return
			} else {
				panic(err)
			}
		}
		game.errors = make(map[direction]error)
		//Print to terminal
		//TODO(sjon): Print to display
		log.Println(game.grid)
		//Print score to display
		app.drawMove(game)
	}

	builder.ConnectSignals(signalmap)
	//Done with signal handlers

	//Start at the default state
	if err := game.spawn(); err != nil {
		fmt.Println(game.grid)
	}
	app.window.Show()
	gtk.Main()
}

func (a app) drawMove(game gameState) {
	a.scoreCounter.set(game.score)
	a.tiles.set(game.grid)
}

func randomImages(g *grid, t [tileCount]tile) {
	for i := 0; i < tileCount; i++ {
		for k := range nums {
			g[i/columnSize][i%rowSize] = k
			t[i].set(g[i/columnSize][i%rowSize])
			break
		}
	}

}

//Moves if it is legal to do so
//TODO(sjon): find the error in this code or verify whether it exists
func (g *gameState) move(d direction) error {
	var canmove bool

	switch d {
	case goLeft:
		for y := 0; y <= 3; y++ { //vertical
			for x := 3; x >= 0; x-- { //horizontal
				if g.grid[y][x] != 0 { //If the square is empty we don't have to move it
					if x == 0 { //If we are at the left can't go further left
						continue
					}
					if g.grid[y][x-1] == 0 { //If the square to the left is empty move the current square left
						g.grid[y][x-1], g.grid[y][x] = g.grid[y][x], g.grid[y][x-1]
						canmove = true
					} else if g.grid[y][x-1] == g.grid[y][x] { //If the square to the left is full and the same merge
						g.grid[y][x-1] *= 2
						g.grid[y][x] = 0
						g.score += g.grid[y][x-1]
						x-- //We made a merge and can't make another merge using that tile
						canmove = true
					} //If it is another value don't move
				}
			}
			for x := 0; x < 3; x++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x+1] = g.grid[y][x+1], g.grid[y][x]
				}
			}
			for x := 0; x < 3; x++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x+1] = g.grid[y][x+1], g.grid[y][x]
				}
			}
			for x := 0; x < 3; x++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x+1] = g.grid[y][x+1], g.grid[y][x]
				}
			}
			for x := 0; x < 3; x++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x+1] = g.grid[y][x+1], g.grid[y][x]
				}
			}
		}
	case goDown:
		//		for x := 0; x <= 3; x++ { //horizontal
		//			for y := 3; y >= 0; y++ { //vertical
		//			}
		//		}
		for x := 0; x <= 3; x++ { //horizontal
			for y := 0; y <= 3; y++ { //vertical
				if g.grid[y][x] != 0 { //If the square is empty we don't have to move it
					if y == 3 { //If we are at the left can't go further left
						continue
					}
					if g.grid[y+1][x] == 0 { //If the square to the left is empty move the current square left
						g.grid[y+1][x], g.grid[y][x] = g.grid[y][x], g.grid[y+1][x]
						canmove = true
					} else if g.grid[y+1][x] == g.grid[y][x] { //If the square to the left is full and the same merge
						g.grid[y+1][x] *= 2
						g.grid[y][x] = 0
						g.score += g.grid[y+1][x]
						y-- //We made a merge and can't make another merge using that tile
						canmove = true
					} //If it is another value don't move
				}
			}
			for y := 3; y > 0; y-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y-1][x] = g.grid[y-1][x], g.grid[y][x]
				}
			}
			for y := 3; y > 0; y-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y-1][x] = g.grid[y-1][x], g.grid[y][x]
				}
			}
			for y := 3; y > 0; y-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y-1][x] = g.grid[y-1][x], g.grid[y][x]
				}
			}
			for y := 3; y > 0; y-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y-1][x] = g.grid[y-1][x], g.grid[y][x]
				}
			}
		}
	case goUp:
		//		for x := 0; x <= 3; x++ { //horizontal
		//			for y := 3; y >= 0; y++ { //vertical
		//			}
		//		}
		for x := 0; x <= 3; x++ { //vertical
			for y := 3; y >= 0; y-- { //horizontal
				if g.grid[y][x] != 0 { //If the square is empty we don't have to move it
					if y == 0 { //If we are at the left can't go further left
						continue
					}
					if g.grid[y-1][x] == 0 { //If the square to the left is empty move the current square left
						g.grid[y-1][x], g.grid[y][x] = g.grid[y][x], g.grid[y-1][x]
						canmove = true
					} else if g.grid[y-1][x] == g.grid[y][x] { //If the square to the left is full and the same merge
						g.grid[y-1][x] *= 2
						g.grid[y][x] = 0
						g.score += g.grid[y-1][x]
						y-- //We made a merge and can't make another merge using that tile
						canmove = true
					} //If it is another value don't move
				}
			}
			for y := 0; y < 3; y++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y+1][x] = g.grid[y+1][x], g.grid[y][x]
				}
			}
			for y := 0; y < 3; y++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y+1][x] = g.grid[y+1][x], g.grid[y][x]
				}
			}
			for y := 0; y < 3; y++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y+1][x] = g.grid[y+1][x], g.grid[y][x]
				}
			}
			for y := 0; y < 3; y++ {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y+1][x] = g.grid[y+1][x], g.grid[y][x]
				}
			}
		}
	case goRight:
		//		for y := 0; y <= 3; y++ { //vertical
		//			for x := 3; x >= 0; x-- { //horizontal
		//			}
		//		}
		for y := 0; y <= 3; y++ { //vertical
			for x := 0; x <= 3; x++ { //horizontal
				if g.grid[y][x] != 0 { //If the square is empty we don't have to move it
					if x == 3 { //If we are at the RIGHT can't go further RIGHT
						continue
					}
					if g.grid[y][x+1] == 0 { //If the square to the RIGHT is empty move the current square RIGHT
						g.grid[y][x+1], g.grid[y][x] = g.grid[y][x], g.grid[y][x+1]
						canmove = true
					} else if g.grid[y][x+1] == g.grid[y][x] { //If the square to the RIGHT is full and the same merge
						g.grid[y][x+1] *= 2
						g.grid[y][x] = 0
						g.score += g.grid[y][x+1]
						x++ //We made a merge and can't make another merge using that tile
						canmove = true
					} //If it is another value don't move
				}
			}
			for x := 3; x > 0; x-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x-1] = g.grid[y][x-1], g.grid[y][x]
				}
			}
			for x := 3; x > 0; x-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x-1] = g.grid[y][x-1], g.grid[y][x]
				}
			}
			for x := 3; x > 0; x-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x-1] = g.grid[y][x-1], g.grid[y][x]
				}
			}
			for x := 3; x > 0; x-- {
				if g.grid[y][x] == 0 { //Make them all go fully left
					g.grid[y][x], g.grid[y][x-1] = g.grid[y][x-1], g.grid[y][x]
				}
			}
		}
	default:
		panic(inputError)
	}
	if canmove {
		//We can't spawn if we are full or didn't make a legal move
		if err := g.spawn(); err != nil {
			panic(err)
		}
		return nil
	}
	for i := 0; i < tileCount; i++ {
		if g.grid[i/columnSize][i%rowSize] == 0 {
			return moveError
		}
	}
	//Well we couldn't find an empty slot
	//And we didn't make a successfull move this round
	//But that only described one of the four directions we can't move in
	//We aren't game over until all directions have a fullerror
	return fullError
}

//Spawn a new tile in a random empty square
func (g *gameState) spawn() error {
	var options []int
	for i := 0; i < tileCount; i++ {
		if g.grid[i/columnSize][i%rowSize] == 0 {
			options = append(options, i)
		}
	}
	if len(options) == 0 {
		return fullError
	}
	toSpawn := options[rand.Intn(len(options))]
	if r := rand.Float64(); r > spawnRate {
		g.grid[toSpawn/columnSize][toSpawn%rowSize] = 4
	} else {
		g.grid[toSpawn/columnSize][toSpawn%rowSize] = 2

	}
	return nil
}
