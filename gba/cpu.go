package gba

type CPU struct {
	*Motherboard
	CPURegisters
}

func NewCPU(m *Motherboard) *CPU {
	return &CPU{Motherboard: m}
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

func (c CPU) registerAddr(r uint32) *uint32 {
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
		}[*c.cpsrAddr()],
		9: map[uint32]*uint32{
			0x10: &c.R9,
			0x11: &c.R9_fiq,
			0x12: &c.R9,
			0x13: &c.R9,
			0x17: &c.R9,
			0x1B: &c.R9,
		}[*c.cpsrAddr()],
		10: map[uint32]*uint32{
			0x10: &c.R10,
			0x11: &c.R10_fiq,
			0x12: &c.R10,
			0x13: &c.R10,
			0x17: &c.R10,
			0x1B: &c.R10,
		}[*c.cpsrAddr()],
		11: map[uint32]*uint32{
			0x10: &c.R11,
			0x11: &c.R11_fiq,
			0x12: &c.R11,
			0x13: &c.R11,
			0x17: &c.R11,
			0x1B: &c.R11,
		}[*c.cpsrAddr()],
		12: map[uint32]*uint32{
			0x10: &c.R12,
			0x11: &c.R12_fiq,
			0x12: &c.R12,
			0x13: &c.R12,
			0x17: &c.R12,
			0x1B: &c.R12,
		}[*c.cpsrAddr()],
		13: map[uint32]*uint32{
			0x10: &c.R13,
			0x11: &c.R13_fiq,
			0x12: &c.R13_irq,
			0x13: &c.R13_svc,
			0x17: &c.R13_abt,
			0x1B: &c.R13_und,
		}[*c.cpsrAddr()],
		14: map[uint32]*uint32{
			0x10: &c.R14,
			0x11: &c.R14_fiq,
			0x12: &c.R14_irq,
			0x13: &c.R14_svc,
			0x17: &c.R14_abt,
			0x1B: &c.R14_und,
		}[*c.cpsrAddr()],
		15: &c.R15,
	}[r]
}

func (c CPU) cpsrAddr() *uint32 {
	return &c.CPSR
}

func (c CPU) spsrAddr() *uint32 {
	return map[uint32]*uint32{
		0x11: &c.SPSR_fiq,
		0x12: &c.SPSR_irq,
		0x13: &c.SPSR_svc,
		0x17: &c.SPSR_abt,
		0x1B: &c.SPSR_und,
	}[*c.cpsrAddr()]
}

type Register struct {
	R [16]uint32
}

func (c CPU) Do(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Thumb,
		1: c.Arm,
	}[1](instruction) // todo
}

func (c CPU) Arm(instruction uint32) {
	map[uint32]func(uint32){
		0: c.ArmBranch,
		1: c.ArmALU,
		2: c.ArmMul,
	}[1](instruction) // todo
}

func (c CPU) ArmBranch(instruction uint32) {
	map[uint32]func(uint32){}[0](instruction) // todo
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
	}[ReadBits(instruction, 21, 4)](instruction)
}

func (c CPU) Arm_AND(instruction uint32) { // Rd = Rn AND Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) & *c.registerAddr(Op2)
}

func (c CPU) Arm_EOR(instruction uint32) { // Rd = Rn XOR Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) ^ *c.registerAddr(Op2)
}

func (c CPU) Arm_SUB(instruction uint32) { // Rd = Rn-Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) - *c.registerAddr(Op2)
}

func (c CPU) Arm_RSB(instruction uint32) { // Rd = Op2-Rn
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Op2) - *c.registerAddr(Rn)
}

func (c CPU) Arm_ADD(instruction uint32) { // Rd = Rn+Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) + *c.registerAddr(Op2)
}

func (c CPU) Arm_ADC(instruction uint32) { // Rd = Rn+Op2+Cy
	//Rd := ReadBits(instruction, 12, 4)
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rd) = *c.registerAddr(Rn) + *c.registerAddr(Op2) + *c.registerAddr(Cy)
}

func (c CPU) Arm_SBC(instruction uint32) { // Rd = Rn-Op2+Cy-1
	//Rd := ReadBits(instruction, 12, 4)
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rd) = *c.registerAddr(Rn) - *c.registerAddr(Op2) + *c.registerAddr(Cy) - 1
}

func (c CPU) Arm_RSC(instruction uint32) { // Rd = Op2-Rn+Cy-1
	//Rd := ReadBits(instruction, 12, 4)
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rd) = *c.registerAddr(Op2) - *c.registerAddr(Rn) + *c.registerAddr(Cy) - 1
}

func (c CPU) Arm_TST(instruction uint32) { // Void = Rn AND Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) & *c.registerAddr(Op2)
}

func (c CPU) Arm_TEQ(instruction uint32) { // Void = Rn XOR Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) ^ *c.registerAddr(Op2)
}

func (c CPU) Arm_CMP(instruction uint32) { // Void = Rn-Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) - *c.registerAddr(Op2)
}

func (c CPU) Arm_CMN(instruction uint32) { // Void = Rn+Op2
	//Rn := ReadBits(instruction, 16, 4)
	//Op2 := ReadBits(instruction, 0, 4)
	//*c.registerAddr(Rn) + *c.registerAddr(Op2)
}

func (c CPU) Arm_ORR(instruction uint32) { // Rd = Rn OR Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) | *c.registerAddr(Op2)
}

func (c CPU) Arm_MOV(instruction uint32) { // Rd = Op2
	Rd := ReadBits(instruction, 12, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Op2)
}

func (c CPU) Arm_BIC(instruction uint32) { // Rd = Rn AND NOT Op2
	Rd := ReadBits(instruction, 12, 4)
	Rn := ReadBits(instruction, 16, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = *c.registerAddr(Rn) & ^*c.registerAddr(Op2)
}

func (c CPU) Arm_MVN(instruction uint32) { // Rd = NOT Op2
	Rd := ReadBits(instruction, 12, 4)
	Op2 := ReadBits(instruction, 0, 4)
	*c.registerAddr(Rd) = ^*c.registerAddr(Op2)
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
		//0b1001: c.ARM_SMLAWy,
		//0b1001: c.ARM_SMULWy,
		0b1010: c.ARM_SMLALxy,
		0b1011: c.ARM_SMULxy,
	}[ReadBits(instruction, 21, 4)](instruction)
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
