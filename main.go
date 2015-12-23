package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(&os.Args)

	builder, err := gtk.BuilderNew()
	if err != nil {
		panic(err)
	}
	signalmap := make(map[string]interface{})
	builder.AddFromFile("ui.glade")

	var resetButton *gtk.Button
	var aboutButton *gtk.Button
	var scoreCounter *gtk.Label
	var window *gtk.Window

	obj, err := builder.GetObject("button1")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		resetButton = b
	}

	obj, err = builder.GetObject("button2")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Button); ok {
		aboutButton = b
	}

	obj, err = builder.GetObject("label3")
	if err != nil {
		panic(err)
	}
	if b, ok := obj.(*gtk.Label); ok {
		scoreCounter = b
	}

	obj, err = builder.GetObject("window1")
	if err != nil {
		panic(err)
	}
	if w, ok := obj.(*gtk.Window); ok {
		window = w
	}

	aboutButton.Connect("clicked", func() {
		s, err := scoreCounter.GetText()
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

		scoreCounter.SetText(ns)
	})

	resetButton.Connect("clicked", func() {})

	signalmap["removeWindow"] = func() {
		gtk.MainQuit()
	}
	builder.ConnectSignals(signalmap)
	fmt.Println(signalmap)
	window.Show()
	gtk.Main()
}
