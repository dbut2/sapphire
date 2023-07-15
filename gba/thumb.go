package gba

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
	case instruction&0b1111_1000_0000_0000 == 0b1110_0000_0000_0000:
		c.ThumbBranch2(instruction)
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

	value, _ := Shift(LSL, c.R[uint32(Rs)], uint32(Offset))

	c.R[Rd] = value
}

func (c *CPU) Thumb_LSR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Offset := ReadBits(instruction, 6, 5)

	value, _ := Shift(LSR, c.R[Rs], Offset)

	c.R[Rd] = value
}

func (c *CPU) Thumb_ASR(instruction uint32) {
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Offset := ReadBits(instruction, 6, 5)

	value, _ := Shift(ASR, c.R[Rs], Offset)

	c.R[Rd] = value
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

	value := uint64(nn)

	c.R[Rd] = uint32(value)

	N, Z, C, V := FlagLogic(c.R[Rd], nn, value)

	c.cpsrSetN(N)
	c.cpsrSetZ(Z)
	c.cpsrSetC(C)
	c.cpsrSetV(V)
}

func (c *CPU) Thumb_CMPImm(instruction uint32) { // Void = Rd - nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := c.R[Rd] - nn
	_ = value
}

func (c *CPU) Thumb_ADDImm(instruction uint32) { // Rd = Rd + nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := c.R[Rd] + nn

	c.R[Rd] = value
}

func (c *CPU) Thumb_SUBImm(instruction uint32) { // Rd = Rd - nn
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := c.R[Rd] - nn

	c.R[Rd] = value
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

	value := c.R[Rd] & c.R[Rs]
	c.R[Rd] = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_EOR(instruction uint32) { // Rd = Rd XOR Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] ^ c.R[Rs]
	c.R[Rd] = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_LSLALU(instruction uint32) { // Rd = Rd << (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] << c.R[Rs] & 0xFF
	c.R[Rd] = value
}

func (c *CPU) Thumb_LSRALU(instruction uint32) { // Rd = Rd >> (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] >> c.R[Rs] & 0xFF
	c.R[Rd] = value
}

func (c *CPU) Thumb_ASRALU(instruction uint32) { // Rd = Rd SAR (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value, _ := Shift(ASR, c.R[Rd], c.R[Rs]&0xFF)

	c.R[Rd] = value
}

func (c *CPU) Thumb_ADC(instruction uint32) { // Rd = Rd + Rs + Cy
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Cy := ReadBits(c.CPSR, 29, 1)

	value := c.R[Rd] + c.R[Rs] + Cy
	c.R[Rd] = value
}

func (c *CPU) Thumb_SBC(instruction uint32) { // Rd = Rd - Rs - NOT Cy
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Cy := ReadBits(c.CPSR, 29, 1)

	value := c.R[Rd] - c.R[Rs] - ^Cy
	c.R[Rd] = value
}

func (c *CPU) Thumb_ROR(instruction uint32) { // Rd = Rd ROR (Rs AND 0FFh)
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value, _ := Shift(ROR, c.R[Rd], c.R[Rs]&0xFF)
	c.R[Rd] = value
}

func (c *CPU) Thumb_TST(instruction uint32) { // Void = Rd AND Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)
	Cy := ReadBits(c.CPSR, 29, 1)

	value := c.R[Rd]&c.R[Rs] + Cy

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_NEG(instruction uint32) { // Rd = 0 - Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := 0 - c.R[Rs]
	c.R[Rd] = value
}

func (c *CPU) Thumb_CMP(instruction uint32) { // Void = Rd - Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] - c.R[Rs]
	_ = value
}

func (c *CPU) Thumb_CMN(instruction uint32) { // Void = Rd + Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] + c.R[Rs]
	_ = value
}

func (c *CPU) Thumb_ORR(instruction uint32) { // Rd = Rd OR Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] | c.R[Rs]
	c.R[Rd] = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_MUL(instruction uint32) { // Rd = Rd * Rss
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] * c.R[Rs]
	c.R[Rd] = value
}

func (c *CPU) Thumb_BIC(instruction uint32) { // Rd = Rd AND NOT Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := c.R[Rd] & ^c.R[Rs]
	c.R[Rd] = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_MVN(instruction uint32) { // Rd = NOT Rs
	Rd := ReadBits(instruction, 0, 3)
	Rs := ReadBits(instruction, 3, 3)

	value := ^c.R[Rs]
	c.R[Rd] = value

	c.Thumb_SetCPSRLogic(value)
}

func (c *CPU) Thumb_SetCPSRArithAdd(instruction uint32, left, right uint32, value uint64) {
	c.cpsrSetN(int32(value) < 0)
	c.cpsrSetZ(uint32(value) == 0)
	c.cpsrSetC(value >= 1<<32)
	c.cpsrSetV(^(left^right)&(left^uint32(value))>>31 == 1)
}

func (c *CPU) Thumb_SetCPSRArithSub(instruction uint32, left, right uint32, value uint64) {
	c.cpsrSetN(int32(value) < 0)
	c.cpsrSetZ(uint32(value) == 0)
	c.cpsrSetC(value < 1<<32)
	c.cpsrSetV((left^right)&(left^uint32(value))>>31 == 1)
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
	map[uint32]func(uint32){
		0: c.Thumb_BXHi,
		1: c.Thumb_BLXHi,
	}[ReadBits(instruction, 7, 1)](instruction)

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

func (c *CPU) ThumbBranch(instruction uint32) {
	if !c.cond(ReadBits(instruction, 8, 4)) {
		return
	}

	offset := signify(ReadBits(instruction, 0, 8), 8) << 1
	c.R[15] = addInt(c.R[15], offset)
	c.prefetchFlush()
}

func (c *CPU) ThumbBranch2(instruction uint32) {
	offset := signify(ReadBits(instruction, 0, 11), 1) << 1
	c.R[15] = addInt(c.R[15], offset)
	c.prefetchFlush()
}

func (c *CPU) ThumbMemoryPCRel(instruction uint32) {
	Rd := ReadBits(instruction, 8, 3)
	nn := ReadBits(instruction, 0, 8)

	value := c.Memory.Access32(c.R[15] + nn<<2)
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
