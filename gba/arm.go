package gba

import (
	"fmt"
)

func (c *CPU) Arm(instruction uint32) {
	if !c.cond(instruction >> 28) {
		return
	}
	armTable[(instruction>>16)&0xFF0|(instruction>>4)&0xF](c, instruction)
}

func (c *CPU) armSlow(instruction uint32) {
	switch {
	case instruction&0b0000_1111_1111_1111_1111_1111_0000_0000 == 0b0000_0001_0010_1111_1111_1111_0000_0000:
		c.ArmBranchX(instruction)
	case instruction&0b0000_1111_1011_1111_0000_1111_1111_1111 == 0b0000_0001_0000_1111_0000_0000_0000_0000:
		c.ArmMRS(instruction)
	case instruction&0b0000_1101_1011_0000_1111_0000_0000_0000 == 0b0000_0001_0010_0000_1111_0000_0000_0000:
		c.ArmMSR(instruction)
	case instruction&0b0000_1111_1011_0000_0000_1111_1111_0000 == 0b0000_0001_0000_0000_0000_0000_1001_0000:
		c.ArmSwap(instruction)
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1111_0000_0000_0000_0000_0000_0000:
		c.ArmSWI(instruction)
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1000_0000_0000_0000_0000_0000_0000:
		c.ArmMemoryBlock(instruction)
	case instruction&0b0000_1111_1100_0000_0000_0000_1111_0000 == 0b0000_0000_0000_0000_0000_0000_1001_0000:
		c.ArmMultiply(instruction)
	case instruction&0b0000_1111_1000_0000_0000_0000_1111_0000 == 0b0000_0000_1000_0000_0000_0000_1001_0000:
		c.ArmMultiplyLong(instruction)
	case instruction&0b0000_1110_0000_0000_0000_0000_1001_0000 == 0b0000_0000_0000_0000_0000_0000_1001_0000:
		c.Arm_MemoryHalf(instruction)
	case instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1010_0000_0000_0000_0000_0000_0000:
		c.ArmBranch(instruction)
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0100_0000_0000_0000_0000_0000_0000:
		c.ArmMemory(instruction)
	case instruction&0b0000_1111_0000_0000_0000_0000_0000_0000 == 0b0000_1110_0000_0000_0000_0000_0000_0000,
		instruction&0b0000_1110_0000_0000_0000_0000_0000_0000 == 0b0000_1100_0000_0000_0000_0000_0000_0000,
		instruction&0b0000_1111_1110_0000_0000_0000_0000_0000 == 0b0000_1100_0100_0000_0000_0000_0000_0000:
		return
	case instruction&0b0000_1100_0000_0000_0000_0000_0000_0000 == 0b0000_0000_0000_0000_0000_0000_0000_0000:
		c.ArmALU(instruction)
	default:
		c.noins(instruction)
	}
}

var armTable = buildArmTable()

func buildArmTable() (table [4096]func(*CPU, uint32)) {
	armNop := func(c *CPU, instruction uint32) {}
	cases := []struct {
		mask, value uint32
		handler     func(*CPU, uint32)
	}{
		{0x0FFFFF00, 0x012FFF00, (*CPU).ArmBranchX},
		{0x0FBF0FFF, 0x010F0000, (*CPU).ArmMRS},
		{0x0DB0F000, 0x0120F000, (*CPU).ArmMSR},
		{0x0FB00FF0, 0x01000090, (*CPU).ArmSwap},
		{0x0F000000, 0x0F000000, (*CPU).ArmSWI},
		{0x0E000000, 0x08000000, (*CPU).ArmMemoryBlock},
		{0x0FC000F0, 0x00000090, (*CPU).ArmMultiply},
		{0x0F8000F0, 0x00800090, (*CPU).ArmMultiplyLong},
		{0x0E000090, 0x00000090, (*CPU).Arm_MemoryHalf},
		{0x0E000000, 0x0A000000, (*CPU).ArmBranch},
		{0x0C000000, 0x04000000, (*CPU).ArmMemory},
		{0x0F000000, 0x0E000000, armNop},
		{0x0E000000, 0x0C000000, armNop},
		{0x0FE00000, 0x0C400000, armNop},
		{0x0C000000, 0x00000000, (*CPU).ArmALU},
	}
	const indexBits = 0x0FF000F0
	for i := range table {
		pattern := uint32(i>>4)<<20 | uint32(i&0xF)<<4
		table[i] = (*CPU).noins
		for _, cs := range cases {
			if pattern&cs.mask&indexBits == cs.value&indexBits {
				if cs.mask&^indexBits == 0 {
					table[i] = cs.handler
				} else {
					table[i] = (*CPU).armSlow
				}
				break
			}
		}
	}
	return
}

func (c *CPU) ArmALU(instruction uint32) {
	Opcode := ReadBits(instruction, 21, 4)
	Rd := ReadBits(instruction, 12, 4)
	Cy := ReadBits(c.CPSR, 29, 1)
	Rn := c.Arm_Rn(instruction)
	Op2 := c.Arm_Op2(instruction)
	S := ReadBits(instruction, 20, 1)

	var value uint64
	logic := false
	void := false

	switch Opcode {
	case 0b0000:
		value = uint64(Rn & Op2)
		logic = true
	case 0b0001:
		value = uint64(Rn ^ Op2)
		logic = true
	case 0b0010:
		value = uint64(Rn) - uint64(Op2)
	case 0b0011:
		value = uint64(Op2) - uint64(Rn)
	case 0b0100:
		value = uint64(Rn) + uint64(Op2)
	case 0b0101:
		value = uint64(Rn) + uint64(Op2) + uint64(Cy)
	case 0b0110:
		value = uint64(Rn) - uint64(Op2) + uint64(Cy) - 1
	case 0b0111:
		value = uint64(Op2) - uint64(Rn) + uint64(Cy) - 1
	case 0b1000:
		value = uint64(Rn & Op2)
		logic = true
		void = true
	case 0b1001:
		value = uint64(Rn ^ Op2)
		logic = true
		void = true
	case 0b1010:
		value = uint64(Rn) - uint64(Op2)
		void = true
	case 0b1011:
		value = uint64(Rn) + uint64(Op2)
		void = true
	case 0b1100:
		value = uint64(Rn | Op2)
		logic = true
	case 0b1101:
		value = uint64(Op2)
		logic = true
	case 0b1110:
		value = uint64(Rn &^ Op2)
		logic = true
	case 0b1111:
		value = uint64(^Op2)
		logic = true
	}

	if !void {
		c.R[Rd] = uint32(value)

		if Rd == 15 && S == 0 {
			c.prefetchFlush()
		}
	}

	if S == 0 {
		return
	}

	if Rd == 15 {
		c.restoreCpsr()
		if !void {
			c.prefetchFlush()
		}
		return
	}

	if logic {
		c.cpsrSetZ(uint32(value) == 0)
		c.cpsrSetN(uint32(value)>>31 == 1)
		return
	}

	var N, Z, C, V bool
	switch Opcode {
	case 0b0010, 0b0110, 0b1010:
		N, Z, C, V = FlagArithSub(Rn, Op2, value)
	case 0b0011, 0b0111:
		N, Z, C, V = FlagArithReSub(Rn, Op2, value)
	default:
		N, Z, C, V = FlagArithAdd(Rn, Op2, value)
	}
	c.cpsrSetV(V)
	c.cpsrSetC(C)
	c.cpsrSetZ(Z)
	c.cpsrSetN(N)
}

func (c *CPU) Arm_Rn(instruction uint32) uint32 {
	Rn := ReadBits(instruction, 16, 4)
	return c.Arm_Rx(instruction, Rn)
}

func (c *CPU) Arm_Rm(instruction uint32) uint32 {
	Rm := ReadBits(instruction, 0, 4)
	return c.Arm_Rx(instruction, Rm)
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
			return c.ArmShift(st, Rm, Is, S, 1)
		case 1:
			c.cycle(1)
			Rs := ReadBits(instruction, 8, 4)
			return c.ArmShift(st, Rm, c.R[Rs]&0xFF, S, I)
		default:
			c.noins(instruction)
			return 0
		}
	case 1:
		Is := ReadBits(instruction, 8, 4) * 2
		nn := ReadBits(instruction, 0, 8)
		return c.ArmShift(ROR, nn, Is, S, 0)
	default:
		c.noins(instruction)
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
			if S == 1 {
				c.cpsrSetC(ReadBits(value, 0, 1) == 1)
			}
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
	Opcode := ReadBits(instruction, 24, 1)

	switch Opcode {
	case 0:
		c.Arm_B(instruction)
	case 1:
		c.Arm_BL(instruction)
	}

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
	Opcde := ReadBits(instruction, 4, 4)

	switch Opcde {
	case 0b0001:
		c.Arm_BX(instruction)
	case 0b0011:
		c.Arm_BLX(instruction)
	default:
		c.noins(instruction)
	}

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

func (c *CPU) ArmMRS(instruction uint32) {
	Rd := ReadBits(instruction, 12, 4)
	Psr := ReadBits(instruction, 22, 1)

	switch Psr {
	case 0:
		c.R[Rd] = c.CPSR
	case 1:
		c.R[Rd] = *c.spsrAddr(c.cpsrMode())
	}
}

func (c *CPU) ArmSwap(instruction uint32) {
	B := ReadBits(instruction, 22, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rd := ReadBits(instruction, 12, 4)
	Rm := ReadBits(instruction, 0, 4)
	addr := c.R[Rn]

	if B == 0 {
		tmp := c.Memory.Read32(addr, true, false)
		c.Memory.Set32(addr, c.R[Rm], true, false)
		c.R[Rd] = tmp
	} else {
		tmp := uint32(c.Memory.Read8(addr, true, false))
		c.Memory.Set8(addr, uint8(c.R[Rm]), true, false)
		c.R[Rd] = tmp
	}
}

func (c *CPU) ArmMSR(instruction uint32) {
	I := ReadBits(instruction, 25, 1)
	Psr := ReadBits(instruction, 22, 1)

	var fieldMask uint32
	if ReadBits(instruction, 19, 1) == 1 {
		fieldMask |= 0xFF000000
	}
	if ReadBits(instruction, 18, 1) == 1 {
		fieldMask |= 0x00FF0000
	}
	if ReadBits(instruction, 17, 1) == 1 {
		fieldMask |= 0x0000FF00
	}
	if ReadBits(instruction, 16, 1) == 1 {
		fieldMask |= 0x000000FF
	}

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
		fieldMask &= 0xFFFFFF00
	}

	if Psr == 0 {
		cpsr := (c.CPSR & ^fieldMask) | (Op & fieldMask)

		mode := ReadBits(cpsr, 0, 5)
		c.cpsrSetMode(mode)

		c.CPSR = cpsr
	} else {
		*c.spsrAddr(c.cpsrMode()) = (*c.spsrAddr(c.cpsrMode()) & ^fieldMask) | (Op & fieldMask)
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
		Offset = c.ArmShift(ShiftType, c.R[Rm], Is, 0, 1)
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
			c.R[Rd] = uint32(c.Memory.Read8(addr, true, false))
		} else {
			c.R[Rd] = c.Memory.Read32(addr, true, false)
		}
	} else {
		value := c.R[Rd]
		if Rd == 15 {
			value += 4
		}
		if B == 1 {
			c.Memory.Set8(addr, uint8(value), true, false)
		} else {
			c.Memory.Set32(addr, value, true, false)
		}
	}

	if P == 0 {
		addr += Offset
	}

	if (P == 0 || ReadBits(instruction, 21, 1) == 1) && !(L == 1 && Rn == Rd) {
		c.R[Rn] = addr
	}

	if L == 1 && Rd == 15 {
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

	if Rlist == 0 {
		c.R[15] = c.Memory.Read32(emptyRlistAddr(address, P, U), true, false)
		c.prefetchFlush()
		if W == 1 {
			c.R[Rn] = emptyRlistBase(oldRn, U)
		}
		if S == 1 {
			c.cpsrSetMode(oldMode)
		}
		return
	}

	switch {
	case P == 0 && U == 0: // DA
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				c.R[i] = c.Memory.Read32(address&^3, true, false)
				address -= 4
			}
		}
	case P == 1 && U == 0: // DB
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				address -= 4
				c.R[i] = c.Memory.Read32(address&^3, true, false)
			}
		}
	case P == 0 && U == 1: // IA
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				c.R[i] = c.Memory.Read32(address&^3, true, false)
				address += 4
			}
		}
	case P == 1 && U == 1: // IB
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				address += 4
				c.R[i] = c.Memory.Read32(address&^3, true, false)
			}
		}
	}

	if W == 1 && (Rlist>>Rn)&1 == 0 {
		switch U {
		case 0:
			c.R[Rn] = oldRn - setBitCount(Rlist)*4
		case 1:
			c.R[Rn] = oldRn + setBitCount(Rlist)*4
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

	if Rlist == 0 {
		c.Memory.Set32(emptyRlistAddr(address, P, U), c.R[15]+4, true, false)
		if W == 1 {
			c.R[Rn] = emptyRlistBase(oldRn, U)
		}
		if S == 1 {
			c.cpsrSetMode(oldMode)
		}
		return
	}

	var finalBase uint32
	if U == 1 {
		finalBase = oldRn + setBitCount(Rlist)*4
	} else {
		finalBase = oldRn - setBitCount(Rlist)*4
	}
	lowest := -1
	for i := 0; i <= 15; i++ {
		if (Rlist>>i)&1 == 1 {
			lowest = i
			break
		}
	}

	storeReg := func(address uint32, i int) {
		value := c.R[i]
		if i == 15 {
			value += 4
		}
		if W == 1 && uint32(i) == Rn && i != lowest {
			value = finalBase
		}
		c.Memory.Set32(address, value, true, false)
	}

	switch {
	case P == 0 && U == 0: // DA
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				storeReg(address, i)
				address -= 4
			}
		}
	case P == 1 && U == 0: // DB
		for i := 15; i >= 0; i-- {
			if (Rlist>>i)&1 == 1 {
				address -= 4
				storeReg(address, i)
			}
		}
	case P == 0 && U == 1: // IA
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				storeReg(address, i)
				address += 4
			}
		}
	case P == 1 && U == 1: // IB
		for i := 0; i <= 15; i++ {
			if (Rlist>>i)&1 == 1 {
				address += 4
				storeReg(address, i)
			}
		}
	}

	if W == 1 {
		switch U {
		case 0:
			c.R[Rn] = oldRn - setBitCount(Rlist)*4
		case 1:
			c.R[Rn] = oldRn + setBitCount(Rlist)*4
		}
	}

	if S == 1 {
		c.cpsrSetMode(oldMode)
	}
}

func emptyRlistAddr(base uint32, P, U uint32) uint32 {
	switch {
	case P == 0 && U == 1:
		return base
	case P == 1 && U == 1:
		return base + 4
	case P == 0 && U == 0:
		return base - 0x3C
	default:
		return base - 0x40
	}
}

func emptyRlistBase(base uint32, U uint32) uint32 {
	if U == 1 {
		return base + 0x40
	}
	return base - 0x40
}

func (c *CPU) loadHalf(addr uint32) uint32 {
	value := uint32(c.Memory.Read16(addr, true, false))
	if addr&1 == 1 {
		value = value>>8 | value<<24
	}
	return value
}

func (c *CPU) loadHalfSigned(addr uint32) uint32 {
	if addr&1 == 1 {
		return uint32(signify(uint32(c.Memory.Read8(addr, true, false)), 8))
	}
	return uint32(signify(uint32(c.Memory.Read16(addr, true, false)), 16))
}

func (c *CPU) Arm_MemoryHalf(instruction uint32) {
	P := ReadBits(instruction, 24, 1)
	U := ReadBits(instruction, 23, 1)
	I := ReadBits(instruction, 22, 1)
	W := ReadBits(instruction, 21, 1)
	L := ReadBits(instruction, 20, 1)
	Rn := ReadBits(instruction, 16, 4)
	Rd := ReadBits(instruction, 12, 4)
	Opcode := ReadBits(instruction, 5, 2)

	var Offset uint32

	if I == 0 {
		Rm := ReadBits(instruction, 0, 4)
		Offset = c.R[Rm]
	} else {
		Offset = ReadBits(instruction, 8, 4)<<4 + ReadBits(instruction, 0, 4)
	}

	if U == 0 {
		Offset = -Offset
	}
	addr := c.R[Rn]

	if P == 1 {
		addr += Offset
	}

	var setRegisters uint16

	switch L {
	case 0:
		switch Opcode {
		case 0b01: // STRH
			c.Memory.Set16(addr, uint16(c.R[Rd]), true, false)
		case 0b10: // LDRD
			addr &= ^uint32(8)
			c.R[Rd] = c.Memory.Read32(addr, true, false)
			c.R[Rd+1] = c.Memory.Read32(addr+4, true, false)
			setRegisters |= 1 << Rd
			setRegisters |= 1 << (Rd + 1)
		case 0b11: //STRD
			addr &= ^uint32(8)
			c.Memory.Set32(addr, c.R[Rd], true, false)
			c.Memory.Set32(addr+4, c.R[Rd+1], true, false)
		default:
			c.noins(instruction)
		}
	case 1:
		switch Opcode {
		case 0b01:
			c.R[Rd] = c.loadHalf(addr)
			setRegisters |= 1 << Rd
		case 0b10:
			c.R[Rd] = uint32(signify(uint32(c.Memory.Read8(addr, true, false)), 8))
			setRegisters |= 1 << Rd
		case 0b11:
			c.R[Rd] = c.loadHalfSigned(addr)
			setRegisters |= 1 << Rd
		default:
			c.noins(instruction)
		}
	}

	if P == 0 {
		addr += Offset
	}

	if (P == 0 || W == 1) && !(L == 1 && Rn == Rd) {
		c.R[Rn] = addr
		setRegisters |= 1 << Rn
	}

	if setRegisters&(1<<15) != 0 {
		c.prefetchFlush()
	}
}

func (c *CPU) ArmMultiply(instruction uint32) {
	A := ReadBits(instruction, 21, 1)
	S := ReadBits(instruction, 20, 1)
	Rd := ReadBits(instruction, 16, 4)
	Rn := ReadBits(instruction, 12, 4)
	Rs := ReadBits(instruction, 8, 4)
	Rm := ReadBits(instruction, 0, 4)

	result := c.R[Rm] * c.R[Rs]
	if A == 1 {
		result += c.R[Rn]
	}
	c.R[Rd] = result

	if S == 1 {
		c.cpsrSetN(result>>31 == 1)
		c.cpsrSetZ(result == 0)
	}
}

func (c *CPU) ArmMultiplyLong(instruction uint32) {
	U := ReadBits(instruction, 22, 1)
	A := ReadBits(instruction, 21, 1)
	S := ReadBits(instruction, 20, 1)
	RdHi := ReadBits(instruction, 16, 4)
	RdLo := ReadBits(instruction, 12, 4)
	Rs := ReadBits(instruction, 8, 4)
	Rm := ReadBits(instruction, 0, 4)

	var result int64
	if U == 0 {
		result = int64(uint64(c.R[Rm]) * uint64(c.R[Rs]))
	} else {
		result = int64(int32(c.R[Rm])) * int64(int32(c.R[Rs]))
	}

	if A == 1 {
		acc := int64(uint64(c.R[RdHi])<<32 | uint64(c.R[RdLo]))
		result += acc
	}

	c.R[RdLo] = uint32(result)
	c.R[RdHi] = uint32(result >> 32)

	if S == 1 {
		c.cpsrSetN(c.R[RdHi]>>31 == 1)
		c.cpsrSetZ(c.R[RdHi] == 0 && c.R[RdLo] == 0)
	}
}

func (c *CPU) ArmSWI(instruction uint32) {
	nn := ReadBits(instruction, 16, 8)
	c.SWI(nn)
}
