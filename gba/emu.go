package gba

import (
	"image"
	"time"
)

func NewEmu(gamepak, bios []byte) *Emulator {
	motherboard := NewMotherboard(gamepak, bios)

	img := image.NewRGBA(image.Rect(0, 0, 240, 160))
	motherboard.LCD.setImage(img)
	motherboard.LCD.SetDraw(func() {})

	return &Emulator{Motherboard: motherboard}
}

func (e *Emulator) PreBoot() {
	e.CPU.R13_irq = 0x03007FA0
	e.CPU.R13_svc = 0x03007FE0

	e.CPU.CPSR = SYS
	e.CPU.R[13] = 0x03007F00

	SetIORegister(e.Memory, POSTFLG, uint8(1))

	e.CPU.R[15] = 0x08000000
	e.CPU.prefetchFlush()
}

func (e *Emulator) Boot() {
	e.PreBoot()
	e.Run()
}

func (e *Emulator) Run() {
	const frameDur = 16739 * time.Microsecond
	ticker := time.NewTicker(frameDur)
	lastDraw := time.Now()
	for {
		if e.FastForward {
			if time.Since(lastDraw) >= frameDur {
				lastDraw = time.Now()
				e.runFrame(false)
			} else {
				e.runFrame(true)
			}
			continue
		}
		<-ticker.C
		e.runFrame(false)
		lastDraw = time.Now()
	}
}

func (e *Emulator) frame() {
	e.runFrame(false)
}

func (e *Emulator) runFrame(skipDraw bool) {
	e.skipDraw = skipDraw
	for line := uint16(0); line < 228; line++ {
		e.scanline(line)
	}

	if skipDraw {
		e.LCD.CountFrame()
	} else {
		e.LCD.DrawFrame()
	}
	e.Flash.Flush()
}

func (e *Emulator) scanline(line uint16) {
	SetIORegister16(e.Memory, VCOUNT, line)

	dispstat := ReadIORegister16(e.Memory, DISPSTAT)

	prevVBlank := ReadBits(dispstat, 0, 1)
	VBlank := (159 - (line % 227)) >> 15 // 0: 0-159, 1: 160-226, 0: 227
	LYC := ReadBits(dispstat, 8, 8)
	VCounter := isEqual(line, LYC)

	dispstat = SetBits(dispstat, 0, 1, VBlank)
	dispstat = SetBits(dispstat, 2, 1, VCounter)

	SetIORegister16(e.Memory, DISPSTAT, dispstat)

	if VBlank == 1 && prevVBlank == 0 {
		if ReadBits(dispstat, 3, 1) == 1 {
			ifReg := ReadIORegister16(e.Memory, IF)
			SetIORegister16(e.Memory, IF, ifReg|0x0001)
		}
		e.DMA.transfer(DMAVBlank)
	}

	if line == 0 {
		e.LCD.LatchAffineRefs()
	}

	if VCounter == 1 && ReadBits(dispstat, 5, 1) == 1 {
		ifReg := ReadIORegister16(e.Memory, IF)
		SetIORegister16(e.Memory, IF, ifReg|0x0004)
	}

	for e.CPU.cycles = e.CPU.cycles % 1232; e.CPU.cycles < 1232; {
		e.step()
	}

	blank := ReadBits(ReadIORegister16(e.Memory, DISPCNT), 7, 1)

	if line < 160 {
		if !e.skipDraw {
			e.LCD.DrawLine(line, blank)
		}
		e.LCD.IncrementAffineRefs()
	}

	if ReadBits(dispstat, 4, 1) == 1 {
		ifReg := ReadIORegister16(e.Memory, IF)
		SetIORegister16(e.Memory, IF, ifReg|0x0002)
	}
	if line < 160 {
		e.DMA.transfer(DMAHBlank)
	}
}

func (e *Emulator) step() {
	hblank := e.CPU.cycles > 1005
	if hblank != e.hblank {
		e.hblank = hblank
		dispstat := ReadIORegister16(e.Memory, DISPSTAT)
		var bit uint16
		if hblank {
			bit = 1
		}
		SetIORegister16(e.Memory, DISPSTAT, SetBits(dispstat, 1, 1, bit))
	}

	if ifReg := ReadIORegister16(e.Memory, IF); ifReg != 0 {
		if e.CPU.cpsrIRQDisable() == 0 {
			ime := ReadIORegister16(e.Memory, IME)
			ie := ReadIORegister16(e.Memory, IE)
			if ime > 0 && ie&ifReg > 0 {
				e.CPU.halted = false
				e.CPU.exception(0x18)
			}
		} else if e.CPU.halted {
			ime := ReadIORegister16(e.Memory, IME)
			ie := ReadIORegister16(e.Memory, IE)
			if ime > 0 && ie&ifReg > 0 {
				e.CPU.halted = false
			}
		}
	}

	if e.CPU.halted {
		jump := uint32(8)
		if e.CPU.cycles < 1232 {
			jump = 1232 - e.CPU.cycles
		}
		if until := e.Timer.untilDeadline(); until < jump {
			jump = until
		}
		if jump == 0 {
			jump = 1
		}
		e.CPU.cycle(jump)
		e.Timer.Tick(jump)
		e.APU.Tick(jump)
		return
	}

	preCount := e.CPU.cycles
	e.stepCPU()
	postCount := e.CPU.cycles

	elapsed := postCount - preCount
	e.Timer.Tick(elapsed)
	e.APU.Tick(elapsed)
}
