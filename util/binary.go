package util

import "io"

// ReadNullTerminatedString は io.Reader からヌル終端文字列を読み込んで返す
func ReadNullTerminatedString(r io.Reader)(string,error){
	str := make([]byte,0)
	for {
		c := make([]byte,1)
		_,err := r.Read(c)
		if err == io.EOF{
			break
		}
		if err != nil{
			return string(str),err
		}
		if c[0] == 0 {
			break
		}
		str = append(str,c[0])
	}

	return string(str), nil
}
