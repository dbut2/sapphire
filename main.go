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

	go func() {
		testSetup3(emu)
		drawing := false

		ticker := time.NewTicker(time.Second / time.Duration(fps))
		for {
			<-ticker.C
			if drawing {
				continue
			}
			drawing = true

			emu.LCD.WriteTo(img)
			cimg.Refresh()

			go func() {
				testDraw3(emu)
				drawing = false
			}()
		}
	}()

	w.ShowAndRun()
}

func testSetup3(emu gba.Emulator) {
	gba.SetFlag(emu.Memory, gba.BGMODE, 3)
	for i := 0; i < 240; i++ {
		for j := 0; j < 160; j++ {
			index := uint32(i + j*240)
			c := emu.LCD.Color(uint32(32*i/240), 0, uint32(32*j/160), 1)
			emu.Memory.Set16(gba.VRAM[0]+index*2, c)
		}
	}
}

func testDraw3(emu gba.Emulator) {
	for i := uint32(0); i < 65536; i++ {
		r, g, b, a := emu.LCD.RGBA2(emu.Memory.Access16(gba.VRAM[0] + i*2))
		c := emu.LCD.Color(r, g+1, b, a)
		emu.Memory.Set16(gba.VRAM[0]+i*2, c)
	}
}

func testSetup5(emu gba.Emulator) {
	gba.SetFlag(emu.Memory, gba.BGMODE, 5)
	for i := 0; i < 160; i++ {
		for j := 0; j < 128; j++ {
			index := uint32(i + j*160)
			c := emu.LCD.Color(uint32(32*i/160), 0, uint32(32*j/128), 1)
			emu.Memory.Set16(0x06000000+index*2, c)
			emu.Memory.Set16(0x0600A000+index*2, c)
		}
	}
}

func testDraw5(emu gba.Emulator) {
	for i := uint32(0); i < 65536; i++ {
		r, g, b, a := emu.LCD.RGBA2(emu.Memory.Access16(gba.VRAM[0] + i*2))
		c := emu.LCD.Color(r, g+1, b, a)
		emu.Memory.Set16(gba.VRAM[0]+i*2, c)
	}
	gba.SetFlag(emu.Memory, gba.BGFRAME, ^gba.ReadFlag(emu.Memory, gba.BGFRAME))
}
