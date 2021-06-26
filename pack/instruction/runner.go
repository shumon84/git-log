package instruction

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"

	"github.com/shumon84/git-log/object"
	"github.com/shumon84/git-log/util"
)

var (
	ErrInvalidSrcLength = errors.New("invalid src length")
	ErrInvalidDstLength = errors.New("invalid dst length")
)

type Runner struct {
	src        *bytes.Buffer
	srcLength  uint64
	dst        *bytes.Buffer
	dstLength  uint64
	baseObject *object.Object

	instruction byte
}

func NewRunner(deltaObject []byte, baseObject *object.Object) (*Runner, error) {
	src := bytes.NewBuffer(deltaObject)

	srcLength, _, err := util.DecodeSizeEncoding(src)
	if err != nil {
		return nil, err
	}
	dstLength, _, err := util.DecodeSizeEncoding(src)
	if err != nil {
		return nil, err
	}

	dst := bytes.NewBuffer(nil)

	return &Runner{
		src:        src,
		srcLength:  srcLength,
		dst:        dst,
		dstLength:  dstLength,
		baseObject: baseObject,
	}, nil
}

func (r *Runner) Run() (*object.Object, error) {
	for r.next() {
		if err := r.execInstruction(); err != nil {
			return nil, err
		}
	}

	obj := &object.Object{
		Type: r.baseObject.Type,
		Size: int(r.dstLength),
	}

	checkSum := sha1.New()
	tr := io.TeeReader(r.dst, checkSum)

	checkSum.Write(obj.Header())
	data, err := ioutil.ReadAll(tr)
	if err != nil {
		return nil, err
	}

	obj.Hash = checkSum.Sum(nil)
	obj.Data = data
	obj.Size = len(data)

	if uint64(r.baseObject.Size) != r.srcLength {
		return nil, ErrInvalidSrcLength
	}
	if uint64(len(data)) != r.dstLength {
		return nil, ErrInvalidDstLength
	}

	return obj, nil
}

func (r *Runner) next() bool {
	instruction, err := r.src.ReadByte()
	if err != nil {
		return false
	}
	r.instruction = instruction
	return true
}

func (r *Runner) execInstruction() error {
	if r.instruction&0x80 > 0 {
		return r.execCopyInstruction(r.instruction)
	} else {
		return r.execAddInstruction(r.instruction)
	}
}

func (r *Runner) execCopyInstruction(instruction byte) error {
	args := make([]byte, 7)
	for i := range args {
		if instruction&(0x1<<i) == 0 {
			continue
		}
		arg, err := r.src.ReadByte()
		if err != nil {
			return err
		}
		args[i] = arg
	}

	offset := binary.LittleEndian.Uint32(args[:4])
	size := binary.LittleEndian.Uint32(append(args[4:], 0))
	if size == 0 {
		size = 0x10000
	}
	if _, err := r.dst.Write(r.baseObject.Data[offset : offset+size]); err != nil {
		return err
	}
	return nil
}

func (r *Runner) execAddInstruction(instruction byte) error {
	_, err := io.CopyN(r.dst, r.src, int64(instruction))
	if err != nil {
		return err
	}
	return nil
}
