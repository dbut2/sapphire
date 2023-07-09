package gba

type Memory []byte

func NewMemory() Memory {
	return make(Memory, 0x10000000)
}

func (m Memory) Access8(address uint32) uint8 {
	address = address & ^uint32(0)
	v := uint8(m[address])
	return v
}

func (m Memory) Set8(address uint32, value uint8) {
	address = address & ^uint32(0)
	m[address] = uint8(value)
}

func (m Memory) Access16(address uint32) uint16 {
	address = address & ^uint32(1)
	v := uint16(m[address])
	v += uint16(m[address+1]) << 8
	return v
}

func (m Memory) Set16(address uint32, value uint16) {
	address = address & ^uint32(1)
	m[address] = uint8(value)
	m[address+1] = uint8(value >> 8)
}

func (m Memory) Access32(address uint32) uint32 {
	address = address & ^uint32(3)
	v := uint32(m[address])
	v += uint32(m[address+1]) << 8
	v += uint32(m[address+2]) << 16
	v += uint32(m[address+3]) << 24
	return v
}

func (m Memory) Set32(address uint32, value uint32) {
	address = address & ^uint32(3)
	m[address] = uint8(value)
	m[address+1] = uint8(value >> 8)
	m[address+2] = uint8(value >> 16)
	m[address+3] = uint8(value >> 24)
}

func (m Memory) AccessSlice(address uint32, to uint32) []byte {
	return m[address : to+1]
}

func (m Memory) SetSlice(address uint32, value []byte) {
	for i := uint32(0); i < uint32(len(value)); i++ {
		m[address+i] = value[i]
	}
}

type MemoryBlock [2]uint32

func ReadMemoryBlock(m Memory, mb MemoryBlock) []byte {
	return m.AccessSlice(mb[0], mb[1])
}

func SetMemoryBlock(m Memory, mb MemoryBlock, value []byte) {
	m.SetSlice(mb[0], value)
}

type IORegister[S Size] uint32

type Size interface {
	uint8 | uint16 | uint32 | uint64
}

func ReadIORegister[S Size](m Memory, r IORegister[S]) S {
	switch v := any(*new(S)).(type) {
	case uint8:
		return S(m.Access8(uint32(r)))
	case uint16:
		return S(m.Access16(uint32(r)))
	case uint32:
		return S(m.Access32(uint32(r)))
	default:
		panic(v)
	}
}

func SetIORegister[S Size](m Memory, r IORegister[S], value S) {
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

func ReadFlag[S Size](m Memory, flag IOFlag[S]) S {
	return ReadBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size)
}

func SetFlag[S Size](m Memory, flag IOFlag[S], value S) {
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
