package gba

import (
	"image"
)

const (
	layerBG0      = 0
	layerBG1      = 1
	layerBG2      = 2
	layerBG3      = 3
	layerOBJ      = 4
	layerBackdrop = 5
)

type LCD struct {
	*Motherboard
	img  *image.RGBA
	draw func()

	topColor [240]uint16
	topLayer [240]uint8
	botColor [240]uint16
	botLayer [240]uint8
	objSemi  [240]bool

	winMask   [240]uint8
	useWindow bool

	bg2Xi, bg2Yi int32
	bg3Xi, bg3Yi int32
}

func NewLCD(m *Motherboard) *LCD {
	return &LCD{
		Motherboard: m,
	}
}

func (l *LCD) SetImage(img *image.RGBA) {
	l.img = img
}

func (l *LCD) SetDraw(draw func()) {
	l.draw = draw
}

func (l *LCD) DrawFrame() {
	l.draw()
}

func (l *LCD) LatchAffineRefs() {
	l.bg2Xi = int32(ReadIORegister(l.Memory, BG2X)<<4) >> 4
	l.bg2Yi = int32(ReadIORegister(l.Memory, BG2Y)<<4) >> 4
	l.bg3Xi = int32(ReadIORegister(l.Memory, BG3X)<<4) >> 4
	l.bg3Yi = int32(ReadIORegister(l.Memory, BG3Y)<<4) >> 4
}

func (l *LCD) IncrementAffineRefs() {
	l.bg2Xi += int32(int16(ReadIORegister(l.Memory, BG2PB)))
	l.bg2Yi += int32(int16(ReadIORegister(l.Memory, BG2PD)))
	l.bg3Xi += int32(int16(ReadIORegister(l.Memory, BG3PB)))
	l.bg3Yi += int32(int16(ReadIORegister(l.Memory, BG3PD)))
}

func (l *LCD) OnAffineRefWrite(addr uint32) {
	switch {
	case addr >= uint32(BG2X) && addr < uint32(BG2X)+4:
		l.bg2Xi = int32(ReadIORegister(l.Memory, BG2X)<<4) >> 4
	case addr >= uint32(BG2Y) && addr < uint32(BG2Y)+4:
		l.bg2Yi = int32(ReadIORegister(l.Memory, BG2Y)<<4) >> 4
	case addr >= uint32(BG3X) && addr < uint32(BG3X)+4:
		l.bg3Xi = int32(ReadIORegister(l.Memory, BG3X)<<4) >> 4
	case addr >= uint32(BG3Y) && addr < uint32(BG3Y)+4:
		l.bg3Yi = int32(ReadIORegister(l.Memory, BG3Y)<<4) >> 4
	}
}

func (l *LCD) DrawLine(line uint16, blank uint16) {
	if blank == 1 {
		l.Blank(line)
		return
	}

	switch ReadFlag(l.Memory, BGMODE) {
	case 0:
		l.BGMode0Write(line)
	case 1:
		l.BGMode1Write(line)
	case 2:
		l.BGMode2Write(line)
	case 3:
		l.BGMode3Write(line)
	case 4:
		l.BGMode4Write(line)
	case 5:
		l.BGMode5Write(line)
	}
}

func (l *LCD) Blank(line uint16) {
	lo := uint32(line) * 240
	for i := lo; i < lo+240; i++ {
		l.img.Pix[i*4+0] = 255
		l.img.Pix[i*4+1] = 255
		l.img.Pix[i*4+2] = 255
		l.img.Pix[i*4+3] = 255
	}
}

func (l *LCD) setPixel(x, y uint32, r, g, b uint32) {
	if x >= 240 || y >= 160 {
		return
	}
	idx := (y*240 + x) * 4
	l.img.Pix[idx+0] = uint8(r<<3 | r>>2)
	l.img.Pix[idx+1] = uint8(g<<3 | g>>2)
	l.img.Pix[idx+2] = uint8(b<<3 | b>>2)
	l.img.Pix[idx+3] = 255
}

func (l *LCD) setPixelFromColor(x, y uint32, color uint16) {
	c := uint32(color)
	r := ReadBits(c, 0, 5)
	g := ReadBits(c, 5, 5)
	b := ReadBits(c, 10, 5)
	l.setPixel(x, y, r, g, b)
}

func (l *LCD) initLine(line uint16) {
	palette := l.Memory.ReadMemoryBlock(Palette)
	backdrop := uint16(palette[0]) | uint16(palette[1])<<8
	for x := 0; x < 240; x++ {
		l.topColor[x] = backdrop
		l.topLayer[x] = layerBackdrop
		l.botColor[x] = backdrop
		l.botLayer[x] = layerBackdrop
		l.objSemi[x] = false
	}
	l.computeWindowMask(line)
}

func (l *LCD) computeWindowMask(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.useWindow = false

	win0en := ReadBits(dispcnt, 13, 1) == 1
	win1en := ReadBits(dispcnt, 14, 1) == 1
	objWinEn := ReadBits(dispcnt, 15, 1) == 1

	if !win0en && !win1en && !objWinEn {
		return
	}
	l.useWindow = true

	winout := ReadIORegister(l.Memory, WINOUT)
	outside := uint8(winout & 0x3F)

	for x := 0; x < 240; x++ {
		l.winMask[x] = outside
	}

	if objWinEn {
		objWinMask := uint8((winout >> 8) & 0x3F)
		l.applyOBJWindow(line, objWinMask)
	}

	if win1en {
		winin := ReadIORegister(l.Memory, WININ)
		l.applyWindow(line, WIN1V, WIN1H, uint8((winin>>8)&0x3F))
	}

	if win0en {
		winin := ReadIORegister(l.Memory, WININ)
		l.applyWindow(line, WIN0V, WIN0H, uint8(winin&0x3F))
	}
}

func (l *LCD) applyOBJWindow(line uint16, mask uint8) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	linear := ReadBits(dispcnt, 6, 1) == 1

	oam := l.Memory.ReadMemoryBlock(OAM)
	vram := l.Memory.ReadMemoryBlock(VRAM)

	for i := 0; i < 128; i++ {
		base := i * 8
		attr0 := uint16(oam[base]) | uint16(oam[base+1])<<8
		attr1 := uint16(oam[base+2]) | uint16(oam[base+3])<<8
		attr2 := uint16(oam[base+4]) | uint16(oam[base+5])<<8

		scalerot := ReadBits(attr0, 8, 1) == 1
		doubleSize := ReadBits(attr0, 9, 1) == 1
		if !scalerot && doubleSize {
			continue
		}

		objMode := (attr0 >> 10) & 3
		if objMode != 2 {
			continue
		}

		shape := ReadBits(attr0, 14, 2)
		size := ReadBits(attr1, 14, 2)
		bpp8 := ReadBits(attr0, 13, 1) == 1

		width := objWidth[shape][size]
		height := objHeight[shape][size]
		if width == 0 || height == 0 {
			continue
		}

		canvasW := width
		canvasH := height
		if scalerot && doubleSize {
			canvasW *= 2
			canvasH *= 2
		}

		sprY := int(attr0 & 0xFF)
		if sprY >= 160 {
			sprY -= 256
		}
		if int(line) < sprY || int(line) >= sprY+int(canvasH) {
			continue
		}

		sprX := int(attr1 & 0x1FF)
		if sprX >= 240 {
			sprX -= 512
		}
		if sprX >= 240 || sprX+int(canvasW) <= 0 {
			continue
		}

		tileNum := uint32(attr2 & 0x3FF)

		if scalerot {
			matrixID := uint32(ReadBits(attr1, 9, 5))
			matBase := matrixID * 32
			pa := int32(int16(uint16(oam[matBase+6]) | uint16(oam[matBase+7])<<8))
			pb := int32(int16(uint16(oam[matBase+14]) | uint16(oam[matBase+15])<<8))
			pc := int32(int16(uint16(oam[matBase+22]) | uint16(oam[matBase+23])<<8))
			pd := int32(int16(uint16(oam[matBase+30]) | uint16(oam[matBase+31])<<8))

			cx := int32(canvasW / 2)
			cy := int32(canvasH / 2)
			hx := int32(width / 2)
			hy := int32(height / 2)
			iy := int32(int(line) - sprY)

			for ix := int32(0); ix < int32(canvasW); ix++ {
				screenX := int(ix) + sprX
				if screenX < 0 || screenX >= 240 {
					continue
				}
				dx := ix - cx
				dy := iy - cy
				srcX := (pa*dx+pb*dy)>>8 + hx
				srcY := (pc*dx+pd*dy)>>8 + hy
				if srcX < 0 || srcX >= int32(width) || srcY < 0 || srcY >= int32(height) {
					continue
				}
				colorIdx := objPixel(vram, tileNum, uint32(srcX), uint32(srcY), width, bpp8, linear)
				if colorIdx != 0 {
					l.winMask[screenX] = mask
				}
			}
		} else {
			hFlip := ReadBits(attr1, 12, 1) == 1
			vFlip := ReadBits(attr1, 13, 1) == 1

			localY := uint32(int(line) - sprY)
			if vFlip {
				localY = height - 1 - localY
			}

			for px := uint32(0); px < width; px++ {
				screenX := int(px) + sprX
				if screenX < 0 || screenX >= 240 {
					continue
				}
				localX := px
				if hFlip {
					localX = width - 1 - localX
				}
				colorIdx := objPixel(vram, tileNum, localX, localY, width, bpp8, linear)
				if colorIdx != 0 {
					l.winMask[screenX] = mask
				}
			}
		}
	}
}

func (l *LCD) applyWindow(line uint16, winV, winH IORegister[uint16], mask uint8) {
	v := ReadIORegister(l.Memory, winV)
	top := uint16(v >> 8)
	bottom := uint16(v & 0xFF)

	var inVRange bool
	if top <= bottom {
		inVRange = line >= top && line < bottom
	} else {
		inVRange = line >= top || line < bottom
	}
	if !inVRange {
		return
	}

	h := ReadIORegister(l.Memory, winH)
	left := uint16(h >> 8)
	right := uint16(h & 0xFF)

	if left <= right {
		for x := left; x < right && x < 240; x++ {
			l.winMask[x] = mask
		}
	} else {
		for x := uint16(0); x < right && x < 240; x++ {
			l.winMask[x] = mask
		}
		for x := left; x < 240; x++ {
			l.winMask[x] = mask
		}
	}
}

func (l *LCD) drawBufPixel(x uint32, color uint16, layer uint8) {
	if l.useWindow && l.winMask[x]&(1<<layer) == 0 {
		return
	}
	l.botColor[x] = l.topColor[x]
	l.botLayer[x] = l.topLayer[x]
	l.topColor[x] = color
	l.topLayer[x] = layer
}

func (l *LCD) compositeLine(line uint16) {
	bldcnt := ReadIORegister(l.Memory, BLDCNT)
	mode := (bldcnt >> 6) & 3
	above := bldcnt & 0x3F
	below := (bldcnt >> 8) & 0x3F

	bldalpha := ReadIORegister(l.Memory, BLDALPHA)
	eva := uint32(bldalpha & 0x1F)
	evb := uint32((bldalpha >> 8) & 0x1F)
	if eva > 16 {
		eva = 16
	}
	if evb > 16 {
		evb = 16
	}

	bldy := ReadIORegister(l.Memory, BLDY)
	evy := uint32(bldy & 0x1F)
	if evy > 16 {
		evy = 16
	}

	for x := uint32(0); x < 240; x++ {
		top := l.topColor[x]
		topL := l.topLayer[x]
		bot := l.botColor[x]
		botL := l.botLayer[x]

		finalColor := top
		sfxOk := !l.useWindow || l.winMask[x]&(1<<5) != 0

		if l.objSemi[x] && topL == layerOBJ {
			if sfxOk && (below>>botL)&1 == 1 {
				finalColor = blendAlpha(top, bot, eva, evb)
			}
		} else if sfxOk {
			switch mode {
			case 1:
				if (above>>topL)&1 == 1 && (below>>botL)&1 == 1 {
					finalColor = blendAlpha(top, bot, eva, evb)
				}
			case 2:
				if (above>>topL)&1 == 1 {
					finalColor = blendWhite(top, evy)
				}
			case 3:
				if (above>>topL)&1 == 1 {
					finalColor = blendBlack(top, evy)
				}
			}
		}

		l.setPixelFromColor(x, uint32(line), finalColor)
	}
}

func blendAlpha(above, below uint16, eva, evb uint32) uint16 {
	ar := uint32(above) & 0x1F
	ag := (uint32(above) >> 5) & 0x1F
	ab := (uint32(above) >> 10) & 0x1F
	br := uint32(below) & 0x1F
	bg := (uint32(below) >> 5) & 0x1F
	bb := (uint32(below) >> 10) & 0x1F
	r := min((ar*eva+br*evb)>>4, 31)
	g := min((ag*eva+bg*evb)>>4, 31)
	b := min((ab*eva+bb*evb)>>4, 31)
	return uint16(r | (g << 5) | (b << 10))
}

func blendWhite(color uint16, evy uint32) uint16 {
	r := uint32(color) & 0x1F
	g := (uint32(color) >> 5) & 0x1F
	b := (uint32(color) >> 10) & 0x1F
	r += (31 - r) * evy / 16
	g += (31 - g) * evy / 16
	b += (31 - b) * evy / 16
	return uint16(r | (g << 5) | (b << 10))
}

func blendBlack(color uint16, evy uint32) uint16 {
	r := uint32(color) & 0x1F
	g := (uint32(color) >> 5) & 0x1F
	b := (uint32(color) >> 10) & 0x1F
	r -= r * evy / 16
	g -= g * evy / 16
	b -= b * evy / 16
	return uint16(r | (g << 5) | (b << 10))
}

func (l *LCD) drawTextBGLine(line uint16, bgcnt IORegister[uint16], hofs IORegister[uint16], vofs IORegister[uint16], layer uint8) {
	cnt := ReadIORegister(l.Memory, bgcnt)

	charBase := uint32(ReadBits(cnt, 2, 2)) * 0x4000
	colorMode := ReadBits(cnt, 7, 1)
	screenBase := uint32(ReadBits(cnt, 8, 5)) * 0x0800
	screenSize := ReadBits(cnt, 14, 2)

	scrollX := uint32(ReadIORegister(l.Memory, hofs)) & 0x1FF
	scrollY := uint32(ReadIORegister(l.Memory, vofs)) & 0x1FF

	var bgWidthTiles, bgHeightTiles uint32
	switch screenSize {
	case 0:
		bgWidthTiles, bgHeightTiles = 32, 32
	case 1:
		bgWidthTiles, bgHeightTiles = 64, 32
	case 2:
		bgWidthTiles, bgHeightTiles = 32, 64
	case 3:
		bgWidthTiles, bgHeightTiles = 64, 64
	}

	bgWidthPx := bgWidthTiles * 8
	bgHeightPx := bgHeightTiles * 8

	y := (uint32(line) + scrollY) % bgHeightPx
	tileY := y / 8
	fineY := y % 8

	vram := l.Memory.ReadMemoryBlock(VRAM)
	palette := l.Memory.ReadMemoryBlock(Palette)

	for screenX := uint32(0); screenX < 240; screenX++ {
		x := (screenX + scrollX) % bgWidthPx
		tileX := x / 8
		fineX := x % 8

		screenBlockOffset := uint32(0)
		localTileX := tileX
		localTileY := tileY

		if bgWidthTiles == 64 && tileX >= 32 {
			screenBlockOffset += 0x0800
			localTileX -= 32
		}
		if bgHeightTiles == 64 && tileY >= 32 {
			if bgWidthTiles == 64 {
				screenBlockOffset += 0x1000
			} else {
				screenBlockOffset += 0x0800
			}
			localTileY -= 32
		}

		mapOffset := screenBase + screenBlockOffset + (localTileY*32+localTileX)*2
		if mapOffset+1 >= uint32(len(vram)) {
			continue
		}
		tileEntry := uint16(vram[mapOffset]) | uint16(vram[mapOffset+1])<<8

		tileNum := uint32(tileEntry & 0x3FF)
		hFlip := (tileEntry >> 10) & 1
		vFlip := (tileEntry >> 11) & 1
		palNum := uint32((tileEntry >> 12) & 0xF)

		pixelY := fineY
		pixelX := fineX

		if hFlip == 1 {
			pixelX = 7 - pixelX
		}
		if vFlip == 1 {
			pixelY = 7 - pixelY
		}

		var colorIdx uint32
		if colorMode == 0 {

			tileAddr := charBase + tileNum*32 + pixelY*4 + pixelX/2
			if tileAddr >= uint32(len(vram)) {
				continue
			}
			b := uint32(vram[tileAddr])
			if pixelX%2 == 0 {
				colorIdx = b & 0xF
			} else {
				colorIdx = (b >> 4) & 0xF
			}
			if colorIdx == 0 {
				continue
			}
			colorIdx = palNum*16 + colorIdx
		} else {

			tileAddr := charBase + tileNum*64 + pixelY*8 + pixelX
			if tileAddr >= uint32(len(vram)) {
				continue
			}
			colorIdx = uint32(vram[tileAddr])
			if colorIdx == 0 {
				continue
			}
		}

		palAddr := colorIdx * 2
		color := uint16(palette[palAddr]) | uint16(palette[palAddr+1])<<8
		l.drawBufPixel(screenX, color, layer)
	}
}

func (l *LCD) BGMode0Write(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.initLine(line)

	bgCnts := [4]IORegister[uint16]{BG0CNT, BG1CNT, BG2CNT, BG3CNT}
	bgHofs := [4]IORegister[uint16]{BG0HOFS, BG1HOFS, BG2HOFS, BG3HOFS}
	bgVofs := [4]IORegister[uint16]{BG0VOFS, BG1VOFS, BG2VOFS, BG3VOFS}
	bgLayers := [4]uint8{layerBG0, layerBG1, layerBG2, layerBG3}

	type bgInfo struct {
		bgNum    int
		priority uint16
	}

	var bgs []bgInfo
	for i := 0; i < 4; i++ {
		if ReadBits(dispcnt, uint8(8+i), 1) == 1 {
			cnt := ReadIORegister(l.Memory, bgCnts[i])
			pri := ReadBits(cnt, 0, 2)
			bgs = append(bgs, bgInfo{i, pri})
		}
	}

	objEnabled := ReadBits(dispcnt, 12, 1) == 1

	for pri := uint16(3); pri <= 3; pri-- {
		for bgIdx := 3; bgIdx >= 0; bgIdx-- {
			for _, bg := range bgs {
				if bg.bgNum == bgIdx && bg.priority == pri {
					l.drawTextBGLine(line, bgCnts[bg.bgNum], bgHofs[bg.bgNum], bgVofs[bg.bgNum], bgLayers[bg.bgNum])
				}
			}
		}
		if objEnabled {
			l.drawOBJLine(line, pri)
		}
	}

	l.compositeLine(line)
}

func (l *LCD) BGMode1Write(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.initLine(line)

	bgCnts := [2]IORegister[uint16]{BG0CNT, BG1CNT}
	bgHofs := [2]IORegister[uint16]{BG0HOFS, BG1HOFS}
	bgVofs := [2]IORegister[uint16]{BG0VOFS, BG1VOFS}
	bgLayers := [2]uint8{layerBG0, layerBG1}

	type bgInfo struct {
		idx      int
		priority uint16
	}
	var textBGs []bgInfo
	for i := 0; i < 2; i++ {
		if ReadBits(dispcnt, uint8(8+i), 1) == 1 {
			cnt := ReadIORegister(l.Memory, bgCnts[i])
			pri := ReadBits(cnt, 0, 2)
			textBGs = append(textBGs, bgInfo{i, pri})
		}
	}

	bg2Enabled := ReadBits(dispcnt, 10, 1) == 1
	var bg2Pri uint16
	if bg2Enabled {
		bg2Pri = ReadBits(ReadIORegister(l.Memory, BG2CNT), 0, 2)
	}
	objEnabled := ReadBits(dispcnt, 12, 1) == 1

	for pri := uint16(3); pri <= 3; pri-- {
		if bg2Enabled && bg2Pri == pri {
			l.drawAffineBGLine(line, BG2CNT, BG2PA, BG2PB, BG2PC, BG2PD, BG2X, BG2Y, layerBG2)
		}
		for bgIdx := 1; bgIdx >= 0; bgIdx-- {
			for _, bg := range textBGs {
				if bg.idx == bgIdx && bg.priority == pri {
					l.drawTextBGLine(line, bgCnts[bg.idx], bgHofs[bg.idx], bgVofs[bg.idx], bgLayers[bg.idx])
				}
			}
		}
		if objEnabled {
			l.drawOBJLine(line, pri)
		}
	}

	l.compositeLine(line)
}

func (l *LCD) BGMode2Write(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.initLine(line)

	bg2Enabled := ReadBits(dispcnt, 10, 1) == 1
	bg3Enabled := ReadBits(dispcnt, 11, 1) == 1
	objEnabled := ReadBits(dispcnt, 12, 1) == 1

	var bg2Pri, bg3Pri uint16
	if bg2Enabled {
		bg2Pri = ReadBits(ReadIORegister(l.Memory, BG2CNT), 0, 2)
	}
	if bg3Enabled {
		bg3Pri = ReadBits(ReadIORegister(l.Memory, BG3CNT), 0, 2)
	}

	for pri := uint16(3); pri <= 3; pri-- {
		if bg3Enabled && bg3Pri == pri {
			l.drawAffineBGLine(line, BG3CNT, BG3PA, BG3PB, BG3PC, BG3PD, BG3X, BG3Y, layerBG3)
		}
		if bg2Enabled && bg2Pri == pri {
			l.drawAffineBGLine(line, BG2CNT, BG2PA, BG2PB, BG2PC, BG2PD, BG2X, BG2Y, layerBG2)
		}
		if objEnabled {
			l.drawOBJLine(line, pri)
		}
	}

	l.compositeLine(line)
}

func (l *LCD) BGMode3Write(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.initLine(line)

	if ReadBits(dispcnt, 10, 1) == 1 {
		for i := uint32(0); i < 240; i++ {
			pixel := uint32(line)*240 + i
			color := l.Memory.Read16(VRAM.Start+pixel*2, false, false)
			l.drawBufPixel(i, color, layerBG2)
		}
	}

	if ReadBits(dispcnt, 12, 1) == 1 {
		for pri := uint16(3); pri <= 3; pri-- {
			l.drawOBJLine(line, pri)
		}
	}

	l.compositeLine(line)
}

func (l *LCD) BGMode4Write(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.initLine(line)

	if ReadBits(dispcnt, 10, 1) == 1 {
		frame := [2]uint32{0, 0xA000}[ReadFlag(l.Memory, BGFRAME)]
		vram := l.Memory.ReadMemoryBlock(VRAM)
		palette := l.Memory.ReadMemoryBlock(Palette)
		for i := uint32(0); i < 240; i++ {
			offset := uint32(line)*240 + i
			colorIdx := vram[frame+offset]
			palAddr := uint32(colorIdx) * 2
			color := uint16(palette[palAddr]) | uint16(palette[palAddr+1])<<8
			l.drawBufPixel(i, color, layerBG2)
		}
	}

	if ReadBits(dispcnt, 12, 1) == 1 {
		for pri := uint16(3); pri <= 3; pri-- {
			l.drawOBJLine(line, pri)
		}
	}

	l.compositeLine(line)
}

func (l *LCD) BGMode5Write(line uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	l.initLine(line)

	if ReadBits(dispcnt, 10, 1) == 1 {
		frame := [2]uint32{0x06000000, 0x0600A000}[ReadFlag(l.Memory, BGFRAME)]
		if uint32(line) < 128 {
			for i := uint32(0); i < 160; i++ {
				pixel := uint32(line)*160 + i
				color := l.Memory.Read16(frame+pixel*2, false, false)
				screenX := i + 40
				l.drawBufPixel(screenX, color, layerBG2)
			}
		}
	}

	if ReadBits(dispcnt, 12, 1) == 1 {
		for pri := uint16(3); pri <= 3; pri-- {
			l.drawOBJLine(line, pri)
		}
	}

	l.compositeLine(line)
}

func (l *LCD) drawAffineBGLine(line uint16, bgcnt IORegister[uint16],
	pa IORegister[uint16], pb IORegister[uint16], pc IORegister[uint16], pd IORegister[uint16],
	refX IORegister[uint32], refY IORegister[uint32], layer uint8) {

	cnt := ReadIORegister(l.Memory, bgcnt)
	charBase := uint32(ReadBits(cnt, 2, 2)) * 0x4000
	screenBase := uint32(ReadBits(cnt, 8, 5)) * 0x0800
	wrap := ReadBits(cnt, 13, 1) == 1
	sizeID := ReadBits(cnt, 14, 2)

	bgSize := uint32(128) << sizeID
	tilesPerSide := bgSize / 8

	dxX := int32(int16(ReadIORegister(l.Memory, pa)))
	dxY := int32(int16(ReadIORegister(l.Memory, pc)))

	var curX, curY int32
	if layer == layerBG2 {
		curX = l.bg2Xi
		curY = l.bg2Yi
	} else {
		curX = l.bg3Xi
		curY = l.bg3Yi
	}

	vram := l.Memory.ReadMemoryBlock(VRAM)
	palette := l.Memory.ReadMemoryBlock(Palette)

	for screenX := uint32(0); screenX < 240; screenX++ {

		texX := curX >> 8
		texY := curY >> 8

		curX += dxX
		curY += dxY

		if wrap {
			texX = ((texX % int32(bgSize)) + int32(bgSize)) % int32(bgSize)
			texY = ((texY % int32(bgSize)) + int32(bgSize)) % int32(bgSize)
		} else {
			if texX < 0 || texX >= int32(bgSize) || texY < 0 || texY >= int32(bgSize) {
				continue
			}
		}

		tileX := uint32(texX) / 8
		tileY := uint32(texY) / 8
		fineX := uint32(texX) % 8
		fineY := uint32(texY) % 8

		mapOffset := screenBase + tileY*tilesPerSide + tileX
		if mapOffset >= uint32(len(vram)) {
			continue
		}
		tileNum := uint32(vram[mapOffset])

		tileAddr := charBase + tileNum*64 + fineY*8 + fineX
		if tileAddr >= uint32(len(vram)) {
			continue
		}
		colorIdx := uint32(vram[tileAddr])
		if colorIdx == 0 {
			continue
		}

		palAddr := colorIdx * 2
		color := uint16(palette[palAddr]) | uint16(palette[palAddr+1])<<8
		l.drawBufPixel(screenX, color, layer)
	}
}

var objWidth = [4][4]uint32{
	{8, 16, 32, 64},
	{16, 32, 32, 64},
	{8, 8, 16, 32},
	{0, 0, 0, 0},
}

var objHeight = [4][4]uint32{
	{8, 16, 32, 64},
	{8, 8, 16, 32},
	{16, 32, 32, 64},
	{0, 0, 0, 0},
}

func objPixel(vram []byte, tileNum, srcX, srcY, width uint32, bpp8, linear bool) uint32 {
	tileRow := srcY / 8
	fineY := srcY % 8
	tileCol := srcX / 8
	fineX := srcX % 8

	var tileID uint32
	if linear {
		if bpp8 {
			tileID = tileNum + tileRow*(width/8)*2 + tileCol*2
		} else {
			tileID = tileNum + tileRow*(width/8) + tileCol
		}
	} else {
		if bpp8 {
			tileID = tileNum + tileRow*32 + tileCol*2
		} else {
			tileID = tileNum + tileRow*32 + tileCol
		}
	}

	if bpp8 {
		tileAddr := 0x10000 + tileID*32 + fineY*8 + fineX
		if tileAddr >= uint32(len(vram)) {
			return 0
		}
		return uint32(vram[tileAddr])
	}
	tileAddr := 0x10000 + tileID*32 + fineY*4 + fineX/2
	if tileAddr >= uint32(len(vram)) {
		return 0
	}
	b := uint32(vram[tileAddr])
	if fineX%2 == 0 {
		return b & 0xF
	}
	return (b >> 4) & 0xF
}

func (l *LCD) drawOBJLine(line uint16, targetPri uint16) {
	dispcnt := ReadIORegister(l.Memory, DISPCNT)
	linear := ReadBits(dispcnt, 6, 1) == 1

	oam := l.Memory.ReadMemoryBlock(OAM)
	vram := l.Memory.ReadMemoryBlock(VRAM)
	palette := l.Memory.ReadMemoryBlock(Palette)

	for i := 127; i >= 0; i-- {
		base := i * 8
		attr0 := uint16(oam[base]) | uint16(oam[base+1])<<8
		attr1 := uint16(oam[base+2]) | uint16(oam[base+3])<<8
		attr2 := uint16(oam[base+4]) | uint16(oam[base+5])<<8

		scalerot := ReadBits(attr0, 8, 1) == 1
		doubleSize := ReadBits(attr0, 9, 1) == 1

		if !scalerot && doubleSize {
			continue
		}

		pri := ReadBits(attr2, 10, 2)
		if pri != targetPri {
			continue
		}

		objMode := (attr0 >> 10) & 3
		if objMode == 2 || objMode == 3 {
			continue
		}
		semiTrans := objMode == 1

		shape := ReadBits(attr0, 14, 2)
		size := ReadBits(attr1, 14, 2)
		bpp8 := ReadBits(attr0, 13, 1) == 1

		width := objWidth[shape][size]
		height := objHeight[shape][size]
		if width == 0 || height == 0 {
			continue
		}

		canvasW := width
		canvasH := height
		if scalerot && doubleSize {
			canvasW *= 2
			canvasH *= 2
		}

		sprY := int(attr0 & 0xFF)
		if sprY >= 160 {
			sprY -= 256
		}
		if int(line) < sprY || int(line) >= sprY+int(canvasH) {
			continue
		}

		sprX := int(attr1 & 0x1FF)
		if sprX >= 240 {
			sprX -= 512
		}
		if sprX >= 240 || sprX+int(canvasW) <= 0 {
			continue
		}

		tileNum := uint32(attr2 & 0x3FF)
		palNum := uint32((attr2 >> 12) & 0xF)

		if scalerot {

			matrixID := uint32(ReadBits(attr1, 9, 5))

			matBase := matrixID * 32
			pa := int32(int16(uint16(oam[matBase+6]) | uint16(oam[matBase+7])<<8))
			pb := int32(int16(uint16(oam[matBase+14]) | uint16(oam[matBase+15])<<8))
			pc := int32(int16(uint16(oam[matBase+22]) | uint16(oam[matBase+23])<<8))
			pd := int32(int16(uint16(oam[matBase+30]) | uint16(oam[matBase+31])<<8))

			cx := int32(canvasW / 2)
			cy := int32(canvasH / 2)

			hx := int32(width / 2)
			hy := int32(height / 2)

			iy := int32(int(line) - sprY)

			for ix := int32(0); ix < int32(canvasW); ix++ {
				screenX := int(ix) + sprX
				if screenX < 0 || screenX >= 240 {
					continue
				}

				dx := ix - cx
				dy := iy - cy

				srcX := (pa*dx+pb*dy)>>8 + hx
				srcY := (pc*dx+pd*dy)>>8 + hy

				if srcX < 0 || srcX >= int32(width) || srcY < 0 || srcY >= int32(height) {
					continue
				}

				colorIdx := objPixel(vram, tileNum, uint32(srcX), uint32(srcY), width, bpp8, linear)
				if colorIdx == 0 {
					continue
				}

				var palAddr uint32
				if bpp8 {
					palAddr = 0x200 + colorIdx*2
				} else {
					palAddr = 0x200 + (palNum*16+colorIdx)*2
				}
				color := uint16(palette[palAddr]) | uint16(palette[palAddr+1])<<8
				l.drawBufPixel(uint32(screenX), color, layerOBJ)
				l.objSemi[screenX] = semiTrans
			}
		} else {

			hFlip := ReadBits(attr1, 12, 1) == 1
			vFlip := ReadBits(attr1, 13, 1) == 1

			localY := uint32(int(line) - sprY)
			if vFlip {
				localY = height - 1 - localY
			}

			for px := uint32(0); px < width; px++ {
				screenX := int(px) + sprX
				if screenX < 0 || screenX >= 240 {
					continue
				}

				localX := px
				if hFlip {
					localX = width - 1 - localX
				}

				colorIdx := objPixel(vram, tileNum, localX, localY, width, bpp8, linear)
				if colorIdx == 0 {
					continue
				}

				var palAddr uint32
				if bpp8 {
					palAddr = 0x200 + colorIdx*2
				} else {
					palAddr = 0x200 + (palNum*16+colorIdx)*2
				}
				color := uint16(palette[palAddr]) | uint16(palette[palAddr+1])<<8
				l.drawBufPixel(uint32(screenX), color, layerOBJ)
				l.objSemi[screenX] = semiTrans
			}
		}
	}
}

func (l *LCD) PaletteRGBA(n uint8) (r, g, b, a uint32) {
	c := l.Memory.Read16(Palette.Start+uint32(n)*2, false, false)
	return l.RGBA(c)
}

func (l *LCD) RGBA(d uint16) (r, g, b, a uint32) {
	c := uint32(d)

	r = ReadBits(c, 0, 5)
	g = ReadBits(c, 5, 5)
	b = ReadBits(c, 10, 5)
	a = 1

	return r, g, b, a
}

func (l *LCD) Color(r, g, b, a uint32) uint16 {
	c := (r&31)<<11 + (g&31)<<6 + (b&31)<<1 + a
	return uint16(c)
}
