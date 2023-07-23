package gba

type Memory struct {
	*Motherboard

	Blocks map[MemoryBlock][]byte
}

func NewMemory(mm *Motherboard) *Memory {
	m := &Memory{
		Motherboard: mm,
		Blocks:      make(map[MemoryBlock][]byte),
	}

	blocks := []MemoryBlock{BIOS, WRAM1, WRAM2, IOR, Palette, VRAM, OAM, GPSRAM}

	for _, block := range blocks {
		m.Blocks[block] = make([]byte, block.Size)
	}

	GPRom := make([]byte, GPRom1.Size)
	m.Blocks[GPRom1] = GPRom
	m.Blocks[GPRom2] = GPRom
	m.Blocks[GPRom3] = GPRom

	x4000410 := MemoryBlock{
		Start:  0x4000410,
		End:    0x4000410,
		Size:   1,
		Cycles: [3]uint32{1, 1, 1},
	}
	m.Blocks[x4000410] = make([]byte, 1)

	return m
}

func (m Memory) ReadMemoryBlock(mb MemoryBlock) []byte {
	return m.Blocks[mb]
}

func (m Memory) SetMemoryBlock(mb MemoryBlock, value []byte) {
	copy(m.Blocks[mb], value)
}

func (m Memory) addrMemoryBlock(address uint32) MemoryBlock {
	for mb := range m.Blocks {
		if address < mb.Start || address > mb.End {
			continue
		}

		return mb
	}

	panic(address)
}

func (m Memory) addrBlock(address uint32) ([]byte, uint32) {
	mb := m.addrMemoryBlock(address)
	return m.Blocks[mb], (address - mb.Start) % mb.Size
}

func (m Memory) cycle(address uint32, size uint32) {
	m.CPU.cycle(m.addrMemoryBlock(address).Cycles[size])
}

func (m Memory) addrByte(address uint32) *byte {
	block, offset := m.addrBlock(address)
	return &block[offset]
}

func (m Memory) checkDMA(address uint32) {
	if address < 0x040000B0 || address > 0x040000DF {
		return
	}

	m.DMA.transfer(DMAImmediate)
}

func (m Memory) Get8(address uint32) (value uint8) {
	m.cycle(address, 0)
	return *m.addrByte(address)
}

func (m Memory) Set8(address uint32, value uint8) {
	m.cycle(address, 0)
	defer m.checkDMA(address)
	*m.addrByte(address) = value
}

func (m Memory) Get16(address uint32) (value uint16) {
	address &= ^uint32(1)
	m.cycle(address, 1)
	value = uint16(*m.addrByte(address))
	value |= uint16(*m.addrByte(address + 1)) << 8
	return
}

func (m Memory) Set16(address uint32, value uint16) {
	address &= ^uint32(1)
	m.cycle(address, 1)
	defer m.checkDMA(address)
	*m.addrByte(address) = uint8(value)
	*m.addrByte(address + 1) = uint8(value >> 8)
}

func (m Memory) Get32(address uint32) (value uint32) {
	address &= ^uint32(1)
	m.cycle(address, 2)
	value = uint32(*m.addrByte(address))
	value |= uint32(*m.addrByte(address + 1)) << 8
	value |= uint32(*m.addrByte(address + 2)) << 16
	value |= uint32(*m.addrByte(address + 3)) << 24
	return
}

func (m Memory) Set32(address uint32, value uint32) {
	address &= ^uint32(1)
	m.cycle(address, 2)
	defer m.checkDMA(address)
	*m.addrByte(address) = uint8(value)
	*m.addrByte(address + 1) = uint8(value >> 8)
	*m.addrByte(address + 2) = uint8(value >> 16)
	*m.addrByte(address + 3) = uint8(value >> 24)
}

func ReadIORegister[S Size](m *Memory, r IORegister[S]) S {
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

func SetIORegister[S Size](m *Memory, r IORegister[S], value S) {
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

func ReadFlag[S Size](m *Memory, flag IOFlag[S]) S {
	return ReadBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size)
}

func SetFlag[S Size](m *Memory, flag IOFlag[S], value S) {
	SetIORegister(m, flag.Register, SetBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size, value))
}
