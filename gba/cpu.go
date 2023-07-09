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

	c.cpsrSetMode(SYS)
	c.prefetchFlush()

	c.Run()
}

func (c *CPU) Run() {
	for {
		instruction := c.pipeline[0]

		c.Step(instruction)

		if !c.flushed {
			c.pipeline[0] = c.pipeline[1]
			c.pipeline[1] = c.R15

			c.pcInc()
		}
		c.flushed = false
	}
}

func (c *CPU) pcInc() {
	switch c.cpsrState() {
	case 0:
		c.R15 += 4
	case 1:
		c.R15 += 2
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
	c.pipeline[0] = c.R15
	c.pcInc()
	c.pipeline[1] = c.R15
	c.pcInc()
	c.flushed = true
}

type CPURegisters struct {
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
	CPSR     uint32
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

func (c *CPU) cpsrSetMode(value uint32) {
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

func (c *CPU) cpsrFIQDisable() uint32 {
	return ReadBits(c.CPSR, 6, 1)
}

func (c *CPU) cpsrSetFIQDisable(value uint32) {
	c.CPSR = SetBits(c.CPSR, 6, 1, value)
}

func (c *CPU) cpsrSetN(value uint32) {
	c.CPSR = SetBits(c.CPSR, 31, 1, value)
}

func (c *CPU) cpsrSetZ(value uint32) {
	c.CPSR = SetBits(c.CPSR, 30, 1, value)
}

func (c *CPU) cpsrSetC(value uint32) {
	c.CPSR = SetBits(c.CPSR, 29, 1, value)
}

func (c *CPU) cpsrSetV(value uint32) {
	c.CPSR = SetBits(c.CPSR, 28, 1, value)
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

	*c.registerAddr(14) = c.R15
	*c.spsrAddr() = c.CPSR

	c.cpsrSetState(0)
	c.cpsrSetIRQDisable(1)
	switch vector {
	case 0x00, 0x1C:
		c.cpsrSetFIQDisable(1)
	}

	c.R15 = vector
}

func (c *CPU) Arm(instruction uint32) {
	if !c.cond(ReadBits(instruction, 28, 4)) {
		return
	}

	// 0b0000_0000_0000_0000_0000_0000_0000_0000
	switch {
	case instruction&0b0000_1111_1111_1111_1111_1111_0000_0000 == 0b0000_0001_0010_1111_1111_1111_0000_0000:
		c.ArmBranchX(instruction)
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0000_0000_0000_0000_0000_0000_0000:
		c.ArmALU(instruction)
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1010_0000_0000_0000_0000_0000_0000:
		c.ArmBranch(instruction)
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0100_0000_0000_0000_0000_0000_0000:
		c.ArmMemory(instruction)
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1000_0000_0000_0000_0000_0000_0000:
		c.ArmMemoryBlock(instruction)
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1110_0000_0000_0000_0000_0000_0000,
		instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1100_0000_0000_0000_0000_0000_0000,
		instruction&0b0000_1111_1110_0000_0000_0000_0000_0000 == 0b0000_1100_0100_0000_0000_0000_0000_0000:
		noins(instruction)
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1111_0000_0000_0000_0000_0000_0000:
		c.ArmSWI(instruction)
	default:
		noins(instruction)
	}
}

func noins(instruction any) {
	panic(fmt.Sprintf("nothing to do for: %0.32b", instruction))
}

func (c *CPU) ArmALU(instruction uint32) {
	map[uint32]func(uint32){
		0x0: c.Arm_AND,
		0x1: c.Arm_EOR,
		0x2: c.Arm_SUB,
		0x3: c.Arm_RSB,
		0x4: c.Arm_ADD,
		0x5: c.Arm_ADC,
		0x6: c.Arm_SBC,
		0x7: c.Arm_RSC,
		0x8: c.Arm_TST,
		0x9: c.Arm_TEQ,
		0xA: c.Arm_CMP,
		0xB: c.Arm_CMN,
		0xC: c.Arm_ORR,
		0xD: c.Arm_MOV,
		0xE: c.Arm_BIC,
		0xF: c.Arm_MVN,
	}[ReadBits(instruction, 21, 4)](instruction)
}

func (c *CPU) Arm_AND(instruction uint32) { // Rd = Rn AND Op2
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := Rn & Op2
	*c.registerAddr(Rd) = value

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_EOR(instruction uint32) { // Rd = Rn XOR Op2
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := Rn ^ Op2
	*c.registerAddr(Rd) = value

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_SUB(instruction uint32) { // Rd = Rn-Op2
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Rn) - uint64(Op2)
	*c.registerAddr(Rd) = uint32(value)

	c.Arm_SetCPSRArithSub(instruction, Rn, Op2, value)
}

func (c *CPU) Arm_RSB(instruction uint32) { // Rd = Op2-Rn
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Op2) - uint64(Rn)
	*c.registerAddr(Rd) = uint32(value)

	c.Arm_SetCPSRArithSub(instruction, Op2, Rn, value)
}

func (c *CPU) Arm_ADD(instruction uint32) { // Rd = Rn+Op2
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Rn) + uint64(Op2)
	*c.registerAddr(Rd) = uint32(value)

	c.Arm_SetCPSRArithAdd(instruction, Rn, Op2, value)
}

func (c *CPU) Arm_ADC(instruction uint32) { // Rd = Rn+Op2+Cy
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Cy := ReadBits(c.CPSR, 29, 1)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Rn) + uint64(Op2) + uint64(Cy)
	*c.registerAddr(Rd) = uint32(value)

	c.Arm_SetCPSRArithAdd(instruction, Rn, Op2, value)
}

func (c *CPU) Arm_SBC(instruction uint32) { // Rd = Rn-Op2+Cy-1
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Cy := ReadBits(c.CPSR, 29, 1)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Rn) - uint64(Op2) + uint64(Cy) - 1
	*c.registerAddr(Rd) = uint32(value)

	c.Arm_SetCPSRArithSub(instruction, Rn, Op2, value)
}

func (c *CPU) Arm_RSC(instruction uint32) { // Rd = Op2-Rn+Cy-1
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Cy := ReadBits(c.CPSR, 29, 1)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Op2) - uint64(Rn) + uint64(Cy) - 1
	*c.registerAddr(Rd) = uint32(value)

	c.Arm_SetCPSRArithSub(instruction, Op2, Rn, value)
}

func (c *CPU) Arm_TST(instruction uint32) { // Void = Rn AND Op2
	Rn := c.Arm_Rn(instruction)
	Op2 := c.Arm_Op2(instruction)

	value := Rn & Op2

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_TEQ(instruction uint32) { // Void = Rn XOR Op2
	Rn := c.Arm_Rn(instruction)
	Op2 := c.Arm_Op2(instruction)

	value := Rn ^ Op2

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_CMP(instruction uint32) { // Void = Rn-Op2
	Rn := c.Arm_Rn(instruction)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Rn) - uint64(Op2)

	c.Arm_SetCPSRArithSub(instruction, Rn, Op2, value)
}

func (c *CPU) Arm_CMN(instruction uint32) { // Void = Rn+Op2
	Rn := c.Arm_Rn(instruction)
	Op2 := c.Arm_Op2(instruction)

	value := uint64(Rn) + uint64(Op2)

	c.Arm_SetCPSRArithAdd(instruction, Rn, Op2, value)
}

func (c *CPU) Arm_ORR(instruction uint32) { // Rd = Rn OR Op2
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := Rn | Op2
	*c.registerAddr(Rd) = value

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_MOV(instruction uint32) { // Rd = Op2
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := Op2
	*c.registerAddr(Rd) = value

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_BIC(instruction uint32) { // Rd = Rn AND NOT Op2
	Rn := c.Arm_Rn(instruction)
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := Rn & ^Op2
	*c.registerAddr(Rd) = value

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_MVN(instruction uint32) { // Rd = NOT Op2
	Rd := ReadBits(instruction, 12, 4)
	Op2 := c.Arm_Op2(instruction)

	value := ^Op2
	*c.registerAddr(Rd) = value

	c.Arm_SetCPSRLogic(instruction, value)
}

func (c *CPU) Arm_Rn(instruction uint32) uint32 {
	Rn := ReadBits(instruction, 16, 4)
	if Rn == 15 {
		I := ReadBits(instruction, 25, 1)
		R := ReadBits(instruction, 4, 1)
		if I == 0 && R == 1 {
			return *c.registerAddr(Rn) + 4
		}
	}
	return *c.registerAddr(Rn)
}

func (c *CPU) Arm_Rm(instruction uint32) uint32 {
	Rm := ReadBits(instruction, 0, 4)
	if Rm == 15 {
		I := ReadBits(instruction, 25, 1)
		R := ReadBits(instruction, 4, 1)
		if I == 0 && R == 1 {
			return *c.registerAddr(Rm) + 4
		}
	}
	return *c.registerAddr(Rm)
}

func (c *CPU) Arm_Op2(instruction uint32) uint32 {
	S := ReadBits(instruction, 20, 1)
	I := ReadBits(instruction, 25, 1)
	switch I {
	case 0:
		st := ReadBits(instruction, 5, 2)
		R := ReadBits(instruction, 4, 1)
		Rm := c.Arm_Rm(instruction)

		switch R {
		case 0:
			Is := ReadBits(instruction, 7, 5)
			return c.shift(st, S, Rm, Is)
		case 1:
			Rs := ReadBits(instruction, 8, 4) & 0b11111111
			return shift(st, Rm, *c.registerAddr(Rs))
		default:
			noins(instruction)
			return 0
		}
	case 1:
		Is := ReadBits(instruction, 8, 4) * 2
		nn := ReadBits(instruction, 0, 8)
		return c.shift(3, S, nn, Is)
	default:
		noins(instruction)
		return 0
	}
}

func (c *CPU) Arm_SetCPSRLogic(instruction uint32, value uint32) {
	S := ReadBits(instruction, 20, 1)
	if S == 1 {
		c.cpsrSetN(ReadBits(value, 31, 1))

		if value == 0 {
			c.cpsrSetZ(1)
		} else {
			c.cpsrSetZ(0)
		}
	}
}

func (c *CPU) Arm_SetCPSRArithAdd(instruction uint32, left, right uint32, value uint64) {
	S := ReadBits(instruction, 20, 1)
	if S == 1 {
		c.cpsrSetN(ReadBits(uint32(value), 31, 1))

		if value == 0 {
			c.cpsrSetZ(1)
		} else {
			c.cpsrSetZ(0)
		}

		if value >= 1<<32 {
			c.cpsrSetC(1)
		} else {
			c.cpsrSetC(0)
		}

		c.cpsrSetV(^(left ^ right) & (left ^ uint32(value)) >> 31)
	}
}

func (c *CPU) Arm_SetCPSRArithSub(instruction uint32, left, right uint32, value uint64) {
	S := ReadBits(instruction, 20, 1)
	if S == 1 {
		c.cpsrSetN(ReadBits(uint32(value), 31, 1))

		if value == 0 {
			c.cpsrSetZ(1)
		} else {
			c.cpsrSetZ(0)
		}

		if value < 1<<32 {
			c.cpsrSetC(1)
		} else {
			c.cpsrSetC(0)
		}

		c.cpsrSetV((left ^ right) & (left ^ uint32(value)) >> 31)
	}
}

func (c *CPU) shift(shiftType uint32, S uint32, value uint32, amount uint32) uint32 {
	if S == 1 {
		switch shiftType {
		case LSL:
			if value&(1<<(32-amount)) > 0 {
				c.cpsrSetC(1)
			} else {
				c.cpsrSetC(0)
			}
		case LSR, ASR:
			if value&(1<<(amount-1)) > 0 {
				c.cpsrSetC(1)
			} else {
				c.cpsrSetC(0)
			}
		case ROR:
			if (value>>(amount-1))&1 > 0 {
				c.cpsrSetC(1)
			} else {
				c.cpsrSetC(0)
			}
		}
	}

	return shift(shiftType, value, amount)
}

const (
	LSL uint32 = iota
	LSR
	ASR
	ROR
)

func shift(shiftType uint32, value uint32, amount uint32) uint32 {
	switch shiftType {
	case 0: // LSL
		return value << amount
	case 1: // LSR
		return value >> amount
	case 2: // ASR
		if value>>31 == 1 {
			return (value >> amount) | (^uint32(0) << (32 - amount))
		} else {
			return value >> amount
		}
	case 3: // ROR
		return (value >> amount) | (value << (32 - amount))
	default:
		panic(fmt.Sprintf("bad shift: %d", shiftType))
	}
}

func (c *CPU) cond(cond uint32) bool {
	N := ReadBits(c.CPSR, 31, 1) // Sign flag
	Z := ReadBits(c.CPSR, 30, 1) // Zero flag
	C := ReadBits(c.CPSR, 29, 1) // Carry flag
	V := ReadBits(c.CPSR, 28, 1) // Overflow flag

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

func (c *CPU) ArmBranch(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Arm_B,
		1: c.Arm_BL,
	}[ReadBits(instruction, 24, 1)](instruction)

	c.prefetchFlush()
}

func (c *CPU) Arm_B(instruction uint32) {
	nn := Signify(ReadBits(instruction, 0, 24), 24) << 2
	c.R15 = addInt(c.R15, nn)
}

func (c *CPU) Arm_BL(instruction uint32) {
	nn := Signify(ReadBits(instruction, 0, 24), 24) << 2
	*c.registerAddr(14) = c.R15
	c.R15 = addInt(c.R15, nn)
}

func addInt(a uint32, b int32) uint32 {
	if b < 0 {
		return a - uint32(-b)
	}
	return a + uint32(b)
}

func (c *CPU) ArmBranchX(instruction uint32) {
	map[uint32]func(uint32){
		0b0001: c.Arm_BX,
		0b0011: c.Arm_BLX,
	}[ReadBits(instruction, 4, 4)](instruction)

	c.prefetchFlush()
}

func (c *CPU) Arm_BX(instruction uint32) {
	c.cpsrSetState(1)
	Rn := ReadBits(instruction, 0, 4)
	value := SetBits(*c.registerAddr(Rn), 0, 1, 1)
	c.R15 = value - 1
}

func (c *CPU) Arm_BLX(instruction uint32) {
	c.cpsrSetState(1)
	Rn := ReadBits(instruction, 0, 4)
	value := SetBits(*c.registerAddr(Rn), 0, 1, 1)
	*c.registerAddr(14) = c.R15
	c.R15 = value - 1
}

func (c *CPU) ArmMemory(instruction uint32) {
	I := ReadBits(instruction, 25, 1)
	P := ReadBits(instruction, 24, 1)
	U := ReadBits(instruction, 23, 1)
	B := ReadBits(instruction, 22, 1)
	L := ReadBits(instruction, 20, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rd := ReadBits(instruction, 12, 4)

	Offset := uint32(0)
	if I == 0 {
		Offset = ReadBits(instruction, 0, 12)
	} else {
		Is := ReadBits(instruction, 7, 5)
		ShiftType := ReadBits(instruction, 5, 2)
		Rm := ReadBits(instruction, 0, 4)
		Offset = shift(ShiftType, *c.registerAddr(Rm), Is)
	}

	if U == 0 {
		Offset = -Offset
	}
	addr := *c.registerAddr(Rn) + Offset

	if L == 1 {
		if B == 1 {
			*c.registerAddr(Rd) = uint32(c.Memory.Access8(addr))
		} else {
			*c.registerAddr(Rd) = c.Memory.Access32(addr)
		}
	} else {
		if B == 1 {
			c.Memory.Set8(addr, uint8(*c.registerAddr(Rd)))
		} else {
			c.Memory.Set32(addr, *c.registerAddr(Rd))
		}
	}

	if P == 0 || ReadBits(instruction, 21, 1) == 1 {
		*c.registerAddr(Rn) = addr
	}
}

func (c *CPU) ArmMemoryBlock(instruction uint32) {
	P := ReadBits(instruction, 24, 1)
	U := ReadBits(instruction, 23, 1)
	S := ReadBits(instruction, 22, 1)
	W := ReadBits(instruction, 21, 1)
	L := ReadBits(instruction, 20, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rlist := ReadBits(instruction, 0, 16)

	address := *c.registerAddr(Rn)

	for i := uint32(0); i < 16; i++ {
		if (Rlist>>i)&1 == 1 {
			if P == 1 {
				if U == 1 {
					address += 4
				} else {
					address -= 4
				}
			}

			if L == 1 {
				*c.registerAddr(i) = c.Memory.Access32(address)
				if S == 1 && i == 15 && c.CPSR>>29&1 == 0 {
					c.CPSR = *c.spsrAddr()
				}
			} else {
				c.Memory.Set32(address, *c.registerAddr(i))
			}

			if P == 0 {
				if U == 1 {
					address += 4
				} else {
					address -= 4
				}
			}
		}
	}

	if W == 1 {
		*c.registerAddr(Rn) = address
	}
}

func (c *CPU) ArmSWI(instruction uint32) {
	noins(instruction)
}

func (c *CPU) Thumb(instruction uint32) {
	switch {
	case instruction&0b1111_1100_0000_0000 == 0b0100_0000_0000_0000:
		c.ThumbALU(instruction)
	case instruction&0b1111_1100_0000_0000 == 0b0100_0100_0000_0000:
		c.ThumbHiReg(instruction)
	case instruction&0b1111_1000_0000_0000 == 0b0001_1000_0000_0000:
		c.ThumbAddSub(instruction)
	case instruction&0b1111_1000_0000_0000 == 0b0100_1000_0000_0000:
		c.ThumbMemoryPCRel(instruction)
	case instruction&0b1111_0010_0000_0000 == 0b0101_0000_0000_0000:
		c.ThumbMemoryReg(instruction)
	case instruction&0b1111_0010_0000_0000 == 0b0101_0010_0000_0000:
		c.ThumbMemorySign(instruction)
	case instruction&0b1110_0000_0000_0000 == 0b0110_0000_0000_0000:
		c.ThumbMemoryImm(instruction)
	case instruction&0b1111_0000_0000_0000 == 0b1000_0000_0000_0000:
		c.ThumbMemoryHalf(instruction)
	case instruction&0b1111_0000_0000_0000 == 0b1001_0000_0000_0000:
		c.ThumbMemorySPRel(instruction)
	case instruction&0b1110_0000_0000_0000 == 0b0000_0000_0000_0000:
		c.ThumbShift(instruction)
	case instruction&0b1110_0000_0000_0000 == 0b0010_0000_0000_0000:
		c.ThumbImm(instruction)
	case instruction&0b1111_0000_0000_0000 == 0b1101_0000_0000_0000:
		c.ThumbBranch(instruction)
	default:
		noins(instruction)
	}
}

func (c *CPU) ThumbShift(instruction uint32) {
	map[uint32]func(uint32){
		0b00: c.Thumb_LSL,
		0b01: c.Thumb_LSR,
		0b10: c.Thumb_ASR,
	}[ReadBits(instruction, 11, 2)](instruction)
}

func (c *CPU) Thumb_LSL(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Offset := ReadBits(instruction, 6, 5)

	value := shift(LSL, *c.registerAddr(uint32(Rs)), uint32(Offset))

	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_LSR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Offset := ReadBits(instruction, 6, 5)

	value := shift(LSR, *c.registerAddr(Rs), Offset)

	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_ASR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Offset := ReadBits(instruction, 6, 5)

	value := shift(ASR, *c.registerAddr(Rs), Offset)

	*c.registerAddr(Rd) = value
}

func (c *CPU) ThumbAddSub(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Thumb_ADD,
		1: c.Thumb_ADD,
	}[ReadBits(instruction, 10, 1)](instruction)
}

func (c *CPU) Thumb_ADD(instruction uint32) { // Rd=Rs+Rn / Rd=Rs+nn
	imm := ReadBits(instruction, 9, 1)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	var value uint32
	var op1 uint32
	var op2 uint32

	switch imm {
	case 0:
		Rn := ReadBits(instruction, 6, 3)
		op1 = *c.registerAddr(Rs)
		op2 = *c.registerAddr(Rn)
		value = op1 + op2
	case 1:
		nn := ReadBits(instruction, 6, 3)
		op1 = *c.registerAddr(Rs)
		op2 = nn
		value = op1 + nn
	}

	*c.registerAddr(Rd) = value

	c.cpsrSetN(ReadBits(value, 31, 1))
	if value == 0 {
		c.cpsrSetZ(1)
	} else {
		c.cpsrSetZ(0)
	}

	if (value < op1) || (value < op2) {
		c.cpsrSetC(1)
	} else {
		c.cpsrSetC(0)
	}

	if (op1^op2 < 0) && (op1^value >= 0) {
		c.cpsrSetV(1)
	} else {
		c.cpsrSetV(0)
	}
}

func (c *CPU) Thumb_SUB(instruction uint32) { // Rd=Rs-Rn / Rd=Rs-nn
	imm := ReadBits(instruction, 9, 1)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	var value uint32
	var op1 uint32
	var op2 uint32

	switch imm {
	case 0:
		Rn := ReadBits(instruction, 6, 3)
		op1 = *c.registerAddr(Rs)
		op2 = *c.registerAddr(Rn)
		value = op1 - op2
	case 1:
		nn := ReadBits(instruction, 6, 3)
		op1 = *c.registerAddr(Rs)
		op2 = nn
		value = op1 - nn
	}

	*c.registerAddr(Rd) = value

	c.cpsrSetN(ReadBits(value, 31, 1))
	if value == 0 {
		c.cpsrSetZ(1)
	} else {
		c.cpsrSetZ(0)
	}

	if op1 >= op2 {
		c.cpsrSetC(1)
	} else {
		c.cpsrSetC(0)
	}

	if (op1^op2 >= 0) && (op1^value < 0) {
		c.cpsrSetV(1)
	} else {
		c.cpsrSetV(0)
	}
}

func (c *CPU) ThumbImm(instruction uint32) {
	map[uint32]func(uint32){
		0b00: c.Thumb_MOVImm,
		0b01: c.Thumb_CMPImm,
		0b10: c.Thumb_ADDImm,
		0b11: c.Thumb_SUBImm,
	}[ReadBits(instruction, 11, 2)](instruction)
}

func (c *CPU) Thumb_MOVImm(instruction uint32) { // Rd = nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := nn

	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_CMPImm(instruction uint32) { // Void = Rd - nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := *c.registerAddr(Rd) - nn
	_ = value
}

func (c *CPU) Thumb_ADDImm(instruction uint32) { // Rd = Rd + nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := *c.registerAddr(Rd) + nn

	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_SUBImm(instruction uint32) { // Rd = Rd - nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := *c.registerAddr(Rd) - nn

	*c.registerAddr(Rd) = value
}

func (c *CPU) ThumbALU(instruction uint32) {
	map[uint32]func(uint32){
		0x0: c.Thumb_AND,
		0x1: c.Thumb_EOR,
		0x2: c.Thumb_LSLALU,
		0x3: c.Thumb_LSRALU,
		0x4: c.Thumb_ASRALU,
		0x5: c.Thumb_ADC,
		0x6: c.Thumb_SBC,
		0x7: c.Thumb_ROR,
		0x8: c.Thumb_TST,
		0x9: c.Thumb_NEG,
		0xA: c.Thumb_CMP,
		0xB: c.Thumb_CMN,
		0xC: c.Thumb_ORR,
		0xD: c.Thumb_MUL,
		0xE: c.Thumb_BIC,
		0xF: c.Thumb_MVN,
	}[ReadBits(instruction, 6, 4)](instruction)
}

func (c *CPU) Thumb_AND(instruction uint32) { // Rd = Rd AND Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) & *c.registerAddr(Rs)
	*c.registerAddr(Rd) = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_EOR(instruction uint32) { // Rd = Rd XOR Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) ^ *c.registerAddr(Rs)
	*c.registerAddr(Rd) = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_LSLALU(instruction uint32) { // Rd = Rd << (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) << *c.registerAddr(Rs) & 0xFF
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_LSRALU(instruction uint32) { // Rd = Rd >> (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) >> *c.registerAddr(Rs) & 0xFF
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_ASRALU(instruction uint32) { // Rd = Rd SAR (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := shift(ASR, *c.registerAddr(Rd), *c.registerAddr(Rs)&0xFF)

	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_ADC(instruction uint32) { // Rd = Rd + Rs + Cy
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Cy := ReadBits(c.CPSR, 29, 1)

	value := *c.registerAddr(Rd) + *c.registerAddr(Rs) + Cy
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_SBC(instruction uint32) { // Rd = Rd - Rs - NOT Cy
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Cy := ReadBits(c.CPSR, 29, 1)

	value := *c.registerAddr(Rd) - *c.registerAddr(Rs) - ^Cy
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_ROR(instruction uint32) { // Rd = Rd ROR (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := shift(ROR, *c.registerAddr(Rd), *c.registerAddr(Rs)&0xFF)
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_TST(instruction uint32) { // Void = Rd AND Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Cy := ReadBits(c.CPSR, 29, 1)

	value := *c.registerAddr(Rd)&*c.registerAddr(Rs) + Cy

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_NEG(instruction uint32) { // Rd = 0 - Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := 0 - *c.registerAddr(Rs)
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_CMP(instruction uint32) { // Void = Rd - Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) - *c.registerAddr(Rs)
	_ = value
}

func (c *CPU) Thumb_CMN(instruction uint32) { // Void = Rd + Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) + *c.registerAddr(Rs)
	_ = value
}

func (c *CPU) Thumb_ORR(instruction uint32) { // Rd = Rd OR Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) | *c.registerAddr(Rs)
	*c.registerAddr(Rd) = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_MUL(instruction uint32) { // Rd = Rd * Rss
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) * *c.registerAddr(Rs)
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_BIC(instruction uint32) { // Rd = Rd AND NOT Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := *c.registerAddr(Rd) & ^*c.registerAddr(Rs)
	*c.registerAddr(Rd) = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_MVN(instruction uint32) { // Rd = NOT Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := ^*c.registerAddr(Rs)
	*c.registerAddr(Rd) = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_SetCPSRArithAdd(instruction uint32, left, right uint32, value uint64) {
	c.cpsrSetN(ReadBits(uint32(value), 31, 1))

	if uint32(value) == 0 {
		c.cpsrSetZ(1)
	} else {
		c.cpsrSetZ(0)
	}

	if value >= 1<<32 {
		c.cpsrSetC(1)
	} else {
		c.cpsrSetC(0)
	}

	c.cpsrSetV(^(left ^ right) & (left ^ uint32(value)) >> 31)
}

func (c *CPU) Thumb_SetCPSRArithSub(instruction uint32, left, right uint32, value uint64) {
	c.cpsrSetN(ReadBits(uint32(value), 31, 1))

	if uint32(value) == 0 {
		c.cpsrSetZ(1)
	} else {
		c.cpsrSetZ(0)
	}

	if value < 1<<32 {
		c.cpsrSetC(1)
	} else {
		c.cpsrSetC(0)
	}

	c.cpsrSetV((left ^ right) & (left ^ uint32(value)) >> 31)
}

func (c *CPU) ThumbHiReg(instruction uint32) {
	map[uint32]func(uint32){
		0x0: c.Thumb_ADDHi,
		0x1: c.Thumb_CMPHi,
		0x2: c.Thumb_MOVHi,
		0x3: c.ThumbBranchHi,
	}[ReadBits(instruction, 8, 2)](instruction)
}

func (c *CPU) Thumb_ADDHi(instruction uint32) { // Rd = Rd+Rs
	Rd := ReadBits(instruction, 0, 3) + ReadBits(instruction, 7, 1)<<3
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := *c.registerAddr(Rd) + *c.registerAddr(Rs)
	if Rd == 15 {
		value += 4
	}
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_CMPHi(instruction uint32) { // Void = Rd-Rs
	Rd := ReadBits(instruction, 0, 3) + ReadBits(instruction, 7, 1)<<3
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := *c.registerAddr(Rd) - *c.registerAddr(Rs)
	if Rd == 15 {
		value += 4
	}
	_ = value
}

func (c *CPU) Thumb_MOVHi(instruction uint32) { // Rd = Rs
	if instruction == 0b0100_0110_1100_0000 { // NOP
		return
	}

	Rd := ReadBits(instruction, 0, 3) + ReadBits(instruction, 7, 1)<<3
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := *c.registerAddr(Rs)
	if Rd == 15 {
		value += 4
	}
	*c.registerAddr(Rd) = value
}

func (c *CPU) ThumbBranchHi(instruction uint32) { // PC = Rs
	map[uint32]func(uint32){
		0: c.Thumb_BXHi,
		1: c.Thumb_BLXHi,
	}[ReadBits(instruction, 7, 1)](instruction)

	c.prefetchFlush()
}

func (c *CPU) Thumb_BXHi(instruction uint32) {
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := *c.registerAddr(Rs)
	T := ReadBits(value, 0, 1)
	value &= ^T

	c.R15 = value

	c.cpsrSetState(T)
}

func (c *CPU) Thumb_BLXHi(instruction uint32) {
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := *c.registerAddr(Rs)
	T := ReadBits(value, 0, 1)
	value &= ^T

	*c.registerAddr(14) = c.R15
	c.R15 = value

	c.cpsrSetState(T)
}

func (c *CPU) Thumb_SetCPSRLogic(value uint32) {
	c.cpsrSetN(ReadBits(value, 31, 1))

	if value == 0 {
		c.cpsrSetZ(1)
	} else {
		c.cpsrSetZ(0)
	}
}

func (c *CPU) ThumbBranch(instruction uint32) {
	if !c.cond(ReadBits(instruction, 6, 4)) {
		return
	}

	c.R15 += ReadBits(instruction, 0, 8) << 1

	c.prefetchFlush()
}

func (c *CPU) ThumbMemoryPCRel(instruction uint32) {
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := c.Memory.Access32(c.R15 + nn<<2)
	*c.registerAddr(Rd) = value
}

func (c *CPU) ThumbMemoryReg(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Thumb_STR,
		1: c.Thumb_STRB,
		2: c.Thumb_LDR,
		3: c.Thumb_LDRB,
	}[ReadBits(instruction, 10, 2)](instruction)
}

func (c *CPU) Thumb_STR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := *c.registerAddr(Rd)
	c.Memory.Set32(*c.registerAddr(Rb)+*c.registerAddr(Ro), value)
}

func (c *CPU) Thumb_STRB(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := uint8(*c.registerAddr(Rd))
	c.Memory.Set8(*c.registerAddr(Rb)+*c.registerAddr(Ro), value)
}

func (c *CPU) Thumb_LDR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := c.Memory.Access32(*c.registerAddr(Rb) + *c.registerAddr(Ro))
	*c.registerAddr(Rd) = value
}

func (c *CPU) Thumb_LDRB(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := uint32(c.Memory.Access8(*c.registerAddr(Rb) + *c.registerAddr(Ro)))
	*c.registerAddr(Rd) = value
}

func (c *CPU) ThumbMemoryImm(instruction uint32) {
	noins(instruction) // todo
}

func (c *CPU) ThumbMemorySign(instruction uint32) {
	noins(instruction) // todo
}

func (c *CPU) ThumbMemoryHalf(instruction uint32) {
	noins(instruction) // todo
}

func (c *CPU) ThumbMemorySPRel(instruction uint32) {
	noins(instruction) // todo
}

func Signify(value uint32, size uint32) int32 {
	shiftValue := 32 - size
	return int32(value<<shiftValue) >> shiftValue
}
