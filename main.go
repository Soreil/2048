package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/gotk3/gotk3/gtk"
)

type app struct {
	window       *gtk.Window
	resetButton  *gtk.Button
	aboutButton  *gtk.Button
	scoreCounter *gtk.Label
	images       [16]*gtk.Image
	statusBar    *gtk.Statusbar
	sync.Mutex
}

func main() {
	gtk.Init(&os.Args)

	//The window object
	builder, err := gtk.BuilderNew()
	if err != nil {
		panic(err)
	}

	//Signal map containing functions for the signals
	signalmap := make(map[string]interface{})
	builder.AddFromFile("ui.glade")

	var app app

	obj, err := builder.GetObject("button1")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		app.resetButton = b
	}

	obj, err = builder.GetObject("button2")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		app.aboutButton = b
	}

	obj, err = builder.GetObject("label3")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Label); ok {
		app.scoreCounter = b
	}

	obj, err = builder.GetObject("window1")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.Window); ok {
		app.window = w
	}

	signalmap["aboutClicked"] = func() {
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

	signalmap["resetClicked"] = func() {
		fmt.Println("Reset clicked!")
	}

	signalmap["removeWindow"] = func() {
		gtk.MainQuit()
	}
	builder.ConnectSignals(signalmap)
	fmt.Println(signalmap)
	app.window.Show()
	gtk.Main()
}
