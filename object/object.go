package object

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/shumon84/git-log/sha"
	"github.com/shumon84/git-log/util"
)

type Object struct {
	Hash sha.SHA1
	Type Type
	Size int
	Data []byte
}

func (o *Object) Header() []byte {
	return []byte(fmt.Sprintf("%s %d\x00", o.Type, o.Size))
}

func ReadObject(r io.Reader) (*Object, error) {
	checkSum := sha1.New()
	tr := io.TeeReader(r, checkSum)

	objectType, size, err := readHeader(tr)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(tr)
	if err != nil {
		return nil, err
	}

	hash := checkSum.Sum(nil)

	object := &Object{
		Hash: hash,
		Type: objectType,
		Size: size,
		Data: data,
	}

	return object, nil
}

func readHeader(r io.Reader) (Type, int, error) {
	headerString, err := util.ReadNullTerminatedString(r)
	if err != nil {
		return UndefinedObject, 0, err
	}

	header := strings.Split(headerString, " ")
	if len(header) != 2 {
		return UndefinedObject, 0, ErrInvalidObject
	}
	objectTypeString := header[0]
	sizeString := header[1]

	objectType, err := NewType(objectTypeString)
	if err != nil {
		return UndefinedObject, 0, err
	}

	size := 0
	if _, err := fmt.Sscanf(sizeString, " %d", &size); err != nil {
		return UndefinedObject, 0, err
	}

	return objectType, size, nil
}
