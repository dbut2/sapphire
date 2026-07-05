package gba

type Emulator struct {
	*Motherboard
	FastForward bool
	hblank      bool
}

func (e *Emulator) stepCPU() {
	e.CPU.Step()
}
