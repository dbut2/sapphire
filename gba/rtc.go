package gba

import "time"

type rtcPin = uint8

const (
	rtcPinSCK rtcPin = iota
	rtcPinSIO
	rtcPinCS
)

const (
	rtcRegForceReset = 0x0
	rtcRegDatetime   = 0x2
	rtcRegForceIRQ   = 0x3
	rtcRegCtrl       = 0x4
	rtcRegTime       = 0x6
)

var rtcCommandParamLen = [8]uint8{0, 0, 7, 0, 1, 0, 3, 0}

type RTC struct {
	state uint8

	sck bool
	sio uint8
	cs  bool

	shiftCount  uint8
	data        uint8
	reg         uint8
	currentByte uint8
	buf         [7]uint8

	ctrl uint8

	*Motherboard
}

func NewRTC(m *Motherboard) *RTC {
	return &RTC{
		state:       0xFF,
		Motherboard: m,
	}
}

func (r *RTC) Read() uint8 {
	if r.cs {
		return r.sio << rtcPinSIO
	}
	return 0
}

func (r *RTC) Write(data uint8, mask uint8) {
	cs := r.cs
	sck := r.sck

	if mask>>rtcPinSCK&1 != 0 {
		r.sck = data>>rtcPinSCK&1 != 0
	}
	if mask>>rtcPinSIO&1 != 0 {
		r.sio = (data >> rtcPinSIO) & 1
	}
	if mask>>rtcPinCS&1 != 0 {
		r.cs = data>>rtcPinCS&1 != 0
	}

	if r.cs {
		if !cs {
			r.state = 0
			return
		}

		if !sck && r.sck {
			switch r.state {
			case 0:
				r.receiveCommand()
			case 1:
				r.receiveParameter()
			case 2:
				r.fetchRegisterData()
			}
		}
	}
}

func (r *RTC) readSIO() bool {
	r.data &^= 1 << r.shiftCount
	r.data |= r.sio << r.shiftCount

	r.shiftCount++
	if r.shiftCount == 8 {
		r.shiftCount = 0
		return true
	}
	return false
}

func (r *RTC) receiveCommand() {
	if !r.readSIO() {
		return
	}

	r.data = rtcReverseCommand(r.data)
	r.reg = (r.data >> 4) & 0b111
	r.shiftCount = 0
	r.currentByte = 0

	mode := (r.data >> 7) & 1
	switch mode {
	case 0:
		if rtcCommandParamLen[r.reg] > 0 {
			r.state = 1
		} else {
			r.writeReg()
			r.state = 0xFF
		}
	case 1:
		r.readReg()
		if rtcCommandParamLen[r.reg] > 0 {
			r.state = 2
		} else {
			r.state = 0xFF
		}
	}
}

func (r *RTC) receiveParameter() {
	if r.currentByte < rtcCommandParamLen[r.reg] && r.readSIO() {
		r.buf[r.currentByte] = r.data
		r.currentByte++
		if r.currentByte == rtcCommandParamLen[r.reg] {
			r.writeReg()
			r.state = 0xFF
		}
	}
}

func (r *RTC) fetchRegisterData() {
	r.sio = r.buf[r.currentByte] & 1
	r.buf[r.currentByte] >>= 1

	r.shiftCount++
	if r.shiftCount == 8 {
		r.shiftCount = 0
		r.currentByte++
		if r.currentByte == rtcCommandParamLen[r.reg] {
			r.state = 0xFF
		}
	}
}

func rtcReverseCommand(cmd uint8) uint8 {
	switch {
	case (cmd >> 4) == 0b0110:
		cmd = (cmd << 4) | (cmd >> 4)
		cmd = ((cmd & 0x33) << 2) | ((cmd & 0xCC) >> 2)
		cmd = ((cmd & 0x55) << 1) | ((cmd & 0xAA) >> 1)
	case (cmd & 0b1111) == 0b0110:
	}
	return cmd
}

func (r *RTC) readReg() {
	switch r.reg {
	case rtcRegDatetime:
		now := time.Now()
		r.buf[0] = rtcBCD(uint8(now.Year() - 2000))
		r.buf[1] = rtcBCD(uint8(now.Month()))
		r.buf[2] = rtcBCD(uint8(now.Day()))
		r.buf[3] = rtcBCD(uint8(now.Weekday()))
		r.buf[4] = rtcBCD(r.hour(uint8(now.Hour())))
		r.buf[5] = rtcBCD(uint8(now.Minute()))
		r.buf[6] = rtcBCD(uint8(now.Second()))
	case rtcRegCtrl:
		r.buf[0] = r.ctrl
	case rtcRegTime:
		now := time.Now()
		r.buf[0] = rtcBCD(r.hour(uint8(now.Hour())))
		r.buf[1] = rtcBCD(uint8(now.Minute()))
		r.buf[2] = rtcBCD(uint8(now.Second()))
	}
}

func (r *RTC) writeReg() {
	switch r.reg {
	case rtcRegForceReset:
		r.ctrl = 0
	case rtcRegForceIRQ:
		ifReg := ReadIORegister(r.Memory, IF)
		SetIORegister(r.Memory, IF, ifReg|1<<13)
	case rtcRegCtrl:
		r.ctrl = r.buf[0] & 0x7F
	}
}

func rtcBCD(x uint8) uint8 {
	y := uint8(0)
	e := uint8(1)
	for x > 0 {
		y += (x % 10) * e
		e *= 16
		x /= 10
	}
	return y
}

func (r *RTC) hour(h24 uint8) uint8 {
	if r.ctrl>>6&1 == 0 && h24 >= 12 {
		return (h24 - 12) | 0x80
	}
	return h24
}
