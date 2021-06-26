package pack

import (
	"encoding/binary"
	"io"
)

type Header struct {
	Signature []byte // 4byte
	Version   uint32
	Entries   uint32
}

func ReadHeader(r io.Reader) (*Header, error) {
	signature := make([]byte, 4)
	if _, err := r.Read(signature); err != nil {
		return nil, err
	}
	if string(signature) != "PACK" {
		return nil, ErrInvalidPackFile
	}

	var version uint32
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return nil, err
	}
	var entries uint32
	if err := binary.Read(r, binary.BigEndian, &entries); err != nil {
		return nil, err
	}
	return &Header{
		Signature: signature,
		Version:   version,
		Entries:   entries,
	}, nil
}
