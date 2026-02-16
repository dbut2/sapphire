package gba

type GPIO struct {
	Data      uint8
	Direction uint8
	Readable  bool
}

func NewGPIO() *GPIO {
	return &GPIO{}
}

func (g *GPIO) Read(addr uint32) uint8 {
	if !g.Readable {
		return 0
	}
	switch addr &^ 1 {
	case 0x080000C4:
		return g.Data & ^g.Direction & 0x0F
	case 0x080000C6:
		return g.Direction & 0x0F
	case 0x080000C8:
		if g.Readable {
			return 1
		}
		return 0
	}
	return 0
}

func (g *GPIO) Write(addr uint32, data uint8) {
	switch addr &^ 1 {
	case 0x080000C4:
		g.Data = data & 0x0F
	case 0x080000C6:
		g.Direction = data & 0x0F
	case 0x080000C8:
		g.Readable = data&1 == 1
	}
}
