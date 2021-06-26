package util

import (
	"io"
)

func ReadNullTerminatedString(r io.Reader) (string, error) {
	str := make([]byte, 0)
	for {
		c := make([]byte, 1)
		_, err := r.Read(c)
		if err == io.EOF {
			break
		}
		if err != nil {
			return string(str), err
		}
		if c[0] == 0 {
			break
		}
		str = append(str, c[0])
	}
	return string(str), nil
}
