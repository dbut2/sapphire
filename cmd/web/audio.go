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
    this.cap = 4096;
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

	if nav := global.Get("navigator"); !nav.IsUndefined() {
		if session := nav.Get("audioSession"); !session.IsUndefined() && !session.IsNull() {
			func() {
				defer func() { _ = recover() }()
				session.Set("type", "playback")
			}()
		}
	}

	sampleRate := e.APU.SampleRate()
	var ctx, node js.Value
	ready := false

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

	resumeIfNeeded := func() {
		if ctx.IsUndefined() {
			return
		}
		if ctx.Get("state").String() != "running" {
			ctx.Call("resume")
		}
	}

	unlock := func() {
		if ctx.IsUndefined() {
			return
		}
		defer func() { _ = recover() }()
		buf := ctx.Call("createBuffer", 1, 1, 22050)
		src := ctx.Call("createBufferSource")
		src.Set("buffer", buf)
		src.Call("connect", ctx.Get("destination"))
		src.Call("start", 0)
	}

	start := func() {
		if !ctx.IsUndefined() {
			resumeIfNeeded()
			return
		}
		defer func() { _ = recover() }()
		opts := global.Get("Object").New()
		opts.Set("sampleRate", sampleRate)
		func() {
			defer func() {
				if recover() != nil {
					ctx = AudioCtx.New()
				}
			}()
			ctx = AudioCtx.New(opts)
		}()
		resumeIfNeeded()
		unlock()
		worklet := ctx.Get("audioWorklet")
		if worklet.IsUndefined() {
			return
		}
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
			onReady.Release()
			return nil
		})
		worklet.Call("addModule", url).Call("then", onReady)
	}

	doc := global.Get("document")
	events := []string{"pointerdown", "touchstart", "click", "keydown"}
	var gesture js.Func
	gesture = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		start()
		if ready && !ctx.IsUndefined() && ctx.Get("state").String() == "running" {
			for _, ev := range events {
				doc.Call("removeEventListener", ev, gesture)
			}
			gesture.Release()
		}
		return nil
	})
	for _, ev := range events {
		doc.Call("addEventListener", ev, gesture)
	}

	muted := false
	gain := js.Value{}
	toggleMute := func() bool {
		if !ready {
			return muted
		}
		if gain.IsUndefined() || gain.IsNull() {
			gain = ctx.Call("createGain")
			node.Call("disconnect")
			node.Call("connect", gain)
			gain.Call("connect", ctx.Get("destination"))
		}
		muted = !muted
		v := float32(1)
		if muted {
			v = 0
		}
		gain.Get("gain").Call("setValueAtTime", v, ctx.Get("currentTime"))
		return muted
	}
	doc.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 || args[0].Get("code").String() != "KeyM" {
			return nil
		}
		toggleMute()
		return nil
	}))
	global.Set("sapphireToggleMute", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return toggleMute()
	}))

	e.APU.SetOutput(func(samples []int16) {
		if !ready {
			return
		}
		send(samples)
	})
}
