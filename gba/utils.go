package gba

import (
	"fmt"
	"math/bits"
)

func ReadBits[S Size](v S, bit uint8, size uint8) S {
	return (v >> bit) & (1<<size - 1)
}

type Size interface {
	uint8 | uint16 | uint32 | uint64
}

func SetBits[S Size](v S, bit uint8, size uint8, value S) S {
	mask := S(1<<size-1) << bit
	v &= ^mask
	v |= value << bit
	return v
}

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

const (
	LSL uint32 = iota
	LSR
	ASR
	ROR
)

func Shift(shiftType uint32, value, amount uint32) (uint32, bool) {
	switch shiftType {
	case LSL:
		return ShiftLSL(value, amount)
	case LSR:
		return ShiftLSR(value, amount)
	case ASR:
		return ShiftASR(value, amount)
	case ROR:
		return ShiftROR(value, amount)
	default:
		panic(fmt.Sprintf("bad shift: %d", shiftType))
	}
}

func ShiftLSL(value, amount uint32) (uint32, bool) {
	return value << amount, value&(1<<(32-amount)) > 0
}

func ShiftLSR(value, amount uint32) (uint32, bool) {
	return value >> amount, value&(1<<(amount-1)) > 0
}

func ShiftASR(value, amount uint32) (uint32, bool) {
	s := value & (1 << 31)
	for i := uint32(0); i < amount; i++ {
		value = (value >> 1) | s
	}
	return value, value&(1<<(amount-1)) > 0
}

func ShiftROR(value, amount uint32) (uint32, bool) {
	return value>>(amount%32) | value<<(32-(amount%32)), (value>>(amount-1))&1 > 0
}

func addInt(a uint32, b int32) uint32 {
	if b < 0 {
		return a - uint32(-b)
	}
	return a + uint32(b)
}

func signify(value uint32, size uint32) int32 {
	shiftValue := 32 - size
	return int32(value<<shiftValue) >> shiftValue
}

func setBitCount(value uint32) uint32 {
	return uint32(bits.OnesCount32(value))
}

const (
	_ uint32 = 1 << (10 * iota)
	k
	m
)

func isEqual[S Size](a, b S) S {
	if a^b == 0 {
		return 1
	}
	return 0
}
