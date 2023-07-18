package main

import (
	_ "embed"
	"image"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"github.com/dbut2/sapphire/gba"
)

//go:embed main.gba
var gamepak []byte

func main() {
	a := app.New()
	w := a.NewWindow("Sapphire")

	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	cimg := canvas.NewImageFromImage(img)
	cimg.ScaleMode = canvas.ImageScalePixels

	w.SetContent(cimg)
	w.Resize(fyne.NewSize(240, 160))

	fps := 30
	ticker := time.NewTicker(time.Second / time.Duration(fps))

	emu := gba.NewEmu(gamepak)

	setup := testSetupNone
	draw := testDrawNone

	go func() {
		setup(emu)
		drawing := false

		for {
			<-ticker.C
			if drawing {
				continue
			}
			drawing = true

			emu.LCD.WriteTo(img)
			cimg.Refresh()

			go func() {
				draw(emu)
				drawing = false
			}()
		}
	}()

	w.SetMainMenu(&fyne.MainMenu{
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
										setup, draw = testSetup3, testDraw3
										img = image.NewRGBA(image.Rect(0, 0, 240, 160))
										cimg.Image = img
										setup(emu)
									},
								},
								{
									Label: "BGMODE5",
									Action: func() {
										setup, draw = testSetup5, testDraw5
										img = image.NewRGBA(image.Rect(0, 0, 240, 160))
										cimg.Image = img
										setup(emu)
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
										ticker = time.NewTicker(time.Second / 1)
									},
								},
								{
									Label: "30",
									Action: func() {
										ticker = time.NewTicker(time.Second / 30)
									},
								},
								{
									Label: "120",
									Action: func() {
										ticker = time.NewTicker(time.Second / 120)
									},
								},
								{
									Label: "1000",
									Action: func() {
										ticker = time.NewTicker(time.Second / 1000)
									},
								},
							},
						},
						Label: "FPS",
					},
				},
			},
		},
	})

	go emu.CPU.Boot()
	w.ShowAndRun()
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
