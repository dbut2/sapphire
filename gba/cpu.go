package gba

import (
	"fmt"
)

type CPU struct {
	*Motherboard
	CPURegisters
	pipeline [2]uint32
	flushed  bool
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

	c.Run()
}

func (c *CPU) Run() {
	for {
		instruction := c.pipeline[0]

		c.Step(instruction)

		if !c.flushed {
			c.pipeline[0] = c.pipeline[1]
			c.pipeline[1] = c.R[15]

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
		instruction := c.Memory.Access32(curr)
		c.Arm(instruction)
	case 1:
		instruction := c.Memory.Access16(curr)
		c.Thumb(uint32(instruction))
	}
}

func (c *CPU) prefetchFlush() {
	c.pipeline[0] = c.R[15]
	c.pcInc()
	c.pipeline[1] = c.R[15]
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

func (c *CPU) registerAddr(r uint32) *uint32 {
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
			0x10: &c.R8,
			0x11: &c.R8_fiq,
			0x12: &c.R8,
			0x13: &c.R8,
			0x17: &c.R8,
			0x1B: &c.R8,
			0x1F: &c.R8,
		}[c.cpsrMode()],
		9: map[uint32]*uint32{
			0x10: &c.R9,
			0x11: &c.R9_fiq,
			0x12: &c.R9,
			0x13: &c.R9,
			0x17: &c.R9,
			0x1B: &c.R9,
			0x1F: &c.R9,
		}[c.cpsrMode()],
		10: map[uint32]*uint32{
			0x10: &c.R10,
			0x11: &c.R10_fiq,
			0x12: &c.R10,
			0x13: &c.R10,
			0x17: &c.R10,
			0x1B: &c.R10,
			0x1F: &c.R10,
		}[c.cpsrMode()],
		11: map[uint32]*uint32{
			0x10: &c.R11,
			0x11: &c.R11_fiq,
			0x12: &c.R11,
			0x13: &c.R11,
			0x17: &c.R11,
			0x1B: &c.R11,
			0x1F: &c.R11,
		}[c.cpsrMode()],
		12: map[uint32]*uint32{
			0x10: &c.R12,
			0x11: &c.R12_fiq,
			0x12: &c.R12,
			0x13: &c.R12,
			0x17: &c.R12,
			0x1B: &c.R12,
			0x1F: &c.R12,
		}[c.cpsrMode()],
		13: map[uint32]*uint32{
			0x10: &c.R13,
			0x11: &c.R13_fiq,
			0x12: &c.R13_irq,
			0x13: &c.R13_svc,
			0x17: &c.R13_abt,
			0x1B: &c.R13_und,
			0x1F: &c.R13,
		}[c.cpsrMode()],
		14: map[uint32]*uint32{
			0x10: &c.R14,
			0x11: &c.R14_fiq,
			0x12: &c.R14_irq,
			0x13: &c.R14_svc,
			0x17: &c.R14_abt,
			0x1B: &c.R14_und,
			0x1F: &c.R14,
		}[c.cpsrMode()],
		15: &c.R15,
	}[r]
}

func (c *CPU) spsrAddr() *uint32 {
	return map[uint32]*uint32{
		0x11: &c.SPSR_fiq,
		0x12: &c.SPSR_irq,
		0x13: &c.SPSR_svc,
		0x17: &c.SPSR_abt,
		0x1B: &c.SPSR_und,
	}[c.cpsrMode()]
}

func (c *CPU) cpsrMode() uint32 {
	return ReadBits(c.CPSR, 0, 5)
}

func (c *CPU) cpsrInitMode(value uint32) {
	c.CPSR = SetBits(c.CPSR, 0, 5, value)
}

func (c *CPU) cpsrSetMode(value uint32) {
	for i := uint32(0); i <= 15; i++ {
		*c.registerAddr(i) = c.R[i]
	}
	if c.cpsrMode() != 0x10 && c.cpsrMode() != 0x1F {
		*c.spsrAddr() = c.SPSR
	}

	c.CPSR = SetBits(c.CPSR, 0, 5, value)

	for i := uint32(0); i <= 15; i++ {
		c.R[i] = *c.registerAddr(i)
	}
	if c.cpsrMode() != 0x10 && c.cpsrMode() != 0x1F {
		c.SPSR = *c.spsrAddr()
	}
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

func (c *CPU) cpsrFIQDisable() uint32 {
	return ReadBits(c.CPSR, 6, 1)
}

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

const (
	USR uint32 = 0x10
	FIQ uint32 = 0x11
	IRQ uint32 = 0x12
	SVC uint32 = 0x13
	ABT uint32 = 0x17
	UND uint32 = 0x1B
	SYS uint32 = 0x1F
)

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

	c.R[14] = c.R[15]
	c.SPSR = c.CPSR

	c.cpsrSetState(0)
	c.cpsrSetIRQDisable(1)
	switch vector {
	case 0x00, 0x1C:
		c.cpsrSetFIQDisable(1)
	}

	c.R[15] = vector
}

func ShiftLSL(value, amount uint32) (uint32, bool) {
	return value << amount, value&(1<<(32-amount)) > 0
}

func ShiftLSR(value, amount uint32) (uint32, bool) {
	return value >> amount, value&(1<<(amount-1)) > 0
}

func ShiftASR(value, amount uint32) (uint32, bool) {
	s := value & 0x8000_0000
	for i := uint32(0); i < amount; i++ {
		value = (value >> 1) | s
	}
	return value, value&(1<<(amount-1)) > 0
}

func ShiftROR(value, amount uint32) (uint32, bool) {
	return value>>(amount%32) | value<<(32-(amount%32)), (value>>(amount-1))&1 > 0
}

const (
	LSL uint32 = iota
	LSR
	ASR
	ROR
)

func (c *CPU) cond(cond uint32) bool {
	N := c.cpsrN()
	Z := c.cpsrZ()
	C := c.cpsrC()
	V := c.cpsrV()

	switch cond {
	case 0x0: // EQ Z=1 equal (zero) (same)
		return Z == 1
	case 0x1: // NE Z=1 not equal (nonzero) (not same)
		return Z == 0
	case 0x2: // CS/HS C=1 unsigned higher or same (carry set)
		return C == 1
	case 0x3: // CC/LO C=0 unsigned lower (carry cleared)
		return C == 0
	case 0x4: // MI N=1 signed negative (minus)
		return N == 1
	case 0x5: // PL N=0 signed positive or zero (plus)
		return N == 0
	case 0x6: // VS V=1 signed overflow (V set)
		return V == 1
	case 0x7: // VC V=0 signed no overflow (V cleared)
		return V == 0
	case 0x8: // HI C=1 and Z=0 unsigned higher
		return C == 1 && Z == 0
	case 0x9: // LS C=0 or Z=1 unsigned lower or same
		return C == 0 || Z == 1
	case 0xA: // GE N=V signed greater or equal
		return N == V
	case 0xB: // LT N<>V signed less than
		return N != V
	case 0xC: // GT Z=0 and N=V signed greater than
		return Z == 0 && N == V
	case 0xD: // LE Z=1 or N<>V signed less or equal
		return Z == 1 || N != V
	case 0xE: // AL - always (the "AL" suffix can be omitted)
		return true
	default:
		return false
	}
}

func addInt(a uint32, b int32) uint32 {
	if b < 0 {
		return a - uint32(-b)
	}
	return a + uint32(b)
}

func Signify(value uint32, size uint32) int32 {
	shiftValue := 32 - size
	return int32(value<<shiftValue) >> shiftValue
}
