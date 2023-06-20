package gba

type Memory []byte

func NewMemory() Memory {
	return make(Memory, 0x10000000)
}

func (m Memory) Access8(address uint32) uint8 {
	v := uint8(m[address])
	return v
}

func (m Memory) Set8(address uint32, value uint8) {
	m[address] = uint8(value)
}

func (m Memory) Access16(address uint32) uint16 {
	v := uint16(m[address])
	v += uint16(m[address+1]) << 8
	return v
}

func (m Memory) Set16(address uint32, value uint16) {
	m[address] = uint8(value)
	m[address+1] = uint8(value >> 8)
}

func (m Memory) Access32(address uint32) uint32 {
	v := uint32(m[address])
	v += uint32(m[address+1]) << 8
	v += uint32(m[address+2]) << 16
	v += uint32(m[address+3]) << 24
	return v
}

func (m Memory) Set32(address uint32, value uint32) {
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

type IORegister[S RegisterSize] uint32

type RegisterSize interface {
	uint8 | uint16 | uint32
}

func ReadIORegister[S RegisterSize](m Memory, r IORegister[S]) S {
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

func SetIORegister[S RegisterSize](m Memory, r IORegister[S], value S) {
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

type IOFlag[S RegisterSize] struct {
	Register IORegister[S]
	Bit      uint8
	Size     uint8
}

func Flag[S RegisterSize](r IORegister[S], bit uint8, size uint8) IOFlag[S] {
	return IOFlag[S]{
		Register: r,
		Bit:      bit,
		Size:     size,
	}
}

func ReadFlag[S RegisterSize](m Memory, flag IOFlag[S]) S {
	v := ReadIORegister(m, flag.Register)
	return v >> flag.Bit & (1<<flag.Size - 1)
}

func SetFlag[S RegisterSize](m Memory, flag IOFlag[S], value S) {
	v := ReadIORegister(m, flag.Register)
	var mask S = (1<<flag.Size - 1) << flag.Bit
	v &= ^mask
	v |= value << flag.Bit
	SetIORegister(m, flag.Register, v)
}
