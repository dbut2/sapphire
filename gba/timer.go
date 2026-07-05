package gba

type Timer struct {
	*Motherboard

	prescalerAccum [4]uint32
	reloads        [4]uint16
	counters       [4]uint16
	control        [4]uint16
	enabled        [4]bool

	acc      uint32
	deadline uint32
}

func NewTimer(m *Motherboard) *Timer {
	return &Timer{Motherboard: m, deadline: idleDeadline}
}

const idleDeadline = 1 << 30

func (t *Timer) untilDeadline() uint32 {
	if t.acc >= t.deadline {
		return 1
	}
	return t.deadline - t.acc
}

func (t *Timer) Tick(cycles uint32) {
	t.acc += cycles
	if t.acc >= t.deadline {
		t.flush()
	}
}

func (t *Timer) flush() {
	cycles := t.acc
	t.acc = 0

	t.prescalerAccum[0] += cycles
	inc0 := t.prescalerAccum[0]
	t.prescalerAccum[0] = 0

	t.prescalerAccum[1] += cycles
	inc1 := t.prescalerAccum[1] >> 6
	t.prescalerAccum[1] &= 63

	t.prescalerAccum[2] += cycles
	inc2 := t.prescalerAccum[2] >> 8
	t.prescalerAccum[2] &= 255

	t.prescalerAccum[3] += cycles
	inc3 := t.prescalerAccum[3] >> 10
	t.prescalerAccum[3] &= 1023

	incs := [4]uint32{inc0, inc1, inc2, inc3}
	overflowed := t.tick(0, incs, false)
	overflowed = t.tick(1, incs, overflowed)
	overflowed = t.tick(2, incs, overflowed)
	t.tick(3, incs, overflowed)

	t.recomputeDeadline()
}

var prescalerShift = [4]uint32{0, 6, 8, 10}

func (t *Timer) recomputeDeadline() {
	deadline := uint32(idleDeadline)
	for i := 0; i < 4; i++ {
		if !t.enabled[i] || (t.control[i]>>2)&1 == 1 {
			continue
		}
		prescaler := t.control[i] & 3
		until := (0x10000-uint32(t.counters[i]))<<prescalerShift[prescaler] - t.prescalerAccum[prescaler]
		if until < deadline {
			deadline = until
		}
	}
	t.deadline = deadline
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
		if t.APU != nil && idx < 2 {
			t.APU.OnTimerOverflow(idx)
		}
		return true
	}

	return false
}

func (t *Timer) SyncToMemory() {
	t.flush()
	SetIORegister16(t.Memory, TM0CNT_L, t.counters[0])
	SetIORegister16(t.Memory, TM1CNT_L, t.counters[1])
	SetIORegister16(t.Memory, TM2CNT_L, t.counters[2])
	SetIORegister16(t.Memory, TM3CNT_L, t.counters[3])
}

func (t *Timer) Set(address uint32, value uint16) {
	t.flush()
	t.reloads[timerAddrIndex(address)] = value
}

func (t *Timer) Reload(address uint32) {
	idx := timerAddrIndex(address)
	t.counters[idx] = t.reloads[idx]
	SetIORegister16(t.Memory, indexTimer[idx], t.reloads[idx])
	t.recomputeDeadline()
}

func (t *Timer) OnControlWrite(address uint32, value uint16) {
	t.flush()
	idx := timerAddrIndex(address)
	wasEnabled := t.enabled[idx]
	t.control[idx] = value
	t.enabled[idx] = (value>>7)&1 == 1

	if !wasEnabled && t.enabled[idx] {
		t.counters[idx] = t.reloads[idx]
		SetIORegister16(t.Memory, indexTimer[idx], t.reloads[idx])
	}
	t.recomputeDeadline()
}

var indexTimer = [4]IORegister[uint16]{TM0CNT_L, TM1CNT_L, TM2CNT_L, TM3CNT_L}

func timerAddrIndex(address uint32) int {
	return int((address >> 2) & 3)
}
