package main

import (
	"image"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

func clr(r, g, b, a uint32) []byte {
	c := r&31<<11 + g&31<<6 + b&31<<1 + a
	return []byte{byte(c >> 8), byte(c & 255)}
}

func rgba(d []byte) (r, g, b, a uint32) {
	var c uint32

	for i := range d {
		c <<= 8
		c += uint32(d[i])
	}

	r = (c & (0b11111 << 11)) >> 11
	g = (c & (0b11111 << 6)) >> 6
	b = (c & (0b11111 << 1)) >> 1
	a = c & 0b1

	return r, g, b, a
}

func main() {

	a := app.New()
	w := a.NewWindow("Sapphire")

	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	cimg := canvas.NewImageFromImage(img)

	cimg.ScaleMode = canvas.ImageScalePixels

	vramSize := 240 * 160 * 2
	vram := NewMemory(vramSize * 2)

	video := Video{
		Memory: vram[:vramSize],
		Buffer: vram[vramSize:],
	}

	for i := 0; i < 240; i++ {
		for j := 0; j < 160; j++ {
			index := (i + j*240) * 2
			video.Buffer.Set(index, clr(uint32(32*i/240), 0, uint32(32*j/160), 1))
		}
	}

	fps := 30

	go func() {
		timer := time.NewTicker(time.Second / time.Duration(fps))

		for true {
			_ = <-timer.C
			if video.Drawing {
				continue
			}
			video.Drawing = true

			video.Draw(img)
			cimg.Refresh()

			go func() {
				for i := range video.Memory {
					if i%2 == 1 {
						continue
					}

					r, g, b, a := rgba(video.Memory[i : i+2])
					video.Buffer.Set(i, clr(r, g+1, b, a))
				}
				video.Drawing = false
			}()
		}
	}()

	w.SetContent(cimg)

	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(240, 160))
	w.ShowAndRun()
}

type Video struct {
	Drawing bool
	Memory  Memory
	Buffer  Memory
}

func (v *Video) Draw(img *image.RGBA) {
	v.Memory, v.Buffer = v.Buffer, v.Memory
	for i := 0; i < 240; i++ {
		for j := 0; j < 160; j++ {
			index := (i + j*240) * 2
			r, g, b, a := rgba(v.Memory[index : index+2])
			img.Pix[index*2+0] = five2eight[r]
			img.Pix[index*2+1] = five2eight[g]
			img.Pix[index*2+2] = five2eight[b]
			img.Pix[index*2+3] = uint8(a) * 255
		}
	}
}

type Memory []byte

func NewMemory(size int) Memory {
	return make([]byte, size)
}

func (m Memory) Set(at int, b []byte) {
	for i := range b {
		m[at+i] = b[i]
	}
}

var five2eight = map[uint32]uint8{
	0:  0,
	1:  8,
	2:  16,
	3:  24,
	4:  33,
	5:  41,
	6:  49,
	7:  57,
	8:  66,
	9:  74,
	10: 82,
	11: 90,
	12: 99,
	13: 107,
	14: 115,
	15: 123,
	16: 132,
	17: 140,
	18: 148,
	19: 156,
	20: 165,
	21: 173,
	22: 181,
	23: 189,
	24: 198,
	25: 206,
	26: 214,
	27: 222,
	28: 231,
	29: 239,
	30: 247,
	31: 255,
}
