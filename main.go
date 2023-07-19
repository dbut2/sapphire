package main

import (
	_ "embed"
	"image"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"github.com/dbut2/sapphire/gba"
)

func main() {
	a := app.New()

	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	cimg := canvas.NewImageFromImage(img)
	cimg.ScaleMode = canvas.ImageScalePixels

	win := window{
		setup:  testSetupNone,
		draw:   testDrawNone,
		ticker: time.NewTicker(time.Second / time.Duration(30)),
		window: a.NewWindow("Sapphire"),
		img:    img,
		cimg:   cimg,
	}
	win.initEmu()

	win.window.SetContent(cimg)
	win.window.Resize(fyne.NewSize(240, 160))

	win.run()

	win.selectGame()

	win.window.SetMainMenu(win.defaultMainMenu())

	win.window.ShowAndRun()
}

type window struct {
	emu         *gba.Emulator
	setup, draw func(emu *gba.Emulator)
	ticker      *time.Ticker

	window fyne.Window

	img  *image.RGBA
	cimg *canvas.Image
}

func (w *window) initEmu() {
	w.emu = gba.NewEmu(w.selectGame())
}

func (w *window) selectGame() []byte {
	filename := "sapphire.gba"
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	return bytes
}

func (w *window) run() {
	go w.emu.CPU.Boot()
	go func() {
		w.setup(w.emu)
		drawing := false

		for {
			<-w.ticker.C
			if drawing {
				continue
			}
			drawing = true

			w.emu.LCD.WriteTo(w.img)
			w.cimg.Refresh()

			go func() {
				w.draw(w.emu)
				drawing = false
			}()
		}
	}()
}

func (w *window) defaultMainMenu() *fyne.MainMenu {
	return &fyne.MainMenu{
		Items: []*fyne.Menu{
			{
				Label: "Debug",
				Items: []*fyne.MenuItem{
					{
						ChildMenu: &fyne.Menu{
							Items: []*fyne.MenuItem{
								{
									Label: "BGMODE3",
									Action: func() {
										w.setup, w.draw = testSetup3, testDraw3
										w.img = image.NewRGBA(image.Rect(0, 0, 240, 160))
										w.cimg.Image = w.img
										w.setup(w.emu)
									},
								},
								{
									Label: "BGMODE5",
									Action: func() {
										w.setup, w.draw = testSetup5, testDraw5
										w.img = image.NewRGBA(image.Rect(0, 0, 240, 160))
										w.cimg.Image = w.img
										w.setup(w.emu)
									},
								},
							},
						},
						Label: "Video tests",
					},
					{
						ChildMenu: &fyne.Menu{
							Items: []*fyne.MenuItem{
								{
									Label: "1",
									Action: func() {
										w.ticker = time.NewTicker(time.Second / 1)
									},
								},
								{
									Label: "30",
									Action: func() {
										w.ticker = time.NewTicker(time.Second / 30)
									},
								},
								{
									Label: "120",
									Action: func() {
										w.ticker = time.NewTicker(time.Second / 120)
									},
								},
								{
									Label: "1000",
									Action: func() {
										w.ticker = time.NewTicker(time.Second / 1000)
									},
								},
							},
						},
						Label: "FPS",
					},
				},
			},
		},
	}
}

func testSetupNone(emu *gba.Emulator) {

}

func testDrawNone(emu *gba.Emulator) {

}

func testSetup3(emu *gba.Emulator) {
	gba.SetFlag(emu.Memory, gba.BGMODE, 3)
	for i := 0; i < 240; i++ {
		for j := 0; j < 160; j++ {
			index := uint32(i + j*240)
			c := emu.LCD.Color(uint32(32*i/240), 0, uint32(32*j/160), 1)
			emu.Memory.Set16(gba.VRAM[0]+index*2, c)
		}
	}
}

func testDraw3(emu *gba.Emulator) {
	for i := gba.VRAM[0]; i <= gba.VRAM[1]; i += 2 {
		r, g, b, a := emu.LCD.RGBA(emu.Memory.Get16(i))
		c := emu.LCD.Color(r, g+1, b, a)
		emu.Memory.Set16(i, c)
	}
}

func testSetup5(emu *gba.Emulator) {
	gba.SetFlag(emu.Memory, gba.BGMODE, 5)
	for i := 0; i < 160; i++ {
		for j := 0; j < 128; j++ {
			index := uint32(i + j*160)
			c1 := emu.LCD.Color(uint32(32*i/160), 16, uint32(32*j/128), 1)
			c2 := emu.LCD.Color(uint32(32*i/160), 0, uint32(32*j/128), 1)
			emu.Memory.Set16(0x06000000+index*2, c1)
			emu.Memory.Set16(0x0600A000+index*2, c2)
		}
	}
}

func testDraw5(emu *gba.Emulator) {
	for i := gba.VRAM[0]; i <= gba.VRAM[1]; i += 2 {
		r, g, b, a := emu.LCD.RGBA(emu.Memory.Get16(i))
		c := emu.LCD.Color(r, g+1, b, a)
		emu.Memory.Set16(i, c)
	}
	gba.SetFlag(emu.Memory, gba.BGFRAME, ^gba.ReadFlag(emu.Memory, gba.BGFRAME))
}
