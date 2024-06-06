package gba

import (
	"image"
	"time"
)

func NewEmu(gamepak []byte) *Emulator {
	motherboard := NewMotherboard(gamepak)

	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	motherboard.LCD.SetImage(img)
	motherboard.LCD.SetDraw(func() {})

	return &Emulator{Motherboard: motherboard}
}

func (e *Emulator) Boot() {
	e.CPU.R13 = 0x03007F00
	e.CPU.R13_svc = 0x03007FE0
	e.CPU.R13_irq = 0x03007FA0

	e.CPU.cpsrInitMode(SYS)
	e.CPU.prefetchFlush()

	SetIORegister(e.CPU.Memory, DISPCNT, 0x80)
	e.CPU.exception(0x08)

	e.Run()
}

func (e *Emulator) Run() {
	ticker := time.NewTicker(16739000 * time.Nanosecond)
	for {
		<-ticker.C
		e.frame()
	}
}

func (e *Emulator) frame() {
	for line := uint16(0); line < 228; line++ {
		e.scanline(line)
	}

	e.LCD.DrawFrame()
}

func (e *Emulator) scanline(line uint16) {
	SetIORegister(e.Memory, VCOUNT, line)

	dispstat := ReadIORegister(e.Memory, DISPSTAT)

	VBlank := (159 - (line % 227)) >> 15 // 0: 0-159, 1: 160-226, 0: 227
	LYC := ReadBits(dispstat, 8, 8)
	VCounter := isEqual(line, LYC)

	dispstat = SetBits(dispstat, 0, 1, VBlank)
	dispstat = SetBits(dispstat, 2, 1, VCounter)

	SetIORegister(e.Memory, DISPSTAT, dispstat)

	blank := ReadBits(ReadIORegister(e.Memory, DISPCNT), 7, 1)

	for e.CPU.cycles = e.CPU.cycles % 1232; e.CPU.cycles < 1232; {
		e.step()
	}

	if line < 160 {
		e.LCD.DrawLine(line, blank)
	}
}

func (e *Emulator) step() {
	dispstat := ReadIORegister(e.Memory, DISPSTAT)
	HBlank := (1005 - e.CPU.cycles) >> 31 // 0: 0-1005, 1: 1006-1231
	dispstat = SetBits(dispstat, 1, 1, uint16(HBlank))
	SetIORegister(e.Memory, DISPSTAT, dispstat)

	preCount := e.CPU.cycles
	e.stepCPU()
	postCount := e.CPU.cycles

	e.Timer.Tick(postCount - preCount)
}
