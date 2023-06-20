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

func main() {
	a := app.New()
	w := a.NewWindow("Sapphire")
	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	cimg := canvas.NewImageFromImage(img)
	cimg.ScaleMode = canvas.ImageScalePixels
	w.SetContent(cimg)
	w.Resize(fyne.NewSize(240, 160))
	fps := 30

	var gamepak []byte
	emu := gba.NewEmu(gamepak)

	drawing := false
	go func() {
		timer := time.NewTicker(time.Second / time.Duration(fps))

		testSetup(emu)
		for true {
			<-timer.C
			if drawing {
				continue
			}
			drawing = true

			emu.LCD.WriteTo(img)
			cimg.Refresh()

			go func() {
				testDraw(emu)
				drawing = false
			}()
		}
	}()

	w.ShowAndRun()
}

func testSetup(emu gba.Emulator) {
	gba.SetIORegister(emu.Memory, gba.DISPCNT, 0b0000000000000011)
	for i := 0; i < 240; i++ {
		for j := 0; j < 160; j++ {
			index := uint32(i + j*240)
			c := emu.LCD.Color(uint32(32*i/240), 0, uint32(32*j/160), 1)
			emu.Memory.SetSlice(gba.VRAM[0]+index*2, c)
		}
	}
}

func testDraw(emu gba.Emulator) {
	for i := uint32(0); i < 240*160; i++ {
		r, g, b, a := emu.LCD.RGBA2(gba.ReadMemoryBlock(emu.Memory, gba.VRAM)[i*2 : i*2+2])
		c := emu.LCD.Color(r, g+1, b, a)
		emu.Memory.SetSlice(gba.VRAM[0]+i*2, c)
	}
}
