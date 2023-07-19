package gba

type Memory struct {
	*Motherboard

	BIOS    []byte
	WRAM1   []byte
	WRAM2   []byte
	Palette []byte
	VRAM    []byte
	OAM     []byte
	GPROM1  []byte
	GPROM2  []byte
	GPROM3  []byte
	SRAM    []byte
	IO      []byte
}

func NewMemory(mm *Motherboard) Memory {
	return Memory{
		Motherboard: mm,
		BIOS:        make([]byte, 16*k),
		WRAM1:       make([]byte, 256*k),
		WRAM2:       make([]byte, 32*k),
		Palette:     make([]byte, 1*k),
		VRAM:        make([]byte, 96*k),
		OAM:         make([]byte, 1*k),
		GPROM1:      make([]byte, 32*m),
		GPROM2:      make([]byte, 32*m),
		GPROM3:      make([]byte, 32*m),
		SRAM:        make([]byte, 64*k),
		IO:          make([]byte, 1*k),
	}
}

func (m Memory) addrBlock(address uint32) ([]byte, uint32) {
	switch {
	case address <= 0x00003FFF:
		return m.BIOS[:], 0x00000000
	case address >= 0x02000000 && address <= 0x02FFFFFF:
		return m.WRAM1[:], address & ^(256*k - 1)
	case address >= 0x03000000 && address <= 0x03FFFFFF:
		return m.WRAM2[:], address & ^(32*k - 1)
	case address >= 0x05000000 && address <= 0x05FFFFFF:
		return m.Palette[:], address & ^(1*k - 1)
	case address >= 0x06000000 && address <= 0x06017FFF:
		return m.VRAM[:], 0x06000000
	case address >= 0x07000000 && address <= 0x07FFFFFF:
		return m.OAM[:], address & ^(64*k - 1)
	case address >= 0x08000000 && address <= 0x09FFFFFF:
		return m.GPROM1[:], 0x08000000
	case address >= 0x0A000000 && address <= 0x0BFFFFFF:
		return m.GPROM2[:], 0x0A000000
	case address >= 0x0C000000 && address <= 0x0DFFFFFF:
		return m.GPROM3[:], 0x0C000000
	case address >= 0x0E000000:
		return m.SRAM[:], address & ^(64*k - 1)
	case address >= 0x04000000 && address <= 0x040003FE: // todo: investigate io
		return m.IO[:], 0x04000000
	case address == 0x04000410:
		return make([]byte, 4), 0x4000410
	default:
		panic(address)
	}
}

func (m Memory) addrByte(address uint32) *byte {
	block, offset := m.addrBlock(address)
	return &block[address-offset]
}

func (m Memory) checkDMA(address uint32) {
	if address < 0x040000B0 || address > 0x040000DF {
		return
	}

	m.DMA.transfer(DMAImmediate)
}

func (m Memory) Get8(address uint32) (value uint8) {
	return *m.addrByte(address)
}

func (m Memory) Set8(address uint32, value uint8) {
	defer m.checkDMA(address)
	*m.addrByte(address) = value
}

func (m Memory) Get16(address uint32) (value uint16) {
	address &= ^uint32(1)
	value = uint16(*m.addrByte(address))
	value |= uint16(*m.addrByte(address + 1)) << 8
	return
}

func (m Memory) Set16(address uint32, value uint16) {
	defer m.checkDMA(address)
	address &= ^uint32(1)
	*m.addrByte(address) = uint8(value)
	*m.addrByte(address + 1) = uint8(value >> 8)
}

func (m Memory) Get32(address uint32) (value uint32) {
	address &= ^uint32(1)
	value = uint32(*m.addrByte(address))
	value |= uint32(*m.addrByte(address + 1)) << 8
	value |= uint32(*m.addrByte(address + 2)) << 16
	value |= uint32(*m.addrByte(address + 3)) << 24
	return
}

func (m Memory) Set32(address uint32, value uint32) {
	defer m.checkDMA(address)
	address &= ^uint32(1)
	*m.addrByte(address) = uint8(value)
	*m.addrByte(address + 1) = uint8(value >> 8)
	*m.addrByte(address + 2) = uint8(value >> 16)
	*m.addrByte(address + 3) = uint8(value >> 24)
}

func (m Memory) GetSlice(address uint32, size uint32) (value []byte) {
	block, offset := m.addrBlock(address)
	address -= offset
	return block[address : address+size]
}

func (m Memory) SetSlice(address uint32, value []byte) {
	defer m.checkDMA(address)
	block, offset := m.addrBlock(address)
	address -= offset
	for i := uint32(0); i < uint32(len(value)); i++ {
		block[address+i] = value[i]
	}
}

type MemoryBlock [2]uint32

func ReadMemoryBlock(m Memory, mb MemoryBlock) []byte {
	return m.GetSlice(mb[0], mb[1]-mb[0])
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
		return S(m.Get8(uint32(r)))
	case uint16:
		return S(m.Get16(uint32(r)))
	case uint32:
		return S(m.Get32(uint32(r)))
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
