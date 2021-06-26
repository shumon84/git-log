package pack

import (
	"github.com/shumon84/git-log/object"
)

type EntryType int

const (
	_ EntryType = iota
	OBJ_COMMIT
	OBJ_TREE
	OBJ_BLOB
	OBJ_TAG
	_
	OBJ_OFS_DELTA
	OBJ_REF_DELTA
)

func (e EntryType) String() string {
	switch e {
	case OBJ_COMMIT:
		return "OBJ_COMMIT"
	case OBJ_TREE:
		return "OBJ_TREE"
	case OBJ_BLOB:
		return "OBJ_BLOB"
	case OBJ_TAG:
		return "OBJ_TAG"
	case OBJ_OFS_DELTA:
		return "OBJ_OFS_DELTA"
	case OBJ_REF_DELTA:
		return "	OBJ_OFS_DELTA"
	}
	entryTypeString := []string{
		"Undefined",
		"OBJ_COMMIT",
		"OBJ_TREE",
		"OBJ_BLOB",
		"OBJ_TAG",
		"Undefined",
		"OBJ_OFS_DELTA",
		"OBJ_REF_DELTA",
	}
	return entryTypeString[e]
}

func (e EntryType) ToObjectType() (objectType object.Type, err error) {
	switch e {
	case OBJ_COMMIT:
		objectType = object.CommitObject
	case OBJ_TREE:
		objectType = object.TreeObject
	case OBJ_BLOB:
		objectType = object.BlobObject
	case OBJ_TAG:
		objectType = object.TagObject
	default:
		objectType = object.UndefinedObject
		err = ErrCantBeConvertedToObject
	}
	return
}
