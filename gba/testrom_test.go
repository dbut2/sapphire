package gba

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

var testROMs = []struct {
	name   string
	frames int
}{
	{"arm", 300},
	{"thumb", 300},
	{"memory", 300},
	{"shades", 10},
	{"stripes", 10},
	{"hello", 10},
	{"irq_delay", 60},
	{"timer_change", 60},
	{"line_timing", 60},
}

func runROM(t *testing.T, path string, frames int) *image.RGBA {
	t.Helper()
	gamepak, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("rom not available: %v", err)
	}
	emu := NewEmu(gamepak)
	emu.LCD.ShowFPS = false
	emu.PreBoot()
	for i := 0; i < frames; i++ {
		emu.frame()
	}
	return emu.LCD.Front()
}

func TestROMGoldens(t *testing.T) {
	for _, rom := range testROMs {
		t.Run(rom.name, func(t *testing.T) {
			img := runROM(t, filepath.Join("testdata", "roms", rom.name+".gba"), rom.frames)
			checkGolden(t, rom.name, img)
		})
	}
}

func TestSapphireGolden(t *testing.T) {
	img := runROM(t, "../sapphire.gba", 600)
	checkGolden(t, "sapphire", img)
}

func checkGolden(t *testing.T, name string, img *image.RGBA) {
	t.Helper()
	goldenPath := filepath.Join("testdata", "golden", name+".png")

	if os.Getenv("UPDATE_GOLDEN") != "" {
		_ = os.MkdirAll(filepath.Dir(goldenPath), 0755)
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}

	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("no golden (run with UPDATE_GOLDEN=1): %v", err)
	}
	defer f.Close()
	want, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	wantRGBA, ok := want.(*image.RGBA)
	if !ok {
		wantRGBA = image.NewRGBA(want.Bounds())
		for y := want.Bounds().Min.Y; y < want.Bounds().Max.Y; y++ {
			for x := want.Bounds().Min.X; x < want.Bounds().Max.X; x++ {
				wantRGBA.Set(x, y, want.At(x, y))
			}
		}
	}
	if !bytes.Equal(wantRGBA.Pix, img.Pix) {
		diffPath := filepath.Join(t.TempDir(), name+"_got.png")
		f, _ := os.Create(diffPath)
		_ = png.Encode(f, img)
		f.Close()
		t.Fatalf("framebuffer differs from golden %s (got: %s)", goldenPath, diffPath)
	}
}
