package gba

import (
	"fmt"
	"testing"
)

func Test_Thumb(t *testing.T) {

	ins := uint32(0b0100001001001001)

	do := (&CPU{}).ParseThumb(ins)
	fmt.Println()
	do(ins)
}
