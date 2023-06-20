package gba

import (
	_ "embed"
)

//go:embed bios.gba
var bios []byte

type Emulator struct {
	CPU    CPU
	Memory Memory
	LCD    LCD
}

func NewEmu(gamepak []byte) Emulator {
	c := CPU{}
	m := NewMemory()
	l := LCD{Memory: m}

	SetMemoryBlock(m, BIOS, bios)
	SetMemoryBlock(m, GPRom1, gamepak)

	return Emulator{CPU: c, Memory: m, LCD: l}
}
