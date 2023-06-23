package gba

import (
	_ "embed"
)

//go:embed bios.gba
var bios []byte

type Motherboard struct {
	CPU    *CPU
	Memory Memory
	LCD    *LCD
}

func NewMotherboard(gamepak []byte) *Motherboard {
	m := &Motherboard{}

	m.CPU = NewCPU(m)
	m.Memory = NewMemory()
	m.LCD = NewLCD(m)

	SetMemoryBlock(m.Memory, BIOS, bios)
	SetMemoryBlock(m.Memory, GPRom1, gamepak)

	return m
}
