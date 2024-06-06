//go:build debug

package main

import (
	"os"

	"github.com/dbut2/sapphire/gba"
)

func main() {
	gameData, err := os.ReadFile("sapphire.gba")
	if err != nil {
		panic(err.Error())
	}
	e := gba.NewEmu(gameData)

	e.Hooks.RegisterHook(gba.PreStepCPUEmuHook, func(emulator *gba.Emulator) {
		// Do something before CPU step
	})
	e.Boot()
}
