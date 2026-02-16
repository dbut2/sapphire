package gba

type Emulator struct {
	*Motherboard
	FastForward bool
}

func (e *Emulator) stepCPU() {
	e.CPU.Step()
}
