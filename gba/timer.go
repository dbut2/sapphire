package gba

type Timer struct {
	*Motherboard

	reloads [4]uint16
}

func NewTimer(m *Motherboard) *Timer {
	return &Timer{Motherboard: m}
}

func (t *Timer) Tick(pre, post uint32) {
	timer, overflowed := t.tick(TM0CNT_L, TM0CNT_H, pre, post, false)
	SetIORegister(t.Memory, TM0CNT_L, timer)

	timer, overflowed = t.tick(TM1CNT_L, TM1CNT_H, pre, post, overflowed)
	SetIORegister(t.Memory, TM1CNT_L, timer)

	timer, overflowed = t.tick(TM2CNT_L, TM2CNT_H, pre, post, overflowed)
	SetIORegister(t.Memory, TM2CNT_L, timer)

	timer, _ = t.tick(TM3CNT_L, TM3CNT_H, pre, post, overflowed)
	SetIORegister(t.Memory, TM3CNT_L, timer)
}

func (t *Timer) tick(regL, regH IORegister[uint16], pre, post uint32, prevOverflowed bool) (uint16, bool) {
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
		switch prescaler {
		case 0:
			inc = (post>>0 - pre>>0) << 0
		case 1:
			inc = (post>>6 - pre>>6) << 6
		case 2:
			inc = (post>>8 - pre>>8) << 8
		case 3:
			inc = (post>>10 - pre>>10) << 10

		}
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

func (t Timer) set(address uint32, value uint16) {
	t.reloads[timerAddrIndex[address]] = value
}

var timerIndex = map[IORegister[uint16]]int{
	TM0CNT_L: 0,
	TM1CNT_L: 1,
	TM2CNT_L: 2,
	TM3CNT_L: 3,
}

var timerAddrIndex = map[uint32]int{
	uint32(TM0CNT_L): 0,
	uint32(TM1CNT_L): 1,
	uint32(TM2CNT_L): 2,
	uint32(TM3CNT_L): 3,
}
