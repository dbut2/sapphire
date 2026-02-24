//go:build js && wasm

package main

import (
	_ "embed"
	"encoding/base64"
	"os"
	"syscall/js"

	"dbut.dev/sapphire/gba"
)

const saveKey = "sapphire.sav"

var keyMap = map[string]uint16{
	"KeyZ":       1 << 0, // A
	"KeyX":       1 << 1, // B
	"Backspace":  1 << 2, // Select
	"Enter":      1 << 3, // Start
	"ArrowRight": 1 << 4, // Right
	"ArrowLeft":  1 << 5, // Left
	"ArrowUp":    1 << 6, // Up
	"ArrowDown":  1 << 7, // Down
	"KeyS":       1 << 8, // R
	"KeyA":       1 << 9, // L
}

func main() {
	canvasID := "sapphire-canvas"
	for i, arg := range os.Args {
		if arg == "--canvas" && i+1 < len(os.Args) {
			canvasID = os.Args[i+1]
		}
	}

	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", canvasID)
	if canvasEl.IsNull() {
		panic("canvas element #" + canvasID + " not found")
	}

	ctx := canvasEl.Call("getContext", "2d")
	ctx.Set("imageSmoothingEnabled", false)

	width := 240
	height := 160
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	e := gba.NewEmu(game)

	localStorage := js.Global().Get("localStorage")
	if encoded := localStorage.Call("getItem", saveKey); !encoded.IsNull() {
		if data, err := base64.StdEncoding.DecodeString(encoded.String()); err == nil {
			e.Flash.LoadData(data)
		}
	}
	e.Flash.OnSave(func(data []byte) {
		encoded := base64.StdEncoding.EncodeToString(data)
		localStorage.Call("setItem", saveKey, encoded)
	})

	front := e.LCD.Front()
	imageData := ctx.Call("createImageData", width, height)
	jsPixels := js.Global().Get("Uint8Array").New(width * height * 4)

	e.LCD.SetDraw(func() {
		js.CopyBytesToJS(jsPixels, front.Pix)
		imageData.Get("data").Call("set", jsPixels)
		ctx.Call("putImageData", imageData, 0, 0)
	})

	var pressed uint16

	keydownHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		code := args[0].Get("code").String()
		if mask, ok := keyMap[code]; ok {
			args[0].Call("preventDefault")
			pressed |= mask
			gba.SetIORegister(e.Memory, gba.KEYINPUT, 0x03FF&^pressed)
		}
		return nil
	})

	keyupHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		code := args[0].Get("code").String()
		if mask, ok := keyMap[code]; ok {
			pressed &^= mask
			gba.SetIORegister(e.Memory, gba.KEYINPUT, 0x03FF&^pressed)
		}
		return nil
	})

	doc.Call("addEventListener", "keydown", keydownHandler)
	doc.Call("addEventListener", "keyup", keyupHandler)

	go e.Boot()

	select {}
}

// WASM requires a gamepak to be built
// Please source a game and place at ./gamepak.gba
//
//go:embed gamepak.gba
var game []byte
