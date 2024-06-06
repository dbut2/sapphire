//go:build !debug

package gba

type Emulator struct {
	*Motherboard
}

func (e *Emulator) stepCPU() {
	e.CPU.Step()
}
