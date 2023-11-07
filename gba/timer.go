package gba

type Timer struct {
	*Motherboard

	timers  [4]uint32
	reloads [4]uint16
}

func NewTimer(m *Motherboard) *Timer {
	return &Timer{Motherboard: m}
}

var prescalerValues = [4]uint32{1, 64, 256, 1024}

func (t *Timer) Tick(cycles uint32) {
	incs := [4]uint32{}

	for i := range prescalerValues {
		t.timers[i] += cycles

		if t.timers[i] >= prescalerValues[i] {
			incs[i] = prescalerValues[i]
			t.timers[i] %= prescalerValues[i]
		}
	}

	timer, overflowed := t.tick(TM0CNT_L, TM0CNT_H, incs, false)
	SetIORegister(t.Memory, TM0CNT_L, timer)

	timer, overflowed = t.tick(TM1CNT_L, TM1CNT_H, incs, overflowed)
	SetIORegister(t.Memory, TM1CNT_L, timer)

	timer, overflowed = t.tick(TM2CNT_L, TM2CNT_H, incs, overflowed)
	SetIORegister(t.Memory, TM2CNT_L, timer)

	timer, _ = t.tick(TM3CNT_L, TM3CNT_H, incs, overflowed)
	SetIORegister(t.Memory, TM3CNT_L, timer)
}

func (t *Timer) tick(regL, regH IORegister[uint16], incs [4]uint32, prevOverflowed bool) (uint16, bool) {
	cntL := ReadIORegister(t.Memory, regL)
	cntH := ReadIORegister(t.Memory, regH)

	prescaler := ReadBits(cntH, 0, 2)
	countUpTiming := ReadBits(cntH, 2, 1)
	irqEnable := ReadBits(cntH, 6, 1)
	startStop := ReadBits(cntH, 7, 1)

	if startStop == 0 {
		return cntL, false
	}

	var inc uint32
	var overflowed bool

	switch countUpTiming {
	case 0:
		inc = incs[prescaler]
	case 1:
		if prevOverflowed {
			inc = 1
		} else {
			inc = 0
		}
	}

	inced := uint32(cntL) + inc
	cntL = uint16(inced)
	if inced > 1<<16-1 {
		overflowed = true

		cntL = t.reloads[timerIndex[regL]]

		if irqEnable == 1 {
			t.CPU.exception(0x18)
		}
	}

	return cntL, overflowed
}

func (t *Timer) Set(address uint32, value uint16) {
	t.reloads[timerAddrIndex[address]] = value
}

func (t *Timer) Reload(address uint32) {
	SetIORegister(t.Memory, indexTimer[timerAddrIndex[address]], t.reloads[timerAddrIndex[address]])
}

var timerIndex = map[IORegister[uint16]]int{
	TM0CNT_L: 0,
	TM1CNT_L: 1,
	TM2CNT_L: 2,
	TM3CNT_L: 3,
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
}
