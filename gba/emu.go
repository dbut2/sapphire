package gba

type Emulator struct {
	*Motherboard
}

func NewEmu(gamepak []byte) *Emulator {
	return &Emulator{Motherboard: NewMotherboard(gamepak)}
}
