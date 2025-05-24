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
}

func NewMotherboard(gamepak []byte) *Motherboard {
	m := &Motherboard{}

	m.CPU = NewCPU(m)
	m.Memory = NewMemory(m)
	m.LCD = NewLCD(m)
	m.DMA = NewDMA(m)
	m.Timer = NewTimer(m)

	for i := range bios {
		bios[i] ^= 0x69
	}

	m.Memory.SetMemoryBlock(BIOS, bios[:])
	m.Memory.SetMemoryBlock(GPRom1, gamepak)

	return m
}
