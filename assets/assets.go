package assets

import _ "embed"

//go:embed bios.gba
var BIOS []byte

//go:embed gamepak.gba
var Gamepak []byte
