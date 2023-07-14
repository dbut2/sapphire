package gba

func FlagLogic(left, right uint32, value uint64) (N, Z, C, V bool) {
	N = uint32(value)>>31 == 1
	Z = uint32(value) == 0
	C = false
	V = false

	return N, Z, C, V
}

func FlagArithAdd(left, right uint32, value uint64) (N, Z, C, V bool) {
	N = uint32(value)>>31 == 1
	Z = uint32(value) == 0
	C = value > 0xFFFFFFFF
	V = ^(left^right)&(left^uint32(value))&0x80000000 != 0

	return N, Z, C, V
}

func FlagArithSub(left, right uint32, value uint64) (N, Z, C, V bool) {
	N = uint32(value)>>31 == 1
	Z = uint32(value) == 0
	C = value < 0x100000000
	V = (left^right)&(left^uint32(value))&0x80000000 != 0

	return N, Z, C, V
}

func FlagArithReSub(left, right uint32, value uint64) (N, Z, C, V bool) {
	return FlagArithSub(right, left, value)
}
