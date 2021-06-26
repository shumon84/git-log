package pack

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/shumon84/git-log/object"
	"github.com/shumon84/git-log/sha"
)

type Entry struct {
	Type   EntryType
	Size   uint64
	Data   []byte
	Offset int64
	Hash   sha.SHA1
}

func (e *Entry) ToObject() (*object.Object, error) {
	zr, err := zlib.NewReader(bytes.NewBuffer(e.Data))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	objectType, err := e.Type.ToObjectType()
	if err != nil {
		return nil, err
	}
	size := len(data)

	obj := &object.Object{
		Hash: e.Hash,
		Type: objectType,
		Size: size,
		Data: data,
	}

	checkSum := sha1.New()
	checkSum.Write(obj.Header())
	checkSum.Write(data)

	if string(obj.Hash) != string(checkSum.Sum(nil)) {
		return nil, object.ErrInvalidObject
	}

	return obj, nil
}

func (e *Entry) IsDeltaObject() bool {
	return e.Type == OBJ_REF_DELTA || e.Type == OBJ_OFS_DELTA
}

func readEntry(pack io.ReadSeeker, idxEntry *IdxEntry) (*Entry, error) {
	offset := int64(idxEntry.Offset)
	if idxEntry.HasLargeOffset() {
		offset = int64(idxEntry.LargeOffset)
	}

	if _, err := pack.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	header := make([]uint8, 0)
	for {
		var u8 uint8
		binary.Read(pack, binary.BigEndian, &u8)
		header = append(header, u8)
		if u8 < 0x80 {
			break
		}
	}

	entryType := EntryType((header[0] >> 4) & 0x7)

	size := uint64(header[0] & 0xf)
	for i, base := range header[1:] {
		size |= uint64(base&0x7f) << (i*7 + 4)
	}

	data := make([]byte, size)
	if _, err := pack.Read(data); err != nil {
		return nil, err
	}

	entry := &Entry{
		Type:   entryType,
		Size:   size,
		Data:   data,
		Offset: offset,
	}

	return entry, nil
}
