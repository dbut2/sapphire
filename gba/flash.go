package gba

import "bytes"

const (
	flashReady = iota
	flashCmd1
	flashCmd2
	flashIDMode
	flashWriteMode
	flashEraseMode
	flashEraseCmd1
	flashEraseCmd2
	flashBankSelect
)

type Flash struct {
	*Motherboard

	state uint8
	bank  uint8

	manufacturer uint8
	device       uint8
	size         int

	Data [128 * 1024]byte

	saveFunc func(data []byte)
	dirty    bool
}

func NewFlash(m *Motherboard, gamepak []byte) *Flash {
	f := &Flash{
		Motherboard: m,
	}

	f.detectFlashType(gamepak)

	for i := 0; i < f.size*1024; i++ {
		f.Data[i] = 0xFF
	}
	return f
}

func (f *Flash) detectFlashType(rom []byte) {
	has128 := false
	has64 := false
	idx128 := bytes.Index(rom, []byte("FLASH1M_V"))
	idx64 := bytes.Index(rom, []byte("FLASH_V1"))
	idx512 := bytes.Index(rom, []byte("FLASH512_V"))
	best := -1
	for _, idx := range []int{idx128, idx64, idx512} {
		if idx >= 0 && (best < 0 || idx < best) {
			best = idx
		}
	}
	if best >= 0 {
		has128 = best == idx128
		has64 = !has128
	}

	if has128 {
		f.manufacturer = 0x62
		f.device = 0x13
		f.size = 128
	} else if has64 {
		f.manufacturer = 0x32
		f.device = 0x1B
		f.size = 64
	} else {
		f.manufacturer = 0x62
		f.device = 0x13
		f.size = 128
	}
}

func (f *Flash) Read(address uint32) uint8 {
	addr := address & 0xFFFF
	if f.state == flashIDMode {
		switch addr {
		case 0x0000:
			return f.manufacturer
		case 0x0001:
			return f.device
		default:
			return 0
		}
	}

	offset := uint32(f.bank)*0x10000 + uint32(addr)
	if offset >= uint32(f.size)*1024 {
		return 0xFF
	}
	return f.Data[offset]
}

func (f *Flash) Write(address uint32, value uint8) {
	addr := address & 0xFFFF

	switch f.state {
	case flashReady:
		if addr == 0x5555 && value == 0xAA {
			f.state = flashCmd1
		}
	case flashCmd1:
		if addr == 0x2AAA && value == 0x55 {
			f.state = flashCmd2
		} else {
			f.state = flashReady
		}
	case flashCmd2:
		if addr == 0x5555 {
			switch value {
			case 0x90:
				f.state = flashIDMode
			case 0xF0:
				f.state = flashReady
			case 0x80:
				f.state = flashEraseMode
			case 0xA0:
				f.state = flashWriteMode
			case 0xB0:
				if f.size == 128 {
					f.state = flashBankSelect
				} else {
					f.state = flashReady
				}
			default:
				f.state = flashReady
			}
		} else {
			f.state = flashReady
		}
	case flashIDMode:
		if addr == 0x5555 && value == 0xAA {
			f.state = flashCmd1
		} else if value == 0xF0 {
			f.state = flashReady
		}
	case flashWriteMode:
		offset := uint32(f.bank)*0x10000 + uint32(addr)
		if offset < uint32(f.size)*1024 {
			f.Data[offset] = value
			f.dirty = true
		}
		f.state = flashReady
	case flashEraseMode:
		if addr == 0x5555 && value == 0xAA {
			f.state = flashEraseCmd1
		} else {
			f.state = flashReady
		}
	case flashEraseCmd1:
		if addr == 0x2AAA && value == 0x55 {
			f.state = flashEraseCmd2
		} else {
			f.state = flashReady
		}
	case flashEraseCmd2:
		if addr == 0x5555 && value == 0x10 {
			for i := 0; i < f.size*1024; i++ {
				f.Data[i] = 0xFF
			}
			f.dirty = true
			f.state = flashReady
		} else if value == 0x30 {
			sector := uint32(addr&0xF000) + uint32(f.bank)*0x10000
			maxAddr := uint32(f.size) * 1024
			for i := uint32(0); i < 0x1000 && sector+i < maxAddr; i++ {
				f.Data[sector+i] = 0xFF
			}
			f.dirty = true
			f.state = flashReady
		} else {
			f.state = flashReady
		}
	case flashBankSelect:
		if addr == 0x0000 {
			f.bank = value & 1
		}
		f.state = flashReady
	}
}

func (f *Flash) LoadData(data []byte) {
	n := len(data)
	if n > f.size*1024 {
		n = f.size * 1024
	}
	copy(f.Data[:n], data[:n])
}

func (f *Flash) OnSave(fn func(data []byte)) {
	f.saveFunc = fn
}

func (f *Flash) Flush() {
	if !f.dirty || f.saveFunc == nil {
		return
	}
	f.saveFunc(f.Data[:f.size*1024])
	f.dirty = false
}
