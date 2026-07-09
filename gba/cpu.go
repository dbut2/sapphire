package gba

import (
	"encoding/binary"
	"fmt"
	"math"
)

type CPU struct {
	*Motherboard
	CPURegisters

	curr, next uint32
	flushed    bool

	cycles uint32
	halted bool

	fetchData  []byte
	fetchStart uint32
	fetchMask  uint32
	fetchTop   uint32
	fetchCyc16 uint32
	fetchCyc32 uint32
}

func NewCPU(m *Motherboard) *CPU {
	return &CPU{Motherboard: m}
}

func (c *CPU) cycle(n uint32) {
	c.cycles += n
}

func (c *CPU) pcInc() {
	switch c.cpsrState() {
	case 0:
		c.R[15] += 4
	case 1:
		c.R[15] += 2
	}
}

func (c *CPU) Step() {
	curr := c.curr

	switch c.cpsrState() {
	case 0:
		instruction := c.fetch32(curr &^ 3)
		c.Arm(instruction)
	case 1:
		instruction := c.fetch16(curr &^ 1)
		c.Thumb(uint32(instruction))
	}

	if !c.flushed {
		c.curr = c.next
		c.next = c.R[15]

		c.pcInc()
	}
	c.flushed = false
}

func (c *CPU) fetch32(addr uint32) uint32 {
	if addr>>24 != c.fetchTop || c.fetchData == nil {
		if !c.cacheFetchBlock(addr) {
			return c.Memory.Read32(addr, true, false)
		}
	}
	c.cycles += c.fetchCyc32
	off := (addr - c.fetchStart) & c.fetchMask
	return binary.LittleEndian.Uint32(c.fetchData[off:])
}

func (c *CPU) fetch16(addr uint32) uint16 {
	if addr>>24 != c.fetchTop || c.fetchData == nil {
		if !c.cacheFetchBlock(addr) {
			return c.Memory.Read16(addr, true, false)
		}
	}
	c.cycles += c.fetchCyc16
	off := (addr - c.fetchStart) & c.fetchMask
	return binary.LittleEndian.Uint16(c.fetchData[off:])
}

func (c *CPU) cacheFetchBlock(addr uint32) bool {
	top := addr >> 24
	if top >= 0x04 && top <= 0x07 {
		c.fetchData = nil
		return false
	}
	bd := c.Memory.findBlock(addr)
	if bd == nil || bd.MemoryBlock.Mask == 0 {
		c.fetchData = nil
		return false
	}
	c.fetchData = bd.Data
	c.fetchStart = bd.MemoryBlock.Start
	c.fetchMask = bd.MemoryBlock.Mask
	c.fetchTop = top
	c.fetchCyc16 = bd.MemoryBlock.Cycles[1]
	c.fetchCyc32 = bd.MemoryBlock.Cycles[2]
	return true
}

func (c *CPU) noins(instruction uint32) {
	fmt.Printf("UND: instr=%08X PC=%08X CPSR=%08X state=%d\n", instruction, c.curr, c.CPSR, c.cpsrState())
	c.exception(0x04)
}

func (c *CPU) prefetchFlush() {
	switch c.cpsrState() {
	case 0:
		c.R[15] &= ^uint32(3)
	case 1:
		c.R[15] &= ^uint32(1)
	}
	c.curr = c.R[15]
	c.pcInc()
	c.next = c.R[15]
	c.pcInc()
	c.flushed = true
}

type CPURegisters struct {
	R    [16]uint32
	CPSR uint32

	R0       uint32
	R1       uint32
	R2       uint32
	R3       uint32
	R4       uint32
	R5       uint32
	R6       uint32
	R7       uint32
	R8       uint32
	R9       uint32
	R10      uint32
	R11      uint32
	R12      uint32
	R13      uint32
	R14      uint32
	R15      uint32
	R8_fiq   uint32
	R9_fiq   uint32
	R10_fiq  uint32
	R11_fiq  uint32
	R12_fiq  uint32
	R13_fiq  uint32
	R13_svc  uint32
	R13_abt  uint32
	R13_irq  uint32
	R13_und  uint32
	R14_fiq  uint32
	R14_svc  uint32
	R14_abt  uint32
	R14_irq  uint32
	R14_und  uint32
	SPSR_fiq uint32
	SPSR_svc uint32
	SPSR_abt uint32
	SPSR_irq uint32
	SPSR_und uint32
}

func (c *CPU) registerAddr(mode uint32, r uint32) *uint32 {
	switch r {
	case 0:
		return &c.R0
	case 1:
		return &c.R1
	case 2:
		return &c.R2
	case 3:
		return &c.R3
	case 4:
		return &c.R4
	case 5:
		return &c.R5
	case 6:
		return &c.R6
	case 7:
		return &c.R7
	case 8:
		if mode == FIQ {
			return &c.R8_fiq
		}
		return &c.R8
	case 9:
		if mode == FIQ {
			return &c.R9_fiq
		}
		return &c.R9
	case 10:
		if mode == FIQ {
			return &c.R10_fiq
		}
		return &c.R10
	case 11:
		if mode == FIQ {
			return &c.R11_fiq
		}
		return &c.R11
	case 12:
		if mode == FIQ {
			return &c.R12_fiq
		}
		return &c.R12
	case 13:
		switch mode {
		case FIQ:
			return &c.R13_fiq
		case IRQ:
			return &c.R13_irq
		case SVC:
			return &c.R13_svc
		case ABT:
			return &c.R13_abt
		case UND:
			return &c.R13_und
		default:
			return &c.R13
		}
	case 14:
		switch mode {
		case FIQ:
			return &c.R14_fiq
		case IRQ:
			return &c.R14_irq
		case SVC:
			return &c.R14_svc
		case ABT:
			return &c.R14_abt
		case UND:
			return &c.R14_und
		default:
			return &c.R14
		}
	default:
		return &c.R15
	}
}

func (c *CPU) spsrAddr(mode uint32) *uint32 {
	switch mode {
	case FIQ:
		return &c.SPSR_fiq
	case IRQ:
		return &c.SPSR_irq
	case SVC:
		return &c.SPSR_svc
	case ABT:
		return &c.SPSR_abt
	case UND:
		return &c.SPSR_und
	default:
		return &c.CPSR
	}
}

const (
	USR uint32 = 0b10000
	FIQ uint32 = 0b10001
	IRQ uint32 = 0b10010
	SVC uint32 = 0b10011
	ABT uint32 = 0b10111
	UND uint32 = 0b11011
	SYS uint32 = 0b11111
)

func (c *CPU) restoreCpsr() {
	cpsr := *c.spsrAddr(c.cpsrMode())
	newMode := ReadBits(cpsr, 0, 5)
	c.cpsrSetMode(newMode)
	c.CPSR = cpsr
}

func (c *CPU) cpsrMode() uint32 {
	return ReadBits(c.CPSR, 0, 5)
}

//func (c *CPU) cpsrInitMode(value uint32) {
//	c.CPSR = SetBits(c.CPSR, 0, 5, value)
//}

func (c *CPU) cpsrSetMode(value uint32) {
	prevMode := c.cpsrMode() | 0b10000
	nextMode := value | 0b10000

	if prevMode == nextMode {
		c.CPSR = SetBits(c.CPSR, 0, 5, value)
		return
	}

	for i := uint32(0); i <= 15; i++ {
		*c.registerAddr(prevMode, i) = c.R[i]
	}

	for i := uint32(0); i <= 15; i++ {
		c.R[i] = *c.registerAddr(nextMode, i)
	}

	c.CPSR = SetBits(c.CPSR, 0, 5, value)
}

func (c *CPU) cpsrState() uint32 {
	return ReadBits(c.CPSR, 5, 1)
}

func (c *CPU) cpsrSetState(value uint32) {
	c.CPSR = SetBits(c.CPSR, 5, 1, value)
}

func (c *CPU) cpsrIRQDisable() uint32 {
	return ReadBits(c.CPSR, 7, 1)
}

func (c *CPU) cpsrSetIRQDisable(value uint32) {
	c.CPSR = SetBits(c.CPSR, 7, 1, value)
}

//func (c *CPU) cpsrFIQDisable() uint32 {
//	return ReadBits(c.CPSR, 6, 1)
//}

func (c *CPU) cpsrSetFIQDisable(value uint32) {
	c.CPSR = SetBits(c.CPSR, 6, 1, value)
}

func (c *CPU) cpsrN() uint32 {
	return ReadBits(c.CPSR, 31, 1)
}

func (c *CPU) cpsrSetN(value bool) {
	c.CPSR = SetBits(c.CPSR, 31, 1, bool2uint32(value))
}

func (c *CPU) cpsrZ() uint32 {
	return ReadBits(c.CPSR, 30, 1)
}

func (c *CPU) cpsrSetZ(value bool) {
	c.CPSR = SetBits(c.CPSR, 30, 1, bool2uint32(value))
}

func (c *CPU) cpsrC() uint32 {
	return ReadBits(c.CPSR, 29, 1)
}

func (c *CPU) cpsrSetC(value bool) {
	c.CPSR = SetBits(c.CPSR, 29, 1, bool2uint32(value))
}

func (c *CPU) cpsrV() uint32 {
	return ReadBits(c.CPSR, 28, 1)
}

func (c *CPU) cpsrSetV(value bool) {
	c.CPSR = SetBits(c.CPSR, 28, 1, bool2uint32(value))
}

func (c *CPU) exception(vector uint32) {
	oldCPSR := c.CPSR

	var newMode uint32
	switch vector {
	case 0x00: // reset
		newMode = SVC
	case 0x04: // undefined
		newMode = UND
	case 0x08: // swi
		newMode = SVC
	case 0x0C: // prefetch abort
		newMode = ABT
	case 0x10: // data
		newMode = ABT
	case 0x14: // address exceed
		newMode = SVC
	case 0x18: // irq
		newMode = IRQ
	case 0x1C: // fiq
		newMode = FIQ
	}

	c.cpsrSetMode(newMode)
	*c.spsrAddr(newMode) = oldCPSR

	switch vector {
	case 0x08, 0x04:
		c.R[14] = c.next
	default:
		c.R[14] = c.curr + 4
	}

	c.cpsrSetState(0)
	c.cpsrSetIRQDisable(1)
	switch vector {
	case 0x00, 0x1C:
		c.cpsrSetFIQDisable(1)
	}

	c.R[15] = vector
	c.prefetchFlush()
}

func (c *CPU) cond(cond uint32) bool {
	return condTable[cond<<4|c.CPSR>>28]
}

var condTable = func() (t [256]bool) {
	for cond := uint32(0); cond < 16; cond++ {
		for flags := uint32(0); flags < 16; flags++ {
			t[cond<<4|flags] = evalCond(cond, flags>>3&1, flags>>2&1, flags>>1&1, flags&1)
		}
	}
	return
}()

func evalCond(cond, N, Z, C, V uint32) bool {
	switch cond {
	case 0b0000: // EQ Z=1 equal (zero) (same)
		return Z == 1
	case 0b0001: // NE Z=1 not equal (nonzero) (not same)
		return Z == 0
	case 0b0010: // CS/HS C=1 unsigned higher or same (carry set)
		return C == 1
	case 0b0011: // CC/LO C=0 unsigned lower (carry cleared)
		return C == 0
	case 0b0100: // MI N=1 signed negative (minus)
		return N == 1
	case 0b0101: // PL N=0 signed positive or zero (plus)
		return N == 0
	case 0b0110: // VS V=1 signed overflow (V set)
		return V == 1
	case 0b0111: // VC V=0 signed no overflow (V cleared)
		return V == 0
	case 0b1000: // HI C=1 and Z=0 unsigned higher
		return C == 1 && Z == 0
	case 0b1001: // LS C=0 or Z=1 unsigned lower or same
		return C == 0 || Z == 1
	case 0b1010: // GE N=V signed greater or equal
		return N == V
	case 0b1011: // LT N<>V signed less than
		return N != V
	case 0b1100: // GT Z=0 and N=V signed greater than
		return Z == 0 && N == V
	case 0b1101: // LE Z=1 or N<>V signed less or equal
		return Z == 1 || N != V
	case 0b1110: // AL - always (the "AL" suffix can be omitted)
		return true
	default:
		return false
	}
}

const (
	SoftReset         uint32 = 0x00
	RegisterRamReset  uint32 = 0x01
	SWIHalt           uint32 = 0x02
	SWIIntrWait       uint32 = 0x04
	SWIVBlankIntrWait uint32 = 0x05
	SWIDiv            uint32 = 0x06
	SWIDivArm         uint32 = 0x07
	SWISqrt           uint32 = 0x08
	SWIArcTan         uint32 = 0x09
	SWIArcTan2        uint32 = 0x0A
	CpuSet            uint32 = 0x0B
	SWICpuFastSet     uint32 = 0x0C
	SWIBitUnPack      uint32 = 0x10
	SWILz77UnCompWram uint32 = 0x11
	SWILz77UnCompVram uint32 = 0x12
	SWIRLUnCompWram   uint32 = 0x14
	SWIRLUnCompVram   uint32 = 0x15
	SWIObjAffineSet   uint32 = 0x0F
	SWIBgAffineSet    uint32 = 0x0E
)

func (c *CPU) SWI(comment uint32) {
	switch comment {
	case SoftReset:
		c.R13 = 0x03007F00
		c.R13_svc = 0x03007FE0
		c.R13_irq = 0x03007FA0
		flag := c.Memory.Read8(0x3007FFA, true, false)
		for i := uint32(0x3007E00); i <= 0x3007FFF; i++ {
			c.Memory.Set8(i, 0, true, false)
		}
		if flag == 0 {
			c.R[14] = 0x08000000
		} else {
			c.R[14] = 0x02000000
		}
		c.cpsrSetState(0)
		c.R[15] = c.R[14]
		c.prefetchFlush()
		return

	case RegisterRamReset:
		flags := c.R[0]
		if flags&0x01 != 0 {
			c.Memory.ClearBlock(WRAM1)
		}
		if flags&0x02 != 0 {
			c.Memory.ClearBlock(WRAM2)
		}
		if flags&0x04 != 0 {
			c.Memory.ClearBlock(Palette)
		}
		if flags&0x08 != 0 {
			c.Memory.ClearBlock(VRAM)
		}
		if flags&0x10 != 0 {
			c.Memory.ClearBlock(OAM)
		}
		SetIORegister(c.Memory, DISPCNT, uint16(0x0080))

	case SWIHalt:
		c.halted = true

	case SWIIntrWait:
		if c.R[0] == 1 {
			ifReg := ReadIORegister(c.Memory, IF)
			ifReg &= ^uint16(c.R[1])
			SetIORegister(c.Memory, IF, ifReg)
		}
		c.halted = true

	case SWIVBlankIntrWait:
		ifReg := ReadIORegister(c.Memory, IF)
		ifReg &= ^uint16(1)
		SetIORegister(c.Memory, IF, ifReg)
		c.halted = true

	case SWIDiv:
		num := int32(c.R[0])
		den := int32(c.R[1])
		if den == 0 {
			c.R[0] = 0
			c.R[1] = 0
			c.R[3] = 0
		} else {
			c.R[0] = uint32(num / den)
			c.R[1] = uint32(num % den)
			result := num / den
			if result < 0 {
				c.R[3] = uint32(-result)
			} else {
				c.R[3] = uint32(result)
			}
		}

	case SWIDivArm:
		num := int32(c.R[1])
		den := int32(c.R[0])
		if den == 0 {
			c.R[0] = 0
			c.R[1] = 0
			c.R[3] = 0
		} else {
			c.R[0] = uint32(num / den)
			c.R[1] = uint32(num % den)
			result := num / den
			if result < 0 {
				c.R[3] = uint32(-result)
			} else {
				c.R[3] = uint32(result)
			}
		}

	case SWISqrt:
		c.R[0] = uint32(math.Sqrt(float64(c.R[0])))

	case SWIArcTan:
		tan := float64(int16(c.R[0])) / 16384.0
		result := math.Atan(tan) / (2 * math.Pi) * 0x10000
		c.R[0] = uint32(int16(result))

	case SWIArcTan2:
		x := float64(int16(c.R[0]))
		y := float64(int16(c.R[1]))
		result := math.Atan2(y, x) / (2 * math.Pi) * 0x10000
		if result < 0 {
			result += 0x10000
		}
		c.R[0] = uint32(uint16(result))

	case CpuSet:
		source := c.R[0]
		destination := c.R[1]
		datasize := ReadBits(c.R[2], 26, 1)
		fill := ReadBits(c.R[2], 24, 1)
		count := ReadBits(c.R[2], 0, 21)

		switch {
		case fill == 0 && datasize == 0:
			for i := uint32(0); i < count; i++ {
				offset := i << 1
				value := c.Memory.Read16(source+offset, false, false)
				c.Memory.Set16(destination+offset, value, false, false)
			}
		case fill == 0 && datasize == 1:
			for i := uint32(0); i < count; i++ {
				offset := i << 2
				value := c.Memory.Read32(source+offset, false, false)
				c.Memory.Set32(destination+offset, value, false, false)
			}
		case fill == 1 && datasize == 0:
			value := c.Memory.Read16(source, false, false)
			for i := uint32(0); i < count; i++ {
				offset := i << 1
				c.Memory.Set16(destination+offset, value, false, false)
			}
		case fill == 1 && datasize == 1:
			value := c.Memory.Read32(source, false, false)
			for i := uint32(0); i < count; i++ {
				offset := i << 2
				c.Memory.Set32(destination+offset, value, false, false)
			}
		}

	case SWICpuFastSet:
		source := c.R[0]
		destination := c.R[1]
		fill := ReadBits(c.R[2], 24, 1)
		count := ReadBits(c.R[2], 0, 21) &^ uint32(7)

		if fill == 0 {
			for i := uint32(0); i < count; i++ {
				offset := i << 2
				value := c.Memory.Read32(source+offset, false, false)
				c.Memory.Set32(destination+offset, value, false, false)
			}
		} else {
			value := c.Memory.Read32(source, false, false)
			for i := uint32(0); i < count; i++ {
				offset := i << 2
				c.Memory.Set32(destination+offset, value, false, false)
			}
		}

	case SWIBitUnPack:
		c.swibitUnPack()

	case SWILz77UnCompWram:
		c.swilz77UnComp(false)
	case SWILz77UnCompVram:
		c.swilz77UnComp(true)

	case SWIRLUnCompWram:
		c.swirlUnComp(false)
	case SWIRLUnCompVram:
		c.swirlUnComp(true)

	case SWIObjAffineSet:
		c.swiobjAffineSet()

	case SWIBgAffineSet:
		c.swibgAffineSet()

	default:
		fmt.Printf("Warning: unhandled SWI 0x%02X\n", comment)
	}
}

func (c *CPU) swilz77UnComp(toVRAM bool) {
	src := c.R[0]
	dst := c.R[1]

	header := c.Memory.Read32(src, false, false)
	decompSize := header >> 8
	src += 4

	var buf [2]uint8
	bufIdx := 0

	writeByte := func(b uint8) {
		if toVRAM {
			buf[bufIdx] = b
			bufIdx++
			if bufIdx == 2 {
				c.Memory.Set16(dst&^1, uint16(buf[0])|uint16(buf[1])<<8, false, false)
				bufIdx = 0
			}
		} else {
			c.Memory.Set8(dst, b, false, false)
		}
		dst++
	}

	written := uint32(0)
	for written < decompSize {
		flags := c.Memory.Read8(src, false, false)
		src++

		for i := 7; i >= 0 && written < decompSize; i-- {
			if flags&(1<<i) == 0 {
				writeByte(c.Memory.Read8(src, false, false))
				src++
				written++
			} else {
				b1 := uint32(c.Memory.Read8(src, false, false))
				b2 := uint32(c.Memory.Read8(src+1, false, false))
				src += 2

				length := (b1 >> 4) + 3
				offset := ((b1 & 0xF) << 8) | b2

				for j := uint32(0); j < length && written < decompSize; j++ {
					writeByte(c.Memory.Read8(dst-offset-1, false, false))
					written++
				}
			}
		}
	}
}

func (c *CPU) swirlUnComp(toVRAM bool) {
	src := c.R[0]
	dst := c.R[1]

	header := c.Memory.Read32(src, false, false)
	decompSize := header >> 8
	src += 4

	var buf [2]uint8
	bufIdx := 0

	writeByte := func(b uint8) {
		if toVRAM {
			buf[bufIdx] = b
			bufIdx++
			if bufIdx == 2 {
				c.Memory.Set16(dst&^1, uint16(buf[0])|uint16(buf[1])<<8, false, false)
				bufIdx = 0
			}
		} else {
			c.Memory.Set8(dst, b, false, false)
		}
		dst++
	}

	written := uint32(0)
	for written < decompSize {
		flags := c.Memory.Read8(src, false, false)
		src++

		if flags&0x80 != 0 {
			length := uint32(flags&0x7F) + 3
			data := c.Memory.Read8(src, false, false)
			src++
			for j := uint32(0); j < length && written < decompSize; j++ {
				writeByte(data)
				written++
			}
		} else {
			length := uint32(flags&0x7F) + 1
			for j := uint32(0); j < length && written < decompSize; j++ {
				writeByte(c.Memory.Read8(src, false, false))
				src++
				written++
			}
		}
	}
}

func (c *CPU) swibitUnPack() {
	src := c.R[0]
	dst := c.R[1]
	info := c.R[2]

	srcLen := uint32(c.Memory.Read16(info, false, false))
	srcWidth := uint32(c.Memory.Read8(info+2, false, false))
	dstWidth := uint32(c.Memory.Read8(info+3, false, false))
	dataOffset := c.Memory.Read32(info+4, false, false)
	zeroFlag := dataOffset & 0x80000000
	dataOffset &= 0x7FFFFFFF

	srcBits := uint32(0)
	srcBuf := uint32(0)
	dstBuf := uint32(0)
	dstBits := uint32(0)

	for i := uint32(0); i < srcLen; i++ {
		srcBuf = uint32(c.Memory.Read8(src+i, false, false))
		for srcBits = 0; srcBits < 8; srcBits += srcWidth {
			data := (srcBuf >> srcBits) & ((1 << srcWidth) - 1)
			if data != 0 || zeroFlag != 0 {
				data += dataOffset
			}
			dstBuf |= data << dstBits
			dstBits += dstWidth
			if dstBits >= 32 {
				c.Memory.Set32(dst, dstBuf, false, false)
				dst += 4
				dstBuf = 0
				dstBits = 0
			}
		}
	}
}

func (c *CPU) swiobjAffineSet() {
	src := c.R[0]
	dst := c.R[1]
	count := c.R[2]
	offset := c.R[3]

	for i := uint32(0); i < count; i++ {
		sx := float64(int16(c.Memory.Read16(src, false, false))) / 256.0
		sy := float64(int16(c.Memory.Read16(src+2, false, false))) / 256.0
		angle := float64(uint16(c.Memory.Read16(src+4, false, false))) / 65536.0 * 2 * math.Pi
		src += 8

		sin := math.Sin(angle)
		cos := math.Cos(angle)

		pa := int16(cos * sx * 256)
		pb := int16(-sin * sx * 256)
		pc := int16(sin * sy * 256)
		pd := int16(cos * sy * 256)

		c.Memory.Set16(dst, uint16(pa), false, false)
		dst += offset
		c.Memory.Set16(dst, uint16(pb), false, false)
		dst += offset
		c.Memory.Set16(dst, uint16(pc), false, false)
		dst += offset
		c.Memory.Set16(dst, uint16(pd), false, false)
		dst += offset
	}
}

func (c *CPU) swibgAffineSet() {
	src := c.R[0]
	dst := c.R[1]
	count := c.R[2]

	for i := uint32(0); i < count; i++ {
		origX := int32(c.Memory.Read32(src, false, false))
		origY := int32(c.Memory.Read32(src+4, false, false))
		dispX := int16(c.Memory.Read16(src+8, false, false))
		dispY := int16(c.Memory.Read16(src+10, false, false))
		sx := float64(int16(c.Memory.Read16(src+12, false, false))) / 256.0
		sy := float64(int16(c.Memory.Read16(src+14, false, false))) / 256.0
		angle := float64(uint16(c.Memory.Read16(src+16, false, false))) / 65536.0 * 2 * math.Pi
		src += 20

		sin := math.Sin(angle)
		cos := math.Cos(angle)

		pa := int16(cos * sx * 256)
		pb := int16(-sin * sx * 256)
		pc := int16(sin * sy * 256)
		pd := int16(cos * sy * 256)

		startX := origX - int32(pa)*int32(dispX) - int32(pb)*int32(dispY)
		startY := origY - int32(pc)*int32(dispX) - int32(pd)*int32(dispY)

		c.Memory.Set16(dst, uint16(pa), false, false)
		c.Memory.Set16(dst+2, uint16(pb), false, false)
		c.Memory.Set16(dst+4, uint16(pc), false, false)
		c.Memory.Set16(dst+6, uint16(pd), false, false)
		c.Memory.Set32(dst+8, uint32(startX), false, false)
		c.Memory.Set32(dst+12, uint32(startY), false, false)
		dst += 16
	}
}
