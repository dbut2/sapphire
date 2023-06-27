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
	c.R15 = 4

	for {
		c.Step()
	}
}

func (c *CPU) Step() {
	c.curr = c.next
	c.next = c.R15
	c.R15 = c.curr + 0b1000>>c.cpsrState()

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
		}[c.cpsrMode()],
		9: map[uint32]*uint32{
			0x10: &c.R9,
			0x11: &c.R9_fiq,
			0x12: &c.R9,
			0x13: &c.R9,
			0x17: &c.R9,
			0x1B: &c.R9,
		}[c.cpsrMode()],
		10: map[uint32]*uint32{
			0x10: &c.R10,
			0x11: &c.R10_fiq,
			0x12: &c.R10,
			0x13: &c.R10,
			0x17: &c.R10,
			0x1B: &c.R10,
		}[c.cpsrMode()],
		11: map[uint32]*uint32{
			0x10: &c.R11,
			0x11: &c.R11_fiq,
			0x12: &c.R11,
			0x13: &c.R11,
			0x17: &c.R11,
			0x1B: &c.R11,
		}[c.cpsrMode()],
		12: map[uint32]*uint32{
			0x10: &c.R12,
			0x11: &c.R12_fiq,
			0x12: &c.R12,
			0x13: &c.R12,
			0x17: &c.R12,
			0x1B: &c.R12,
		}[c.cpsrMode()],
		13: map[uint32]*uint32{
			0x10: &c.R13,
			0x11: &c.R13_fiq,
			0x12: &c.R13_irq,
			0x13: &c.R13_svc,
			0x17: &c.R13_abt,
			0x1B: &c.R13_und,
		}[c.cpsrMode()],
		14: map[uint32]*uint32{
			0x10: &c.R14,
			0x11: &c.R14_fiq,
			0x12: &c.R14_irq,
			0x13: &c.R14_svc,
			0x17: &c.R14_abt,
			0x1B: &c.R14_und,
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

func (c *CPU) cpsrState() uint32 {
	return ReadBits(c.CPSR, 5, 1)
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

func (c *CPU) Arm(instruction uint32) {
	if !c.cond(instruction) {
		return
	}

	// 0b0000_0000_0000_0000_0000_0000_0000_0000
	switch {
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
}

func (c *CPU) Arm_SetCPSRArithSub(instruction uint32, left, right uint32, value uint64) {
	S := ReadBits(instruction, 20, 1)
	if S == 1 {
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
}

func (c *CPU) shift(shiftType uint32, S uint32, value uint32, amount uint32) uint32 {
	if S == 1 {
		switch shiftType {
		case 0:
			if value&(1<<(32-amount)) > 0 {
				c.cpsrSetC(1)
			} else {
				c.cpsrSetC(0)
			}
		case 1, 2:
			if value&(1<<(amount-1)) > 0 {
				c.cpsrSetC(1)
			} else {
				c.cpsrSetC(0)
			}
		case 3:
			if (value>>(amount-1))&1 > 0 {
				c.cpsrSetC(1)
			} else {
				c.cpsrSetC(0)
			}
		}
	}

	return shift(shiftType, value, amount)
}

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

func (c *CPU) cond(instruction uint32) bool {
	cond := ReadBits(instruction, 28, 4)

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

func (c *CPU) ArmBranch(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Arm_B,
		1: c.Arm_BL,
	}[ReadBits(instruction, 24, 1)](instruction)
}

func (c *CPU) Arm_B(instruction uint32) {
	nn := int32(ReadBits(instruction, 0, 24)<<8) >> 6
	c.R15 = addInt(c.curr+8, nn)
}

type Uint interface {
	uint8 | uint16 | uint32 | uint64
}

type Int interface {
	int8 | int16 | int32 | int64
}

func addInt[U Uint, I Int](a U, b I) U {
	if b > 0 {
		a += U(b)
	} else {
		a -= U(-b)
	}
	return a
}

func (c *CPU) Arm_BL(instruction uint32) {
	nn := int32(ReadBits(instruction, 0, 24)<<8) >> 6
	c.R14 = c.curr + 4
	c.R15 = addInt(c.curr+8, nn)
	noins(instruction)
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
	//
}

func (c *CPU) Thumb(instruction uint16) {
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
