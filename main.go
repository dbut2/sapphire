package main

import (
	_ "embed"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/dbut2/dialog"
	"github.com/spf13/cobra"

	"dbut.dev/sapphire/gba"
)

func main() {
	c := cmd(run)
	err := c.Execute()
	if err != nil {
		panic(err.Error())
	}
}

func run(game string, gamepak []byte) {
	savePath := deriveSavePath(game)

	a := app.New()
	emu := gba.NewEmu(gamepak)

	if data, err := os.ReadFile(savePath); err == nil {
		emu.Flash.LoadData(data)
	}
	emu.Flash.OnSave(func(data []byte) {
		_ = os.WriteFile(savePath, data, 0644)
	})

	win := window{
		emu:    emu,
		window: a.NewWindow("Sapphire"),
	}
	win.Start()
}

func deriveSavePath(game string) string {
	ext := filepath.Ext(game)
	if strings.EqualFold(ext, ".gba") {
		return game[:len(game)-len(ext)] + ".sav"
	}
	return game + ".sav"
}

func cmd(run func(game string, gamepak []byte)) *cobra.Command {
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

			run(game, gamepak)

			return nil
		},
	}
	c.Flags().StringP("game", "g", "", "Game to load")
	return c
}

type window struct {
	emu    *gba.Emulator
	window fyne.Window

	mu      sync.Mutex
	pressed map[fyne.KeyName]bool
}

var keyMap = map[fyne.KeyName]uint16{
	fyne.KeyZ:         1 << 0, // A
	fyne.KeyX:         1 << 1, // B
	fyne.KeyBackspace: 1 << 2, // Select
	fyne.KeyReturn:    1 << 3, // Start
	fyne.KeyRight:     1 << 4, // Right
	fyne.KeyLeft:      1 << 5, // Left
	fyne.KeyUp:        1 << 6, // Up
	fyne.KeyDown:      1 << 7, // Down
	fyne.KeyS:         1 << 8, // R
	fyne.KeyA:         1 << 9, // L
}

func (w *window) updateKeyInput() {
	w.mu.Lock()
	var mask uint16
	for key := range w.pressed {
		mask |= keyMap[key]
	}
	w.mu.Unlock()
	gba.SetIORegister(w.emu.Memory, gba.KEYINPUT, 0x03FF&^mask)
}

func (w *window) Start() {
	w.pressed = make(map[fyne.KeyName]bool)

	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	w.emu.LCD.SetImage(img)
	cimg := canvas.NewImageFromImage(img)
	cimg.ScaleMode = canvas.ImageScalePixels

	w.emu.LCD.SetDraw(func() {
		w.updateKeyInput()
		fyne.Do(func() {
			cimg.Refresh()
		})
	})

	w.window.SetContent(cimg)
	w.window.Resize(fyne.NewSize(480, 320))

	if dc, ok := w.window.Canvas().(desktop.Canvas); ok {
		dc.SetOnKeyDown(func(event *fyne.KeyEvent) {
			if event.Name == fyne.KeySpace {
				w.emu.FastForward = true
				return
			}
			if _, mapped := keyMap[event.Name]; mapped {
				w.mu.Lock()
				w.pressed[event.Name] = true
				w.mu.Unlock()
			}
		})
		dc.SetOnKeyUp(func(event *fyne.KeyEvent) {
			if event.Name == fyne.KeySpace {
				w.emu.FastForward = false
				return
			}
			w.mu.Lock()
			delete(w.pressed, event.Name)
			w.mu.Unlock()
		})
	}

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
