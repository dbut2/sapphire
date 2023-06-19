package gba

const (
	b = 1 << (iota * 10)
	kb
	mb
)

type Memory [0x10000000]byte

func (m Memory) Access8(address uint32) uint8 {
	v := uint8(0)
	v += uint8(m[address])
	return v
}

func (m Memory) Access16(address uint32) uint16 {
	v := uint16(0)
	v += uint16(m[address])
	v += uint16(m[address+1]) << 8
	return v
}

func (m Memory) Access32(address uint32) uint32 {
	v := uint32(0)
	v += uint32(m[address])
	v += uint32(m[address+1]) << 8
	v += uint32(m[address+2]) << 16
	v += uint32(m[address+3]) << 24
	return v
}

type Register[S Size] uint32

type Size interface {
	uint8 | uint16 | uint32
}

func ReadRegister[S Size](m Memory, r Register[S]) S {
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
