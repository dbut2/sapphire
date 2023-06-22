package gba

type CPU struct {
	Register Register
}

type Register struct {
	R [16]uint32
}

func (c CPU) Do(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Thumb,
		1: c.Arm,
	}[0](instruction)
}

func (c CPU) Arm(instruction uint32) {
	map[uint32]func(uint32){
		0: c.ArmBranch,
		1: c.ArmALU,
		2: c.ArmMul,
	}[0](instruction)
}

func (c CPU) ArmBranch(instruction uint32) {
	map[uint32]func(uint32){}[0](instruction)
}

func (c CPU) ArmALU(instruction uint32) {
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
	}[ReadFlag(instruction, 21, 4)](instruction)
}

func (c CPU) Arm_AND(instruction uint32) { // Rd = Rn AND Op2
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] & c.Register.R[Op2]
}

func (c CPU) Arm_EOR(instruction uint32) { // Rd = Rn XOR Op2
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] ^ c.Register.R[Op2]
}

func (c CPU) Arm_SUB(instruction uint32) { // Rd = Rn-Op2
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] - c.Register.R[Op2]
}

func (c CPU) Arm_RSB(instruction uint32) { // Rd = Op2-Rn
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Op2] - c.Register.R[Rn]
}

func (c CPU) Arm_ADD(instruction uint32) { // Rd = Rn+Op2
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] + c.Register.R[Op2]
}

func (c CPU) Arm_ADC(instruction uint32) { // Rd = Rn+Op2+Cy
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] + c.Register.R[Op2] + c.Register.R[Cy]
}

func (c CPU) Arm_SBC(instruction uint32) { // Rd = Rn-Op2+Cy-1
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] - c.Register.R[Op2] + c.Register.R[Cy] - 1

}

func (c CPU) Arm_RSC(instruction uint32) { // Rd = Op2-Rn+Cy-1
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Op2] - c.Register.R[Rn] + c.Register.R[Cy] - 1
}

func (c CPU) Arm_TST(instruction uint32) { // Void = Rn AND Op2
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rn] & c.Register.R[Op2]

}

func (c CPU) Arm_TEQ(instruction uint32) { // Void = Rn XOR Op2
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rn] ^ c.Register.R[Op2]
}

func (c CPU) Arm_CMP(instruction uint32) { // Void = Rn-Op2
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rn] - c.Register.R[Op2]
}

func (c CPU) Arm_CMN(instruction uint32) { // Void = Rn+Op2
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rn] + c.Register.R[Op2]
}

func (c CPU) Arm_ORR(instruction uint32) { // Rd = Rn OR Op2
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] | c.Register.R[Op2]
}

func (c CPU) Arm_MOV(instruction uint32) { // Rd = Op2
	Rd := ReadFlag(instruction, 12, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Op2]
}

func (c CPU) Arm_BIC(instruction uint32) { // Rd = Rn AND NOT Op2
	Rd := ReadFlag(instruction, 12, 4)
	Rn := ReadFlag(instruction, 16, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = c.Register.R[Rn] & ^c.Register.R[Op2]
}

func (c CPU) Arm_MVN(instruction uint32) { // Rd = NOT Op2
	Rd := ReadFlag(instruction, 12, 4)
	Op2 := ReadFlag(instruction, 0, 4)
	c.Register.R[Rd] = ^c.Register.R[Op2]
}

func (c CPU) ArmMul(instruction uint32) {
	map[uint32]func(uint32){
		0b0000: c.ARM_MUL,
		0b0001: c.ARM_MLA,
		0b0010: c.ARM_UMAAL,
		0b0100: c.ARM_UMULL,
		0b0101: c.ARM_UMLAL,
		0b0110: c.ARM_SMULL,
		0b0111: c.ARM_SMLAL,
		0b1000: c.ARM_SMLAxy,
		0b1001: c.ARM_SMLAWy,
		0b1001: c.ARM_SMULWy,
		0b1010: c.ARM_SMLALxy,
		0b1011: c.ARM_SMULxy,
	}[ReadFlag(instruction, 21, 4)](instruction)
}

func (c CPU) ARM_MUL(instruction uint32) {

}

func (c CPU) ARM_MLA(instruction uint32) {

}

func (c CPU) ARM_UMAAL(instruction uint32) {

}

func (c CPU) ARM_UMULL(instruction uint32) {

}

func (c CPU) ARM_UMLAL(instruction uint32) {

}

func (c CPU) ARM_SMULL(instruction uint32) {

}

func (c CPU) ARM_SMLAL(instruction uint32) {

}

func (c CPU) ARM_SMLAxy(instruction uint32) {

}

func (c CPU) ARM_SMLAWy(instruction uint32) {

}

func (c CPU) ARM_SMULWy(instruction uint32) {

}

func (c CPU) ARM_SMLALxy(instruction uint32) {

}

func (c CPU) ARM_SMULxy(instruction uint32) {

}

func (c CPU) Thumb(instruction uint32) {

}
