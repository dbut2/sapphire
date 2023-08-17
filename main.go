package main

import (
	_ "embed"
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/dbut2/dialog"

	"github.com/dbut2/sapphire/gba"
)

func main() {
	a := app.New()

	gamepak := loadGame()

	win := window{
		emu:    gba.NewEmu(gamepak),
		window: a.NewWindow("Sapphire"),
	}
	win.Start()
}

type window struct {
	emu    *gba.Emulator
	window fyne.Window
}

func (w *window) Start() {
	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	w.emu.LCD.SetImage(img)
	cimg := canvas.NewImageFromImage(img)
	cimg.ScaleMode = canvas.ImageScalePixels

	w.emu.LCD.SetDraw(func() {
		cimg.Refresh()
	})

	w.window.SetContent(cimg)
	w.window.Resize(fyne.NewSize(240, 160))

	go w.emu.Boot()

	w.window.ShowAndRun()
}

func loadGame() []byte {
	filename := selectGame()
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	return bytes
}

func selectGame() string {
	filename := os.Getenv("game")
	if filename != "" {
		return filename
	}
	filename, err := dialog.File().Load()
	if err != nil {
		panic(err.Error())
	}
	return filename
}
