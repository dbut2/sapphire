//go:build debug

package gba

import (
	"github.com/dbut2/sapphire/debugger/hooks"
)

type Emulator struct {
	*Motherboard

	Hooks hooks.HookService[EmuHook, *Emulator]
}

type EmuHook int

const (
	PreStepCPUEmuHook EmuHook = iota
	PostStepCPUEmuHook
)

func (e *Emulator) stepCPU() {
	e.Hooks.Hook(PreStepCPUEmuHook, e)
	e.CPU.Step()
	e.Hooks.Hook(PostStepCPUEmuHook, e)
}
