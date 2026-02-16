//go:build js && wasm && embed_gamepak

package main

import _ "embed"

//go:embed gamepak.gba
var game []byte
