// generated by stringer -type Color; DO NOT EDIT

package engine

import "fmt"

const _Color_name = "NoColorBlackWhite"

var _Color_index = [...]uint8{7, 12, 17}

func (i Color) String() string {
	if i >= Color(len(_Color_index)) {
		return fmt.Sprintf("Color(%d)", i)
	}
	hi := _Color_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _Color_index[i-1]
	}
	return _Color_name[lo:hi]
}
