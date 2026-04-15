//go:build js && wasm

package main

import (
	"math"
	"syscall/js"

	"dbut.dev/sapphire/gba"
)

const workletSource = `
class SapphireProcessor extends AudioWorkletProcessor {
  constructor(opts) {
    super();
    this.srcRate = (opts.processorOptions && opts.processorOptions.sampleRate) || 32768;
    this.cap = 32768;
    this.ring = new Float32Array(this.cap);
    this.head = 0; this.tail = 0; this.size = 0;
    this.readPos = 0; this.curL = 0; this.curR = 0;
    this.port.onmessage = (ev) => {
      const buf = ev.data;
      const n = buf.length;
      for (let i = 0; i < n; i++) {
        if (this.size >= this.cap) { this.head = (this.head + 1) % this.cap; this.size--; }
        this.ring[this.tail] = buf[i];
        this.tail = (this.tail + 1) % this.cap;
        this.size++;
      }
    };
  }
  read() {
    if (this.size <= 0) return 0;
    const v = this.ring[this.head];
    this.head = (this.head + 1) % this.cap;
    this.size--;
    return v;
  }
  process(_, outs) {
    const out = outs[0];
    const L = out[0], R = out[1];
    const frames = L.length;
    const step = this.srcRate / sampleRate;
    for (let i = 0; i < frames; i++) {
      const adv = Math.floor(this.readPos + step) - Math.floor(this.readPos);
      for (let j = 0; j < adv; j++) { this.curL = this.read(); this.curR = this.read(); }
      this.readPos += step;
      L[i] = this.curL; R[i] = this.curR;
    }
    if (this.readPos > 1e9) this.readPos = 0;
    return true;
  }
}
registerProcessor('sapphire', SapphireProcessor);
`

func setupAudio(e *gba.Emulator) {
	global := js.Global()
	AudioCtx := global.Get("AudioContext")
	if AudioCtx.IsUndefined() {
		AudioCtx = global.Get("webkitAudioContext")
	}
	if AudioCtx.IsUndefined() {
		return
	}

	sampleRate := e.APU.SampleRate()
	var ctx, node js.Value
	ready := false
	pending := [][]byte{}

	var jsU8, jsF32 js.Value
	bytes := make([]byte, 0, 8192)

	send := func(samples []int16) {
		n := len(samples)
		need := n * 4
		bytes = bytes[:need]
		for i, s := range samples {
			bits := math.Float32bits(float32(s) / 32768)
			bytes[i*4+0] = byte(bits)
			bytes[i*4+1] = byte(bits >> 8)
			bytes[i*4+2] = byte(bits >> 16)
			bytes[i*4+3] = byte(bits >> 24)
		}
		if jsU8.IsUndefined() || jsU8.Get("length").Int() < need {
			jsU8 = global.Get("Uint8Array").New(need)
			jsF32 = global.Get("Float32Array").New(jsU8.Get("buffer"))
		}
		js.CopyBytesToJS(jsU8, bytes)
		node.Get("port").Call("postMessage", jsF32.Call("subarray", 0, n))
	}

	start := func() {
		if !ctx.IsUndefined() {
			if ctx.Get("state").String() == "suspended" {
				ctx.Call("resume")
			}
			return
		}
		opts := global.Get("Object").New()
		opts.Set("sampleRate", sampleRate)
		ctx = AudioCtx.New(opts)
		blobParts := global.Get("Array").New(1)
		blobParts.SetIndex(0, workletSource)
		blobOpts := global.Get("Object").New()
		blobOpts.Set("type", "application/javascript")
		blob := global.Get("Blob").New(blobParts, blobOpts)
		url := global.Get("URL").Call("createObjectURL", blob)
		procOpts := global.Get("Object").New()
		procOpts.Set("sampleRate", sampleRate)
		nodeOpts := global.Get("Object").New()
		nodeOpts.Set("numberOfInputs", 0)
		nodeOpts.Set("numberOfOutputs", 1)
		outCh := global.Get("Array").New(1)
		outCh.SetIndex(0, 2)
		nodeOpts.Set("outputChannelCount", outCh)
		nodeOpts.Set("processorOptions", procOpts)

		var onReady js.Func
		onReady = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			node = global.Get("AudioWorkletNode").New(ctx, "sapphire", nodeOpts)
			node.Call("connect", ctx.Get("destination"))
			ready = true
			for _, p := range pending {
				bytes = p
				need := len(p)
				if jsU8.IsUndefined() || jsU8.Get("length").Int() < need {
					jsU8 = global.Get("Uint8Array").New(need)
					jsF32 = global.Get("Float32Array").New(jsU8.Get("buffer"))
				}
				js.CopyBytesToJS(jsU8, p)
				node.Get("port").Call("postMessage", jsF32.Call("subarray", 0, need/4))
			}
			pending = nil
			onReady.Release()
			return nil
		})
		ctx.Get("audioWorklet").Call("addModule", url).Call("then", onReady)
	}

	doc := global.Get("document")
	var gesture js.Func
	gesture = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		start()
		doc.Call("removeEventListener", "pointerdown", gesture)
		doc.Call("removeEventListener", "keydown", gesture)
		gesture.Release()
		return nil
	})
	doc.Call("addEventListener", "pointerdown", gesture)
	doc.Call("addEventListener", "keydown", gesture)

	e.APU.SetOutput(func(samples []int16) {
		if !ready {
			n := len(samples)
			b := make([]byte, n*4)
			for i, s := range samples {
				bits := math.Float32bits(float32(s) / 32768)
				b[i*4+0] = byte(bits)
				b[i*4+1] = byte(bits >> 8)
				b[i*4+2] = byte(bits >> 16)
				b[i*4+3] = byte(bits >> 24)
			}
			pending = append(pending, b)
			if len(pending) > 128 {
				pending = pending[1:]
			}
			return
		}
		send(samples)
	})
}
