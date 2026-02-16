package gba

type DMAController struct {
	*Motherboard
	transferring bool

	internalSrc [4]uint32
	internalDst [4]uint32
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

func (d *DMAController) LatchAddresses(ch int) {
	SADs := [4]IORegister[uint32]{DMA0SAD, DMA1SAD, DMA2SAD, DMA3SAD}
	DADs := [4]IORegister[uint32]{DMA0DAD, DMA1DAD, DMA2DAD, DMA3DAD}
	d.internalSrc[ch] = ReadIORegister(d.Memory, SADs[ch])
	d.internalDst[ch] = ReadIORegister(d.Memory, DADs[ch])
}

func (d *DMAController) transfer(timing uint16) {
	if d.transferring {
		return
	}
	d.transferring = true
	defer func() { d.transferring = false }()

	CNT_Ls := [4]IORegister[uint16]{DMA0CNT_L, DMA1CNT_L, DMA2CNT_L, DMA3CNT_L}
	CNT_Hs := [4]IORegister[uint16]{DMA0CNT_H, DMA1CNT_H, DMA2CNT_H, DMA3CNT_H}

	for i := 0; i < 4; i++ {
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

		stepSize := uint32(2)
		if ttype == 1 {
			stepSize = 4
		}

		count := uint32(cntl)
		if count == 0 {
			if i == 3 {
				count = 0x10000
			} else {
				count = 0x4000
			}
		} else if i < 3 {
			count &= 0x3FFF
		}

		src := d.internalSrc[i]
		des := d.internalDst[i]

		if repeat == 1 && desCnt == 0b11 {
			DADs := [4]IORegister[uint32]{DMA0DAD, DMA1DAD, DMA2DAD, DMA3DAD}
			des = ReadIORegister(d.Memory, DADs[i])
		}

		for j := uint32(0); j < count; j++ {
			switch stepSize {
			case 2:
				d.Memory.Set16(des, d.Memory.Read16(src, false, false), false, false)
			case 4:
				d.Memory.Set32(des, d.Memory.Read32(src, false, false), false, false)
			}

			switch srcCnt {
			case 0b00:
				src += stepSize
			case 0b01:
				src -= stepSize
			case 0b10:
			}

			switch desCnt {
			case 0b00:
				des += stepSize
			case 0b01:
				des -= stepSize
			case 0b10:
			case 0b11:
				des += stepSize
			}
		}

		d.internalSrc[i] = src
		d.internalDst[i] = des

		if irq == 1 {
			ifReg := ReadIORegister(d.Memory, IF)
			SetIORegister(d.Memory, IF, ifReg|uint16(1<<(8+i)))
		}

		SetIORegister(d.Memory, CNT_Hs[i], SetBits(cnth, 15, 1, repeat))
	}
}
