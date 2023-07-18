package gba

import (
	"fmt"
)

type CPU struct {
	*Motherboard
	CPURegisters

	curr, next uint32
	flushed    bool
}

func NewCPU(m *Motherboard) *CPU {
	return &CPU{Motherboard: m}
}

func (c *CPU) Boot() {
	c.R13 = 0x03007F00
	c.R13_svc = 0x03007FE0
	c.R13_irq = 0x03007FA0

	c.cpsrInitMode(SYS)
	c.prefetchFlush()

	SetIORegister(c.Memory, DISPCNT, 0x80)
	c.exception(0x08)

	c.Run()
}

func (c *CPU) Run() {
	for {
		curr := c.curr

		c.Step(curr)

		if !c.flushed {
			c.curr = c.next
			c.next = c.R[15]

			c.pcInc()
		}
		c.flushed = false
	}
}

func (c *CPU) pcInc() {
	switch c.cpsrState() {
	case 0:
		c.R[15] += 4
	case 1:
		c.R[15] += 2
	}
}

func (c *CPU) Step(curr uint32) {
	switch c.cpsrState() {
	case 0:
		instruction := c.Memory.Get32(curr)
		c.Arm(instruction)
	case 1:
		instruction := c.Memory.Get16(curr)
		c.Thumb(uint32(instruction))
	}
}

func noins(instruction uint32) {
	panic(fmt.Sprintf("nothing to do for: %032b", instruction))
}

func (c *CPU) prefetchFlush() {
	c.curr = c.R[15]
	c.pcInc()
	c.next = c.R[15]
	c.pcInc()
	c.flushed = true
}

type CPURegisters struct {
	// registers to interact
	R    [16]uint32
	CPSR uint32
	SPSR uint32

	// registers to be swapped on mode change
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
	return map[uint32]*uint32{
		0: &c.R0,
		1: &c.R1,
		2: &c.R2,
		3: &c.R3,
		4: &c.R4,
		5: &c.R5,
		6: &c.R6,
		7: &c.R7,
		8: map[uint32]*uint32{
			USR: &c.R8,
			FIQ: &c.R8_fiq,
			IRQ: &c.R8,
			SVC: &c.R8,
			ABT: &c.R8,
			UND: &c.R8,
			SYS: &c.R8,
		}[mode],
		9: map[uint32]*uint32{
			USR: &c.R9,
			FIQ: &c.R9_fiq,
			IRQ: &c.R9,
			SVC: &c.R9,
			ABT: &c.R9,
			UND: &c.R9,
			SYS: &c.R9,
		}[mode],
		10: map[uint32]*uint32{
			USR: &c.R10,
			FIQ: &c.R10_fiq,
			IRQ: &c.R10,
			SVC: &c.R10,
			ABT: &c.R10,
			UND: &c.R10,
			SYS: &c.R10,
		}[mode],
		11: map[uint32]*uint32{
			USR: &c.R11,
			FIQ: &c.R11_fiq,
			IRQ: &c.R11,
			SVC: &c.R11,
			ABT: &c.R11,
			UND: &c.R11,
			SYS: &c.R11,
		}[mode],
		12: map[uint32]*uint32{
			USR: &c.R12,
			FIQ: &c.R12_fiq,
			IRQ: &c.R12,
			SVC: &c.R12,
			ABT: &c.R12,
			UND: &c.R12,
			SYS: &c.R12,
		}[mode],
		13: map[uint32]*uint32{
			USR: &c.R13,
			FIQ: &c.R13_fiq,
			IRQ: &c.R13_irq,
			SVC: &c.R13_svc,
			ABT: &c.R13_abt,
			UND: &c.R13_und,
			SYS: &c.R13,
		}[mode],
		14: map[uint32]*uint32{
			USR: &c.R14,
			FIQ: &c.R14_fiq,
			IRQ: &c.R14_irq,
			SVC: &c.R14_svc,
			ABT: &c.R14_abt,
			UND: &c.R14_und,
			SYS: &c.R14,
		}[mode],
		15: &c.R15,
	}[r]
}

func (c *CPU) spsrAddr(mode uint32) *uint32 {
	return map[uint32]*uint32{
		FIQ: &c.SPSR_fiq,
		IRQ: &c.SPSR_irq,
		SVC: &c.SPSR_svc,
		ABT: &c.SPSR_abt,
		UND: &c.SPSR_und,
	}[mode]
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
	cpsr := c.SPSR
	newMode := ReadBits(cpsr, 0, 5)
	c.cpsrSetMode(newMode)
	c.CPSR = cpsr
}

func (c *CPU) cpsrMode() uint32 {
	return ReadBits(c.CPSR, 0, 5)
}

func (c *CPU) cpsrInitMode(value uint32) {
	c.CPSR = SetBits(c.CPSR, 0, 5, value)
}

func (c *CPU) cpsrSetMode(value uint32) {
	prevMode := c.cpsrMode()
	nextMode := value

	wasPrivileged := prevMode != USR && prevMode != SYS
	nowPrivileged := nextMode != USR && nextMode != SYS

	if wasPrivileged {
		*c.spsrAddr(prevMode) = c.SPSR
	}
	if nowPrivileged {
		c.SPSR = *c.spsrAddr(nextMode)
	}
	if nowPrivileged && !wasPrivileged {
		c.SPSR = c.CPSR
	}

	for i := uint32(0); i <= 15; i++ {
		*c.registerAddr(prevMode, i) = c.R[i]
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
	v := map[bool]uint32{false: 0, true: 1}[value]
	c.CPSR = SetBits(c.CPSR, 31, 1, v)
}

func (c *CPU) cpsrZ() uint32 {
	return ReadBits(c.CPSR, 30, 1)
}

func (c *CPU) cpsrSetZ(value bool) {
	v := map[bool]uint32{false: 0, true: 1}[value]
	c.CPSR = SetBits(c.CPSR, 30, 1, v)
}

func (c *CPU) cpsrC() uint32 {
	return ReadBits(c.CPSR, 29, 1)
}

func (c *CPU) cpsrSetC(value bool) {
	v := map[bool]uint32{false: 0, true: 1}[value]
	c.CPSR = SetBits(c.CPSR, 29, 1, v)
}

func (c *CPU) cpsrV() uint32 {
	return ReadBits(c.CPSR, 28, 1)
}

func (c *CPU) cpsrSetV(value bool) {
	v := map[bool]uint32{false: 0, true: 1}[value]
	c.CPSR = SetBits(c.CPSR, 28, 1, v)
}

func (c *CPU) exception(vector uint32) {
	switch vector {
	case 0x00: // reset
		c.cpsrSetMode(SVC)
	case 0x04: // undefined
		c.cpsrSetMode(UND)
	case 0x08: // swi
		c.cpsrSetMode(SVC)
	case 0x0C: // prefetch abort
		c.cpsrSetMode(ABT)
	case 0x10: // data
		c.cpsrSetMode(ABT)
	case 0x14: // address exceed
		c.cpsrSetMode(SVC)
	case 0x18: // irq
		c.cpsrSetMode(IRQ)
	case 0x1C: // fiq
		c.cpsrSetMode(FIQ)
	}

	c.R[14] = c.next
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
	N := c.cpsrN()
	Z := c.cpsrZ()
	C := c.cpsrC()
	V := c.cpsrV()

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
	SoftReset        uint32 = 0x00
	RegisterRamReset uint32 = 0x01
	CpuSet           uint32 = 0x0B
)

func (c *CPU) SWI(comment uint32) {
	switch comment {
	case SoftReset:
		c.R13 = 0x03007F00
		c.R13_svc = 0x03007FE0
		c.R13_irq = 0x03007FA0
		flag := c.Memory.Get8(0x3007FFA)
		for i := uint32(0x3007E00); i <= 0x3007FFF; i++ {
			c.Memory.Set8(i, 0)
		}
		if flag == 0 {
			c.R[14] = 0x02000000
		} else {
			c.R[14] = 0x08000000
		}
		c.R[15] = c.R[14]
		c.prefetchFlush()
		return
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
				value := c.Memory.Get16(source + offset)
				c.Memory.Set16(destination+offset, value)
			}
		case fill == 0 && datasize == 1:
			for i := uint32(0); i < count; i++ {
				offset := i << 2
				value := c.Memory.Get32(source + offset)
				c.Memory.Set32(destination+offset, value)
			}
		case fill == 1 && datasize == 0:
			value := c.Memory.Get16(source)
			for i := uint32(0); i < count; i++ {
				offset := i << 1
				c.Memory.Set16(destination+offset, value)
			}
		case fill == 1 && datasize == 1:
			value := c.Memory.Get32(source)
			for i := uint32(0); i < count; i++ {
				offset := i << 2
				c.Memory.Set32(destination+offset, value)
			}
		}
	case RegisterRamReset:
		c.exception(0x08)
	default:
		noComment(comment)
	}
}

func noComment(comment uint32) {
	panic(fmt.Sprintf("nothing to do for comment: 0x%02x", comment))
}
