//go:build embed_bios

package gba

import _ "embed"

//go:embed bios.gba
var bios []byte
