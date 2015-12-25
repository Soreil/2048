package main

import (
	"math/rand"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
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

//All elements from the builder output we will work with
type app struct {
	window        *gtk.Window
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

func main() {
	gtk.Init(&os.Args)

	//For autoconnect usage
	signalmap := make(map[string]interface{})
	var app app

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
	//Done getting elements

	//Context is needed so we can add messages
	//We will only have one Context
	app.statusID = app.statusBar.GetContextId("arrow keys")

	//Signal handlers

	//Placeholder
	signalmap["helpClicked"] = func() {
		s, err := app.scoreCounter.GetText()
		if err != nil {
			panic(err)
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		ns := strconv.Itoa(n + 1)
		if err != nil {
			panic(err)
		}

		app.scoreCounter.SetText(ns)
	}

	signalmap["aboutClicked"] = func() {
		r := gtk.ResponseType(app.about.Run())
		if r == gtk.RESPONSE_DELETE_EVENT || r == gtk.RESPONSE_CANCEL {
			app.about.Hide()
		}
	}

	//TODO(sjon): Change to storing images in memory instead of loading from disk every single time
	signalmap["resetClicked"] = func() {
		for i := 0; i < 16; i++ {
			app.tiles[i].image.SetFromFile("img/empty.png")
		}
		app.scoreCounter.SetLabel("0")
	}

	//Quit singal handler
	signalmap["removeWindow"] = func() {
		gtk.MainQuit()
	}

	//Currently only for moving the tiles
	//TODO(sjon): Change to not duplicate code for different input schemes
	signalmap["inputHandler"] = func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}
		switch keyEvent.KeyVal() {
		case keyLeft:
			app.statusBar.Push(app.statusID, "left pressed")
			app.tiles[0+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyH:
			app.statusBar.Push(app.statusID, "left pressed")
			app.tiles[0+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyDown:
			app.statusBar.Push(app.statusID, "down pressed")
			app.tiles[1+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyJ:
			app.statusBar.Push(app.statusID, "down pressed")
			app.tiles[1+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyUp:
			app.statusBar.Push(app.statusID, "up pressed")
			app.tiles[2+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyK:
			app.statusBar.Push(app.statusID, "up pressed")
			app.tiles[2+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyRight:
			app.statusBar.Push(app.statusID, "right pressed")
			app.tiles[3+rand.Intn(4)*4].image.SetFromFile(randomImage())
		case keyL:
			app.statusBar.Push(app.statusID, "right pressed")
			app.tiles[3+rand.Intn(4)*4].image.SetFromFile(randomImage())
		default:
			app.statusBar.Push(app.statusID, "fam I don't know")
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
func randomImage() string {
	for _, v := range nums {
		return v
	}
	return "error"
}
