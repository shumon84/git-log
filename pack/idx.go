package pack

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"sort"

	"github.com/shumon84/git-log/sha"
)

type Idx struct {
	MagicNumber uint32      // Version2 only. MagicNumber == 0xff744f63
	Version     uint32      // Version2 only
	FanOut      []uint32    // len(FanOut) == 256
	Entries     []*IdxEntry // len(Entries == FanOut[255])
	Trailer     *Trailer
}

type Trailer struct {
	PackFileCheckSum sha.SHA1
	IdxFileCheckSum  sha.SHA1
}

type IdxEntry struct {
	Offset      uint32
	Hash        sha.SHA1
	CRC         uint32 // Version2 only
	LargeOffset uint64 // Version2 only
}

func (idx *Idx) Find(sha1 sha.SHA1) (*IdxEntry, error) {
	left := uint32(0)
	if sha1[0] != 0x00 {
		left = idx.FanOut[int(sha1[0]-1)]
	}
	right := idx.FanOut[int(sha1[0])]

	index := int(left) + sort.Search(len(idx.Entries[left:right]), func(i int) bool {
		return string(idx.Entries[int(left)+i].Hash) >= string(sha1)
	})
	if index >= int(idx.FanOut[255]) || string(idx.Entries[index].Hash) != string(sha1) {
		return nil, ErrNotFoundIdxEntry
	}

	return idx.Entries[index], nil
}

func (idx *Idx) FindByOffset(offset int64) (*IdxEntry, error) {
	for _, idxEntry := range idx.Entries {
		entryOffset := int64(idxEntry.Offset)
		if idxEntry.HasLargeOffset() {
			offset = int64(idxEntry.LargeOffset)
		}
		if entryOffset == offset {
			return idxEntry, nil
		}
	}
	sortedEntries := make([]uint32, len(idx.Entries))
	for i, v := range idx.Entries {
		sortedEntries[i] = v.Offset
	}
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i] > sortedEntries[j]
	})

	return nil, ErrNotFoundIdxEntry
}

func (e *IdxEntry) HasLargeOffset() bool {
	return e.Offset&0x80000000 == 0x80000000
}

var (
	ErrInvalidIdxFile = errors.New("invalid idx file")
)

func ReadIdxFile(rs io.ReadSeeker) (*Idx, error) {
	var header uint32
	if err := binary.Read(rs, binary.BigEndian, &header); err != nil {
		return nil, err
	}
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if header == 0xff744f63 {
		return readIdxFileV2(rs)
	}
	return readIdxFileV1(rs)
}

func readIdxFileV1(r io.Reader) (*Idx, error) {
	checkSum := sha1.New()
	io.TeeReader(r, checkSum)

	info := &Idx{
		Version: 1,
		FanOut:  make([]uint32, 256),
		Trailer: &Trailer{},
	}

	for i := range info.FanOut {
		if err := binary.Read(r, binary.BigEndian, &info.FanOut[i]); err != nil {
			return nil, err
		}
	}

	info.Entries = make([]*IdxEntry, info.FanOut[255])

	for i := range info.Entries {
		info.Entries[i] = &IdxEntry{}
		if err := binary.Read(r, binary.BigEndian, &info.Entries[i].Offset); err != nil {
			return nil, err
		}
		info.Entries[i].Hash = make(sha.SHA1, 20)
		if err := binary.Read(r, binary.BigEndian, info.Entries[i].Hash); err != nil {
			return nil, err
		}
	}

	info.Trailer.PackFileCheckSum = make(sha.SHA1, 20)
	if err := binary.Read(r, binary.BigEndian, &info.Trailer.PackFileCheckSum); err != nil {
		return nil, err
	}

	hash := checkSum.Sum(nil)
	info.Trailer.IdxFileCheckSum = make(sha.SHA1, 20)
	if err := binary.Read(r, binary.BigEndian, &info.Trailer.IdxFileCheckSum); err != nil {
		return nil, err
	}
	if string(hash) != string(info.Trailer.IdxFileCheckSum) {
		return nil, ErrInvalidIdxFile
	}

	return info, nil
}

func readIdxFileV2(r io.Reader) (*Idx, error) {
	checkSum := sha1.New()
	tr := io.TeeReader(r, checkSum)

	info := &Idx{
		FanOut:  make([]uint32, 256),
		Trailer: &Trailer{},
	}

	if err := binary.Read(tr, binary.BigEndian, &info.MagicNumber); err != nil {
		return nil, err
	}
	if err := binary.Read(tr, binary.BigEndian, &info.Version); err != nil {
		return nil, err
	}

	for i := range info.FanOut {
		if err := binary.Read(tr, binary.BigEndian, &info.FanOut[i]); err != nil {
			return nil, err
		}
	}

	info.Entries = make([]*IdxEntry, info.FanOut[255])

	for i := range info.Entries {
		info.Entries[i] = &IdxEntry{}
		info.Entries[i].Hash = make(sha.SHA1, 20)
		if err := binary.Read(tr, binary.BigEndian, info.Entries[i].Hash); err != nil {
			return nil, err
		}
	}
	for i := range info.Entries {
		if err := binary.Read(tr, binary.BigEndian, &info.Entries[i].CRC); err != nil {
			return nil, err
		}
	}
	largeOffsetMap := map[int]struct{}{}
	for i := range info.Entries {
		if err := binary.Read(tr, binary.BigEndian, &info.Entries[i].Offset); err != nil {
			return nil, err
		}
		if info.Entries[i].HasLargeOffset() {
			largeOffsetMap[i] = struct{}{}
		}
	}

	for index := range largeOffsetMap {
		if err := binary.Read(tr, binary.BigEndian, &info.Entries[index].LargeOffset); err != nil {
			return nil, err
		}
	}

	info.Trailer.PackFileCheckSum = make(sha.SHA1, 20)
	if _, err := tr.Read(info.Trailer.PackFileCheckSum); err != nil {
		return nil, err
	}

	hash := checkSum.Sum(nil)
	info.Trailer.IdxFileCheckSum = make(sha.SHA1, 20)
	if _, err := tr.Read(info.Trailer.IdxFileCheckSum); err != nil {
		return nil, err
	}
	if string(hash) != string(info.Trailer.IdxFileCheckSum) {
		return nil, ErrInvalidIdxFile
	}

	return info, nil
}
