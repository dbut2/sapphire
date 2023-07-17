package gba

func AND(left, right, carry uint32) (value uint64) { // value = left AND right
	return uint64(left & right)
}

func EOR(left, right, carry uint32) (value uint64) { // value = left XOR right
	return uint64(left ^ right)
}

func SUB(left, right, carry uint32) (value uint64) { // value = left - right
	return uint64(left) - uint64(right)
}

func RSB(left, right, carry uint32) (value uint64) { // value = right - left
	return uint64(right) - uint64(left)
}

func ADD(left, right, carry uint32) (value uint64) { // value = left + right
	return uint64(left) + uint64(right)
}

func ADC(left, right, carry uint32) (value uint64) { // value = left + right + carry
	return uint64(left) + uint64(right) + uint64(carry)
}

func SBCArm(left, right, carry uint32) (value uint64) { // value = left - right + carry - 1
	return uint64(left) - uint64(right) + uint64(carry) - 1
}

func SBCThumb(left, right, carry uint32) (value uint64) { // value = left - right - NOT carry
	return uint64(left) - uint64(right) - uint64(^carry)
}

func RSC(left, right, carry uint32) (value uint64) { // value = right - left + carry - 1
	return uint64(right) - uint64(left) + uint64(carry) - 1
}

func TST(left, right, carry uint32) (value uint64) { // Void = left AND right
	return uint64(left & right)
}

func TEQ(left, right, carry uint32) (value uint64) { // Void = left XOR right
	return uint64(left ^ right)
}

func CMP(left, right, carry uint32) (value uint64) { // Void = left - right
	return uint64(left) - uint64(right)
}

func CMN(left, right, carry uint32) (value uint64) { // Void = left + right
	return uint64(left) + uint64(right)
}

func ORR(left, right, carry uint32) (value uint64) { // value = left OR right
	return uint64(left | right)
}

func MOV(left, right, carry uint32) (value uint64) { // value = right
	return uint64(right)
}

func BIC(left, right, carry uint32) (value uint64) { // value = left AND NOT right
	return uint64(left & ^right)
}

func MVN(left, right, carry uint32) (value uint64) { // value = NOT right
	return uint64(^right)
}

func MUL(left, right, carry uint32) (value uint64) {
	return uint64(left) * uint64(right)
}
