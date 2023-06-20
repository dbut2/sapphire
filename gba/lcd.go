package gba

import (
	"image"
)

type LCD struct {
	Memory Memory
}

func (l LCD) WriteTo(img *image.RGBA) {
	map[uint16]func(*image.RGBA){
		0: l.WriteToBGMode0,
		1: l.WriteToBGMode1,
		2: l.WriteToBGMode2,
		3: l.WriteToBGMode3,
		4: l.WriteToBGMode4,
		5: l.WriteToBGMode5,
	}[ReadIORegister(l.Memory, DISPCNT)&0b0000000000000111](img)
}

func (l LCD) WriteToBGMode0(img *image.RGBA) {

}

func (l LCD) WriteToBGMode1(img *image.RGBA) {

}

func (l LCD) WriteToBGMode2(img *image.RGBA) {

}

func (l LCD) WriteToBGMode3(img *image.RGBA) {
	bytes := ReadMemoryBlock(l.Memory, VRAM)
	for i := 0; i < 240*160; i++ {
		r, g, b, a := l.RGBA2(bytes[i*2 : i*2+2])
		img.Pix[i*4+0] = uint8(r<<3 + r>>5)
		img.Pix[i*4+1] = uint8(g<<3 + g>>5)
		img.Pix[i*4+2] = uint8(b<<3 + b>>3)
		img.Pix[i*4+3] = uint8(a * 255)
	}
}

func (l LCD) PaletteRGBA(n uint32) (r, g, b, a uint32) {
	return l.RGBA2(ReadMemoryBlock(l.Memory, Palette)[n : n+2])
}

func (l LCD) RGBA2(d []byte) (r, g, b, a uint32) {
	var c uint32

	for i := range d {
		c <<= 8
		c += uint32(d[i])
	}

	r = c >> 11 & 0b11111
	g = c >> 6 & 0b11111
	b = c >> 1 & 0b11111
	a = c & 0b1

	return r, g, b, a
}

func (l LCD) Color(r, g, b, a uint32) []byte {
	c := (r&31)<<11 + (g&31)<<6 + (b&31)<<1 + a
	return []byte{byte(c >> 8), byte(c & 255)}
}

func (l LCD) WriteToBGMode4(img *image.RGBA) {

}

func (l LCD) WriteToBGMode5(img *image.RGBA) {

}
