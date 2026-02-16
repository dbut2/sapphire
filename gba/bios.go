package gba

import _ "embed"

// Source a GBA bios and place here at ./bios.gba
//
//go:embed bios.gba
var bios []byte
