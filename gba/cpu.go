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

func (c CPU) cpsrAddr() *uint32 {
	return &c.CPSR
}

func (c CPU) cpsrMode() uint32 {
	return ReadBits(*c.cpsrAddr(), 0, 5)
}

func (c CPU) spsrAddr() *uint32 {
	return map[uint32]*uint32{
		0x11: &c.SPSR_fiq,
		0x12: &c.SPSR_irq,
		0x13: &c.SPSR_svc,
		0x17: &c.SPSR_abt,
		0x1B: &c.SPSR_und,
	}[c.cpsrMode()]
}

type Register struct {
	R [16]uint32
}

//\d+: c.(\w+),
//func (c CPU) $1(instruction uint32) {\n\n}\n

//\t\t(\w+)(\{cond\})?(\w+).*
//\t\t: c.Arm_$1$3,

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
		3: c.ArmMemory,
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
		0b0000: c.Arm_MUL,
		0b0001: c.Arm_MLA,
		0b0010: c.Arm_UMAAL,
		0b0100: c.Arm_UMULL,
		0b0101: c.Arm_UMLAL,
		0b0110: c.Arm_SMULL,
		0b0111: c.Arm_SMLAL,
		0b1000: c.Arm_SMLAxy,
		//0b1001: c.Arm_SMLAWy,
		//0b1001: c.Arm_SMULWy,
		0b1010: c.Arm_SMLALxy,
		0b1011: c.Arm_SMULxy,
	}[ReadBits(instruction, 21, 4)](instruction)
}

func (c CPU) Arm_MUL(instruction uint32) {

}

func (c CPU) Arm_MLA(instruction uint32) {

}

func (c CPU) Arm_UMAAL(instruction uint32) {

}

func (c CPU) Arm_UMULL(instruction uint32) {

}

func (c CPU) Arm_UMLAL(instruction uint32) {

}

func (c CPU) Arm_SMULL(instruction uint32) {

}

func (c CPU) Arm_SMLAL(instruction uint32) {

}

func (c CPU) Arm_SMLAxy(instruction uint32) {

}

func (c CPU) Arm_SMLAWy(instruction uint32) {

}

func (c CPU) Arm_SMULWy(instruction uint32) {

}

func (c CPU) Arm_SMLALxy(instruction uint32) {

}

func (c CPU) Arm_SMULxy(instruction uint32) {

}

func (c CPU) ArmMemory(instruction uint32) {
	map[uint32]func(uint32){
		0:  c.Arm_LDRB,
		1:  c.Arm_LDRT,
		2:  c.Arm_LDRH,
		3:  c.Arm_LDRD,
		4:  c.Arm_LDRSB,
		5:  c.Arm_LDRSH,
		6:  c.Arm_LDM,
		7:  c.Arm_STRB,
		8:  c.Arm_STRT,
		9:  c.Arm_STRH,
		10: c.Arm_STRD,
		11: c.Arm_STM,
		12: c.Arm_SWPB,
		13: c.Arm_PLD,
	}[0](instruction) // todo
}

func (c CPU) Arm_LDRB(instruction uint32) {

}

func (c CPU) Arm_LDRT(instruction uint32) {

}

func (c CPU) Arm_LDRH(instruction uint32) {

}

func (c CPU) Arm_LDRD(instruction uint32) {

}

func (c CPU) Arm_LDRSB(instruction uint32) {

}

func (c CPU) Arm_LDRSH(instruction uint32) {

}

func (c CPU) Arm_LDM(instruction uint32) {

}

func (c CPU) Arm_STRB(instruction uint32) {

}

func (c CPU) Arm_STRT(instruction uint32) {

}

func (c CPU) Arm_STRH(instruction uint32) {

}

func (c CPU) Arm_STRD(instruction uint32) {

}

func (c CPU) Arm_STM(instruction uint32) {

}

func (c CPU) Arm_SWPB(instruction uint32) {

}

func (c CPU) Arm_PLD(instruction uint32) {

}

func (c CPU) ArmJump(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Arm_B,
		1: c.Arm_BL,
		2: c.Arm_BX,
		3: c.Arm_BLX,
		4: c.Arm_BLX,
		5: c.Arm_MRS,
		6: c.Arm_MSR,
		7: c.Arm_SWI,
		8: c.Arm_BKPT,
		//The Undefined Instruction    2S+1I+1N ----  PC=4, ARM Und mode, LR=$+4
		//cond=false                   1S       ----  Any opcode with condition=false
		11: c.Arm_NOP,
		12: c.Arm_CLZ,
		13: c.Arm_QADD,
		14: c.Arm_QSUB,
		15: c.Arm_QDADD,
		16: c.Arm_QDSUB,
	}[0](instruction) // todo

}

func (c CPU) Arm_B(instruction uint32) {

}

func (c CPU) Arm_BL(instruction uint32) {

}

func (c CPU) Arm_BX(instruction uint32) {

}

func (c CPU) Arm_BLX(instruction uint32) {

}

func (c CPU) Arm_BLX(instruction uint32) {

}

func (c CPU) Arm_MRS(instruction uint32) {

}

func (c CPU) Arm_MSR(instruction uint32) {

}

func (c CPU) Arm_SWI(instruction uint32) {

}

func (c CPU) Arm_BKPT(instruction uint32) {

}

//The Undefined Instruction    2S+1I+1N ----  PC=4, ARM Und mode, LR=$+4
//cond=false                   1S       ----  Any opcode with condition=false

func (c CPU) Arm_NOP(instruction uint32) {

}

func (c CPU) Arm_CLZ(instruction uint32) {

}

func (c CPU) Arm_QADD(instruction uint32) {

}

func (c CPU) Arm_QSUB(instruction uint32) {

}

func (c CPU) Arm_QDADD(instruction uint32) {

}

func (c CPU) Arm_QDSUB(instruction uint32) {

}

func (c CPU) ArmCoprocessor(instruction uint32) {
	map[uint32]func(uint32){
		0: c.Arm_CDP,
		1: c.Arm_STC,
		2: c.Arm_LDC,
		3: c.Arm_MCR,
		4: c.Arm_MRC,
		5: c.Arm_MCRR,
		6: c.Arm_MRRC,
	}[0](instruction) // todo
}

func (c CPU) Arm_CDP(instruction uint32) {

}

func (c CPU) Arm_STC(instruction uint32) {

}

func (c CPU) Arm_LDC(instruction uint32) {

}

func (c CPU) Arm_MCR(instruction uint32) {

}

func (c CPU) Arm_MRC(instruction uint32) {

}

func (c CPU) Arm_MCRR(instruction uint32) {

}

func (c CPU) Arm_MRRC(instruction uint32) {

}

func (c CPU) Thumb(instruction uint32) {
	map[uint32]func(uint32){}[0](instruction)
}
