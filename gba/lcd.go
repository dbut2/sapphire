package gba

import (
	"image"
)

type LCD struct {
	*Motherboard
	img  *image.RGBA
	draw func()
}

func NewLCD(m *Motherboard) *LCD {
	return &LCD{
		Motherboard: m,
	}
}

func (l *LCD) SetImage(img *image.RGBA) {
	l.img = img
}

func (l *LCD) SetDraw(draw func()) {
	l.draw = draw
}

func (l *LCD) DrawFrame() {
	l.draw()
}

func (l *LCD) DrawLine(line uint16, blank uint16) {
	if blank == 1 {
		l.Blank(line)
		return
	}

	map[uint16]func(uint16){
		0: l.BGMode0Write,
		1: l.BGMode1Write,
		2: l.BGMode2Write,
		3: l.BGMode3Write,
		4: l.BGMode4Write,
		5: l.BGMode5Write,
	}[ReadFlag(l.Memory, BGMODE)](line)
}

func (l *LCD) Blank(line uint16) {
	lo := uint32(line) * 240
	for i := lo; i < lo+240; i++ {
		l.img.Pix[i*4+0] = 255
		l.img.Pix[i*4+1] = 255
		l.img.Pix[i*4+2] = 255
		l.img.Pix[i*4+3] = 1
	}
}

func (l *LCD) BGMode0Write(line uint16) {
	panic("unimplemented") // todo
}

func (l *LCD) BGMode1Write(line uint16) {
	panic("unimplemented") // todo
}

func (l *LCD) BGMode2Write(line uint16) {
	panic("unimplemented") // todo
}

func (l *LCD) BGMode3Write(line uint16) {
	for i := uint32(0); i < 240; i++ {
		pixel := uint32(line)*160 + i
		r, g, b, a := l.RGBA(l.Memory.Read16(VRAM.Start+pixel*2, false, false))
		l.img.Pix[pixel*4+0] = uint8(r<<3 + r>>5)
		l.img.Pix[pixel*4+1] = uint8(g<<3 + g>>5)
		l.img.Pix[pixel*4+2] = uint8(b<<3 + b>>5)
		l.img.Pix[pixel*4+3] = uint8(a * 255)
	}
}

func (l *LCD) BGMode4Write(line uint16) {
	frame := [2]uint32{0x06000000, 0x0600A000}[ReadFlag(l.Memory, BGFRAME)]
	bytes := l.Memory.ReadMemoryBlock(VRAM)
	for i := uint32(0); i < 240*160; i++ {
		pixel := uint32(line)*160 + i
		r, g, b, a := l.PaletteRGBA(bytes[frame+pixel])
		l.img.Pix[pixel*4+0] = uint8(r<<3 + r>>5)
		l.img.Pix[pixel*4+1] = uint8(g<<3 + g>>5)
		l.img.Pix[pixel*4+2] = uint8(b<<3 + b>>5)
		l.img.Pix[pixel*4+3] = uint8(a * 255)
	}
}

func (l *LCD) BGMode5Write(line uint16) {
	frame := [2]uint32{0x06000000, 0x0600A000}[ReadFlag(l.Memory, BGFRAME)]
	for i := uint32(0); i < 160*128; i++ {
		pixel := uint32(line)*160 + i
		index := pixel + (pixel/160)*80 + 240*16 + 40
		r, g, b, a := l.RGBA(l.Memory.Read16(frame+uint32(pixel)*2, false, false))
		l.img.Pix[index*4+0] = uint8(r<<3 + r>>5)
		l.img.Pix[index*4+1] = uint8(g<<3 + g>>5)
		l.img.Pix[index*4+2] = uint8(b<<3 + b>>5)
		l.img.Pix[index*4+3] = uint8(a * 255)
	}
}

func (l *LCD) PaletteRGBA(n uint8) (r, g, b, a uint32) {
	c := l.Memory.Read16(Palette.Start+uint32(n)*2, false, false)
	return l.RGBA(c)
}

func (l *LCD) RGBA(d uint16) (r, g, b, a uint32) {
	c := uint32(d)

	r = ReadBits(c, 0, 5)
	g = ReadBits(c, 5, 5)
	b = ReadBits(c, 10, 5)
	a = ReadBits(c, 15, 1)

	return r, g, b, a
}

func (l *LCD) Color(r, g, b, a uint32) uint16 {
	c := (r&31)<<11 + (g&31)<<6 + (b&31)<<1 + a
	return uint16(c)
}
