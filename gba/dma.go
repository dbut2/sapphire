package gba

type DMAController struct {
	*Motherboard
}

func NewDMA(m *Motherboard) *DMAController {
	return &DMAController{Motherboard: m}
}

const (
	DMAImmediate uint16 = iota
	DMAVBlank
	DMAHBlank
	DMASpecial
)

func (d *DMAController) transfer(timing uint16) {
	SADs := [4]IORegister[uint32]{DMA0SAD, DMA1SAD, DMA2SAD, DMA3SAD}
	DADs := [4]IORegister[uint32]{DMA0DAD, DMA1DAD, DMA2DAD, DMA3DAD}
	CNT_Ls := [4]IORegister[uint16]{DMA0CNT_L, DMA1CNT_L, DMA2CNT_L, DMA3CNT_L}
	CNT_Hs := [4]IORegister[uint16]{DMA0CNT_H, DMA1CNT_H, DMA2CNT_H, DMA3CNT_H}

	for i := 0; i < 4; i++ {
		src := ReadIORegister(d.Memory, SADs[i])
		des := ReadIORegister(d.Memory, DADs[i])
		cntl := ReadIORegister(d.Memory, CNT_Ls[i])
		cnth := ReadIORegister(d.Memory, CNT_Hs[i])

		enabled := ReadBits(cnth, 15, 1)
		cntTiming := ReadBits(cnth, 12, 2)

		if enabled != 1 || cntTiming != timing {
			continue
		}

		irq := ReadBits(cnth, 14, 1)
		ttype := ReadBits(cnth, 10, 1)
		repeat := ReadBits(cnth, 9, 1)
		srcCnt := ReadBits(cnth, 7, 2)
		desCnt := ReadBits(cnth, 5, 2)

		size := map[uint16]int{0: 16, 1: 32}[ttype]

		for j := uint16(0); j < cntl; j++ {
			switch size {
			case 16:
				d.Memory.Set16(des, d.Memory.Read16(src, false, false), false, false)
			case 32:
				d.Memory.Set32(des, d.Memory.Read32(src, false, false), false, false)
			}

			switch srcCnt {
			case 0b00:
				src += uint32(size)
			case 0b01:
				src -= uint32(size)
			case 0b10:
			}

			switch desCnt {
			case 0b00:
				des += uint32(size)
			case 0b01:
				des -= uint32(size)
			case 0b10:
			}
		}

		if irq == 1 {
			panic("oops")
		}

		SetIORegister(d.Memory, CNT_Hs[i], SetBits(cnth, 15, 1, repeat)) // store repeat bit in enable flag
	}
}
