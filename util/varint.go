package tris

import (
	"errors"
	"io"
)

/*
these two functions are copies from the std lib with the addition to return
the nr of bytes the returned (u)int64 occupied.
*/

var overflow = errors.New("tris.reply: varint overflows a 64-bit integer")

func ReadUvarint(r io.ByteReader) (uint64, int, error) {
	var x uint64
	var s uint
	for i := 1; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, 0, err
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return x, i, overflow
			}
			return x | uint64(b)<<s, i, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}

func ReadVarint(r io.ByteReader) (int64, int, error) {
	ux, i, err := ReadUvarint(r) // ok to continue in presence of error
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, i, err
}
