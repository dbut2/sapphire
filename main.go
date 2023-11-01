package main

import (
	_ "embed"
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/dbut2/dialog"
	"github.com/spf13/cobra"

	"github.com/dbut2/sapphire/gba"
)

func main() {
	c := cmd(run)
	err := c.Execute()
	if err != nil {
		panic(err.Error())

	}
}

func run(gamepak []byte) {
	a := app.New()
	win := window{
		emu:    gba.NewEmu(gamepak),
		window: a.NewWindow("Sapphire"),
	}
	win.Start()
}

func cmd(run func(gamepak []byte)) *cobra.Command {
	c := &cobra.Command{
		Use: "sapphire",
		RunE: func(cmd *cobra.Command, args []string) error {
			game, err := cmd.Flags().GetString("game")
			if err != nil {
				return err
			}
			if game == "" {
				game = selectGame()
			}
			gamepak, err := loadGame(game)
			if err != nil {
				return err
			}

			run(gamepak)

			return nil
		},
	}
	c.Flags().StringP("game", "g", "", "Game to load")
	return c
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

func loadGame(game string) ([]byte, error) {
	bytes, err := os.ReadFile(game)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func selectGame() string {
	filename, err := dialog.File().Load()
	if err != nil {
		panic(err.Error())
	}
	return filename
}
