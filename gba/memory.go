package gba

import "encoding/binary"

type Memory struct {
	*Motherboard

	Blocks  []BlockData
	blocks  [0x0F]*BlockData
	gpsram  *BlockData
	ioBlock []byte // direct reference to IOR data for fast IO access
}

type BlockData struct {
	MemoryBlock MemoryBlock
	Data        []byte
}

func NewMemory(mm *Motherboard) *Memory {
	m := &Memory{
		Motherboard: mm,
	}

	x4000410 := mb(0x4000400, 0x40004FF, 0x100, [3]bool{true, true, true}, [3]bool{true, true, true}, [3]uint32{1, 1, 1})

	mblocks := []MemoryBlock{BIOS, WRAM1, WRAM2, IOR, Palette, VRAM, OAM, x4000410}

	m.Blocks = make([]BlockData, 0, len(mblocks)+4) // +4 for gpsram + 3 gprom

	for _, block := range mblocks {
		m.Blocks = append(m.Blocks, BlockData{block, make([]byte, block.Size)})
		bd := &m.Blocks[len(m.Blocks)-1]
		idx := bd.MemoryBlock.Start >> 24
		if m.blocks[idx] == nil {
			m.blocks[idx] = bd
		}
	}
	m.ioBlock = m.blocks[0x04].Data

	m.Blocks = append(m.Blocks, BlockData{GPSRAM, make([]byte, GPSRAM.Size)})
	m.gpsram = &m.Blocks[len(m.Blocks)-1]

	GPRom := make([]byte, GPRom1.Size)
	m.Blocks = append(m.Blocks, BlockData{GPRom1, GPRom})
	gprom1 := &m.Blocks[len(m.Blocks)-1]
	m.Blocks = append(m.Blocks, BlockData{GPRom2, GPRom})
	gprom2 := &m.Blocks[len(m.Blocks)-1]
	m.Blocks = append(m.Blocks, BlockData{GPRom3, GPRom})
	gprom3 := &m.Blocks[len(m.Blocks)-1]
	m.blocks[0x08] = gprom1
	m.blocks[0x09] = gprom1
	m.blocks[0x0A] = gprom2
	m.blocks[0x0B] = gprom2
	m.blocks[0x0C] = gprom3
	m.blocks[0x0D] = gprom3

	return m
}

func (m *Memory) findBlock(add uint32) *BlockData {
	if add >= 0x0E000000 {
		return m.gpsram
	}
	return m.blocks[add>>24]
}

func (m *Memory) ReadMemoryBlock(mb MemoryBlock) []byte {
	return m.addrBlockData(mb.Start).Data
}

func (m *Memory) SetMemoryBlock(mb MemoryBlock, value []byte) {
	copy(m.addrBlockData(mb.Start).Data, value)
}

func (m *Memory) addrBlockData(address uint32) *BlockData {
	bd := m.findBlock(address)
	if bd != nil {
		return bd
	}

	return &m.Blocks[0]
}

func vramOffset(address uint32) uint32 {
	offset := address & 0x1FFFF
	if offset >= 0x18000 {
		offset -= 0x8000
	}
	return offset
}

func (m *Memory) block(bd *BlockData, address uint32) ([]byte, uint32) {
	if address>>24 == 0x06 {
		return bd.Data, vramOffset(address)
	}
	offset := address - bd.MemoryBlock.Start
	if bd.MemoryBlock.Mask != 0 {
		offset &= bd.MemoryBlock.Mask
	} else {
		offset %= bd.MemoryBlock.Size
	}
	return bd.Data, offset
}

func (m *Memory) cycle(bd *BlockData, size uint32) {
	m.CPU.cycle(bd.MemoryBlock.Cycles[size])
}

func (m *Memory) checkAffineRefWrite(address uint32) {
	if (address >= 0x4000028 && address <= 0x400002F) ||
		(address >= 0x4000038 && address <= 0x400003F) {
		line := ReadIORegister16(m, VCOUNT)
		if line < 160 {
			m.LCD.OnAffineRefWrite(address)
		}
	}
}

func (m *Memory) checkDMAEnable(address uint32, value uint16) {
	if address < 0x040000BA || address > 0x040000DE {
		return
	}
	var ch int
	switch address {
	case 0x040000BA:
		ch = 0
	case 0x040000C6:
		ch = 1
	case 0x040000D2:
		ch = 2
	case 0x040000DE:
		ch = 3
	default:
		return
	}
	oldVal := m.Read16(address, false, true)
	oldEnable := (oldVal >> 15) & 1
	newEnable := (value >> 15) & 1
	if oldEnable == 0 && newEnable == 1 {
		m.DMA.LatchAddresses(ch)
	}
}

func (m *Memory) checkDMA(address uint32) {
	if address < 0x040000B0 || address > 0x040000DF {
		return
	}

	m.DMA.transfer(DMAImmediate)
}

func (m *Memory) checkIFWrite(address uint32, value uint16) bool {
	ifAddr := uint32(IF)
	if address != ifAddr {
		return false
	}
	current := m.Read16(ifAddr, false, true)
	current &= ^value
	block, offset := m.block(m.addrBlockData(ifAddr), ifAddr)
	block[offset] = uint8(current)
	block[offset+1] = uint8(current >> 8)
	return true
}

func (m *Memory) checkIFWrite32(address uint32, value uint32) bool {
	ieAddr := uint32(IE)
	if address != ieAddr {
		return false
	}
	ieVal := uint16(value)
	bd := m.addrBlockData(ieAddr)
	block, offset := m.block(bd, ieAddr)
	block[offset] = uint8(ieVal)
	block[offset+1] = uint8(ieVal >> 8)

	ifVal := uint16(value >> 16)
	ifAddr := uint32(IF)
	current := m.Read16(ifAddr, false, true)
	current &= ^ifVal
	block2, offset2 := m.block(m.addrBlockData(ifAddr), ifAddr)
	block2[offset2] = uint8(current)
	block2[offset2+1] = uint8(current >> 8)
	return true
}

func (m *Memory) setTimerL(address uint32, value uint16, forceAddr bool) bool {
	switch address {
	case uint32(TM0CNT_L), uint32(TM1CNT_L), uint32(TM2CNT_L), uint32(TM3CNT_L):
		m.Timer.Set(address, value)
		return true
	default:
		return false
	}
}

func (m *Memory) checkSoundWrite(address uint32, value uint32, size uint32) bool {
	if m.APU == nil {
		return false
	}
	if address >= 0x040000A0 && address < 0x040000A8 {
		m.APU.OnFIFOWrite(address, value, size)
		return true
	}
	return false
}

func (m *Memory) afterSoundWrite(address uint32, size uint32) {
	if m.APU == nil || address < 0x04000060 || address > 0x04000088 {
		return
	}
	m.dispatchSound(address &^ 1)
	if size == 4 {
		m.dispatchSound((address &^ 1) + 2)
	}
}

func (m *Memory) dispatchSound(aligned uint32) {
	offset := aligned & 0x3FF
	v16 := uint16(m.ioBlock[offset]) | uint16(m.ioBlock[offset+1])<<8
	switch aligned {
	case uint32(SOUND1CNT_L):
		m.APU.OnSOUND1CNT_L(v16)
	case uint32(SOUND1CNT_H):
		m.APU.OnSOUND1CNT_H(v16)
	case uint32(SOUND1CNT_X):
		m.APU.OnSOUND1CNT_X(v16)
	case uint32(SOUND2CNT_L):
		m.APU.OnSOUND2CNT_L(v16)
	case uint32(SOUND2CNT_H):
		m.APU.OnSOUND2CNT_H(v16)
	case uint32(SOUND3CNT_L):
		m.APU.OnSOUND3CNT_L(v16)
	case uint32(SOUND3CNT_H):
		m.APU.OnSOUND3CNT_H(v16)
	case uint32(SOUND3CNT_X):
		m.APU.OnSOUND3CNT_X(v16)
	case uint32(SOUND4CNT_L):
		m.APU.OnSOUND4CNT_L(v16)
	case uint32(SOUND4CNT_H):
		m.APU.OnSOUND4CNT_H(v16)
	case uint32(SOUNDCNT_H):
		m.APU.OnSOUNDCNT_H(v16)
	}
}

func (m *Memory) checkTimerH(address uint32, value uint16) {
	switch address {
	case uint32(TM0CNT_H), uint32(TM1CNT_H), uint32(TM2CNT_H), uint32(TM3CNT_H):
		m.Timer.OnControlWrite(address, value)
	}
}

func (m *Memory) Read8(address uint32, cycle bool, forceAddr bool) (value uint8) {
	bd := m.addrBlockData(address)
	if cycle {
		m.cycle(bd, 0)
	}
	if address >= 0x0E000000 && address < 0x10000000 {
		return m.Flash.Read(address)
	}
	if address >= 0x080000C4 && address < 0x080000CA && m.GPIO.Readable {
		return m.GPIO.Read(address)
	}
	block, offset := m.block(bd, address)
	return block[offset]
}

func (m *Memory) Set8(address uint32, value uint8, cycle bool, forceAddr bool) {
	bd := m.addrBlockData(address)
	if cycle {
		m.cycle(bd, 0)
	}
	if address>>24 <= 0x03 {
		if !bd.MemoryBlock.Writes[0] {
			return
		}
		block, offset := m.block(bd, address)
		block[offset] = value
		return
	}
	if address >= 0x0E000000 && address < 0x10000000 {
		m.Flash.Write(address, value)
		return
	}
	if address>>24 == 0x06 {
		if vramOffset(address) >= 0x10000 {
			return // OBJ VRAM ignores 8-bit writes
		}
		halfword := uint16(value) | uint16(value)<<8
		aligned := address &^ 1
		block, offset := m.block(bd, aligned)
		block[offset] = uint8(halfword)
		block[offset+1] = uint8(halfword >> 8)
		return
	}
	if address >= 0x05000000 && address < 0x05000400 {
		halfword := uint16(value) | uint16(value)<<8
		aligned := address &^ 1
		block, offset := m.block(bd, aligned)
		block[offset] = uint8(halfword)
		block[offset+1] = uint8(halfword >> 8)
		return
	}
	if address >= 0x080000C4 && address < 0x080000CA {
		m.GPIO.Write(address&^1, value)
		return
	}
	if address == 0x04000301 {
		if value&0x80 == 0 {
			m.CPU.halted = true
		}
		return
	}
	if !bd.MemoryBlock.Writes[0] {
		return
	}
	if m.setTimerL(address, uint16(value), forceAddr) {
		return // 8-bit writes to timer registers are ignored on GBA
	}
	m.checkTimerH(address, uint16(value))
	if m.checkSoundWrite(address, uint32(value), 1) {
		return
	}
	block, offset := m.block(bd, address)
	block[offset] = value
	m.afterSoundWrite(address, 1)
	m.checkAffineRefWrite(address)
	m.checkDMA(address)
}

func (m *Memory) Read16(address uint32, cycle bool, forceAddr bool) (value uint16) {
	bd := m.addrBlockData(address)
	address &= ^uint32(1)
	if cycle {
		m.cycle(bd, 1)
	}
	if address >= 0x080000C4 && address < 0x080000CA && m.GPIO.Readable {
		return uint16(m.GPIO.Read(address))
	}
	if address >= 0x04000100 && address <= 0x0400010E && address&3 == 0 {
		m.Timer.SyncToMemory()
	}
	block, offset := m.block(bd, address)
	value = binary.LittleEndian.Uint16(block[offset:])
	return
}

func (m *Memory) Set16(address uint32, value uint16, cycle bool, forceAddr bool) {
	bd := m.addrBlockData(address)
	if top := address >> 24; top != 0x04 && top != 0x08 {
		if !bd.MemoryBlock.Writes[1] {
			return
		}
		address &= ^uint32(1)
		if cycle {
			m.cycle(bd, 1)
		}
		block, offset := m.block(bd, address)
		binary.LittleEndian.PutUint16(block[offset:], value)
		return
	}
	if address >= 0x080000C4 && address < 0x080000CA {
		m.GPIO.Write(address&^1, uint8(value))
		return
	}
	if !bd.MemoryBlock.Writes[1] {
		return
	}
	address &= ^uint32(1)
	if cycle {
		m.cycle(bd, 1)
	}
	if m.setTimerL(address, value, forceAddr) {
		return
	}
	m.checkTimerH(address, value)
	m.checkDMAEnable(address, value)
	if !forceAddr && m.checkIFWrite(address, value) {
		return
	}
	if m.checkSoundWrite(address, uint32(value), 2) {
		return
	}
	block, offset := m.block(bd, address)
	block[offset] = uint8(value)
	block[offset+1] = uint8(value >> 8)
	m.afterSoundWrite(address, 2)
	m.checkAffineRefWrite(address)
	m.checkDMA(address)
}

func (m *Memory) Read32(address uint32, cycle bool, forceAddr bool) (value uint32) {
	rotate := (address & 3) * 8
	bd := m.addrBlockData(address)
	address &= ^uint32(3)
	if cycle {
		m.cycle(bd, 2)
	}
	block, offset := m.block(bd, address)
	value = binary.LittleEndian.Uint32(block[offset:])
	if rotate > 0 {
		value = (value >> rotate) | (value << (32 - rotate))
	}
	return
}

func (m *Memory) Set32(address uint32, value uint32, cycle bool, forceAddr bool) {
	bd := m.addrBlockData(address)
	if top := address >> 24; top != 0x04 && top != 0x08 {
		if !bd.MemoryBlock.Writes[2] {
			return
		}
		address &= ^uint32(3)
		if cycle {
			m.cycle(bd, 2)
		}
		block, offset := m.block(bd, address)
		binary.LittleEndian.PutUint32(block[offset:], value)
		return
	}
	if address >= 0x080000C4 && address < 0x080000CA {
		m.GPIO.Write(address&^3, uint8(value))
		m.GPIO.Write((address&^3)+2, uint8(value>>16))
		return
	}
	if !bd.MemoryBlock.Writes[2] {
		return
	}
	address &= ^uint32(3)
	if cycle {
		m.cycle(bd, 2)
	}
	if m.setTimerL(address, uint16(value), forceAddr) {
		m.checkTimerH(address+2, uint16(value>>16))
		block, offset := m.block(bd, address+2)
		block[offset] = uint8(value >> 16)
		block[offset+1] = uint8(value >> 24)
		return
	}
	m.checkTimerH(address, uint16(value))
	m.checkDMAEnable(address+2, uint16(value>>16))
	if !forceAddr && m.checkIFWrite32(address, value) {
		return
	}
	if !forceAddr && m.checkIFWrite(address, uint16(value)) {
		return
	}
	if m.checkSoundWrite(address, value, 4) {
		return
	}
	block, offset := m.block(bd, address)
	block[offset] = uint8(value)
	block[offset+1] = uint8(value >> 8)
	block[offset+2] = uint8(value >> 16)
	block[offset+3] = uint8(value >> 24)
	m.afterSoundWrite(address, 4)
	m.checkAffineRefWrite(address)
	m.checkDMA(address)
}

func (m *Memory) ClearBlock(mb MemoryBlock) {
	clear(m.addrBlockData(mb.Start).Data)
}

func ReadIORegister16(m *Memory, r IORegister[uint16]) uint16 {
	offset := uint32(r) & 0x3FF
	return binary.LittleEndian.Uint16(m.ioBlock[offset:])
}

func SetIORegister16(m *Memory, r IORegister[uint16], value uint16) {
	offset := uint32(r) & 0x3FF
	binary.LittleEndian.PutUint16(m.ioBlock[offset:], value)
}

func ReadIORegister[S Size](m *Memory, r IORegister[S]) S {
	v := *new(S)
	switch t := any(v).(type) {
	case uint8:
		v = S(m.Read8(uint32(r), false, true))
	case uint16:
		v = S(ReadIORegister16(m, IORegister[uint16](r)))
	case uint32:
		v = S(m.Read32(uint32(r), false, true))
	default:
		panic(t)
	}
	return v
}

func SetIORegister[S Size](m *Memory, r IORegister[S], value S) {
	switch t := any(value).(type) {
	case uint8:
		m.Set8(uint32(r), uint8(value), false, true)
	case uint16:
		SetIORegister16(m, IORegister[uint16](r), uint16(value))
	case uint32:
		m.Set32(uint32(r), uint32(value), false, true)
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
	return ReadBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size)
}

func SetFlag[S Size](m *Memory, flag IOFlag[S], value S) {
	SetIORegister(m, flag.Register, SetBits(ReadIORegister(m, flag.Register), flag.Bit, flag.Size, value))
}
