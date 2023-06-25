package gba

import (
	"fmt"
)

type CPU struct {
	*Motherboard
	CPURegisters
	curr uint32
	next uint32
}

func NewCPU(m *Motherboard) *CPU {
	return &CPU{Motherboard: m}
}

func (c *CPU) Boot() {
	c.CPSR = 0x10
	c.next = 0
	c.R15 = 0b100 >> c.cpsrState()

	y := 0

	for {
		c.Step()
		y++
		fmt.Println(y)
	}
}

func (c *CPU) Step() {
	c.curr = c.next
	c.next = c.R15
	c.R15 = c.curr + 0b100>>c.cpsrState()*2

	switch c.cpsrState() {
	case 0:
		instruction := c.Memory.Access32(c.curr)
		c.Arm(instruction)
	case 1:
		instruction := c.Memory.Access16(c.curr)
		c.Thumb(instruction)
	}
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
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		9: map[uint32]*uint32{
			0x10: &c.R9,
			0x11: &c.R9_fiq,
			0x12: &c.R9,
			0x13: &c.R9,
			0x17: &c.R9,
			0x1B: &c.R9,
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		10: map[uint32]*uint32{
			0x10: &c.R10,
			0x11: &c.R10_fiq,
			0x12: &c.R10,
			0x13: &c.R10,
			0x17: &c.R10,
			0x1B: &c.R10,
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		11: map[uint32]*uint32{
			0x10: &c.R11,
			0x11: &c.R11_fiq,
			0x12: &c.R11,
			0x13: &c.R11,
			0x17: &c.R11,
			0x1B: &c.R11,
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		12: map[uint32]*uint32{
			0x10: &c.R12,
			0x11: &c.R12_fiq,
			0x12: &c.R12,
			0x13: &c.R12,
			0x17: &c.R12,
			0x1B: &c.R12,
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		13: map[uint32]*uint32{
			0x10: &c.R13,
			0x11: &c.R13_fiq,
			0x12: &c.R13_irq,
			0x13: &c.R13_svc,
			0x17: &c.R13_abt,
			0x1B: &c.R13_und,
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		14: map[uint32]*uint32{
			0x10: &c.R14,
			0x11: &c.R14_fiq,
			0x12: &c.R14_irq,
			0x13: &c.R14_svc,
			0x17: &c.R14_abt,
			0x1B: &c.R14_und,
		}[ReadBits(*c.cpsrAddr(), 0, 5)],
		15: &c.R15,
	}[r]
}

func (c *CPU) cpsrAddr() *uint32 {
	return &c.CPSR
}

const (
	EQ uint32 = iota
	NE
	CS, HS uint32 = 2, 2
	CC, LO uint32 = 3, 3
	MI     uint32 = iota
	PL
	VS
	VC
	HI
	LS
	GE
	LT
	GT
	LE
	AL
)

func (c *CPU) spsrAddr() *uint32 {
	return map[uint32]*uint32{
		0x11: &c.SPSR_fiq,
		0x12: &c.SPSR_irq,
		0x13: &c.SPSR_svc,
		0x17: &c.SPSR_abt,
		0x1B: &c.SPSR_und,
	}[ReadBits(*c.cpsrAddr(), 0, 5)]
}

func (c *CPU) cpsrState() uint32 {
	return ReadBits(*c.cpsrAddr(), 5, 1)
}

//\t\t(\w+)(\{cond\})?(\w+).*
//\t\t: c.Arm_$1$3,

//\d+: c.(\w+),
//func (c *CPU) $1(instruction uint32) {\n\n}\n

func (c *CPU) Arm(instruction uint32) {
	// 0b0000_0000_0000_0000_0000_0000_0000_0000
	switch {
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0000_0000_0000_0000_0000_0000_0000:
		c.ArmALU(instruction)
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1010_0000_0000_0000_0000_0000_0000:
		c.ArmBranch(instruction)
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1111_0000_0000_0000_0000_0000_0000:
		c.ArmSWI(instruction)
	default:
		noins(instruction)
	}
}

func (c *CPU) ArmSWI(instruction uint32) {
	//
}

func noins(instruction any) {
	panic(fmt.Sprintf("nothing to do for: %b", instruction))
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

//func (c *CPU) Arm_AND(instruction uint32) { // Rd = Rn AND Op2
//	Rd := ReadBits(instruction, 12, 4)
//	Rn := ReadBits(instruction, 16, 4)
//	Op2 := ReadBits(instruction, 0, 4)
//	*c.registerAddr(Rd) = *c.registerAddr(Rn) & *c.registerAddr(Op2)
//}

func (c *CPU) Arm_AND(instruction uint32) {
	condition := ReadBits(instruction, 28, 4)
	immediate := ReadBits(instruction, 25, 1) == 1
	s := ReadBits(instruction, 20, 1) == 1
	rn := ReadBits(instruction, 16, 4)
	rd := ReadBits(instruction, 12, 4)
	operand2 := uint32(0)

	// Check condition
	if !c.checkCondition(condition) {
		return
	}

	// Immediate operand
	if immediate {
		rotateImm := 2 * ReadBits(instruction, 8, 4)
		imm8 := ReadBits(instruction, 0, 8)
		operand2 = (imm8 >> rotateImm) | (imm8 << (32 - rotateImm)) // ROR
	} else { // Register operand
		shiftType := ReadBits(instruction, 5, 2)
		shiftByRegister := ReadBits(instruction, 4, 1) == 1
		rm := ReadBits(instruction, 0, 4)

		if shiftByRegister {
			rs := ReadBits(instruction, 8, 4)
			operand2 = c.shift(*c.registerAddr(rm), shiftType, *c.registerAddr(rs), false)
		} else {
			shiftImm := ReadBits(instruction, 7, 5)
			operand2 = c.shift(*c.registerAddr(rm), shiftType, shiftImm, false)
		}
	}

	// Execute AND operation
	*c.registerAddr(rd) = *c.registerAddr(rn) & operand2

	// Set flags
	if s {
		c.CPSR = SetBits(c.CPSR, 31, 1, *c.registerAddr(rd)>>31) // N
		c.CPSR = SetBits(c.CPSR, 30, 1, 0)                       // Z
		if *c.registerAddr(rd) == 0 {
			c.CPSR = SetBits(c.CPSR, 30, 1, 1) // Z
		}
	}
}

func (c *CPU) shift(value uint32, shiftType uint32, amount uint32, carry bool) uint32 {
	var res uint32
	if amount == 0 {
		if carry {
			c.CPSR = SetBits(c.CPSR, 29, 1, value&1) // C
		}
		return value
	}
	switch shiftType {
	case 0: // LSL
		res = value << amount
		if carry && amount > 0 {
			c.CPSR = SetBits(c.CPSR, 29, 1, (value>>(32-amount))&1)
		}
	case 1: // LSR
		if amount == 32 {
			res = 0
		} else {
			res = value >> amount
		}
		if carry && amount > 0 {
			c.CPSR = SetBits(c.CPSR, 29, 1, (value>>(amount-1))&1)
		}
	case 2: // ASR
		if amount >= 32 || value&0x80000000 != 0 {
			res = 0xFFFFFFFF
		} else {
			res = value >> amount
		}
		if carry && amount > 0 {
			c.CPSR = SetBits(c.CPSR, 29, 1, (value>>(amount-1))&1)
		}
	case 3: // ROR
		amount %= 32
		res = (value >> amount) | (value << (32 - amount))
		if carry && amount > 0 {
			c.CPSR = SetBits(c.CPSR, 29, 1, (value>>(amount-1))&1)
		}
	}
	return res
}

func (c *CPU) checkCondition(cond uint32) bool {
	N := ReadBits(c.CPSR, 31, 1) // Sign flag
	Z := ReadBits(c.CPSR, 30, 1) // Zero flag
	C := ReadBits(c.CPSR, 29, 1) // Carry flag
	V := ReadBits(c.CPSR, 28, 1) // Overflow flag

	switch cond {
	case 0x0: // EQ Z=1 equal (zero) (same)
		return Z == 1
	case 0x1: // EQ Z=1 equal (zero) (same)
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

func (c *CPU) Arm_EOR(instruction uint32) { // Rd = Rn XOR Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) ^ *c.registerAddr(Op2)
}

func (c *CPU) Arm_SUB(instruction uint32) { // Rd = Rn-Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) - *c.registerAddr(Op2)
}

func (c *CPU) Arm_RSB(instruction uint32) { // Rd = Op2-Rn
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Op2) - *c.registerAddr(Rn)
}

func (c *CPU) Arm_ADD(instruction uint32) { // Rd = Rn+Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) + *c.registerAddr(Op2)
}

func (c *CPU) Arm_ADC(instruction uint32) { // Rd = Rn+Op2+Cy
	//Rd := ReadBits(instruction, 12, 4)
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rd) = *c.registerAddr(Rn) + *c.registerAddr(Op2) + *c.registerAddr(Cy)
}

func (c *CPU) Arm_SBC(instruction uint32) { // Rd = Rn-Op2+Cy-1
	//Rd := ReadBits(instruction, 12, 4)
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rd) = *c.registerAddr(Rn) - *c.registerAddr(Op2) + *c.registerAddr(Cy) - 1
}

func (c *CPU) Arm_RSC(instruction uint32) { // Rd = Op2-Rn+Cy-1
	//Rd := ReadBits(instruction, 12, 4)
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rd) = *c.registerAddr(Op2) - *c.registerAddr(Rn) + *c.registerAddr(Cy) - 1
}

func (c *CPU) Arm_TST(instruction uint32) { // Void = Rn AND Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) & *c.registerAddr(Op2)
}

func (c *CPU) Arm_TEQ(instruction uint32) { // Void = Rn XOR Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) ^ *c.registerAddr(Op2)
}

func (c *CPU) Arm_CMP(instruction uint32) { // Void = Rn-Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) - *c.registerAddr(Op2)
}

func (c *CPU) Arm_CMN(instruction uint32) { // Void = Rn+Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) + *c.registerAddr(Op2)
}

func (c *CPU) Arm_ORR(instruction uint32) { // Rd = Rn OR Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) | *c.registerAddr(Op2)
}

func (c *CPU) Arm_MOV(instruction uint32) { // Rd = Op2
	Rd := ReadBits(instruction, 12, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Op2)
}

func (c *CPU) Arm_BIC(instruction uint32) { // Rd = Rn AND NOT Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) & ^*c.registerAddr(Op2)
}

func (c *CPU) Arm_MVN(instruction uint32) { // Rd = NOT Op2
	Rd := ReadBits(instruction, 12, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = ^*c.registerAddr(Op2)
}

func (c *CPU) ArmBranch(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Arm_B,
		1: c.Arm_BL,
	}[ReadBits(instruction, 24, 1)](instruction)
}

func (c *CPU) Arm_B(instruction uint32) {
	nn := ReadBits(instruction, 0, 24)
	c.R15 = c.curr + 8 + nn*4
}

func (c *CPU) Arm_BL(instruction uint32) {
	nn := ReadBits(instruction, 0, 24)
	c.R14 = c.curr + 4
	c.R15 = c.curr + 8 + nn*4
	noins(instruction)
}

//\t\t(\w+)(\{cond\})?(\w+).*
//\t\t: c.Thumb_$1$3,

//\d+: c.(\w+),
//func (c *CPU) $1(instruction uint16) {\n\n}\n

func (c *CPU) Thumb(instruction uint16) {
	// 0b0000_0000_0000_0000
	switch {
	case instruction&0b1111_1100_0000_0000 == 0b0100_0000_0000_0000:
		c.ThumbALU(instruction)
	case instruction&0b1111_1100_0000_0000 == 0b0100_0100_0000_0000:
		c.ThumbHiReg(instruction)
	default:
		panic(instruction)
	}
}

func (c *CPU) ThumbALU(instruction uint16) {
	map[uint32]func(uint16){
		0x0: c.Thumb_AND,
		0x1: c.Thumb_EOR,
		0x2: c.Thumb_LSL,
		0x3: c.Thumb_LSR,
		0x4: c.Thumb_ASR,
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
	}[0](instruction)
}

func (c *CPU) Thumb_AND(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_EOR(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_LSL(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_LSR(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_ASR(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_ADC(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_SBC(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_ROR(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_TST(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_NEG(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_CMP(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_CMN(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_ORR(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_MUL(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_BIC(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_MVN(instruction uint16) {
	noins(instruction)
}

func (c *CPU) ThumbHiReg(instruction uint16) {
	map[uint16]func(uint16){
		0x0: c.Thumb_ADD,
		0x1: c.Thumb_CMP,
		0x2: c.Thumb_MOV,
		0x3: c.Thumb_NOP,
		0x4: c.Thumb_BX,
		0x5: c.Thumb_BLX,
	}[0](instruction) // todo
}

func (c *CPU) Thumb_ADD(instruction uint16) {
	noins(instruction)
}

//func (c *CPU) Thumb_CMP(instruction uint16) {
//
//}

func (c *CPU) Thumb_MOV(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_NOP(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_BX(instruction uint16) {
	noins(instruction)
}

func (c *CPU) Thumb_BLX(instruction uint16) {
	noins(instruction)
}
