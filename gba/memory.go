package gba

import (
	"github.com/dbut2/sapphire/gba/memory"
)

func NewMemory() memory.Memory {
	m := memory.NewMap(parseAddress)

	registerBlock(m, BIOS)
	registerBlock(m, WRAM1)
	registerBlock(m, WRAM2)
	registerBlock(m, IOR)
	registerBlock(m, Palette)
	registerBlock(m, VRAM)
	registerBlock(m, OAM)
	registerBlock(m, GPRom1)
	registerBlock(m, GPRom2)
	registerBlock(m, GPRom3)

	sram := memory.Mirror{Offset: memory.NewOffset(GPSRAM[0], GPSRAM[1]-GPSRAM[0]+1)}
	m.Register(0x0E000000, sram)
	m.Register(0x0F000000, sram)

	return m
}

func registerBlock(m memory.Map[uint32], block MemoryBlock) {
	sub := memory.NewOffset(block[0], block[1]-block[0]+1)
	m.Register(block[0], sub)
}

func parseAddress(address uint32) uint32 {
	return (address >> 24) & 0b1111
}

type MemoryBlock [2]uint32

func ReadMemoryBlock(m memory.Memory, mb MemoryBlock) []byte {
	return m.GetSlice(mb[0], mb[1]-mb[0])
}

func SetMemoryBlock(m memory.Memory, mb MemoryBlock, value []byte) {
	m.SetSlice(mb[0], value)
}

type IORegister[S Size] uint32

type Size interface {
	uint8 | uint16 | uint32 | uint64
}

func ReadIORegister[S Size](m memory.Memory, r IORegister[S]) S {
	switch v := any(*new(S)).(type) {
	case uint8:
		return S(m.Get8(uint32(r)))
	case uint16:
		return S(m.Get16(uint32(r)))
	case uint32:
		return S(m.Get32(uint32(r)))
	default:
		panic(v)
	}
}

func SetIORegister[S Size](m memory.Memory, r IORegister[S], value S) {
	switch v := any(*new(S)).(type) {
	case uint8:
		m.Set8(uint32(r), uint8(value))
	case uint16:
		m.Set16(uint32(r), uint16(value))
	case uint32:
		m.Set32(uint32(r), uint32(value))
	default:
		panic(v)
	}
}

type IOFlag[S Size] struct {
	Register IORegister[S]
	Bit      uint8
	Size     uint8
}

func Flag[S Size](r IORegister[S], bit uint8, size uint8) IOFlag[S] {
	return IOFlag[S]{
		Register: r,
		Bit:      bit,
		Size:     size,
	}
}

func ReadFlag[S Size](m memory.Memory, flag IOFlag[S]) S {
	return ReadBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size)
}

func SetFlag[S Size](m memory.Memory, flag IOFlag[S], value S) {
	SetIORegister(m, flag.Register, SetBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size, value))
}

func ReadBits[S Size](v S, bit uint8, size uint8) S {
	return (v >> bit) & (1<<size - 1)
}

func SetBits[S Size](v S, bit uint8, size uint8, value S) S {
	mask := S(1<<size-1) << bit
	v &= ^mask
	v |= value << bit
	return v
}
