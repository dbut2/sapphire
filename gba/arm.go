package gba

import (
	"fmt"
)

func (c *CPU) Arm(instruction uint32) {
	if !c.cond(ReadBits(instruction, 28, 4)) {
		return
	}

	do := ParseArm(c, instruction)
	do(instruction)
}

func ParseArm(c *CPU, instruction uint32) func(instruction uint32) {
	switch {
	case instruction&0b0000_1111_1111_1111_1111_1111_0000_0000 == 0b0000_0001_0010_1111_1111_1111_0000_0000:
		return c.ArmBranchX
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1111_0000_0000_0000_0000_0000_0000:
		return c.ArmSWI
	case instruction&0b0000_1101_1011_0000_1111_0000_0000_0000 == 0b0000_0001_0010_0000_1111_0000_0000_0000:
		return c.ArmMSR
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0000_0000_0000_0000_0000_0000_0000:
		return c.ArmALU
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1010_0000_0000_0000_0000_0000_0000:
		return c.ArmBranch
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0100_0000_0000_0000_0000_0000_0000:
		return c.ArmMemory
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1000_0000_0000_0000_0000_0000_0000:
		return c.ArmMemoryBlock
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1110_0000_0000_0000_0000_0000_0000,
		instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1100_0000_0000_0000_0000_0000_0000,
		instruction&0b0000_1111_1110_0000_0000_0000_0000_0000 == 0b0000_1100_0100_0000_0000_0000_0000_0000:
		return noins
	default:
		return noins
	}
}

func noins(instruction uint32) {
	panic(fmt.Sprintf("nothing to do for: %0.32b", instruction))
}

func (c *CPU) ArmALU(instruction uint32) {
	Opcode := ReadBits(instruction, 21, 4)

	var doOp func(Rd, Rn, Op2, Cy uint32) (value uint64)
	var flagger func(left, right uint32, value uint64) (N, Z, C, V bool)
	logic := false
	void := false
	switch Opcode {
	case 0x0:
		doOp = c.Arm_AND
		flagger = FlagLogic
		logic = true
	case 0x1:
		doOp = c.Arm_EOR
		flagger = FlagLogic
		logic = true
	case 0x2:
		doOp = c.Arm_SUB
		flagger = FlagArithSub
	case 0x3:
		doOp = c.Arm_RSB
		flagger = FlagArithReSub
	case 0x4:
		doOp = c.Arm_ADD
		flagger = FlagArithAdd
	case 0x5:
		doOp = c.Arm_ADC
		flagger = FlagArithAdd
	case 0x6:
		doOp = c.Arm_SBC
		flagger = FlagArithSub
	case 0x7:
		doOp = c.Arm_RSC
		flagger = FlagArithReSub
	case 0x8:
		doOp = c.Arm_TST
		flagger = FlagLogic
		logic = true
		void = true
	case 0x9:
		doOp = c.Arm_TEQ
		flagger = FlagLogic
		logic = true
		void = true
	case 0xA:
		doOp = c.Arm_CMP
		flagger = FlagArithSub
		void = true
	case 0xB:
		doOp = c.Arm_CMN
		flagger = FlagArithAdd
		void = true
	case 0xC:
		doOp = c.Arm_ORR
		flagger = FlagLogic
		logic = true
	case 0xD:
		doOp = c.Arm_MOV
		flagger = FlagLogic
		logic = true
	case 0xE:
		doOp = c.Arm_BIC
		flagger = FlagLogic
		logic = true
	case 0xF:
		doOp = c.Arm_MVN
		flagger = FlagLogic
		logic = true
	default:
		noins(instruction)
	}

	Rd := ReadBits(instruction, 12, 4)
	Rn := c.Arm_Rn(instruction)
	Op2 := c.Arm_Op2(instruction)
	Cy := ReadBits(c.CPSR, 29, 1)

	S := ReadBits(instruction, 20, 1)

	value := doOp(Rd, Rn, Op2, Cy)

	if !void {
		c.R[Rd] = uint32(value)
	}

	N, Z, C, V := flagger(Rn, Op2, value)

	switch {
	case S == 1 && Rd != 15 && logic:
		c.cpsrSetZ(Z)
		c.cpsrSetN(N)
		c.R[Rd] = uint32(value)
	case S == 1 && Rd != 15 && !logic:
		c.cpsrSetV(V)
		c.cpsrSetC(C)
		c.cpsrSetZ(Z)
		c.cpsrSetN(N)
		c.R[Rd] = uint32(value)
	case S == 1 && Rd == 15 && void:
		c.R[15] = uint32(value)
		c.cpsrSetZ(Z)
		c.cpsrSetN(N)
		c.cpsrSetC(C)
		c.cpsrSetV(V)

		//if c.cpsrMode() != USR {
		//	c.cpsrSetI(I)
		//	c.cpsrSetF(F)
		//	c.cpsrSetM1(M1)
		//	c.cpsrSetM0(M0)
		//}

		cond1 := c.cpsrIRQDisable() == 0
		cond2 := ReadIORegister(c.Memory, IME) > 0
		cond3 := ReadIORegister(c.Memory, IE)&ReadIORegister(c.Memory, IF) > 0
		if cond1 && cond2 && cond3 {
			c.exception(0x18)
		}
	case S == 1 && Rd == 15:
		c.CPSR = c.SPSR
		c.R[15] = uint32(value)
	case S == 0:
	}

	//Using R15 (PC)
	//When using R15 as Destination (Rd), note below CPSR description and Execution time description.

	//Returned CPSR Flags
	//If S=1, Rd<>R15, logical operations (AND,EOR,TST,TEQ,ORR,MOV,BIC,MVN):

	//If S=1, Rd<>R15, arithmetic operations (SUB,RSB,ADD,ADC,SBC,RSC,CMP,CMN):

	//IF S=1, with unused Rd bits=1111b, {P} opcodes (CMPP/CMNP/TSTP/TEQP):

	//If S=1, Rd=R15; should not be used in user mode:

	//If S=0: Flags are not affected (not allowed for CMP,CMN,TEQ,TST).

	return
}

func (c *CPU) Arm_AND(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn AND Op2
	return uint64(Rn & Op2)
}

func (c *CPU) Arm_EOR(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn XOR Op2
	return uint64(Rn ^ Op2)
}

func (c *CPU) Arm_SUB(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn-Op2
	return uint64(Rn) - uint64(Op2)
}

func (c *CPU) Arm_RSB(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Op2-Rn
	return uint64(Op2) - uint64(Rn)
}

func (c *CPU) Arm_ADD(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn+Op2
	return uint64(Rn) + uint64(Op2)
}

func (c *CPU) Arm_ADC(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn+Op2+Cy
	return uint64(Rn) + uint64(Op2) + uint64(Cy)
}

func (c *CPU) Arm_SBC(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn-Op2+Cy-1
	return uint64(Rn) - uint64(Op2) + uint64(Cy) - 1
}

func (c *CPU) Arm_RSC(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Op2-Rn+Cy-1
	return uint64(Op2) - uint64(Rn) + uint64(Cy) - 1
}

func (c *CPU) Arm_TST(Rd, Rn, Op2, Cy uint32) (value uint64) { // Void = Rn AND Op2
	return uint64(Rn & Op2)
}

func (c *CPU) Arm_TEQ(Rd, Rn, Op2, Cy uint32) (value uint64) { // Void = Rn XOR Op2
	return uint64(Rn ^ Op2)
}

func (c *CPU) Arm_CMP(Rd, Rn, Op2, Cy uint32) (value uint64) { // Void = Rn-Op2
	return uint64(Rn) - uint64(Op2)
}

func (c *CPU) Arm_CMN(Rd, Rn, Op2, Cy uint32) (value uint64) { // Void = Rn+Op2
	return uint64(Rn) + uint64(Op2)
}

func (c *CPU) Arm_ORR(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn OR Op2
	return uint64(Rn | Op2)
}

func (c *CPU) Arm_MOV(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Op2
	return uint64(Op2)
}

func (c *CPU) Arm_BIC(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = Rn AND NOT Op2
	return uint64(Rn & ^Op2)
}

func (c *CPU) Arm_MVN(Rd, Rn, Op2, Cy uint32) (value uint64) { // Rd = NOT Op2
	return uint64(^Op2)
}

func (c *CPU) Arm_Rn(instruction uint32) uint32 {
	Rn := ReadBits(instruction, 16, 4)
	return c.Arm_Rx(instruction, Rn)
}

func (c *CPU) Arm_Rx(instruction uint32, Rx uint32) uint32 {
	if Rx == 15 {
		I := ReadBits(instruction, 25, 1)
		R := ReadBits(instruction, 4, 1)
		if I == 0 && R == 1 {
			return c.R[Rx] + 4
		}
	}
	return c.R[Rx]
}

func (c *CPU) Arm_Rm(instruction uint32) uint32 {
	Rm := ReadBits(instruction, 0, 4)
	return c.Arm_Rx(instruction, Rm)
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
			return c.ArmShift(st, Rm, Is, S, I)
		case 1:
			Rs := ReadBits(instruction, 8, 4) & 0b11111111
			return c.ArmShift(st, Rm, c.R[Rs], S, I)
		default:
			noins(instruction)
			return 0
		}
	case 1:
		Is := ReadBits(instruction, 8, 4) * 2
		nn := ReadBits(instruction, 0, 8)
		return c.ArmShift(ROR, nn, Is, S, 0)
	default:
		noins(instruction)
		return 0
	}
}

func (c *CPU) ArmShift(shiftType uint32, value, amount uint32, S uint32, I uint32) uint32 {
	switch shiftType {
	case LSL:
		if amount == 0 && I == 1 {
			return value
		}
		if amount > 32 {
			if S == 1 {
				c.cpsrSetC(false)
			}
			return 0
		}
		v, carry := ShiftLSL(value, amount)
		if amount > 0 && S == 1 {
			c.cpsrSetC(carry)
		}
		return v
	case LSR:
		if amount == 0 && I == 1 {
			amount = 32
		}
		v, carry := ShiftLSR(value, amount)
		if amount > 0 && S == 1 {
			c.cpsrSetC(carry)
		}
		return v
	case ASR:
		if (amount == 0 && I == 1) || amount > 32 {
			amount = 32
		}
		v, carry := ShiftASR(value, amount)
		if amount > 0 && S == 1 {
			c.cpsrSetC(carry)
		}
		return v
	case ROR:
		if amount == 0 && I == 1 {
			oldC := c.cpsrC()
			c.cpsrSetC(ReadBits(value, 0, 1) == 1)
			v, _ := ShiftROR((value & ^(uint32(1)))|oldC, 1)
			return v
		}
		v, carry := ShiftROR(value, amount)
		if amount > 0 && S == 1 {
			c.cpsrSetC(carry)
		}
		return v
	default:
		panic(fmt.Sprintf("bad shift: %d", shiftType))
	}
}

func Shift(shiftType uint32, value, amount uint32) (uint32, bool) {
	switch shiftType {
	case LSL:
		return ShiftLSR(value, amount)
	case LSR:
		return ShiftLSR(value, amount)
	case ASR:
		return ShiftASR(value, amount)
	case ROR:
		return ShiftROR(value, amount)
	default:
		panic(fmt.Sprintf("bad shift: %d", shiftType))
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
	c.R[15] = addInt(c.R[15], nn)
}

func (c *CPU) Arm_BL(instruction uint32) {
	nn := Signify(ReadBits(instruction, 0, 24), 24) << 2
	c.R[14] = c.R[15]
	c.R[15] = addInt(c.R[15], nn)
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
	value := SetBits(c.R[Rn], 0, 1, 1)
	c.R[15] = value - 1
}

func (c *CPU) Arm_BLX(instruction uint32) {
	c.cpsrSetState(1)
	Rn := ReadBits(instruction, 0, 4)
	value := SetBits(c.R[Rn], 0, 1, 1)
	c.R[14] = c.R[15]
	c.R[15] = value - 1
}

func (c *CPU) ArmMSR(instruction uint32) {
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
		Offset, _ = Shift(ShiftType, c.R[Rm], Is)
	}

	if U == 0 {
		Offset = -Offset
	}
	addr := c.R[Rn] + Offset

	if L == 1 {
		if B == 1 {
			c.R[Rd] = uint32(c.Memory.Access8(addr))
		} else {
			c.R[Rd] = c.Memory.Access32(addr)
		}
	} else {
		if B == 1 {
			c.Memory.Set8(addr, uint8(c.R[Rd]))
		} else {
			c.Memory.Set32(addr, c.R[Rd])
		}
	}

	if P == 0 || ReadBits(instruction, 21, 1) == 1 {
		c.R[Rn] = addr
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

	address := c.R[Rn]

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
				c.R[i] = c.Memory.Access32(address)
				if S == 1 && i == 15 && c.CPSR>>29&1 == 0 {
					c.CPSR = c.SPSR
				}
			} else {
				c.Memory.Set32(address, c.R[i])
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
		c.R[Rn] = address
	}
}

func (c *CPU) ArmSWI(instruction uint32) {
	noins(instruction)
}
