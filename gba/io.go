package gba

const (
	DISPCNT     = Register[uint16](0x4000000)
	DISPSTAT    = Register[uint16](0x4000004)
	VCOUNT      = Register[uint16](0x4000006)
	BG0CNT      = Register[uint16](0x4000008)
	BG1CNT      = Register[uint16](0x400000A)
	BG2CNT      = Register[uint16](0x400000C)
	BG3CNT      = Register[uint16](0x400000E)
	BG0HOFS     = Register[uint16](0x4000010)
	BG0VOFS     = Register[uint16](0x4000012)
	BG1HOFS     = Register[uint16](0x4000014)
	BG1VOFS     = Register[uint16](0x4000016)
	BG2HOFS     = Register[uint16](0x4000018)
	BG2VOFS     = Register[uint16](0x400001A)
	BG3HOFS     = Register[uint16](0x400001C)
	BG3VOFS     = Register[uint16](0x400001E)
	BG2PA       = Register[uint16](0x4000020)
	BG2PB       = Register[uint16](0x4000022)
	BG2PC       = Register[uint16](0x4000024)
	BG2PD       = Register[uint16](0x4000026)
	BG2X        = Register[uint32](0x4000028)
	BG2Y        = Register[uint32](0x400002C)
	BG3PA       = Register[uint16](0x4000030)
	BG3PB       = Register[uint16](0x4000032)
	BG3PC       = Register[uint16](0x4000034)
	BG3PD       = Register[uint16](0x4000036)
	BG3X        = Register[uint32](0x4000038)
	BG3Y        = Register[uint32](0x400003C)
	WIN0H       = Register[uint16](0x4000040)
	WIN1H       = Register[uint16](0x4000042)
	WIN0V       = Register[uint16](0x4000044)
	WIN1V       = Register[uint16](0x4000046)
	WININ       = Register[uint16](0x4000048)
	WINOUT      = Register[uint16](0x400004A)
	MOSAIC      = Register[uint16](0x400004C)
	BLDCNT      = Register[uint16](0x4000050)
	BLDALPHA    = Register[uint16](0x4000052)
	BLDY        = Register[uint16](0x4000054)
	SOUND1CNT_L = Register[uint16](0x4000060)
	SOUND1CNT_H = Register[uint16](0x4000062)
	SOUND1CNT_X = Register[uint16](0x4000064)
	SOUND2CNT_L = Register[uint16](0x4000068)
	SOUND2CNT_H = Register[uint16](0x400006C)
	SOUND3CNT_L = Register[uint16](0x4000070)
	SOUND3CNT_H = Register[uint16](0x4000072)
	SOUND3CNT_X = Register[uint16](0x4000074)
	SOUND4CNT_L = Register[uint16](0x4000078)
	SOUND4CNT_H = Register[uint16](0x400007C)
	SOUNDCNT_L  = Register[uint16](0x4000080)
	SOUNDCNT_H  = Register[uint16](0x4000082)
	SOUNDCNT_X  = Register[uint16](0x4000084)
	SOUNDBIAS   = Register[uint16](0x4000088)
	// WAVE_RAM    = Register[uint256](0x4000090)
	FIFO_A      = Register[uint32](0x40000A0)
	FIFO_B      = Register[uint32](0x40000A4)
	DMA0SAD     = Register[uint32](0x40000B0)
	DMA0DAD     = Register[uint32](0x40000B4)
	DMA0CNT_L   = Register[uint16](0x40000B8)
	DMA0CNT_H   = Register[uint16](0x40000BA)
	DMA1SAD     = Register[uint32](0x40000BC)
	DMA1DAD     = Register[uint32](0x40000C0)
	DMA1CNT_L   = Register[uint16](0x40000C4)
	DMA1CNT_H   = Register[uint16](0x40000C6)
	DMA2SAD     = Register[uint32](0x40000C8)
	DMA2DAD     = Register[uint32](0x40000CC)
	DMA2CNT_L   = Register[uint16](0x40000D0)
	DMA2CNT_H   = Register[uint16](0x40000D2)
	DMA3SAD     = Register[uint32](0x40000D4)
	DMA3DAD     = Register[uint32](0x40000D8)
	DMA3CNT_L   = Register[uint16](0x40000DC)
	DMA3CNT_H   = Register[uint16](0x40000DE)
	TM0CNT_L    = Register[uint16](0x4000100)
	TM0CNT_H    = Register[uint16](0x4000102)
	TM1CNT_L    = Register[uint16](0x4000104)
	TM1CNT_H    = Register[uint16](0x4000106)
	TM2CNT_L    = Register[uint16](0x4000108)
	TM2CNT_H    = Register[uint16](0x400010A)
	TM3CNT_L    = Register[uint16](0x400010C)
	TM3CNT_H    = Register[uint16](0x400010E)
	SIODATA32   = Register[uint32](0x4000120)
	SIOMULTI0   = Register[uint16](0x4000120)
	SIOMULTI1   = Register[uint16](0x4000122)
	SIOMULTI2   = Register[uint16](0x4000124)
	SIOMULTI3   = Register[uint16](0x4000126)
	SIOCNT      = Register[uint16](0x4000128)
	SIOMLT_SEND = Register[uint16](0x400012A)
	SIODATA8    = Register[uint16](0x400012A)
	KEYINPUT    = Register[uint16](0x4000130)
	KEYCNT      = Register[uint16](0x4000132)
	RCNT        = Register[uint16](0x4000134)
	IR          = Register[uint16](0x4000136)
	JOYCNT      = Register[uint16](0x4000140)
	JOY_RECV    = Register[uint32](0x4000150)
	JOY_TRANS   = Register[uint32](0x4000154)
	JOYSTAT     = Register[uint16](0x4000158)
	IE          = Register[uint16](0x4000200)
	IF          = Register[uint16](0x4000202)
	WAITCNT     = Register[uint16](0x4000204)
	IME         = Register[uint16](0x4000208)
	POSTFLG     = Register[uint8](0x4000300)
	HALTCNT     = Register[uint8](0x4000301)
)