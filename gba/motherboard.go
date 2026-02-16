package gba

import (
	_ "embed"
)

type Motherboard struct {
	CPU    *CPU
	Memory *Memory
	LCD    *LCD
	DMA    *DMAController
	Timer  *Timer
	Flash  *Flash
	GPIO   *GPIO
}

func NewMotherboard(gamepak []byte) *Motherboard {
	m := &Motherboard{}

	m.CPU = NewCPU(m)
	m.Memory = NewMemory(m)
	m.LCD = NewLCD(m)
	m.DMA = NewDMA(m)
	m.Timer = NewTimer(m)
	m.Flash = NewFlash(m, gamepak)
	m.GPIO = NewGPIO()

	for i := range bios {
		bios[i] ^= 0x69
	}

	m.Memory.SetMemoryBlock(BIOS, bios[:])
	m.Memory.SetMemoryBlock(GPRom1, gamepak)

	SetIORegister(m.Memory, KEYINPUT, uint16(0x03FF))

	SetIORegister(m.Memory, SOUNDBIAS, uint16(0x200))

	return m
}
