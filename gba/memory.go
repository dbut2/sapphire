package gba

type Memory struct {
	*Motherboard

	Blocks []BlockData
}

type BlockData struct {
	MemoryBlock MemoryBlock
	Data        []byte
}

func NewMemory(mm *Motherboard) *Memory {
	m := &Memory{
		Motherboard: mm,
	}

	// BIOS writes here
	x4000410 := MemoryBlock{
		Start:  0x4000410,
		End:    0x4000410,
		Size:   1,
		Cycles: [3]uint32{1, 1, 1},
	}

	blocks := []MemoryBlock{BIOS, WRAM1, WRAM2, IOR, Palette, VRAM, OAM, GPSRAM, x4000410}

	for _, block := range blocks {
		m.Blocks = append(m.Blocks, BlockData{block, make([]byte, block.Size)})
	}

	GPRom := make([]byte, GPRom1.Size)
	m.Blocks = append(m.Blocks, BlockData{GPRom1, GPRom})
	m.Blocks = append(m.Blocks, BlockData{GPRom2, GPRom})
	m.Blocks = append(m.Blocks, BlockData{GPRom3, GPRom})

	return m
}

func (m Memory) ReadMemoryBlock(mb MemoryBlock) []byte {
	return m.addrBlockData(mb.Start).Data
}

func (m Memory) SetMemoryBlock(mb MemoryBlock, value []byte) {
	copy(m.addrBlockData(mb.Start).Data, value)
}

func (m Memory) addrBlockData(address uint32) BlockData {
	for _, bd := range m.Blocks {
		if address < bd.MemoryBlock.Start || address > bd.MemoryBlock.End {
			continue
		}

		return bd
	}

	panic(address)
}

func (m Memory) block(bd BlockData, address uint32) ([]byte, uint32) {
	return bd.Data, (address - bd.MemoryBlock.Start) % bd.MemoryBlock.Size
}

func (m Memory) cycle(bd BlockData, size uint32) {
	m.CPU.cycle(bd.MemoryBlock.Cycles[size])
}

func (m Memory) checkDMA(address uint32) {
	if address < 0x040000B0 || address > 0x040000DF {
		return
	}

	m.DMA.transfer(DMAImmediate)
}

func (m Memory) Get8(address uint32) (value uint8) {
	bd := m.addrBlockData(address)
	//if !bd.MemoryBlock.Reads[0] {
	//	panic(fmt.Sprintf("cannot read 8 bits from %08X", address))
	//}
	m.cycle(bd, 0)
	block, offset := m.block(bd, address)
	return block[offset]
}

func (m Memory) Set8(address uint32, value uint8) {
	bd := m.addrBlockData(address)
	//if !bd.MemoryBlock.Writes[0] {
	//	panic(fmt.Sprintf("cannot write 8 bits to %08X", address))
	//}
	m.cycle(bd, 0)
	block, offset := m.block(bd, address)
	block[offset] = value
	m.checkDMA(address)
}

func (m Memory) Get16(address uint32) (value uint16) {
	bd := m.addrBlockData(address)
	//if !bd.MemoryBlock.Reads[1] {
	//	panic(fmt.Sprintf("cannot read 16 bits from %08X", address))
	//}
	address &= ^uint32(1)
	m.cycle(bd, 1)
	block, offset := m.block(bd, address)
	value = uint16(block[offset])
	block2, offset2 := m.block(bd, address+1)
	value |= uint16(block2[offset2]) << 8
	return
}

func (m Memory) Set16(address uint32, value uint16) {
	bd := m.addrBlockData(address)
	//if !bd.MemoryBlock.Writes[1] {
	//	panic(fmt.Sprintf("cannot write 16 bits to %08X", address))
	//}
	address &= ^uint32(1)
	m.cycle(bd, 1)
	block, offset := m.block(bd, address)
	block[offset] = uint8(value)
	block[offset+1] = uint8(value >> 8)
	m.checkDMA(address)
}

func (m Memory) Get32(address uint32) (value uint32) {
	bd := m.addrBlockData(address)
	//if !bd.MemoryBlock.Reads[2] {
	//	panic(fmt.Sprintf("cannot read 32 bits from %08X", address))
	//}
	address &= ^uint32(1)
	m.cycle(bd, 2)
	block, offset := m.block(bd, address)
	value = uint32(block[offset])
	value |= uint32(block[offset+1]) << 8
	value |= uint32(block[offset+2]) << 16
	value |= uint32(block[offset+3]) << 24
	return
}

func (m Memory) Set32(address uint32, value uint32) {
	bd := m.addrBlockData(address)
	//if !bd.MemoryBlock.Writes[2] {
	//	panic(fmt.Sprintf("cannot write 32 bits to %08X", address))
	//}
	address &= ^uint32(1)
	m.cycle(bd, 2)
	block, offset := m.block(bd, address)
	block[offset] = uint8(value)
	block[offset+1] = uint8(value >> 8)
	block[offset+2] = uint8(value >> 16)
	block[offset+3] = uint8(value >> 24)
	m.checkDMA(address)
}

func (m Memory) ClearBlock(mb MemoryBlock) {
	clear(m.addrBlockData(mb.Start).Data)
}

func GetIORegister[S Size](m *Memory, r IORegister[S]) S {
	v := *new(S)
	switch t := any(v).(type) {
	case uint8:
		v = S(m.Get8(uint32(r)))
	case uint16:
		v = S(m.Get16(uint32(r)))
	case uint32:
		v = S(m.Get32(uint32(r)))
	default:
		panic(t)
	}
	return v
}

func SetIORegister[S Size](m *Memory, r IORegister[S], value S) {
	switch t := any(value).(type) {
	case uint8:
		m.Set8(uint32(r), uint8(value))
	case uint16:
		m.Set16(uint32(r), uint16(value))
	case uint32:
		m.Set32(uint32(r), uint32(value))
	default:
		panic(t)
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
	return ReadBits(GetIORegister(m, flag.Register), flag.Bit, flag.Size)
}

func SetFlag[S Size](m *Memory, flag IOFlag[S], value S) {
	SetIORegister(m, flag.Register, SetBits(GetIORegister(m, flag.Register), flag.Bit, flag.Size, value))
}
