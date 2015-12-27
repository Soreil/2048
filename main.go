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

//Image resides in the EventBox
type tile struct {
	event *gtk.EventBox
	image *gtk.Image
}

//Backing state decoupled from GUI
type gameState struct {
	grid   grid
	score  int
	errors map[direction]error
}

//method receiver for the tiles
type grid [4][4]int

//Output for terminal
func (g grid) String() string {
	//Extra newline added on front for log package default logger
	return fmt.Sprintf("\n%4v\n%4v\n%4v\n%4v", g[0], g[1], g[2], g[3])
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
	scoreCounter  *gtk.Label
	tiles         [16]tile
	statusBar     *gtk.Statusbar
	statusID      uint
}

//Images of tiles, if a score above 4096 happens we will need new tiles!
var nums = map[int]string{
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
	for i := 0; i < 16; i++ {
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
			app.tiles[i].image = b
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
		app.scoreCounter = b
	}

	obj, err = builder.GetObject("statusBar")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.Statusbar); ok {
		app.statusBar = w
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

	//Context is needed so we can add messages
	//We will only have one Context
	app.statusID = app.statusBar.GetContextId("arrow keys")

	//Signal handlers

	//Placeholder
	signalmap["helpClicked"] = func() {
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
		for i := 0; i < 16; i++ {
			app.tiles[i].image.SetFromFile("img/empty.png")
			game.grid[i/4][i%4] = 0
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
	//TODO(sjon): Change to not duplicate code for different input schemes
	//TODO(sjon): Change to not duplicate code for bulk of move
	signalmap["inputHandler"] = func(win *gtk.Window, ev *gdk.Event) {

		keyEvent := &gdk.EventKey{ev}

		defer func() {
			if len(game.errors) == 4 {
				for _, v := range game.errors {
					if v != fullError {
						break
					}
				}
				fmt.Println(gameOver)
				app.encouragement.SetLabel(string(gameOver))
			}
		}()

		switch keyEvent.KeyVal() {
		case keyLeft:
			app.statusBar.Push(app.statusID, "left pressed")
			err := game.move(goLeft)
			if err != nil {
				if err == moveError {
					app.display.Beep()
					return
				} else if err == fullError {
					game.errors[goLeft] = err
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
			app.scoreCounter.SetLabel(strconv.Itoa(game.score))
		//case keyH:
		//	app.statusBar.Push(app.statusID, "left pressed")
		case keyDown:
			app.statusBar.Push(app.statusID, "down pressed")
			err := game.move(goDown)
			if err != nil {
				if err == moveError {
					app.display.Beep()
					return
				} else if err == fullError {
					game.errors[goDown] = err
					return
				} else {
					panic(err)
				}
			}
			//Successful move, reset error counter and process new state
			game.errors = make(map[direction]error)
			log.Println(game.grid)
			app.scoreCounter.SetLabel(strconv.Itoa(game.score))
		//case keyJ:
		//	app.statusBar.Push(app.statusID, "down pressed")
		case keyUp:
			app.statusBar.Push(app.statusID, "up pressed")
			err := game.move(goUp)
			if err != nil {
				if err == moveError {
					app.display.Beep()
					return
				} else if err == fullError {
					game.errors[goUp] = err
					return
				} else {
					panic(err)
				}
			}
			game.errors = make(map[direction]error)
			log.Println(game.grid)
			app.scoreCounter.SetLabel(strconv.Itoa(game.score))
		//case keyK:
		//	app.statusBar.Push(app.statusID, "up pressed")
		case keyRight:
			app.statusBar.Push(app.statusID, "right pressed")
			err := game.move(goRight)
			if err != nil {
				if err == moveError {
					app.display.Beep()
					return
				} else if err == fullError {
					game.errors[goRight] = err
					return
				} else {
					panic(err)
				}
			}
			game.errors = make(map[direction]error)
			log.Println(game.grid)
			app.scoreCounter.SetLabel(strconv.Itoa(game.score))
		//case keyL:
		//	app.statusBar.Push(app.statusID, "right pressed")
		default:
			app.statusBar.Push(app.statusID, fmt.Sprint(inputError))
		}
	}

	builder.ConnectSignals(signalmap)
	//Done with signal handlers

	//Start at the default state
	signalmap["resetClicked"].(func())()

	app.window.Show()
	gtk.Main()
}

//Relies on random map traversal
//UNUSED
func randomImage() string {
	for _, v := range nums {
		return v
	}
	return "error"
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
	}
	if canmove {
		//We can't spawn if we are full or didn't make a legal move
		//TODO(sjon): Handle full differently from an illegal move
		if err := g.spawn(); err != nil {
			panic(err)
		}
		return nil
	} else {
		for i := 0; i < 16; i++ {
			if g.grid[i/4][i%4] == 0 {
				return moveError
			}
		}
		//Well we couldn't find an empty slot
		//And we didn't make a successfull move this round
		//But that only described one of the four directions we can't move in
		//We aren't game over until all directions have a fullerror
		return fullError
	}
}

//Spawn a two in a random empty square

func (g *gameState) spawn() error {
	var options []int
	for i := 0; i < 16; i++ {
		if g.grid[i/4][i%4] == 0 {
			options = append(options, i)
		}
	}
	if len(options) == 0 {
		return fullError
	}
	toSpawn := options[rand.Intn(len(options))]
	if r := rand.Float64(); r > 0.8 {
		g.grid[toSpawn/4][toSpawn%4] = 4
	} else {
		g.grid[toSpawn/4][toSpawn%4] = 2

	}
	return nil
}
