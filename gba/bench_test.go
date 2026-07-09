package gba

import (
	"testing"

	"dbut.dev/sapphire/assets"
)

func BenchmarkFrames(b *testing.B) {
	emu := NewEmu(assets.Gamepak, assets.BIOS)
	emu.LCD.ShowFPS = false
	emu.PreBoot()
	for i := 0; i < 60; i++ {
		emu.frame()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emu.frame()
	}
}
