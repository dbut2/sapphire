package gba

import (
	"image"
)

type LCD struct {
	Memory Memory
}

func (l LCD) WriteTo(img *image.RGBA) {
	map[uint16]func(*image.RGBA){
		0: l.BGMode0WriteTo,
		1: l.BGMode1WriteTo,
		2: l.BGMode2WriteTo,
		3: l.BGMode3WriteTo,
		4: l.BGMode4WriteTo,
		5: l.BGMode5WriteTo,
	}[ReadFlag(l.Memory, BGMODE)](img)
}

func (l LCD) BGMode0WriteTo(img *image.RGBA) {

}

func (l LCD) BGMode1WriteTo(img *image.RGBA) {

}

func (l LCD) BGMode2WriteTo(img *image.RGBA) {

}

func (l LCD) BGMode3WriteTo(img *image.RGBA) {
	frame := MemoryBlock{0x06000000, 0x06013FFF}
	for i := 0; i < 240*160; i++ {
		r, g, b, a := l.RGBA2(l.Memory.Access16(frame[0] + uint32(i)*2))
		img.Pix[i*4+0] = uint8(r<<3 + r>>5)
		img.Pix[i*4+1] = uint8(g<<3 + g>>5)
		img.Pix[i*4+2] = uint8(b<<3 + b>>5)
		img.Pix[i*4+3] = uint8(a * 255)
	}
}

func (l LCD) BGMode4WriteTo(img *image.RGBA) {
	frame := map[uint16]MemoryBlock{
		0: {0x06000000, 0x06009FFF},
		1: {0x0600A000, 0x06013FFF},
	}[ReadFlag(l.Memory, BGFRAME)]
	bytes := ReadMemoryBlock(l.Memory, frame)
	for i := 0; i < 240*160; i++ {
		r, g, b, a := l.PaletteRGBA(bytes[i])
		img.Pix[i*4+0] = uint8(r<<3 + r>>5)
		img.Pix[i*4+1] = uint8(g<<3 + g>>5)
		img.Pix[i*4+2] = uint8(b<<3 + b>>5)
		img.Pix[i*4+3] = uint8(a * 255)
	}
}

func (l LCD) BGMode5WriteTo(img *image.RGBA) {
	frame := map[uint16]MemoryBlock{
		0: {0x06000000, 0x06009FFF},
		1: {0x0600A000, 0x06013FFF},
	}[ReadFlag(l.Memory, BGFRAME)]
	for i := 0; i < 160*128; i++ {
		index := i + (i/160)*80 + 240*16 + 40
		r, g, b, a := l.RGBA2(l.Memory.Access16(frame[0] + uint32(i)*2))
		img.Pix[index*4+0] = uint8(r<<3 + r>>5)
		img.Pix[index*4+1] = uint8(g<<3 + g>>5)
		img.Pix[index*4+2] = uint8(b<<3 + b>>5)
		img.Pix[index*4+3] = uint8(a * 255)
	}
}

func (l LCD) PaletteRGBA(n uint8) (r, g, b, a uint32) {
	c := l.Memory.Access16(Palette[0] + uint32(n)*2)
	return l.RGBA2(c)
}

func (l LCD) RGBA2(d uint16) (r, g, b, a uint32) {
	c := uint32(d)

	r = c >> 11 & 0b11111
	g = c >> 6 & 0b11111
	b = c >> 1 & 0b11111
	a = c & 0b1

	return r, g, b, a
}

func (l LCD) Color(r, g, b, a uint32) uint16 {
	c := (r&31)<<11 + (g&31)<<6 + (b&31)<<1 + a
	return uint16(c)
}
