package gba

import (
	"fmt"
)

func (c *CPU) Arm(instruction uint32) {
	if !c.cond(ReadBits(instruction, 28, 4)) {
		return
	}

	do := c.ParseArm(instruction)
	do(instruction)
}

func (c *CPU) ParseArm(instruction uint32) func(instruction uint32) {
	switch {
	case instruction&0b0000_1111_1111_1111_1111_1111_0000_0000 == 0b0000_0001_0010_1111_1111_1111_0000_0000:
		return c.ArmBranchX
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1111_0000_0000_0000_0000_0000_0000:
		return c.ArmSWI
	case instruction&0b0000_1101_1001_0000_0000_0000_0000_0000 == 0b0000_0001_0000_0000_0000_0000_0000_0000:
		return c.ArmPSR
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

func (c *CPU) ArmALU(instruction uint32) {
	Opcode := ReadBits(instruction, 21, 4)

	var doOp func(Rn, Op2, Cy uint32) (value uint64)
	var flagger func(left, right uint32, value uint64) (N, Z, C, V bool)
	logic := false
	void := false

	switch Opcode {
	case 0x0:
		doOp = AND
		flagger = FlagLogic
		logic = true
	case 0x1:
		doOp = EOR
		flagger = FlagLogic
		logic = true
	case 0x2:
		doOp = SUB
		flagger = FlagArithSub
	case 0x3:
		doOp = RSB
		flagger = FlagArithReSub
	case 0x4:
		doOp = ADD
		flagger = FlagArithAdd
	case 0x5:
		doOp = ADC
		flagger = FlagArithAdd
	case 0x6:
		doOp = SBCArm
		flagger = FlagArithSub
	case 0x7:
		doOp = RSC
		flagger = FlagArithReSub
	case 0x8:
		doOp = TST
		flagger = FlagLogic
		logic = true
		void = true
	case 0x9:
		doOp = TEQ
		flagger = FlagLogic
		logic = true
		void = true
	case 0xA:
		doOp = CMP
		flagger = FlagArithSub
		void = true
	case 0xB:
		doOp = CMN
		flagger = FlagArithAdd
		void = true
	case 0xC:
		doOp = ORR
		flagger = FlagLogic
		logic = true
	case 0xD:
		doOp = MOV
		flagger = FlagLogic
		logic = true
	case 0xE:
		doOp = BIC
		flagger = FlagLogic
		logic = true
	case 0xF:
		doOp = MVN
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

	value := doOp(Rn, Op2, Cy)

	if !void {
		c.R[Rd] = uint32(value)

		if Rd == 15 {
			c.prefetchFlush()
		}
	}

	N, Z, C, V := flagger(Rn, Op2, value)

	switch {
	case S == 1 && Rd != 15 && logic:
		c.cpsrSetZ(Z)
		c.cpsrSetN(N)
	case S == 1 && Rd != 15 && !logic:
		c.cpsrSetV(V)
		c.cpsrSetC(C)
		c.cpsrSetZ(Z)
		c.cpsrSetN(N)
	case S == 1 && Rd == 15 && !void:
		c.cpsrSetZ(Z)
		c.cpsrSetN(N)
		c.cpsrSetC(C)
		c.cpsrSetV(V)

		c.restoreCpsr()

		cond1 := c.cpsrIRQDisable() == 0
		cond2 := ReadIORegister(c.Memory, IME) > 0
		cond3 := ReadIORegister(c.Memory, IE)&ReadIORegister(c.Memory, IF) > 0
		if cond1 && cond2 && cond3 {
			c.exception(0x18)
		}
	}
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

func (c *CPU) ArmBranch(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Arm_B,
		1: c.Arm_BL,
	}[ReadBits(instruction, 24, 1)](instruction)

	c.prefetchFlush()
}

func (c *CPU) Arm_B(instruction uint32) {
	nn := signify(ReadBits(instruction, 0, 24), 24) << 2
	c.R[15] = addInt(c.R[15], nn)
}

func (c *CPU) Arm_BL(instruction uint32) {
	nn := signify(ReadBits(instruction, 0, 24), 24) << 2
	c.R[14] = c.curr + 4
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
	Rn := ReadBits(instruction, 0, 4)
	value := c.R[Rn]
	T := ReadBits(value, 0, 1)
	c.cpsrSetState(T)
	value |= 1
	c.R[15] = value - 1
}

func (c *CPU) Arm_BLX(instruction uint32) {
	Rn := ReadBits(instruction, 0, 4)
	value := c.R[Rn]
	T := ReadBits(value, 0, 1)
	c.cpsrSetState(T)
	value |= 1
	c.R[14] = c.curr + 4
	c.R[15] = value - 1
}

func (c *CPU) ArmPSR(instruction uint32) {
	Opcode := ReadBits(instruction, 21, 1)

	switch Opcode {
	case 0:
		c.ArmMRS(instruction)
	case 1:
		c.ArmMSR(instruction)
	}
}

func (c *CPU) ArmMRS(instruction uint32) {
	Rd := ReadBits(instruction, 12, 4)
	Psr := ReadBits(instruction, 22, 1)

	SWP := ReadBits(instruction, 16, 4)
	if SWP != 0b1111 {
		noins(instruction)
	}

	switch Psr {
	case 0:
		c.R[Rd] = c.CPSR
	case 1:
		c.R[Rd] = c.SPSR
	}
}

func (c *CPU) ArmMSR(instruction uint32) {
	I := ReadBits(instruction, 25, 1)
	Psr := ReadBits(instruction, 22, 1)

	f := ((ReadBits(instruction, 19, 1) ^ 1) - 1) >> 0
	s := ((ReadBits(instruction, 18, 1) ^ 1) - 1) >> 8
	x := ((ReadBits(instruction, 17, 1) ^ 1) - 1) >> 16
	c2 := ((ReadBits(instruction, 16, 1) ^ 1) - 1) >> 24
	fieldMask := f | s | x | c2

	var Op uint32
	switch I {
	case 0:
		Rm := ReadBits(instruction, 0, 4)
		Op = c.R[Rm]
	case 1:
		rotate := 2 * ReadBits(instruction, 8, 4)
		imm := ReadBits(instruction, 0, 8)
		Op, _ = ShiftROR(imm, rotate)
	}

	if c.cpsrMode() == USR {
		fieldMask &= ^c2
	}

	if Psr == 0 {
		cpsr := (c.CPSR & ^fieldMask) | (Op & fieldMask)

		mode := ReadBits(cpsr, 0, 5)
		c.cpsrSetMode(mode)

		c.CPSR = cpsr
	} else {
		c.SPSR = (c.SPSR &^ fieldMask) | (Op & fieldMask)
	}
}

func (c *CPU) ArmMemory(instruction uint32) {
	I := ReadBits(instruction, 25, 1)
	P := ReadBits(instruction, 24, 1)
	U := ReadBits(instruction, 23, 1)
	B := ReadBits(instruction, 22, 1)
	L := ReadBits(instruction, 20, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rd := ReadBits(instruction, 12, 4)

	var Offset uint32
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
	addr := c.R[Rn]

	if P == 1 {
		addr += Offset
	}

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

	if P == 0 {
		addr += Offset
	}

	if P == 0 || ReadBits(instruction, 21, 1) == 1 {
		c.R[Rn] = addr
	}

	if Rd == 15 {
		c.prefetchFlush()
	}
}

func (c *CPU) ArmMemoryBlock(instruction uint32) {
	L := ReadBits(instruction, 20, 1)

	switch L {
	case 0:
		c.Arm_STM(instruction)
	case 1:
		c.Arm_LDM(instruction)
	}
}

func (c *CPU) Arm_LDM(instruction uint32) {
	P := ReadBits(instruction, 24, 1)
	U := ReadBits(instruction, 23, 1)
	S := ReadBits(instruction, 22, 1)
	W := ReadBits(instruction, 21, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rlist := ReadBits(instruction, 0, 16)

	oldMode := c.cpsrMode()
	if S == 1 && (Rlist>>15)&1 == 0 {
		c.cpsrSetMode(USR)
	}

	address := c.R[Rn]
	oldRn := c.R[Rn]

	switch {
	case P == 0 && U == 0: // DA
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				c.R[i] = c.Memory.Access32(address)
				address -= 4
			}
		}
	case P == 1 && U == 0: // DB
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				address -= 4
				c.R[i] = c.Memory.Access32(address)
			}
		}
	case P == 0 && U == 1: // IA
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				c.R[i] = c.Memory.Access32(address)
				address += 4
			}
		}
	case P == 1 && U == 1: // IB
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				address += 4
				c.R[i] = c.Memory.Access32(address)
			}
		}
	}

	if W == 1 {
		switch U {
		case 0:
			c.R[Rn] = oldRn - setBits(Rlist)*4
		case 1:
			c.R[Rn] = oldRn + setBits(Rlist)*4
		}
	}

	if S == 1 && (Rlist>>15)&1 == 0 {
		c.cpsrSetMode(oldMode)
	}

	if (Rlist>>15)&1 == 1 {
		if S == 1 {
			c.restoreCpsr()
		}

		c.prefetchFlush()
	}
}

func (c *CPU) Arm_STM(instruction uint32) {
	P := ReadBits(instruction, 24, 1)
	U := ReadBits(instruction, 23, 1)
	S := ReadBits(instruction, 22, 1)
	W := ReadBits(instruction, 21, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rlist := ReadBits(instruction, 0, 16)

	oldMode := c.cpsrMode()
	if S == 1 {
		c.cpsrSetMode(USR)
	}

	address := c.R[Rn]
	oldRn := c.R[Rn]

	switch {
	case P == 0 && U == 0: // DA
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				c.Memory.Set32(address, c.R[i])
				address -= 4
			}
		}
	case P == 1 && U == 0: // DB
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				address -= 4
				c.Memory.Set32(address, c.R[i])
			}
		}
	case P == 0 && U == 1: // IA
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				c.Memory.Set32(address, c.R[i])
				address += 4
			}
		}
	case P == 1 && U == 1: // IB
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				address += 4
				c.Memory.Set32(address, c.R[i])
			}
		}
	}

	if W == 1 {
		switch U {
		case 0:
			c.R[Rn] = oldRn - setBits(Rlist)*4
		case 1:
			c.R[Rn] = oldRn + setBits(Rlist)*4
		}
	}

	if S == 1 {
		c.cpsrSetMode(oldMode)
	}

	if (Rlist>>15)&1 == 1 {
		if S == 1 {
			c.restoreCpsr()
		}

		c.prefetchFlush()
	}
}

func (c *CPU) ArmSWI(instruction uint32) {
	nn := ReadBits(instruction, 0, 24)
	c.SWI(nn)
}
