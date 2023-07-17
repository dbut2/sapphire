package gba

func (c *CPU) Thumb(instruction uint32) {
	do := c.ParseThumb(instruction)
	do(instruction)
}

func (c *CPU) ParseThumb(instruction uint32) func(uint32) {
	switch {
	case instruction&0b1111_1111_0000_0000 == 0b1101_1111_0000_0000:
		return c.ThumbSWI
	case instruction&0b1111_1100_0000_0000 == 0b0100_0000_0000_0000:
		return c.ThumbALU
	case instruction&0b1111_1100_0000_0000 == 0b0100_0100_0000_0000:
		return c.ThumbHiReg
	case instruction&0b1111_1000_0000_0000 == 0b0001_1000_0000_0000:
		return c.ThumbAddSub
	case instruction&0b1111_1000_0000_0000 == 0b0100_1000_0000_0000:
		return c.ThumbMemoryPCRel
	case instruction&0b1111_0010_0000_0000 == 0b0101_0000_0000_0000:
		return c.ThumbMemoryReg
	case instruction&0b1111_0010_0000_0000 == 0b0101_0010_0000_0000:
		return c.ThumbMemorySign
	case instruction&0b1110_0000_0000_0000 == 0b0110_0000_0000_0000:
		return c.ThumbMemoryImm
	case instruction&0b1111_0000_0000_0000 == 0b0101_0000_0000_0000:
		return c.ThumbMemoryHalfSign
	case instruction&0b1111_0000_0000_0000 == 0b1000_0000_0000_0000:
		return c.ThumbMemoryHalf
	case instruction&0b1111_0000_0000_0000 == 0b1001_0000_0000_0000:
		return c.ThumbMemorySPRel
	case instruction&0b1111_0000_0000_0000 == 0b1010_0000_0000_0000:
		return c.ThumbMemoryPCSP
	case instruction&0b1111_0000_0000_0000 == 0b1100_0000_0000_0000:
		return c.ThumbMemoryBlock
	case instruction&0b1110_0000_0000_0000 == 0b0000_0000_0000_0000:
		return c.ThumbShift
	case instruction&0b1110_0000_0000_0000 == 0b0010_0000_0000_0000:
		return c.ThumbImm
	case instruction&0b1111_0000_0000_0000 == 0b1101_0000_0000_0000:
		return c.ThumbBranchCond
	case instruction&0b1111_1000_0000_0000 == 0b1110_0000_0000_0000:
		return c.ThumbBranchUncond
	case instruction&0b1111_1000_0000_0000 == 0b1111_0000_0000_0000:
		return c.ThumbBranchLink1
	case instruction&0b1110_1000_0000_0000 == 0b1110_1000_0000_0000:
		return c.ThumbBranchLink2
	case instruction&0b1111_0110_0000_0000 == 0b1011_0100_0000_0000:
		return c.ThumbPushPop
	case instruction&0b1111_1111_0000_0000 == 0b1011_0000_0000_0000:
		return c.ThumbAddSP
	default:
		return noins
	}
}

func (c *CPU) ThumbShift(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 2)

	Offset := ReadBits(instruction, 6, 5)
	Rs := ReadBits(instruction, 3, 3)
	Rd := ReadBits(instruction, 0, 3)

	var value uint32
	var carry bool
	switch Opcode {
	case 0b00:
		value, carry = ShiftLSL(c.R[Rs], Offset)
	case 0b01:
		value, carry = ShiftLSR(c.R[Rs], Offset)
	case 0b10:
		value, carry = ShiftASR(c.R[Rs], Offset)
	default:
		noins(instruction)
	}

	c.R[Rd] = value

	n := ReadBits(value, 31, 1) == 1
	z := value == 0

	c.cpsrSetN(n)
	c.cpsrSetZ(z)
	c.cpsrSetC(carry)
}

func (c *CPU) ThumbAddSub(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Thumb_ADD,
		1: c.Thumb_SUB,
		2: c.Thumb_ADD,
		3: c.Thumb_SUB,
	}[ReadBits(instruction, 9, 2)](instruction)
}

func (c *CPU) Thumb_ADD(instruction uint32) { // Rd=Rs+Rn / Rd=Rs+nn
	imm := ReadBits(instruction, 10, 1)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	var value uint64
	var op1 uint32
	var op2 uint32

	switch imm {
	case 0:
		Rn := ReadBits(instruction, 6, 3)
		op1 = c.R[Rs]
		op2 = c.R[Rn]
	case 1:
		nn := ReadBits(instruction, 6, 3)
		op1 = c.R[Rs]
		op2 = nn
	}

	value = uint64(op1) + uint64(op2)
	c.R[Rd] = uint32(value)

	N, Z, C, V := FlagArithAdd(op1, op2, value)

	c.cpsrSetN(N)
	c.cpsrSetZ(Z)
	c.cpsrSetC(C)
	c.cpsrSetV(V)
}

func (c *CPU) Thumb_SUB(instruction uint32) { // Rd=Rs-Rn / Rd=Rs-nn
	imm := ReadBits(instruction, 10, 1)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	var value uint64
	var op1 uint32
	var op2 uint32

	switch imm {
	case 0:
		Rn := ReadBits(instruction, 6, 3)
		op1 = c.R[Rs]
		op2 = c.R[Rn]
	case 1:
		nn := ReadBits(instruction, 6, 3)
		op1 = c.R[Rs]
		op2 = nn
	}

	value = uint64(op1) - uint64(op2)
	c.R[Rd] = uint32(value)

	N, Z, C, V := FlagArithSub(op1, op2, value)

	c.cpsrSetN(N)
	c.cpsrSetZ(Z)
	c.cpsrSetC(C)
	c.cpsrSetV(V)
}

func (c *CPU) ThumbImm(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 2)
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	left := c.R[Rd]
	right := nn

	switch Opcode {
	case 0b00:
		value := MOV(left, right, 0)
		N, Z, _, _ := FlagLogic(left, right, value)
		c.cpsrSetN(N)
		c.cpsrSetZ(Z)
		c.R[Rd] = uint32(value)
	case 0b01:
		value := CMP(left, right, 0)
		N, Z, C, V := FlagArithSub(left, right, value)
		c.cpsrSetN(N)
		c.cpsrSetZ(Z)
		c.cpsrSetC(C)
		c.cpsrSetV(V)
	case 0b10:
		value := ADD(left, right, 0)
		N, Z, C, V := FlagArithAdd(left, right, value)
		c.cpsrSetN(N)
		c.cpsrSetZ(Z)
		c.cpsrSetC(C)
		c.cpsrSetV(V)
		c.R[Rd] = uint32(value)
	case 0b11:
		value := SUB(left, right, 0)
		N, Z, C, V := FlagArithSub(left, right, value)
		c.cpsrSetN(N)
		c.cpsrSetZ(Z)
		c.cpsrSetC(C)
		c.cpsrSetV(V)
		c.R[Rd] = uint32(value)
	}
}

func (c *CPU) ThumbALU(instruction uint32) {
	Opcode := ReadBits(instruction, 6, 4)
	Rs := ReadBits(instruction, 3, 3)
	Rd := ReadBits(instruction, 0, 3)

	left := c.R[Rd]
	right := c.R[Rs]
	Cy := c.cpsrC()

	N := c.cpsrN() == 1
	Z := c.cpsrZ() == 1
	C := c.cpsrC() == 1
	V := c.cpsrV() == 1

	switch Opcode {
	case 0b0000: // AND
		value := AND(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, _, _ = FlagLogic(left, right, value)
	case 0b0001: // EOR
		value := EOR(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, _, _ = FlagLogic(left, right, value)
	case 0b0010: // LSL
		value, carry := ShiftLSL(left, right&0xFF)
		c.R[Rd] = value
		N, Z, _, _ = FlagLogic(left, right, uint64(value))
		C = carry
	case 0b0011: // LSR
		value, carry := ShiftLSR(left, right&0xFF)
		c.R[Rd] = value
		N, Z, _, _ = FlagLogic(left, right, uint64(value))
		C = carry
	case 0b0100: // ASR
		value, carry := ShiftASR(left, right&0xFF)
		c.R[Rd] = value
		N, Z, _, _ = FlagLogic(left, right, uint64(value))
		C = carry
	case 0b0101: // ADC
		value := ADC(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, C, V = FlagArithAdd(left, right, value)
	case 0b0110: // SBC
		value := SBCThumb(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, C, V = FlagArithSub(left, right, value)
	case 0b0111: // ROR
		value, carry := ShiftROR(left, right&0xFF)
		c.R[Rd] = value
		N, Z, _, _ = FlagLogic(left, right, uint64(value))
		C = carry
	case 0b1000: // TST
		value := TST(left, right, Cy)
		N, Z, _, _ = FlagLogic(left, right, value)
	case 0b1001: // NEG
		value := SUB(0, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, C, V = FlagArithSub(0, right, value)
	case 0b1010: // CMP
		value := CMP(left, right, Cy)
		N, Z, C, V = FlagArithSub(left, right, value)
	case 0b1011: // CMN
		value := CMN(left, right, Cy)
		N, Z, C, V = FlagArithAdd(left, right, value)
	case 0b1100: // ORR
		value := ORR(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, _, _ = FlagLogic(left, right, value)
	case 0b1101: // MUL
		value := MUL(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, C, V = FlagArithAdd(left, right, value)
	case 0b1110: // BIC
		value := BIC(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, _, _ = FlagLogic(left, right, value)
	case 0b1111: // MVN
		value := MVN(left, right, Cy)
		c.R[Rd] = uint32(value)
		N, Z, _, _ = FlagLogic(left, right, value)
	}

	c.cpsrSetN(N)
	c.cpsrSetZ(Z)
	c.cpsrSetC(C)
	c.cpsrSetV(V)
}

func (c *CPU) ThumbHiReg(instruction uint32) {
	map[uint32]func(uint32){
		0b00: c.Thumb_ADDHi,
		0b01: c.Thumb_CMPHi,
		0b10: c.Thumb_MOVHi,
		0b11: c.ThumbBranchHi,
	}[ReadBits(instruction, 8, 2)](instruction)
}

func (c *CPU) Thumb_ADDHi(instruction uint32) { // Rd = Rd+Rs
	Rd := ReadBits(instruction, 0, 3) + ReadBits(instruction, 7, 1)<<3
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := c.R[Rd] + c.R[Rs]
	if Rd == 15 {
		value += 4
	}
	c.R[Rd] = value
}

func (c *CPU) Thumb_CMPHi(instruction uint32) { // Void = Rd-Rs
	Rd := ReadBits(instruction, 0, 3) + ReadBits(instruction, 7, 1)<<3
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := c.R[Rd] - c.R[Rs]
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

	value := c.R[Rs]
	if Rd == 15 {
		value += 4
	}
	c.R[Rd] = value
}

func (c *CPU) ThumbBranchHi(instruction uint32) { // PC = Rs
	Opcode := ReadBits(instruction, 7, 1)

	switch Opcode {
	case 0b0:
		c.Thumb_BXHi(instruction)
	case 0b1:
		c.Thumb_BLXHi(instruction)
	}

	c.prefetchFlush()
}

func (c *CPU) Thumb_BXHi(instruction uint32) {
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := c.R[Rs]
	T := ReadBits(value, 0, 1)
	value &= ^T

	c.R[15] = value

	c.cpsrSetState(T)
}

func (c *CPU) Thumb_BLXHi(instruction uint32) {
	Rs := ReadBits(instruction, 3, 3) + ReadBits(instruction, 6, 1)<<3

	value := c.R[Rs]
	T := ReadBits(value, 0, 1)
	value &= ^T

	c.R[14] = c.R[15]
	c.R[15] = value

	c.cpsrSetState(T)
}

func (c *CPU) Thumb_SetCPSRLogic(value uint32) {
	c.cpsrSetN(int32(value) < 0)
	c.cpsrSetZ(uint32(value) == 0)
}

func (c *CPU) ThumbBranchCond(instruction uint32) {
	if !c.cond(ReadBits(instruction, 8, 4)) {
		return
	}

	offset := signify(ReadBits(instruction, 0, 8), 8) << 1
	c.R[15] = addInt(c.R[15], offset)
	c.prefetchFlush()
}

func (c *CPU) ThumbBranchUncond(instruction uint32) {
	offset := signify(ReadBits(instruction, 0, 11), 11) << 1
	c.R[15] = addInt(c.R[15], offset)
	c.prefetchFlush()
}

func (c *CPU) ThumbBranchLink1(instruction uint32) {
	nn := signify(ReadBits(instruction, 0, 11), 11) << 12
	c.R[14] = addInt(c.R[15], nn)
}

func (c *CPU) ThumbBranchLink2(instruction uint32) {
	nn := ReadBits(instruction, 0, 11) << 1
	x := ReadBits(instruction, 12, 0)

	l := c.R[14] + nn
	c.R[14] = c.R[15] - 2 | 1
	c.R[15] = l

	c.cpsrSetState(x ^ 1)
	c.prefetchFlush()
}

func (c *CPU) ThumbSWI(instruction uint32) {
	nn := ReadBits(instruction, 0, 8)
	c.SWI(nn)
}

func (c *CPU) ThumbPushPop(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 1)

	switch Opcode {
	case 0:
		c.ThumbPush(instruction)
	case 1:
		c.ThumbPop(instruction)
	}
}

func (c *CPU) ThumbPush(instruction uint32) {
	Lr := ReadBits(instruction, 8, 1)
	Rlist := ReadBits(instruction, 0, 8)

	if Lr == 1 {
		c.R[13] -= 4
		c.Memory.Set32(c.R[13], c.R[14])
	}
	for i := 7; i >= 0; i-- {
		if Rlist>>i&1 == 1 {
			c.R[13] -= 4
			c.Memory.Set32(c.R[13], c.R[i])
		}
	}
}

func (c *CPU) ThumbPop(instruction uint32) {
	Lr := ReadBits(instruction, 8, 1)
	Rlist := ReadBits(instruction, 0, 8)

	for i := 0; i <= 7; i++ {
		if Rlist>>i&1 == 1 {
			c.R[i] = c.Memory.Access32(c.R[13])
			c.R[13] += 4
		}
	}
	if Lr == 1 {
		c.R[15] = c.Memory.Access32(c.R[13])
		c.R[13] += 4
		c.prefetchFlush()
	}
}

func (c *CPU) ThumbMemorySPRel(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 1)
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8) << 2

	switch Opcode {
	case 0:
		c.Memory.Set32(c.R[13]+nn, c.R[Rd])
	case 1:
		c.R[Rd] = c.Memory.Access32(c.R[13] + nn)
	}
}

func (c *CPU) ThumbAddSP(instruction uint32) {
	Opcode := ReadBits(instruction, 7, 1)
	nn := ReadBits(instruction, 0, 7) << 2

	switch Opcode {
	case 0:
		c.R[13] += nn
	case 1:
		c.R[13] -= nn
	}
}

func (c *CPU) ThumbMemoryPCRel(instruction uint32) {
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8) << 2

	value := c.Memory.Access32(c.R[15] + nn)
	c.R[Rd] = value
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

	value := c.R[Rd]
	c.Memory.Set32(c.R[Rb]+c.R[Ro], value)
}

func (c *CPU) Thumb_STRB(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := uint8(c.R[Rd])
	c.Memory.Set8(c.R[Rb]+c.R[Ro], value)
}

func (c *CPU) Thumb_LDR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := c.Memory.Access32(c.R[Rb] + c.R[Ro])
	c.R[Rd] = value
}

func (c *CPU) Thumb_LDRB(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rb := ReadBits(instruction, 3, 3)
	Ro := ReadBits(instruction, 6, 3)

	value := uint32(c.Memory.Access8(c.R[Rb] + c.R[Ro]))
	c.R[Rd] = value
}

func (c *CPU) ThumbMemoryImm(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 2)
	nn := ReadBits(instruction, 6, 5)
	Rb := ReadBits(instruction, 3, 3)
	Rd := ReadBits(instruction, 0, 3)

	switch Opcode {
	case 0b00: // STR
		nn = nn << 2
		c.Memory.Set32(c.R[Rb]+nn, c.R[Rd])
	case 0b01: // LDR
		nn = nn << 2
		value := c.Memory.Access32(c.R[Rb] + nn)
		c.R[Rd] = value
	case 0b10: // STRB
		c.Memory.Set8(c.R[Rb]+nn, uint8(c.R[Rd]))
	case 0b11: // LDRB
		c.R[Rd] = uint32(c.Memory.Access8(c.R[Rb] + nn))
	}
}

func (c *CPU) ThumbMemoryPCSP(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 1)
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8) << 2

	switch Opcode {
	case 0b0:
		c.R[Rd] = (c.R[15] & ^uint32(2)) + nn
	case 0b1:
		c.R[Rd] = c.R[13] + nn
	}
}

func (c *CPU) ThumbMemoryBlock(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 1)

	Rb := ReadBits(instruction, 8, 3)
	Rlist := ReadBits(instruction, 0, 8)

	switch Opcode {
	case 0b0:
		for i := 0; i <= 7; i++ {
			if (Rlist>>i)&1 == 1 {
				c.Memory.Set32(c.R[Rb], c.R[i])
				c.R[Rb] += 4
			}
		}
	case 0b1:
		for i := 0; i <= 7; i++ {
			if (Rlist>>i)&1 == 1 {
				c.R[i] = c.Memory.Access32(c.R[Rb])
				c.R[Rb] += 4
			}
		}
	}
}

func (c *CPU) ThumbMemorySign(instruction uint32) {
	noins(instruction) // todo
}

func (c *CPU) ThumbMemoryHalfSign(instruction uint32) {
	Opcode := ReadBits(instruction, 10, 2)
	Ro := ReadBits(instruction, 6, 3)
	Rb := ReadBits(instruction, 3, 3)
	Rd := ReadBits(instruction, 0, 3)

	switch Opcode {
	case 0b00: // STRH
		c.Memory.Set16(c.R[Rb]+c.R[Ro], uint16(c.R[Rd]))
	case 0b01: // LDSB
		c.R[Rd] = uint32(signify(uint32(c.Memory.Access8(c.R[Rb]+c.R[Ro])), 8))
	case 0b10: // LDRH
		c.R[Rd] = uint32(c.Memory.Access16(c.R[Rb] + c.R[Ro]))
	case 0b11: // LDSH
		c.R[Rd] = uint32(signify(uint32(c.Memory.Access16(c.R[Rb]+c.R[Ro])), 16))
	}
}

func (c *CPU) ThumbMemoryHalf(instruction uint32) {
	Opcode := ReadBits(instruction, 11, 1)
	nn := ReadBits(instruction, 6, 5) << 1
	Rb := ReadBits(instruction, 3, 3)
	Rd := ReadBits(instruction, 0, 3)

	switch Opcode {
	case 0b0: // STRH
		c.Memory.Set16(c.R[Rb]+nn, uint16(c.R[Rd]))
	case 0b1: // LDRH
		c.R[Rd] = uint32(c.Memory.Access16(c.R[Rb] + nn))
	}
}
