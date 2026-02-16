package gba

type Timer struct {
	*Motherboard

	prescalerAccum [4]uint32
	reloads        [4]uint16
	counters       [4]uint16
	control        [4]uint16
	enabled        [4]bool
}

func NewTimer(m *Motherboard) *Timer {
	return &Timer{Motherboard: m}
}

var prescalerValues = [4]uint32{1, 64, 256, 1024}

func (t *Timer) Tick(cycles uint32) {
	incs := [4]uint32{}

	for i := range prescalerValues {
		t.prescalerAccum[i] += cycles
		incs[i] = t.prescalerAccum[i] / prescalerValues[i]
		t.prescalerAccum[i] %= prescalerValues[i]
	}

	overflowed := t.tick(0, incs, false)
	overflowed = t.tick(1, incs, overflowed)
	overflowed = t.tick(2, incs, overflowed)
	t.tick(3, incs, overflowed)
}

func (t *Timer) tick(idx int, incs [4]uint32, prevOverflowed bool) bool {
	if !t.enabled[idx] {
		return false
	}

	cntH := t.control[idx]
	prescaler := cntH & 3
	countUpTiming := (cntH >> 2) & 1

	var inc uint32
	if countUpTiming == 0 {
		inc = incs[prescaler]
	} else if prevOverflowed {
		inc = 1
	}

	inced := uint32(t.counters[idx]) + inc
	t.counters[idx] = uint16(inced)

	if inced > 0xFFFF {
		t.counters[idx] = t.reloads[idx]

		if (cntH>>6)&1 == 1 { // IRQ enable
			ifReg := ReadIORegister16(t.Memory, IF)
			SetIORegister16(t.Memory, IF, ifReg|uint16(1<<(3+idx)))
		}
		return true
	}

	return false
}

func (t *Timer) SyncToMemory() {
	for i, reg := range [4]IORegister[uint16]{TM0CNT_L, TM1CNT_L, TM2CNT_L, TM3CNT_L} {
		SetIORegister16(t.Memory, reg, t.counters[i])
	}
}

func (t *Timer) Set(address uint32, value uint16) {
	t.reloads[timerAddrIndex[address]] = value
}

func (t *Timer) Reload(address uint32) {
	idx := timerAddrIndex[address]
	t.counters[idx] = t.reloads[idx]
	SetIORegister16(t.Memory, indexTimer[idx], t.reloads[idx])
}

func (t *Timer) OnControlWrite(address uint32, value uint16) {
	idx := timerAddrIndex[address]
	wasEnabled := t.enabled[idx]
	t.control[idx] = value
	t.enabled[idx] = (value>>7)&1 == 1

	if !wasEnabled && t.enabled[idx] {
		t.counters[idx] = t.reloads[idx]
		SetIORegister16(t.Memory, indexTimer[idx], t.reloads[idx])
	}
}

var indexTimer = map[int]IORegister[uint16]{
	0: TM0CNT_L,
	1: TM1CNT_L,
	2: TM2CNT_L,
	3: TM3CNT_L,
}

var timerAddrIndex = map[uint32]int{
	uint32(TM0CNT_L): 0,
	uint32(TM1CNT_L): 1,
	uint32(TM2CNT_L): 2,
	uint32(TM3CNT_L): 3,
	uint32(TM0CNT_H): 0,
	uint32(TM1CNT_H): 1,
	uint32(TM2CNT_H): 2,
	uint32(TM3CNT_H): 3,
}
