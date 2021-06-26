package util

import (
	"encoding/binary"
	"io"
)

func DecodeOffsetEncoding(r io.Reader) (uint64, int, error) {
	buff := make([]uint8, 0)
	for {
		var u8 uint8
		if err := binary.Read(r, binary.BigEndian, &u8); err != nil {
			return 0, 0, err
		}
		buff = append(buff, u8)
		if u8 < 0x80 {
			break
		}
	}
	decode := uint64(0)
	for _, base := range buff {
		base += 1
		decode = decode<<7 | uint64(base&0x7f)
	}
	decode -= 1
	return decode, len(buff), nil
}

func DecodeSizeEncoding(r io.Reader) (uint64, int, error) {
	buff := make([]uint8, 0)
	for {
		var u8 uint8
		if err := binary.Read(r, binary.BigEndian, &u8); err != nil {
			return 0, 0, err
		}
		buff = append(buff, u8)
		if u8 < 0x80 {
			break
		}
	}
	decode := uint64(0)
	for i, base := range buff {
		decode |= uint64(base&0x7f) << (i * 7)
	}
	return decode, len(buff), nil
}
