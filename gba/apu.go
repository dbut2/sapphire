package gba

const (
	fifoCapacity = 32
	apuCPUClock  = 16777216
)

type APU struct {
	*Motherboard

	fifoA    [fifoCapacity]int8
	fifoAHd  int
	fifoALen int
	fifoALat int8

	fifoB    [fifoCapacity]int8
	fifoBHd  int
	fifoBLen int
	fifoBLat int8

	ch1 psgSquare
	ch2 psgSquare
	ch3 psgWave
	ch4 psgNoise

	fsAccum uint32
	fsStep  uint8

	cycleAccum      uint32
	cyclesPerSample uint32
	sampleRate      uint32

	onSamples func(samples []int16)
	outBuf    []int16
	outLimit  int
}

type psgSquare struct {
	enabled   bool
	duty      uint8
	phase     float32
	freqHz    float32
	envVol    uint8
	envTimer  uint8
	envPeriod uint8
	envUp     bool
	lenTimer  uint16
	lenEnable bool

	hasSweep     bool
	swShadow     uint16
	swTimer      uint8
	swPeriod     uint8
	swShift      uint8
	swDec        bool
	swRunning    bool
	swNegApplied bool
}

type psgWave struct {
	enabled   bool
	dacOn     bool
	phase     float32
	freqHz    float32
	volShift  uint8
	forceVol  bool
	bankMode  bool
	bankSel   uint8
	pos       uint8
	lenTimer  uint16
	lenEnable bool
}

type psgNoise struct {
	enabled   bool
	lfsr      uint16
	width7    bool
	envVol    uint8
	envTimer  uint8
	envPeriod uint8
	envUp     bool
	lenTimer  uint8
	lenEnable bool
	clockHz   float32
	tickAccum float32
}

var dutyTable = [4][8]int8{
	{-1, -1, -1, -1, -1, -1, -1, 1},
	{1, -1, -1, -1, -1, -1, -1, 1},
	{1, -1, -1, -1, -1, 1, 1, 1},
	{-1, 1, 1, 1, 1, 1, 1, -1},
}

var waveVolShifts = [4]uint8{4, 0, 1, 2}

func NewAPU(m *Motherboard) *APU {
	sampleRate := uint32(32768)
	return &APU{
		Motherboard:     m,
		sampleRate:      sampleRate,
		cyclesPerSample: apuCPUClock / sampleRate,
		outLimit:        1024,
		ch1:             psgSquare{hasSweep: true},
	}
}

func (a *APU) SampleRate() int { return int(a.sampleRate) }

func (a *APU) SetOutput(cb func([]int16)) {
	a.onSamples = cb
}

func (a *APU) pushFIFOA(b byte) {
	if a.fifoALen < fifoCapacity {
		a.fifoA[(a.fifoAHd+a.fifoALen)%fifoCapacity] = int8(b)
		a.fifoALen++
	}
}

func (a *APU) pushFIFOB(b byte) {
	if a.fifoBLen < fifoCapacity {
		a.fifoB[(a.fifoBHd+a.fifoBLen)%fifoCapacity] = int8(b)
		a.fifoBLen++
	}
}

func (a *APU) popFIFOA() {
	if a.fifoALen > 0 {
		a.fifoALat = a.fifoA[a.fifoAHd]
		a.fifoAHd = (a.fifoAHd + 1) % fifoCapacity
		a.fifoALen--
	}
}

func (a *APU) popFIFOB() {
	if a.fifoBLen > 0 {
		a.fifoBLat = a.fifoB[a.fifoBHd]
		a.fifoBHd = (a.fifoBHd + 1) % fifoCapacity
		a.fifoBLen--
	}
}

func (a *APU) resetFIFOA() {
	a.fifoAHd = 0
	a.fifoALen = 0
	a.fifoALat = 0
}

func (a *APU) resetFIFOB() {
	a.fifoBHd = 0
	a.fifoBLen = 0
	a.fifoBLat = 0
}

func (a *APU) OnFIFOWrite(address uint32, value uint32, size uint32) {
	bytes := size
	if bytes == 0 {
		bytes = 4
	}
	for i := uint32(0); i < bytes; i++ {
		b := byte(value >> (i * 8))
		if address >= uint32(FIFO_B) {
			a.pushFIFOB(b)
		} else {
			a.pushFIFOA(b)
		}
	}
}

func (a *APU) OnSOUNDCNT_H(value uint16) {
	if value&(1<<11) != 0 {
		a.resetFIFOA()
	}
	if value&(1<<15) != 0 {
		a.resetFIFOB()
	}
}

func (a *APU) OnTimerOverflow(idx int) {
	cntH := ReadIORegister16(a.Memory, SOUNDCNT_H)
	timerA := (cntH >> 10) & 1
	timerB := (cntH >> 14) & 1

	if int(timerA) == idx {
		a.popFIFOA()
		if a.fifoALen <= 16 {
			a.DMA.transferSoundFIFO(uint32(FIFO_A))
		}
	}
	if int(timerB) == idx {
		a.popFIFOB()
		if a.fifoBLen <= 16 {
			a.DMA.transferSoundFIFO(uint32(FIFO_B))
		}
	}
}

func (a *APU) OnSOUND1CNT_L(v uint16) {
	a.ch1.swPeriod = uint8((v >> 4) & 7)
	a.ch1.swDec = v&(1<<3) != 0
	a.ch1.swShift = uint8(v & 7)
}

func (a *APU) OnSOUND1CNT_H(v uint16) { configSquareEnv(&a.ch1, v) }
func (a *APU) OnSOUND1CNT_X(v uint16) { configSquareFreq(&a.ch1, v) }
func (a *APU) OnSOUND2CNT_L(v uint16) { configSquareEnv(&a.ch2, v) }
func (a *APU) OnSOUND2CNT_H(v uint16) { configSquareFreq(&a.ch2, v) }

func configSquareEnv(c *psgSquare, v uint16) {
	c.lenTimer = 64 - (v & 0x3F)
	c.duty = uint8((v >> 6) & 3)
	c.envPeriod = uint8((v >> 8) & 7)
	c.envUp = v&(1<<11) != 0
	c.envVol = uint8((v >> 12) & 0xF)
	if (v>>8)&0xFF == 0 {
		c.enabled = false
	}
}

func configSquareFreq(c *psgSquare, v uint16) {
	freqReg := uint16(v & 0x7FF)
	c.freqHz = 131072.0 / float32(2048-int(freqReg))
	c.lenEnable = v&(1<<14) != 0
	if v&(1<<15) != 0 {
		c.enabled = true
		if c.lenTimer == 0 {
			c.lenTimer = 64
		}
		c.envTimer = c.envPeriod
		c.phase = 0
		if c.hasSweep {
			c.swShadow = freqReg
			c.swTimer = c.swPeriod
			if c.swTimer == 0 {
				c.swTimer = 8
			}
			c.swRunning = c.swPeriod != 0 || c.swShift != 0
			c.swNegApplied = false
			if c.swShift != 0 {
				sweepCalc(c)
			}
		}
	}
}

func (a *APU) OnSOUND3CNT_L(v uint16) {
	a.ch3.bankMode = v&(1<<5) != 0
	a.ch3.bankSel = uint8((v >> 6) & 1)
	a.ch3.dacOn = v&(1<<7) != 0
	if !a.ch3.dacOn {
		a.ch3.enabled = false
	}
}

func (a *APU) OnSOUND3CNT_H(v uint16) {
	a.ch3.lenTimer = 256 - (v & 0xFF)
	a.ch3.volShift = waveVolShifts[(v>>13)&3]
	a.ch3.forceVol = v&(1<<15) != 0
}

func (a *APU) OnSOUND3CNT_X(v uint16) {
	freqReg := uint16(v & 0x7FF)
	a.ch3.freqHz = 2097152.0 / float32(2048-int(freqReg))
	a.ch3.lenEnable = v&(1<<14) != 0
	if v&(1<<15) != 0 && a.ch3.dacOn {
		a.ch3.enabled = true
		if a.ch3.lenTimer == 0 {
			a.ch3.lenTimer = 256
		}
		a.ch3.pos = 0
		a.ch3.phase = 0
	}
}

func (a *APU) OnSOUND4CNT_L(v uint16) {
	a.ch4.lenTimer = uint8(64 - (v & 0x3F))
	a.ch4.envPeriod = uint8((v >> 8) & 7)
	a.ch4.envUp = v&(1<<11) != 0
	a.ch4.envVol = uint8((v >> 12) & 0xF)
	if (v>>8)&0xFF == 0 {
		a.ch4.enabled = false
	}
}

func (a *APU) OnSOUND4CNT_H(v uint16) {
	ratio := float32(v & 7)
	if ratio == 0 {
		ratio = 0.5
	}
	shift := (v >> 4) & 0xF
	a.ch4.clockHz = 524288.0 / ratio / float32(uint32(1)<<(shift+1))
	a.ch4.width7 = v&(1<<3) != 0
	a.ch4.lenEnable = v&(1<<14) != 0
	if v&(1<<15) != 0 {
		a.ch4.enabled = true
		if a.ch4.lenTimer == 0 {
			a.ch4.lenTimer = 64
		}
		a.ch4.envTimer = a.ch4.envPeriod
		a.ch4.lfsr = 0x7FFF
		a.ch4.tickAccum = 0
	}
}

func sweepCalc(c *psgSquare) uint16 {
	delta := c.swShadow >> c.swShift
	var newFreq uint16
	if c.swDec {
		newFreq = c.swShadow - delta
		c.swNegApplied = true
	} else {
		newFreq = c.swShadow + delta
	}
	if newFreq > 2047 {
		c.enabled = false
	}
	return newFreq
}

func (a *APU) stepSweep() {
	c := &a.ch1
	if !c.enabled || !c.swRunning {
		return
	}
	if c.swTimer > 0 {
		c.swTimer--
	}
	if c.swTimer == 0 {
		period := c.swPeriod
		if period == 0 {
			period = 8
		}
		c.swTimer = period
		if c.swPeriod != 0 {
			nf := sweepCalc(c)
			if nf <= 2047 && c.swShift != 0 {
				c.swShadow = nf
				c.freqHz = 131072.0 / float32(2048-int(nf))
				sweepCalc(c)
			}
		}
	}
}

func stepEnv(vol *uint8, timer *uint8, period uint8, up bool) {
	if period == 0 {
		return
	}
	if *timer > 0 {
		*timer--
	}
	if *timer == 0 {
		*timer = period
		if up && *vol < 15 {
			*vol++
		} else if !up && *vol > 0 {
			*vol--
		}
	}
}

func (a *APU) stepLength() {
	stepChLen16(&a.ch1.enabled, &a.ch1.lenTimer, a.ch1.lenEnable)
	stepChLen16(&a.ch2.enabled, &a.ch2.lenTimer, a.ch2.lenEnable)
	stepChLen16(&a.ch3.enabled, &a.ch3.lenTimer, a.ch3.lenEnable)
	stepChLen8(&a.ch4.enabled, &a.ch4.lenTimer, a.ch4.lenEnable)
}

func stepChLen16(enabled *bool, timer *uint16, enable bool) {
	if enable && *timer > 0 {
		*timer--
		if *timer == 0 {
			*enabled = false
		}
	}
}

func stepChLen8(enabled *bool, timer *uint8, enable bool) {
	if enable && *timer > 0 {
		*timer--
		if *timer == 0 {
			*enabled = false
		}
	}
}

func (a *APU) stepFS() {
	switch a.fsStep {
	case 0, 4:
		a.stepLength()
	case 2, 6:
		a.stepLength()
		a.stepSweep()
	case 7:
		stepEnv(&a.ch1.envVol, &a.ch1.envTimer, a.ch1.envPeriod, a.ch1.envUp)
		stepEnv(&a.ch2.envVol, &a.ch2.envTimer, a.ch2.envPeriod, a.ch2.envUp)
		stepEnv(&a.ch4.envVol, &a.ch4.envTimer, a.ch4.envPeriod, a.ch4.envUp)
	}
	a.fsStep = (a.fsStep + 1) & 7
}

func (a *APU) squareSample(c *psgSquare) int32 {
	if !c.enabled || c.envVol == 0 || c.freqHz <= 0 {
		return 0
	}
	c.phase += c.freqHz / float32(a.sampleRate)
	if c.phase >= 1 {
		c.phase -= float32(int(c.phase))
	}
	idx := int(c.phase*8) & 7
	return int32(dutyTable[c.duty][idx]) * int32(c.envVol)
}

func (a *APU) waveSample() int32 {
	c := &a.ch3
	if !c.enabled || !c.dacOn || c.volShift == 4 || c.freqHz <= 0 {
		return 0
	}
	step := c.freqHz / float32(a.sampleRate) / 32.0
	c.phase += step
	for c.phase >= 1 {
		c.phase -= 1
		c.pos = (c.pos + 1) & 31
	}
	byteVal := a.Memory.Read8(uint32(WAVE_RAM0_L)+uint32(c.pos>>1), false, true)
	var nibble uint8
	if c.pos&1 == 0 {
		nibble = (byteVal >> 4) & 0xF
	} else {
		nibble = byteVal & 0xF
	}
	sample := int32(nibble) - 8
	if c.forceVol {
		return sample * 3
	}
	return sample >> c.volShift * 4
}

func (a *APU) noiseSample() int32 {
	c := &a.ch4
	if !c.enabled || c.envVol == 0 || c.clockHz <= 0 {
		return 0
	}
	c.tickAccum += c.clockHz / float32(a.sampleRate)
	for c.tickAccum >= 1 {
		c.tickAccum -= 1
		bit := (c.lfsr ^ (c.lfsr >> 1)) & 1
		c.lfsr = (c.lfsr >> 1) | (bit << 14)
		if c.width7 {
			c.lfsr = (c.lfsr &^ (1 << 6)) | (bit << 6)
		}
	}
	out := int32(1)
	if c.lfsr&1 != 0 {
		out = -1
	}
	return out * int32(c.envVol)
}

func (a *APU) Tick(cycles uint32) {
	if a.onSamples == nil {
		return
	}
	a.cycleAccum += cycles
	for a.cycleAccum >= a.cyclesPerSample {
		a.cycleAccum -= a.cyclesPerSample
		a.emitSample()
	}
}

func (a *APU) emitSample() {
	a.fsAccum++
	if a.fsAccum >= a.sampleRate/512 {
		a.fsAccum = 0
		a.stepFS()
	}

	cntL := ReadIORegister16(a.Memory, SOUNDCNT_L)
	cntH := ReadIORegister16(a.Memory, SOUNDCNT_H)
	cntX := ReadIORegister16(a.Memory, SOUNDCNT_X)

	if cntX&(1<<7) == 0 {
		a.outBuf = append(a.outBuf, 0, 0)
		a.flushIfReady()
		return
	}

	psgShiftTab := [4]uint8{2, 1, 0, 0}
	psgVolShift := psgShiftTab[cntH&3]

	s1 := a.squareSample(&a.ch1)
	s2 := a.squareSample(&a.ch2)
	s3 := a.waveSample()
	s4 := a.noiseSample()

	var psgL, psgR int32
	if cntL&(1<<12) != 0 {
		psgL += s1
	}
	if cntL&(1<<8) != 0 {
		psgR += s1
	}
	if cntL&(1<<13) != 0 {
		psgL += s2
	}
	if cntL&(1<<9) != 0 {
		psgR += s2
	}
	if cntL&(1<<14) != 0 {
		psgL += s3
	}
	if cntL&(1<<10) != 0 {
		psgR += s3
	}
	if cntL&(1<<15) != 0 {
		psgL += s4
	}
	if cntL&(1<<11) != 0 {
		psgR += s4
	}

	masterL := int32((cntL>>4)&7) + 1
	masterR := int32(cntL&7) + 1
	psgL = (psgL * masterL) >> psgVolShift
	psgR = (psgR * masterR) >> psgVolShift

	aShift := uint(1)
	if cntH&(1<<2) != 0 {
		aShift = 0
	}
	bShift := uint(1)
	if cntH&(1<<3) != 0 {
		bShift = 0
	}
	aSample := int32(a.fifoALat) >> aShift
	bSample := int32(a.fifoBLat) >> bShift

	var left, right int32
	if cntH&(1<<9) != 0 {
		left += aSample
	}
	if cntH&(1<<8) != 0 {
		right += aSample
	}
	if cntH&(1<<13) != 0 {
		left += bSample
	}
	if cntH&(1<<12) != 0 {
		right += bSample
	}

	left = clamp16((left * 64) + (psgL * 24))
	right = clamp16((right * 64) + (psgR * 24))

	a.outBuf = append(a.outBuf, int16(left), int16(right))
	a.flushIfReady()
}

func (a *APU) flushIfReady() {
	if len(a.outBuf) >= a.outLimit {
		a.onSamples(a.outBuf)
		a.outBuf = a.outBuf[:0]
	}
}

func clamp16(v int32) int32 {
	if v > 32767 {
		return 32767
	}
	if v < -32768 {
		return -32768
	}
	return v
}
