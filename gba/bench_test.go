package gba

import (
	"os"
	"testing"
)

func BenchmarkFrames(b *testing.B) {
	gamepak, err := os.ReadFile("../sapphire.gba")
	if err != nil {
		b.Skipf("rom not available: %v", err)
	}
	emu := NewEmu(gamepak)
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
